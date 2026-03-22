import { Card, Text, Group, ActionIcon, Tooltip } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import type { components } from '../lib/api/schema'

type PromptResponse = components['schemas']['Models.PromptResponse']

interface PromptCardProps {
  prompt: PromptResponse
  onEdit: () => void
  onDelete: () => void
}

function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength) + '...'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function PromptCard({ prompt, onEdit, onDelete }: PromptCardProps) {
  return (
    <Card shadow="xs" padding="md" radius="sm" withBorder>
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <div style={{ flex: 1, minWidth: 0 }}>
          <Text fw={600} size="md">
            {prompt.name}
          </Text>
          {prompt.content && (
            <Text size="sm" c="dimmed" mt={4} lineClamp={2}>
              {truncate(prompt.content, 150)}
            </Text>
          )}
          <Text size="xs" c="dimmed" mt={8}>
            Updated {formatDate(prompt.updatedAt as unknown as string)}
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
