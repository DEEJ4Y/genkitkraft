import { Card, Group, Text, Badge, Code, Button, Stack } from '@mantine/core'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']
type ProviderTypeInfo = components['schemas']['Models.ProviderTypeInfo']

interface ProviderCardProps {
  typeInfo: ProviderTypeInfo
  provider: Provider | undefined
  onConfigure: () => void
  onDelete: () => void
}

export function ProviderCard({ typeInfo, provider, onConfigure, onDelete }: ProviderCardProps) {
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
          <Text fw={600}>{typeInfo.displayName}</Text>
          {typeInfo.comingSoon && (
            <Badge size="xs" variant="light" color="yellow">
              Coming soon
            </Badge>
          )}
        </Group>
        <Badge variant="light" color={configured ? 'green' : 'gray'}>
          {configured ? 'Configured' : 'Not configured'}
        </Badge>
      </Group>

      {configured && (
        <Stack gap="xs" mb="md">
          {provider.apiKey && (
            <Group justify="space-between">
              <Text size="sm" c="dimmed">
                API Key
              </Text>
              <Code>{provider.apiKey}</Code>
            </Group>
          )}
          {provider.baseUrl && (
            <Group justify="space-between">
              <Text size="sm" c="dimmed">
                Base URL
              </Text>
              <Code>{provider.baseUrl}</Code>
            </Group>
          )}
          {provider.config && Object.keys(provider.config).length > 0 && (
            Object.entries(provider.config).map(([key, value]) => {
              const fieldInfo = typeInfo.configFields?.find((f) => f.name === key)
              const isSensitive = fieldInfo?.sensitive
              return (
                <Group key={key} justify="space-between">
                  <Text size="sm" c="dimmed">
                    {fieldInfo?.label ?? key}
                  </Text>
                  <Code>{isSensitive ? '••••••••' : value}</Code>
                </Group>
              )
            })
          )}
        </Stack>
      )}

      <Group gap="xs">
        <Button size="xs" color="dark" onClick={onConfigure} disabled={!!typeInfo.comingSoon}>
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
