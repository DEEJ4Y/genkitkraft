import { useState } from 'react'
import { Title, Text, Stack, Loader, Alert, Center } from '@mantine/core'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import { ProviderCard } from '../components/ProviderCard'
import { ProviderForm } from '../components/ProviderForm'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']
type ProviderType = components['schemas']['Models.ProviderType']

const PROVIDER_TYPES: { type: ProviderType; label: string }[] = [
  { type: 'anthropic', label: 'Anthropic' },
  { type: 'openai', label: 'OpenAI' },
  { type: 'googleai', label: 'Google AI' },
]

export default function SettingsPage() {
  const queryClient = useQueryClient()

  const providersQuery = useQuery({
    queryKey: ['get', '/api/v1/settings/providers'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/settings/providers')
      if (error) throw new Error('Failed to fetch providers')
      return data
    },
  })

  const [editingType, setEditingType] = useState<ProviderType | null>(null)

  function getProviderByType(type: ProviderType): Provider | undefined {
    return providersQuery.data?.providers?.find((p) => p.providerType === type)
  }

  function handleSaved() {
    setEditingType(null)
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/settings/providers'] })
  }

  async function handleDelete(provider: Provider) {
    if (!confirm(`Delete ${provider.name}? This will remove the stored API key.`)) return
    await fetchClient.DELETE('/api/v1/settings/providers/{id}', {
      params: { path: { id: provider.id } },
    })
    queryClient.invalidateQueries({ queryKey: ['get', '/api/v1/settings/providers'] })
  }

  const editingProvider = editingType ? getProviderByType(editingType) : undefined
  const editingLabel = editingType ? PROVIDER_TYPES.find((p) => p.type === editingType)?.label ?? '' : ''

  return (
    <>
      <Title order={2} mb={4}>
        LLM Providers
      </Title>
      <Text size="sm" c="dimmed" mb="lg">
        Configure API keys for the LLM providers you want to use.
      </Text>

      {providersQuery.isPending && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {providersQuery.error && (
        <Alert color="red" variant="light" mb="md">
          Failed to load providers.
        </Alert>
      )}

      <Stack gap="sm">
        {PROVIDER_TYPES.map(({ type, label }) => (
          <ProviderCard
            key={type}
            label={label}
            provider={getProviderByType(type)}
            onConfigure={() => setEditingType(type)}
            onDelete={() => {
              const p = getProviderByType(type)
              if (p) handleDelete(p)
            }}
          />
        ))}
      </Stack>

      {editingType && (
        <ProviderForm
          provider={editingProvider}
          providerType={editingType}
          providerLabel={editingLabel}
          onSaved={handleSaved}
          onCancel={() => setEditingType(null)}
        />
      )}
    </>
  )
}
