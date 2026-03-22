import { Card, Group, Text, Badge, Code, Stack, ActionIcon, Tooltip } from '@mantine/core'
import { IconEdit, IconTrash } from '@tabler/icons-react'
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
    <Card shadow="xs" padding="md" radius="sm" withBorder>
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <div style={{ flex: 1, minWidth: 0 }}>
          <Group gap="xs">
            <div
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: configured && provider.enabled ? 'var(--mantine-color-green-6)' : 'var(--mantine-color-gray-4)',
              }}
            />
            <Text fw={600} size="md">
              {typeInfo.displayName}
            </Text>
            {typeInfo.comingSoon && (
              <Badge size="xs" variant="light" color="yellow">
                Coming soon
              </Badge>
            )}
          </Group>

          <Badge variant="light" color={configured ? 'green' : 'gray'} size="sm" mt={4}>
            {configured ? 'Configured' : 'Not configured'}
          </Badge>
        </div>

        <Group gap="xs" wrap="nowrap">
          <Tooltip label={configured ? 'Edit' : 'Configure'}>
            <ActionIcon variant="subtle" onClick={onConfigure} disabled={!!typeInfo.comingSoon}>
              <IconEdit size={18} />
            </ActionIcon>
          </Tooltip>
          {configured && (
            <Tooltip label="Delete">
              <ActionIcon variant="subtle" color="red" onClick={onDelete}>
                <IconTrash size={18} />
              </ActionIcon>
            </Tooltip>
          )}
        </Group>
      </Group>

      {configured && (
        <Stack gap="xs" mt="xs">
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
    </Card>
  )
}
