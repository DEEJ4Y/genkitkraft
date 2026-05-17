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
}

interface UsePlaygroundChatOptions {
  agentId: string
  sessionId: string | null
  streaming?: boolean
  config?: ChatConfig
  onSessionTitleUpdate?: () => void
}

export function usePlaygroundChat({ agentId, sessionId, streaming = true, config, onSessionTitleUpdate }: UsePlaygroundChatOptions) {
  const [messages, setMessages] = useState<Message[]>([])
  const [streamingContent, setStreamingContent] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortRef = useRef<AbortController | null>(null)

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

  const sendMessage = useCallback(async (content: string) => {
    if (!sessionId || !content.trim() || isStreaming) return

    setError(null)
    const isFirstMessage = messages.length === 0

    // Optimistically add user message
    const userMsg: Message = { id: `temp-${Date.now()}`, role: 'user', content }
    setMessages((prev) => [...prev, userMsg])
    setIsStreaming(true)
    setStreamingContent('')

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
        // Streaming: parse SSE
        const reader = res.body!.getReader()
        const decoder = new TextDecoder()
        let fullContent = ''
        let buffer = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })

          // Process complete SSE lines
          const lines = buffer.split('\n')
          buffer = lines.pop() ?? '' // Keep incomplete line in buffer

          for (const line of lines) {
            if (!line.startsWith('data: ')) continue
            const data = line.slice(6)

            if (data === '[DONE]') continue
            if (data.startsWith('[ERROR]')) {
              setError(data.slice(8))
              continue
            }

            fullContent += data
            setStreamingContent(fullContent)
          }
        }

        // Add assistant message
        if (fullContent) {
          setMessages((prev) => [
            ...prev,
            { id: `assistant-${Date.now()}`, role: 'assistant', content: fullContent },
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
  }, [agentId, sessionId, isStreaming, messages.length, streaming, config, onSessionTitleUpdate])

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort()
  }, [])

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
