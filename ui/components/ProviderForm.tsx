import { useState, FormEvent } from 'react'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']
type ProviderType = components['schemas']['Models.ProviderType']

interface ProviderFormProps {
  provider: Provider | undefined
  providerType: ProviderType
  providerLabel: string
  onSaved: () => void
  onCancel: () => void
}

export function ProviderForm({ provider, providerType, providerLabel, onSaved, onCancel }: ProviderFormProps) {
  const isEdit = !!provider

  const [name, setName] = useState(provider?.name ?? `${providerLabel}`)
  const [apiKey, setApiKey] = useState('')
  const [baseUrl, setBaseUrl] = useState(provider?.baseUrl ?? '')
  const [showKey, setShowKey] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [testing, setTesting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSaving(true)

    try {
      if (isEdit) {
        const body: Record<string, unknown> = { name }
        if (apiKey) body.apiKey = apiKey
        if (providerType === 'openai') body.baseUrl = baseUrl
        const { error: err } = await fetchClient.PUT('/api/v1/settings/providers/{id}', {
          params: { path: { id: provider.id } },
          body: body as any,
        })
        if (err) throw new Error(err.error)
      } else {
        if (!apiKey) {
          setError('API key is required')
          setSaving(false)
          return
        }
        const body: any = { name, providerType, apiKey }
        if (providerType === 'openai' && baseUrl) body.baseUrl = baseUrl
        const { error: err } = await fetchClient.POST('/api/v1/settings/providers', { body })
        if (err) throw new Error(err.error)
      }
      onSaved()
    } catch (err: any) {
      setError(err.message || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  async function handleTest() {
    if (!provider) return
    setTesting(true)
    setTestResult(null)
    try {
      const { data, error: err } = await fetchClient.POST('/api/v1/settings/providers/{id}/test', {
        params: { path: { id: provider.id } },
      })
      if (err) throw new Error(err.error)
      if (data) setTestResult({ success: data.success, message: data.message })
    } catch (err: any) {
      setTestResult({ success: false, message: err.message || 'Test failed' })
    } finally {
      setTesting(false)
    }
  }

  return (
    <div style={styles.overlay}>
      <div style={styles.card}>
        <h2 style={styles.title}>{isEdit ? 'Edit' : 'Configure'} {providerLabel}</h2>
        <form onSubmit={handleSubmit} style={styles.form}>
          <label style={styles.label}>
            Name
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              style={styles.input}
            />
          </label>

          <label style={styles.label}>
            API Key
            <div style={{ position: 'relative' }}>
              <input
                type={showKey ? 'text' : 'password'}
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder={isEdit ? 'Leave blank to keep existing' : 'Enter API key'}
                required={!isEdit}
                style={styles.input}
              />
              <button
                type="button"
                onClick={() => setShowKey(!showKey)}
                style={styles.toggleKey}
              >
                {showKey ? 'Hide' : 'Show'}
              </button>
            </div>
          </label>

          {providerType === 'openai' && (
            <label style={styles.label}>
              Base URL (optional)
              <input
                type="text"
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
                placeholder="https://api.openai.com"
                style={styles.input}
              />
            </label>
          )}

          {error && <p style={styles.error}>{error}</p>}

          {testResult && (
            <p style={{ ...styles.testResult, color: testResult.success ? '#2e7d32' : '#d32f2f' }}>
              {testResult.message}
            </p>
          )}

          <div style={styles.actions}>
            {isEdit && (
              <button type="button" onClick={handleTest} disabled={testing} style={styles.testButton}>
                {testing ? 'Testing...' : 'Test Connection'}
              </button>
            )}
            <div style={{ flex: 1 }} />
            <button type="button" onClick={onCancel} style={styles.cancelButton}>
              Cancel
            </button>
            <button type="submit" disabled={saving} style={styles.saveButton}>
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  overlay: {
    position: 'fixed',
    inset: 0,
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0,0,0,0.4)',
    zIndex: 1000,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: '8px',
    padding: '32px',
    width: '100%',
    maxWidth: '440px',
    boxShadow: '0 4px 24px rgba(0,0,0,0.15)',
  },
  title: {
    margin: '0 0 20px 0',
    fontSize: '18px',
    fontWeight: 600,
  },
  form: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '16px',
  },
  label: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '4px',
    fontSize: '13px',
    fontWeight: 500,
    color: '#333',
  },
  input: {
    padding: '10px 12px',
    fontSize: '14px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box' as const,
  },
  toggleKey: {
    position: 'absolute' as const,
    right: '8px',
    top: '50%',
    transform: 'translateY(-50%)',
    padding: '4px 8px',
    fontSize: '12px',
    backgroundColor: 'transparent',
    border: '1px solid #ddd',
    borderRadius: '4px',
    cursor: 'pointer',
    color: '#666',
  },
  error: {
    margin: 0,
    fontSize: '13px',
    color: '#d32f2f',
  },
  testResult: {
    margin: 0,
    fontSize: '13px',
  },
  actions: {
    display: 'flex',
    gap: '8px',
    alignItems: 'center',
    marginTop: '4px',
  },
  testButton: {
    padding: '8px 16px',
    fontSize: '13px',
    backgroundColor: 'transparent',
    border: '1px solid #ddd',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  cancelButton: {
    padding: '8px 16px',
    fontSize: '13px',
    backgroundColor: 'transparent',
    border: '1px solid #ddd',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  saveButton: {
    padding: '8px 16px',
    fontSize: '13px',
    fontWeight: 600,
    backgroundColor: '#111',
    color: '#fff',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
  },
}
