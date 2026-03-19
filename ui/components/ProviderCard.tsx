import { Card, Group, Text, Badge, Code, Button, Stack } from '@mantine/core'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']

interface ProviderCardProps {
  label: string
  provider: Provider | undefined
  onConfigure: () => void
  onDelete: () => void
}

export function ProviderCard({ label, provider, onConfigure, onDelete }: ProviderCardProps) {
  const configured = !!provider

  return (
    <Card withBorder padding="lg" radius="md">
      <Group justify="space-between" mb="sm">
        <Group gap="xs">
          <div
            style={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              backgroundColor: configured && provider.enabled ? 'var(--mantine-color-green-6)' : 'var(--mantine-color-gray-4)',
            }}
          />
          <Text fw={600}>{label}</Text>
        </Group>
        <Badge variant="light" color={configured ? 'green' : 'gray'}>
          {configured ? 'Configured' : 'Not configured'}
        </Badge>
      </Group>

      {configured && (
        <Stack gap="xs" mb="md">
          <Group justify="space-between">
            <Text size="sm" c="dimmed">
              API Key
            </Text>
            <Code>{provider.apiKey}</Code>
          </Group>
          {provider.baseUrl && (
            <Group justify="space-between">
              <Text size="sm" c="dimmed">
                Base URL
              </Text>
              <Code>{provider.baseUrl}</Code>
            </Group>
          )}
        </Stack>
      )}

      <Group gap="xs">
        <Button size="xs" color="dark" onClick={onConfigure}>
          {configured ? 'Edit' : 'Configure'}
        </Button>
        {configured && (
          <Button size="xs" variant="outline" color="red" onClick={onDelete}>
            Delete
          </Button>
        )}
      </Group>
    </Card>
  )
}
