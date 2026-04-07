import { useCallback, useEffect, useState } from 'react'
import { Alert, Box, Loader, Center } from '@mantine/core'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchClient } from '../../lib/api/client'
import type { components } from '../../lib/api/schema'
import { usePlaygroundChat } from '../../hooks/usePlaygroundChat'
import { SessionSidebar } from './SessionSidebar'
import { ChatView } from './ChatView'
import { PlaygroundConfigBar, type PlaygroundConfig } from './PlaygroundConfigBar'
import { SaveConfigModal } from './SaveConfigModal'

type AgentResponse = components['schemas']['Models.AgentResponse']

interface AgentPlaygroundProps {
  agent: AgentResponse
}

export function AgentPlayground({ agent }: AgentPlaygroundProps) {
  const queryClient = useQueryClient()
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null)
  const [saveModalOpen, setSaveModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const [config, setConfig] = useState<PlaygroundConfig>({
    providerId: agent.providerId,
    modelId: agent.modelId,
    systemPromptId: agent.systemPromptId ?? '',
    temperature: agent.temperature,
    topP: agent.topP,
    topK: agent.topK,
  })

  const sessionsQueryKey = ['get', `/api/v1/agents/${agent.id}/playground/sessions`]

  const sessionsQuery = useQuery({
    queryKey: sessionsQueryKey,
    queryFn: async () => {
      const { data, error } = await fetchClient.GET(
        '/api/v1/agents/{agentId}/playground/sessions',
        { params: { path: { agentId: agent.id } } }
      )
      if (error) throw new Error('Failed to fetch sessions')
      return data
    },
  })

  const sessions = sessionsQuery.data?.sessions ?? []

  const handleSessionTitleUpdate = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: sessionsQueryKey })
  }, [queryClient, sessionsQueryKey])

  const chat = usePlaygroundChat({
    agentId: agent.id,
    sessionId: activeSessionId,
    config: {
      providerId: config.providerId !== agent.providerId ? config.providerId : undefined,
      modelId: config.modelId !== agent.modelId ? config.modelId : undefined,
      systemPromptId: config.systemPromptId !== (agent.systemPromptId ?? '') ? config.systemPromptId : undefined,
      temperature: config.temperature !== agent.temperature ? config.temperature : undefined,
      topP: config.topP !== agent.topP ? config.topP : undefined,
      topK: config.topK !== agent.topK ? config.topK : undefined,
    },
    onSessionTitleUpdate: handleSessionTitleUpdate,
  })

  // Load messages when session changes
  useEffect(() => {
    if (activeSessionId) {
      chat.loadMessages(activeSessionId)
    } else {
      chat.clearMessages()
    }
  }, [activeSessionId]) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleCreateSession() {
    try {
      const { data, error } = await fetchClient.POST(
        '/api/v1/agents/{agentId}/playground/sessions',
        {
          params: { path: { agentId: agent.id } },
          body: {} as any,
        }
      )
      if (error) throw new Error('Failed to create session')
      queryClient.invalidateQueries({ queryKey: sessionsQueryKey })
      if (data) {
        setActiveSessionId(data.id)
      }
    } catch {
      // silently fail
    }
  }

  async function handleDeleteSession(id: string) {
    try {
      await fetchClient.DELETE(
        '/api/v1/agents/{agentId}/playground/sessions/{sessionId}',
        { params: { path: { agentId: agent.id, sessionId: id } } }
      )
      queryClient.invalidateQueries({ queryKey: sessionsQueryKey })
      if (activeSessionId === id) {
        setActiveSessionId(null)
      }
    } catch {
      // silently fail
    }
  }

  function handleOpenSaveModal() {
    setSaveError('')
    if (!config.providerId) {
      setSaveError('A provider must be selected before saving.')
      setSaveModalOpen(true)
      return
    }
    if (!config.modelId.trim()) {
      setSaveError('A model must be selected before saving.')
      setSaveModalOpen(true)
      return
    }
    setSaveModalOpen(true)
  }

  async function handleSaveToAgent() {
    if (!config.providerId || !config.modelId.trim()) return
    setSaving(true)
    try {
      const { error } = await fetchClient.PUT('/api/v1/agents/{id}', {
        params: { path: { id: agent.id } },
        body: {
          providerId: config.providerId,
          modelId: config.modelId,
          systemPromptId: config.systemPromptId || undefined,
          temperature: config.temperature,
          topP: config.topP,
          topK: config.topK,
        } as any,
      })
      if (error) throw new Error('Failed to save')
      queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/agents'] })
      setSaveModalOpen(false)
    } catch {
      // keep modal open on failure
    } finally {
      setSaving(false)
    }
  }

  if (sessionsQuery.isPending) {
    return (
      <Center py="xl">
        <Loader />
      </Center>
    )
  }

  if (sessionsQuery.error) {
    return (
      <Alert color="red" variant="light">
        Failed to load playground sessions.
      </Alert>
    )
  }

  return (
    <>
      <Box
        style={{
          display: 'flex',
          height: 'calc(100vh - 280px)',
          minHeight: 400,
          border: '1px solid var(--mantine-color-gray-3)',
          borderRadius: 8,
          overflow: 'hidden',
        }}
      >
        <SessionSidebar
          sessions={sessions}
          activeSessionId={activeSessionId}
          onSelect={setActiveSessionId}
          onCreate={handleCreateSession}
          onDelete={handleDeleteSession}
        />

        <Box style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <PlaygroundConfigBar
            agent={agent}
            config={config}
            onChange={setConfig}
            onSaveToAgent={handleOpenSaveModal}
          />
          <ChatView
            messages={chat.messages}
            streamingContent={chat.streamingContent}
            isStreaming={chat.isStreaming}
            error={chat.error}
            onSend={chat.sendMessage}
            onStop={chat.stopStreaming}
            hasSession={!!activeSessionId}
          />
        </Box>
      </Box>

      <SaveConfigModal
        opened={saveModalOpen}
        onClose={() => setSaveModalOpen(false)}
        onConfirm={handleSaveToAgent}
        saving={saving}
        error={saveError}
      />
    </>
  )
}
