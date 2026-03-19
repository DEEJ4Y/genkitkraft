import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import { useAuth } from '../lib/auth'
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
  const { user, logout, isAuthRequired } = useAuth()
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
    <div style={styles.page}>
      <header style={styles.header}>
        <div style={styles.headerLeft}>
          <a href="/" style={styles.backLink}>&larr; Back</a>
          <h1 style={styles.heading}>Settings</h1>
        </div>
        {isAuthRequired && user && (
          <div style={styles.userSection}>
            <span style={styles.username}>{user}</span>
            <button onClick={logout} style={styles.signOutButton}>Sign out</button>
          </div>
        )}
      </header>

      <main>
        <h2 style={styles.sectionTitle}>LLM Providers</h2>
        <p style={styles.sectionDesc}>
          Configure API keys for the LLM providers you want to use.
        </p>

        {providersQuery.isPending && <p style={styles.loading}>Loading...</p>}
        {providersQuery.error && <p style={styles.errorText}>Failed to load providers.</p>}

        <div style={styles.cardGrid}>
          {PROVIDER_TYPES.map(({ type, label }) => (
            <ProviderCard
              key={type}
              providerType={type}
              label={label}
              provider={getProviderByType(type)}
              onConfigure={() => setEditingType(type)}
              onDelete={() => {
                const p = getProviderByType(type)
                if (p) handleDelete(p)
              }}
            />
          ))}
        </div>
      </main>

      {editingType && (
        <ProviderForm
          provider={editingProvider}
          providerType={editingType}
          providerLabel={editingLabel}
          onSaved={handleSaved}
          onCancel={() => setEditingType(null)}
        />
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    padding: '24px',
    maxWidth: '800px',
    margin: '0 auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '32px',
  },
  headerLeft: {
    display: 'flex',
    alignItems: 'center',
    gap: '16px',
  },
  backLink: {
    fontSize: '14px',
    color: '#666',
    textDecoration: 'none',
  },
  heading: {
    margin: 0,
    fontSize: '20px',
  },
  userSection: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  username: {
    fontSize: '14px',
    color: '#666',
  },
  signOutButton: {
    padding: '6px 12px',
    fontSize: '13px',
    backgroundColor: 'transparent',
    border: '1px solid #ddd',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  sectionTitle: {
    margin: '0 0 4px 0',
    fontSize: '16px',
    fontWeight: 600,
  },
  sectionDesc: {
    margin: '0 0 20px 0',
    fontSize: '14px',
    color: '#666',
  },
  loading: {
    color: '#666',
    fontSize: '14px',
  },
  errorText: {
    color: '#d32f2f',
    fontSize: '14px',
  },
  cardGrid: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '12px',
  },
}
