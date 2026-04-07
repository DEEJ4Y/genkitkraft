import { useState } from 'react'
import {
  Title,
  Text,
  Stack,
  Group,
  Button,
  Loader,
  Alert,
  Center,
  Pagination,
} from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import { AgentCard } from '../components/AgentCard'
import { AgentEditView } from '../components/AgentEditView'
import { AgentForm } from '../components/AgentForm'
import type { components } from '../lib/api/schema'

type AgentResponse = components['schemas']['Models.AgentResponse']

type View = { mode: 'list' } | { mode: 'create' } | { mode: 'edit'; agentId: string }

const PAGE_SIZE = 20

export default function AgentsPage() {
  const queryClient = useQueryClient()
  const [view, setView] = useState<View>({ mode: 'list' })
  const [page, setPage] = useState(1)

  const offset = (page - 1) * PAGE_SIZE

  const agentsQuery = useQuery({
    queryKey: ['get', '/api/v1/agents', { limit: PAGE_SIZE, offset }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/agents', {
        params: { query: { limit: PAGE_SIZE, offset } },
      })
      if (error) throw new Error('Failed to fetch agents')
      return data
    },
  })

  const editingAgentId = view.mode === 'edit' ? view.agentId : null

  const editingAgentQuery = useQuery({
    queryKey: ['get', '/api/v1/agents', editingAgentId],
    queryFn: async () => {
      if (!editingAgentId) return null
      const { data, error } = await fetchClient.GET('/api/v1/agents/{id}', {
        params: { path: { id: editingAgentId } },
      })
      if (error) throw new Error('Failed to fetch agent')
      return data
    },
    enabled: !!editingAgentId,
  })

  function handleSaved() {
    setView({ mode: 'list' })
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/agents'] })
  }

  async function handleDelete(agent: AgentResponse) {
    if (!confirm(`Delete "${agent.name}"? This cannot be undone.`)) return
    await fetchClient.DELETE('/api/v1/agents/{id}', {
      params: { path: { id: agent.id } },
    })
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/agents'] })
  }

  if (view.mode === 'create') {
    return <AgentForm onSaved={handleSaved} onCancel={() => setView({ mode: 'list' })} />
  }

  if (view.mode === 'edit') {
    if (editingAgentQuery.isPending) {
      return (
        <Center py="xl">
          <Loader />
        </Center>
      )
    }

    if (editingAgentQuery.error || !editingAgentQuery.data) {
      return (
        <Alert color="red" variant="light">
          Failed to load agent.
        </Alert>
      )
    }

    return (
      <AgentEditView
        agent={editingAgentQuery.data}
        onSaved={handleSaved}
        onCancel={() => setView({ mode: 'list' })}
      />
    )
  }

  const agents = agentsQuery.data?.agents ?? []
  const total = agentsQuery.data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <>
      <Group justify="space-between" align="center" mb="lg">
        <Title order={2}>Agents</Title>
        <Button leftSection={<IconPlus size={16} />} onClick={() => setView({ mode: 'create' })}>
          New Agent
        </Button>
      </Group>

      <Text size="sm" c="dimmed" mb="md">
        Configure AI agents with providers, models, and generation settings.
      </Text>

      {agentsQuery.isPending && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {agentsQuery.error && (
        <Alert color="red" variant="light" mb="md">
          Failed to load agents.
        </Alert>
      )}

      {!agentsQuery.isPending && agents.length === 0 && (
        <Text c="dimmed" ta="center" py="xl">
          No agents yet. Create your first agent to get started.
        </Text>
      )}

      <Stack gap="sm">
        {agents.map((agent) => (
          <AgentCard
            key={agent.id}
            agent={agent}
            onEdit={() => setView({ mode: 'edit', agentId: agent.id })}
            onDelete={() => handleDelete(agent)}
          />
        ))}
      </Stack>

      {totalPages > 1 && (
        <Center mt="lg">
          <Pagination total={totalPages} value={page} onChange={setPage} />
        </Center>
      )}
    </>
  )
}
