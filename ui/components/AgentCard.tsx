import { Card, Text, Group, ActionIcon, Tooltip, Badge } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import type { components } from '../lib/api/schema'

type AgentResponse = components['schemas']['Models.AgentResponse']

interface AgentCardProps {
  agent: AgentResponse
  onEdit: () => void
  onDelete: () => void
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function AgentCard({ agent, onEdit, onDelete }: AgentCardProps) {
  return (
    <Card shadow="xs" padding="md" radius="sm" withBorder>
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <div style={{ flex: 1, minWidth: 0 }}>
          <Text fw={600} size="md">
            {agent.name}
          </Text>

          <Group gap="xs" mt={4}>
            <Badge variant="light" size="sm">
              {agent.providerName || agent.providerType}
            </Badge>
            <Badge variant="outline" size="sm">
              {agent.modelId}
            </Badge>
          </Group>

          {agent.systemPromptName ? (
            <Text size="sm" c="dimmed" mt={4}>
              Prompt: {agent.systemPromptName}
            </Text>
          ) : (
            <Text size="sm" c="dimmed" mt={4}>
              No system prompt
            </Text>
          )}

          <Group gap="xs" mt={4}>
            <Text size="xs" c="dimmed">
              temp: {agent.temperature}
            </Text>
            <Text size="xs" c="dimmed">
              topP: {agent.topP}
            </Text>
            <Text size="xs" c="dimmed">
              topK: {agent.topK}
            </Text>
          </Group>

          <Text size="xs" c="dimmed" mt={4}>
            Updated {formatDate(agent.updatedAt as unknown as string)}
          </Text>
        </div>

        <Group gap="xs" wrap="nowrap">
          <Tooltip label="Edit">
            <ActionIcon variant="subtle" onClick={onEdit}>
              <IconEdit size={18} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Delete">
            <ActionIcon variant="subtle" color="red" onClick={onDelete}>
              <IconTrash size={18} />
            </ActionIcon>
          </Tooltip>
        </Group>
      </Group>
    </Card>
  )
}
