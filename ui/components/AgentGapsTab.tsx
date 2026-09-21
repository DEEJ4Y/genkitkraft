import { useState } from 'react'
import {
  Stack,
  Group,
  Text,
  Alert,
  Loader,
  Center,
  Card,
  Badge,
  Button,
  Modal,
  Select,
  Textarea,
  Pagination,
} from '@mantine/core'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type GapResponse = components['schemas']['Models.GapResponse']

const PAGE_SIZE = 20

interface AgentGapsTabProps {
  agentId: string
}

const CATEGORY_COLORS: Record<string, string> = {
  knowledge: 'blue',
  capability: 'orange',
  improvement: 'teal',
}

const STATUS_COLORS: Record<string, string> = {
  open: 'yellow',
  resolved: 'green',
  dismissed: 'gray',
}

const DISMISSAL_REASONS = [
  { value: 'unrelated', label: 'Unrelated' },
  { value: 'insufficient_detail', label: 'Insufficient detail' },
  { value: 'duplicate', label: 'Duplicate' },
  { value: 'other', label: 'Other' },
]

export function AgentGapsTab({ agentId }: AgentGapsTabProps) {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const offset = (page - 1) * PAGE_SIZE
  const [dismissTarget, setDismissTarget] = useState<GapResponse | null>(null)
  const [dismissalCategory, setDismissalCategory] = useState<string | null>(null)
  const [dismissalReason, setDismissalReason] = useState('')

  const gapsQuery = useQuery({
    queryKey: ['get', `/api/v1/agents/${agentId}/gaps`, { limit: PAGE_SIZE, offset }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/agents/{agentId}/gaps', {
        params: { path: { agentId }, query: { limit: PAGE_SIZE, offset } },
      })
      if (error) throw new Error('Failed to fetch gaps')
      return data
    },
  })

  const updateMutation = useMutation({
    mutationFn: async (body: { gapId: string; status: string; dismissalCategory?: string; dismissalReason?: string }) => {
      const { gapId, ...rest } = body
      const { error } = await fetchClient.PUT('/api/v1/agents/{agentId}/gaps/{gapId}', {
        params: { path: { agentId, gapId } },
        body: rest as any,
      })
      if (error) throw new Error('Failed to update gap')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['get', `/api/v1/agents/${agentId}/gaps`] })
      setDismissTarget(null)
      setDismissalCategory(null)
      setDismissalReason('')
    },
  })

  function confirmDismiss() {
    if (!dismissTarget || !dismissalCategory) return
    updateMutation.mutate({
      gapId: dismissTarget.id,
      status: 'dismissed',
      dismissalCategory,
      dismissalReason: dismissalReason || undefined,
    })
  }

  if (gapsQuery.isPending) {
    return (
      <Center py="xl">
        <Loader />
      </Center>
    )
  }

  if (gapsQuery.error) {
    return (
      <Alert color="red" variant="light">
        Failed to load gaps.
      </Alert>
    )
  }

  const gaps = gapsQuery.data?.gaps ?? []
  const total = gapsQuery.data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <Stack>
      <Text size="xs" c="dimmed">
        Gaps this agent has self-reported: questions it couldn&apos;t answer (knowledge), actions it
        couldn&apos;t perform (capability), or ideas to automate more of the flow (improvement).
      </Text>

      {updateMutation.error && (
        <Alert color="red" variant="light">
          {(updateMutation.error as Error).message}
        </Alert>
      )}

      {gaps.length === 0 ? (
        <Text size="sm" c="dimmed">
          No gaps reported yet.
        </Text>
      ) : (
        <Stack gap="sm">
          {gaps.map((g) => {
            const isTerminal = g.status === 'dismissed' && g.dismissalCategory === 'unrelated'
            return (
              <Card key={g.id} padding="sm" withBorder>
                <Group justify="space-between" align="flex-start" wrap="nowrap">
                  <Stack gap={4} style={{ flex: 1 }}>
                    <Group gap="xs">
                      <Badge size="sm" color={CATEGORY_COLORS[g.category] ?? 'gray'} variant="light">
                        {g.category}
                      </Badge>
                      <Badge size="sm" color={STATUS_COLORS[g.status] ?? 'gray'} variant="filled">
                        {g.status}
                      </Badge>
                    </Group>
                    <Text size="sm" fw={500}>
                      {g.context}
                    </Text>
                    <Text size="sm" c="dimmed">
                      {g.details}
                    </Text>
                    {g.suggestedResolution && (
                      <Text size="xs" c="dimmed">
                        Suggestion: {g.suggestedResolution}
                      </Text>
                    )}
                    {g.status === 'dismissed' && g.dismissalCategory && (
                      <Text size="xs" c="dimmed">
                        Dismissed: {g.dismissalCategory}
                        {g.dismissalReason ? ` — ${g.dismissalReason}` : ''}
                      </Text>
                    )}
                    {g.references.length > 0 && (
                      <Text size="xs" c="dimmed">
                        Seen in:{' '}
                        {g.references
                          .map((r) =>
                            r.sessionId ? r.sessionId + (r.messageId ? `/${r.messageId}` : '') : '(stateless call)'
                          )
                          .join(', ')}
                      </Text>
                    )}
                  </Stack>
                  <Group gap="xs" wrap="nowrap">
                    {g.status !== 'resolved' && !isTerminal && (
                      <Button
                        size="xs"
                        variant="light"
                        color="green"
                        onClick={() => updateMutation.mutate({ gapId: g.id, status: 'resolved' })}
                        loading={updateMutation.isPending}
                      >
                        Resolve
                      </Button>
                    )}
                    {g.status !== 'dismissed' && (
                      <Button size="xs" variant="light" color="gray" onClick={() => setDismissTarget(g)}>
                        Dismiss
                      </Button>
                    )}
                    {g.status !== 'open' && !isTerminal && (
                      <Button
                        size="xs"
                        variant="light"
                        onClick={() => updateMutation.mutate({ gapId: g.id, status: 'open' })}
                        loading={updateMutation.isPending}
                      >
                        Reopen
                      </Button>
                    )}
                  </Group>
                </Group>
              </Card>
            )
          })}
        </Stack>
      )}

      {totalPages > 1 && (
        <Center mt="lg">
          <Pagination total={totalPages} value={page} onChange={setPage} />
        </Center>
      )}

      <Modal opened={!!dismissTarget} onClose={() => setDismissTarget(null)} title="Dismiss Gap">
        <Stack>
          <Select
            label="Reason"
            placeholder="Choose a reason"
            data={DISMISSAL_REASONS}
            value={dismissalCategory}
            onChange={setDismissalCategory}
            required
          />
          <Textarea
            label="Additional detail (optional)"
            value={dismissalReason}
            onChange={(e) => setDismissalReason(e.currentTarget.value)}
          />
          {dismissalCategory === 'unrelated' && (
            <Text size="xs" c="orange">
              Dismissing as unrelated is permanent — this gap can never be reopened.
            </Text>
          )}
          <Group justify="flex-end">
            <Button variant="default" onClick={() => setDismissTarget(null)}>
              Cancel
            </Button>
            <Button onClick={confirmDismiss} disabled={!dismissalCategory} loading={updateMutation.isPending}>
              Dismiss
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  )
}
