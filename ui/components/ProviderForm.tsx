import { useState, FormEvent } from 'react'
import { Modal, TextInput, PasswordInput, Button, Group, Alert, Stack, Text } from '@mantine/core'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']
type ProviderTypeInfo = components['schemas']['Models.ProviderTypeInfo']

interface ProviderFormProps {
  provider: Provider | undefined
  typeInfo: ProviderTypeInfo
  onSaved: () => void
  onCancel: () => void
}

export function ProviderForm({ provider, typeInfo, onSaved, onCancel }: ProviderFormProps) {
  const isEdit = !!provider

  const [name, setName] = useState(provider?.name ?? typeInfo.displayName)
  const [apiKey, setApiKey] = useState('')
  const [baseUrl, setBaseUrl] = useState(provider?.baseUrl ?? (typeInfo.baseUrlDefault ?? ''))
  const [configFields, setConfigFields] = useState<Record<string, string>>(() => {
    const initial: Record<string, string> = {}
    for (const field of typeInfo.configFields ?? []) {
      initial[field.name] = provider?.config?.[field.name] ?? ''
    }
    return initial
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [testing, setTesting] = useState(false)

  function setConfigField(fieldName: string, value: string) {
    setConfigFields((prev) => ({ ...prev, [fieldName]: value }))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSaving(true)

    try {
      // Build config object (only non-empty values)
      const config: Record<string, string> = {}
      for (const [key, value] of Object.entries(configFields)) {
        if (value) config[key] = value
      }
      const hasConfig = Object.keys(config).length > 0

      if (isEdit) {
        const body: Record<string, unknown> = { name }
        if (apiKey) body.apiKey = apiKey
        if (typeInfo.requiresBaseUrl || baseUrl) body.baseUrl = baseUrl
        if (hasConfig) body.config = config
        const { error: err } = await fetchClient.PUT('/api/v1/settings/providers/{id}', {
          params: { path: { id: provider.id } },
          body: body as any,
        })
        if (err) throw new Error(err.error)
      } else {
        if (typeInfo.requiresApiKey && !apiKey) {
          setError('API key is required')
          setSaving(false)
          return
        }
        if (typeInfo.requiresBaseUrl && !baseUrl) {
          setError('Base URL is required')
          setSaving(false)
          return
        }
        const body: any = { name, providerType: typeInfo.type }
        if (apiKey) body.apiKey = apiKey
        if (baseUrl) body.baseUrl = baseUrl
        if (hasConfig) body.config = config
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
      title={`${isEdit ? 'Edit' : 'Configure'} ${typeInfo.displayName}`}
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

          {typeInfo.requiresApiKey && (
            <PasswordInput
              label="API Key"
              value={apiKey}
              onChange={(e) => setApiKey(e.currentTarget.value)}
              placeholder={isEdit ? 'Leave blank to keep existing' : 'Enter API key'}
              required={!isEdit}
            />
          )}

          {(typeInfo.requiresBaseUrl || typeInfo.baseUrlDefault) && (
            <TextInput
              label={`Base URL${typeInfo.requiresBaseUrl ? '' : ' (optional)'}`}
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.currentTarget.value)}
              placeholder={typeInfo.baseUrlDefault ?? ''}
              required={typeInfo.requiresBaseUrl}
            />
          )}

          {typeInfo.configFields && typeInfo.configFields.length > 0 && (
            <>
              {typeInfo.configFields.map((field) => {
                const InputComponent = field.sensitive ? PasswordInput : TextInput
                return (
                  <InputComponent
                    key={field.name}
                    label={`${field.label}${field.required ? '' : ' (optional)'}`}
                    value={configFields[field.name] ?? ''}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setConfigField(field.name, e.currentTarget.value)}
                    placeholder={isEdit && field.sensitive ? 'Leave blank to keep existing' : (field.placeholder ?? '')}
                    required={field.required && !isEdit}
                  />
                )
              })}
            </>
          )}

          {typeInfo.envVarHint && (
            <Text size="xs" c="dimmed">
              Env var: {typeInfo.envVarHint}
            </Text>
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
