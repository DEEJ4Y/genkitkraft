import { ActionIcon, Button, Group, NavLink, ScrollArea, Stack, Text } from '@mantine/core'
import { IconPlus, IconTrash } from '@tabler/icons-react'
import type { components } from '../../lib/api/schema'

type SessionResponse = components['schemas']['Models.PlaygroundSessionResponse']

interface SessionSidebarProps {
  sessions: SessionResponse[]
  activeSessionId: string | null
  onSelect: (id: string) => void
  onCreate: () => void
  onDelete: (id: string) => void
}

export function SessionSidebar({
  sessions,
  activeSessionId,
  onSelect,
  onCreate,
  onDelete,
}: SessionSidebarProps) {
  return (
    <Stack
      gap={0}
      style={{
        width: 240,
        minWidth: 240,
        borderRight: '1px solid var(--mantine-color-gray-3)',
        height: '100%',
      }}
    >
      <Group justify="space-between" p="sm" style={{ borderBottom: '1px solid var(--mantine-color-gray-3)' }}>
        <Text size="sm" fw={600}>Sessions</Text>
        <Button
          variant="light"
          size="compact-xs"
          leftSection={<IconPlus size={14} />}
          onClick={onCreate}
        >
          New
        </Button>
      </Group>

      <ScrollArea style={{ flex: 1 }} p="xs">
        {sessions.length === 0 && (
          <Text size="xs" c="dimmed" ta="center" py="md">
            No sessions yet
          </Text>
        )}
        {sessions.map((session) => (
          <Group key={session.id} gap={0} wrap="nowrap" style={{ position: 'relative' }}>
            <NavLink
              label={session.title}
              active={session.id === activeSessionId}
              onClick={() => onSelect(session.id)}
              style={{ flex: 1, borderRadius: 4 }}
              styles={{ label: { fontSize: 13, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' } }}
            />
            <ActionIcon
              variant="subtle"
              color="red"
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onDelete(session.id)
              }}
              style={{ position: 'absolute', right: 4, top: '50%', transform: 'translateY(-50%)' }}
            >
              <IconTrash size={14} />
            </ActionIcon>
          </Group>
        ))}
      </ScrollArea>
    </Stack>
  )
}
