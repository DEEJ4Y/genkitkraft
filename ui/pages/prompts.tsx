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
import { PromptCard } from '../components/PromptCard'
import { PromptForm } from '../components/PromptForm'
import type { components } from '../lib/api/schema'

type PromptResponse = components['schemas']['Models.PromptResponse']

type View = { mode: 'list' } | { mode: 'create' } | { mode: 'edit'; promptId: string }

const PAGE_SIZE = 20

export default function PromptsPage() {
  const queryClient = useQueryClient()
  const [view, setView] = useState<View>({ mode: 'list' })
  const [page, setPage] = useState(1)

  const offset = (page - 1) * PAGE_SIZE

  const promptsQuery = useQuery({
    queryKey: ['get', '/api/v1/prompts', { limit: PAGE_SIZE, offset }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/prompts', {
        params: { query: { limit: PAGE_SIZE, offset } },
      })
      if (error) throw new Error('Failed to fetch prompts')
      return data
    },
  })

  const editingPromptId = view.mode === 'edit' ? view.promptId : null

  const editingPromptQuery = useQuery({
    queryKey: ['get', '/api/v1/prompts', editingPromptId],
    queryFn: async () => {
      if (!editingPromptId) return null
      const { data, error } = await fetchClient.GET('/api/v1/prompts/{id}', {
        params: { path: { id: editingPromptId } },
      })
      if (error) throw new Error('Failed to fetch prompt')
      return data
    },
    enabled: !!editingPromptId,
  })

  function handleSaved() {
    setView({ mode: 'list' })
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/prompts'] })
  }

  async function handleDelete(prompt: PromptResponse) {
    if (!confirm(`Delete "${prompt.name}"? This cannot be undone.`)) return
    await fetchClient.DELETE('/api/v1/prompts/{id}', {
      params: { path: { id: prompt.id } },
    })
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/prompts'] })
  }

  if (view.mode === 'create') {
    return <PromptForm onSaved={handleSaved} onCancel={() => setView({ mode: 'list' })} />
  }

  if (view.mode === 'edit') {
    if (editingPromptQuery.isPending) {
      return (
        <Center py="xl">
          <Loader />
        </Center>
      )
    }

    if (editingPromptQuery.error || !editingPromptQuery.data) {
      return (
        <Alert color="red" variant="light">
          Failed to load prompt.
        </Alert>
      )
    }

    return (
      <PromptForm
        prompt={editingPromptQuery.data}
        onSaved={handleSaved}
        onCancel={() => setView({ mode: 'list' })}
      />
    )
  }

  const prompts = promptsQuery.data?.prompts ?? []
  const total = promptsQuery.data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <>
      <Group justify="space-between" align="center" mb="lg">
        <Title order={2}>Prompts</Title>
        <Button leftSection={<IconPlus size={16} />} onClick={() => setView({ mode: 'create' })}>
          New Prompt
        </Button>
      </Group>

      <Text size="sm" c="dimmed" mb="md">
        Manage system instructions for your agents.
      </Text>

      {promptsQuery.isPending && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {promptsQuery.error && (
        <Alert color="red" variant="light" mb="md">
          Failed to load prompts.
        </Alert>
      )}

      {!promptsQuery.isPending && prompts.length === 0 && (
        <Text c="dimmed" ta="center" py="xl">
          No prompts yet. Create your first prompt to get started.
        </Text>
      )}

      <Stack gap="sm">
        {prompts.map((prompt) => (
          <PromptCard
            key={prompt.id}
            prompt={prompt}
            onEdit={() => setView({ mode: 'edit', promptId: prompt.id })}
            onDelete={() => handleDelete(prompt)}
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
