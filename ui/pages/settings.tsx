import { useState } from 'react'
import { Title, Text, Stack, Loader, Alert, Center, Tabs } from '@mantine/core'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import { ProviderCard } from '../components/ProviderCard'
import { ProviderForm } from '../components/ProviderForm'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']
type ProviderType = components['schemas']['Models.ProviderType']
type ProviderTypeInfo = components['schemas']['Models.ProviderTypeInfo']

function ProvidersTab() {
  const queryClient = useQueryClient()

  const providerTypesQuery = useQuery({
    queryKey: ['get', '/api/v1/provider-types'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/provider-types')
      if (error) throw new Error('Failed to fetch provider types')
      return data
    },
    staleTime: Infinity,
  })

  const providersQuery = useQuery({
    queryKey: ['get', '/api/v1/settings/providers'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/settings/providers')
      if (error) throw new Error('Failed to fetch providers')
      return data
    },
  })

  const [editingType, setEditingType] = useState<ProviderType | null>(null)

  const providerTypes: ProviderTypeInfo[] = providerTypesQuery.data?.providerTypes ?? []

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

  const editingTypeInfo = editingType ? providerTypes.find((pt) => pt.type === editingType) : undefined
  const editingProvider = editingType ? getProviderByType(editingType) : undefined

  const isLoading = providerTypesQuery.isPending || providersQuery.isPending

  return (
    <>
      {isLoading && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {(providerTypesQuery.error || providersQuery.error) && (
        <Alert color="red" variant="light" mb="md">
          Failed to load providers.
        </Alert>
      )}

      <Stack gap="sm">
        {providerTypes.map((pt) => (
          <ProviderCard
            key={pt.type}
            typeInfo={pt}
            provider={getProviderByType(pt.type)}
            onConfigure={() => setEditingType(pt.type)}
            onDelete={() => {
              const p = getProviderByType(pt.type)
              if (p) handleDelete(p)
            }}
          />
        ))}
      </Stack>

      {editingType && editingTypeInfo && (
        <ProviderForm
          provider={editingProvider}
          typeInfo={editingTypeInfo}
          onSaved={handleSaved}
          onCancel={() => setEditingType(null)}
        />
      )}
    </>
  )
}

export default function SettingsPage() {
  return (
    <>
      <Title order={2} mb="lg">
        Settings
      </Title>

      <Tabs defaultValue="providers">
        <Tabs.List mb="md">
          <Tabs.Tab value="providers">LLM Providers</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="providers">
          <Text size="sm" c="dimmed" mb="md">
            Configure API keys for the LLM providers you want to use.
          </Text>
          <ProvidersTab />
        </Tabs.Panel>
      </Tabs>
    </>
  )
}
