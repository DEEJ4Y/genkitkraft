import { useState, FormEvent } from 'react'
import { Modal, TextInput, PasswordInput, Button, Group, Alert, Stack } from '@mantine/core'
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
    <Modal
      opened
      onClose={onCancel}
      title={`${isEdit ? 'Edit' : 'Configure'} ${providerLabel}`}
      size="md"
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            required
          />

          <PasswordInput
            label="API Key"
            value={apiKey}
            onChange={(e) => setApiKey(e.currentTarget.value)}
            placeholder={isEdit ? 'Leave blank to keep existing' : 'Enter API key'}
            required={!isEdit}
          />

          {providerType === 'openai' && (
            <TextInput
              label="Base URL (optional)"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.currentTarget.value)}
              placeholder="https://api.openai.com"
            />
          )}

          {error && (
            <Alert color="red" variant="light">
              {error}
            </Alert>
          )}

          {testResult && (
            <Alert color={testResult.success ? 'green' : 'red'} variant="light">
              {testResult.message}
            </Alert>
          )}

          <Group justify="space-between">
            <div>
              {isEdit && (
                <Button variant="default" onClick={handleTest} loading={testing}>
                  Test Connection
                </Button>
              )}
            </div>
            <Group gap="xs">
              <Button variant="default" onClick={onCancel}>
                Cancel
              </Button>
              <Button type="submit" loading={saving} color="dark">
                Save
              </Button>
            </Group>
          </Group>
        </Stack>
      </form>
    </Modal>
  )
}
