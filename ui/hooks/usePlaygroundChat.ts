import { useState, useCallback, useRef } from 'react'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
}

interface ChatConfig {
  providerId?: string
  modelId?: string
  systemPromptId?: string
  temperatureEnabled?: boolean
  temperature?: number
  topPEnabled?: boolean
  topP?: number
  topKEnabled?: boolean
  topK?: number
  maxToolCalls?: number
  httpToolIds?: string[]
  mcpServers?: { mcpServerId: string; selectAll: boolean; toolNames: string[] }[]
  builtInToolIds?: string[]
}

interface UsePlaygroundChatOptions {
  agentId: string
  sessionId: string | null
  streaming?: boolean
  config?: ChatConfig
  onSessionTitleUpdate?: () => void
}

// SSE read outcomes: 'done'/'error' are terminal sentinels from the server,
// 'aborted' means the user stopped generation (stopStreaming), and
// 'incomplete' means the connection dropped without either — the case that
// triggers a reconnect via Last-Event-ID.
type SSEOutcome = 'done' | 'error' | 'aborted' | 'incomplete'

const MAX_RECONNECT_ATTEMPTS = 5
const RECONNECT_DELAY_MS = 500

export function usePlaygroundChat({ agentId, sessionId, streaming = true, config, onSessionTitleUpdate }: UsePlaygroundChatOptions) {
  const [messages, setMessages] = useState<Message[]>([])
  const [streamingContent, setStreamingContent] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const lastSeqRef = useRef(0)

  const loadMessages = useCallback(async (sid: string) => {
    try {
      const res = await fetch(`/api/v1/agents/${agentId}/playground/sessions/${sid}/messages`, {
        credentials: 'include',
      })
      if (!res.ok) throw new Error('Failed to load messages')
      const data = await res.json()
      setMessages(
        (data.messages ?? []).map((m: any) => ({
          id: m.id,
          role: m.role,
          content: m.content,
        }))
      )
    } catch {
      setMessages([])
    }
  }, [agentId])

  // Reads one SSE response body, appending content deltas onto contentRef and
  // mirroring them into streamingContent as they arrive. Shared by both the
  // initial POST response and any Last-Event-ID reconnect GET response.
  const readSSEBody = useCallback(async (
    body: ReadableStream<Uint8Array>,
    contentRef: { current: string },
    signal: AbortSignal
  ): Promise<SSEOutcome> => {
    const reader = body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let pendingId: number | null = null

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          if (line.startsWith('id: ')) {
            const n = parseInt(line.slice(4), 10)
            if (!Number.isNaN(n)) pendingId = n
            continue
          }
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6)

          if (data === '[DONE]') return 'done'
          if (data.startsWith('[ERROR]')) {
            setError(data.slice(8))
            return 'error'
          }

          if (pendingId !== null) {
            lastSeqRef.current = pendingId
            pendingId = null
          }
          contentRef.current += data
          setStreamingContent(contentRef.current)
        }
      }
    } catch (err: any) {
      if (signal.aborted || err?.name === 'AbortError') return 'aborted'
      return 'incomplete'
    }

    // The reader finished (`done`) without ever seeing a [DONE]/[ERROR]
    // sentinel — the connection dropped mid-stream.
    return 'incomplete'
  }, [])

  const sendMessage = useCallback(async (content: string) => {
    if (!sessionId || !content.trim() || isStreaming) return

    setError(null)
    const isFirstMessage = messages.length === 0

    // Optimistically add user message
    const userMsg: Message = { id: `temp-${Date.now()}`, role: 'user', content }
    setMessages((prev) => [...prev, userMsg])
    setIsStreaming(true)
    setStreamingContent('')
    lastSeqRef.current = 0

    const abortController = new AbortController()
    abortRef.current = abortController

    try {
      const body: any = {
        sessionId,
        content,
        stream: streaming,
      }
      if (config?.providerId) body.providerId = config.providerId
      if (config?.modelId) body.modelId = config.modelId
      if (config?.systemPromptId) body.systemPromptId = config.systemPromptId
      if (config?.temperatureEnabled !== undefined) body.temperatureEnabled = config.temperatureEnabled
      if (config?.temperature !== undefined) body.temperature = config.temperature
      if (config?.topPEnabled !== undefined) body.topPEnabled = config.topPEnabled
      if (config?.topP !== undefined) body.topP = config.topP
      if (config?.topKEnabled !== undefined) body.topKEnabled = config.topKEnabled
      if (config?.topK !== undefined) body.topK = config.topK
      if (config?.maxToolCalls !== undefined) body.maxToolCalls = config.maxToolCalls
      if (config?.httpToolIds) body.httpToolIds = config.httpToolIds
      if (config?.mcpServers) body.mcpServers = config.mcpServers
      if (config?.builtInToolIds) body.builtInToolIds = config.builtInToolIds

      const res = await fetch(`/api/v1/agents/${agentId}/playground/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(body),
        signal: abortController.signal,
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => null)
        throw new Error(errData?.error || `Request failed with status ${res.status}`)
      }

      if (!streaming) {
        // Non-streaming: parse single JSON response
        const data = await res.json()
        if (data.content) {
          setMessages((prev) => [
            ...prev,
            { id: data.id || `assistant-${Date.now()}`, role: 'assistant', content: data.content },
          ])
        }
      } else {
        // Streaming: parse SSE, reconnecting via Last-Event-ID if the
        // connection drops before a terminal sentinel arrives.
        const contentRef = { current: '' }
        let outcome = await readSSEBody(res.body!, contentRef, abortController.signal)

        let attempts = 0
        while (outcome === 'incomplete' && attempts < MAX_RECONNECT_ATTEMPTS) {
          attempts += 1
          await new Promise((resolve) => setTimeout(resolve, RECONNECT_DELAY_MS))
          if (abortController.signal.aborted) break

          try {
            const reconnectRes = await fetch(
              `/api/v1/agents/${agentId}/playground/sessions/${sessionId}/stream`,
              {
                credentials: 'include',
                headers: { 'Last-Event-ID': String(lastSeqRef.current) },
                signal: abortController.signal,
              }
            )
            if (!reconnectRes.ok || !reconnectRes.body) break
            outcome = await readSSEBody(reconnectRes.body, contentRef, abortController.signal)
          } catch {
            break
          }
        }

        // Add assistant message
        if (contentRef.current) {
          setMessages((prev) => [
            ...prev,
            { id: `assistant-${Date.now()}`, role: 'assistant', content: contentRef.current },
          ])
        }
      }

      // Trigger session title refresh on first message
      if (isFirstMessage && onSessionTitleUpdate) {
        onSessionTitleUpdate()
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setError(err.message || 'Failed to send message')
      }
    } finally {
      setIsStreaming(false)
      setStreamingContent('')
      abortRef.current = null
    }
  }, [agentId, sessionId, isStreaming, messages.length, streaming, config, onSessionTitleUpdate, readSSEBody])

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort()
    // Aborting the fetch only ends this response; generation itself runs
    // detached from any single request, so it must be canceled explicitly.
    if (sessionId) {
      fetch(`/api/v1/agents/${agentId}/playground/sessions/${sessionId}/stream/cancel`, {
        method: 'POST',
        credentials: 'include',
      }).catch(() => {})
    }
  }, [agentId, sessionId])

  const clearMessages = useCallback(() => {
    setMessages([])
    setStreamingContent('')
    setError(null)
  }, [])

  return {
    messages,
    streamingContent,
    isStreaming,
    error,
    sendMessage,
    stopStreaming,
    loadMessages,
    clearMessages,
  }
}
