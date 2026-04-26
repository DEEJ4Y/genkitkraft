import { Card, Text, Group, ActionIcon, Tooltip, Badge } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import type { components } from '../lib/api/schema'

type McpServerResponse = components['schemas']['Models.McpServerResponse']

interface McpServerCardProps {
  server: McpServerResponse
  onEdit: () => void
  onDelete: () => void
}

const transportColors: Record<string, string> = {
  sse: 'blue',
  streamableHttp: 'teal',
}

const transportLabels: Record<string, string> = {
  sse: 'SSE',
  streamableHttp: 'Streamable HTTP',
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function McpServerCard({ server, onEdit, onDelete }: McpServerCardProps) {
  return (
    <Card shadow="xs" padding="md" radius="sm" withBorder>
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <div style={{ flex: 1, minWidth: 0 }}>
          <Group gap="sm" mb={4}>
            <Badge color={transportColors[server.transport] ?? 'gray'} variant="light" size="sm">
              {transportLabels[server.transport] ?? server.transport}
            </Badge>
            <Text fw={600} size="md">
              {server.name}
            </Text>
          </Group>
          <Text size="sm" c="dimmed" lineClamp={1} style={{ fontFamily: 'monospace' }}>
            {server.url}
          </Text>
          <Text size="xs" c="dimmed" mt={8}>
            Updated {formatDate(server.updatedAt as unknown as string)}
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
