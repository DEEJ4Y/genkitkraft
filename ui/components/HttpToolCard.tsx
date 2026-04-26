import { Card, Text, Group, ActionIcon, Tooltip, Badge } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import type { components } from '../lib/api/schema'

type HttpToolResponse = components['schemas']['Models.HttpToolResponse']

interface HttpToolCardProps {
  tool: HttpToolResponse
  onEdit: () => void
  onDelete: () => void
}

const methodColors: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'orange',
  PATCH: 'yellow',
  DELETE: 'red',
  HEAD: 'gray',
  OPTIONS: 'grape',
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function HttpToolCard({ tool, onEdit, onDelete }: HttpToolCardProps) {
  return (
    <Card shadow="xs" padding="md" radius="sm" withBorder>
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <div style={{ flex: 1, minWidth: 0 }}>
          <Group gap="sm" mb={4}>
            <Badge color={methodColors[tool.method] ?? 'gray'} variant="light" size="sm">
              {tool.method}
            </Badge>
            <Text fw={600} size="md">
              {tool.name}
            </Text>
          </Group>
          <Text size="sm" c="dimmed" lineClamp={1} style={{ fontFamily: 'monospace' }}>
            {tool.url}
          </Text>
          {tool.description && (
            <Text size="sm" c="dimmed" mt={4} lineClamp={2}>
              {tool.description}
            </Text>
          )}
          <Text size="xs" c="dimmed" mt={8}>
            Updated {formatDate(tool.updatedAt as unknown as string)}
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
