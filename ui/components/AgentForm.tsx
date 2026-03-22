import { useState, useEffect } from 'react'
import {
  TextInput,
  Select,
  NumberInput,
  Slider,
  Button,
  Group,
  Stack,
  Alert,
  Text,
  Anchor,
  Combobox,
  InputBase,
  ScrollArea,
  useCombobox,
} from '@mantine/core'
import { IconArrowLeft } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type AgentResponse = components['schemas']['Models.AgentResponse']
type ProviderResponse = components['schemas']['Models.ProviderResponse']

interface AgentFormProps {
  agent?: AgentResponse
  onSaved: () => void
  onCancel: () => void
}

const MODEL_OPTIONS: Record<string, string[]> = {
  google_ai: [
    'gemini-3.1-pro-preview',
    'gemini-3-flash-preview',
    'gemini-3.1-flash-lite-preview',
    'gemini-2.5-pro',
    'gemini-2.5-flash',
    'gemini-2.5-flash-lite',
  ],
  vertex_ai: [
    'gemini-3.1-pro-preview',
    'gemini-3-flash-preview',
    'gemini-3.1-flash-lite-preview',
    'gemini-2.5-pro',
    'gemini-2.5-flash',
    'gemini-2.5-flash-lite',
  ],
  openai: [
    'gpt-5.4',
    'gpt-5.4-mini',
    'gpt-5.4-nano',
    'gpt-5.3-codex',
    'gpt-4o',
    'gpt-4o-mini',
  ],
  anthropic: [
    'claude-opus-4-6-20250205',
    'claude-sonnet-4-6-20250217',
    'claude-opus-4-5-20251202',
    'claude-sonnet-4-5-20250929',
    'claude-haiku-4-5-20251001',
  ],
  xai: [
    'grok-4.20-beta',
    'grok-4-1-fast-reasoning',
    'grok-4-1-fast-non-reasoning',
    'grok-4',
    'grok-code-fast-1',
  ],
  deepseek: ['deepseek-chat', 'deepseek-reasoner'],
}

export function AgentForm({ agent, onSaved, onCancel }: AgentFormProps) {
  const isEdit = !!agent
  const [name, setName] = useState(agent?.name ?? '')
  const [providerId, setProviderId] = useState(agent?.providerId ?? '')
  const [modelId, setModelId] = useState(agent?.modelId ?? '')
  const [systemPromptId, setSystemPromptId] = useState(agent?.systemPromptId ?? '')
  const [temperature, setTemperature] = useState(agent?.temperature ?? 0.95)
  const [topP, setTopP] = useState(agent?.topP ?? 0.95)
  const [topK, setTopK] = useState(agent?.topK ?? 40)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const providersQuery = useQuery({
    queryKey: ['get', '/api/v1/settings/providers'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/settings/providers')
      if (error) throw new Error('Failed to fetch providers')
      return data
    },
  })

  const promptsQuery = useQuery({
    queryKey: ['get', '/api/v1/prompts', { limit: 100, offset: 0 }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/prompts', {
        params: { query: { limit: 100, offset: 0 } },
      })
      if (error) throw new Error('Failed to fetch prompts')
      return data
    },
  })

  const enabledProviders = (providersQuery.data?.providers ?? []).filter(
    (p: ProviderResponse) => p.enabled
  )

  const selectedProvider = enabledProviders.find(
    (p: ProviderResponse) => p.id === providerId
  )
  const selectedProviderType = selectedProvider?.providerType ?? ''

  // Reset model when provider changes (but not on initial load for edit mode)
  const [prevProviderId, setPrevProviderId] = useState(providerId)
  useEffect(() => {
    if (providerId !== prevProviderId) {
      setPrevProviderId(providerId)
      if (prevProviderId !== '') {
        setModelId('')
        setModelSearch('')
      }
    }
  }, [providerId, prevProviderId])

  const presetModels = MODEL_OPTIONS[selectedProviderType] ?? []

  const modelCombobox = useCombobox({
    onDropdownClose: () => modelCombobox.resetSelectedOption(),
  })
  const [modelSearch, setModelSearch] = useState(agent?.modelId ?? '')

  const providerSelectData = enabledProviders.map((p: ProviderResponse) => ({
    value: p.id,
    label: `${p.name} (${p.providerType})`,
  }))

  const promptSelectData = (promptsQuery.data?.prompts ?? []).map((p) => ({
    value: p.id,
    label: p.name,
  }))

  async function handleSubmit() {
    if (!name.trim()) {
      setError('Name is required')
      return
    }
    if (!providerId) {
      setError('Provider is required')
      return
    }
    if (!modelId.trim()) {
      setError('Model is required')
      return
    }

    setSaving(true)
    setError('')

    try {
      if (isEdit) {
        const { error: err } = await fetchClient.PUT('/api/v1/agents/{id}', {
          params: { path: { id: agent.id } },
          body: {
            name: name.trim(),
            providerId,
            modelId: modelId.trim(),
            systemPromptId: systemPromptId || undefined,
            temperature,
            topP,
            topK,
          } as any,
        })
        if (err) throw new Error((err as any).error)
      } else {
        const { error: err } = await fetchClient.POST('/api/v1/agents', {
          body: {
            name: name.trim(),
            providerId,
            modelId: modelId.trim(),
            systemPromptId: systemPromptId || undefined,
            temperature,
            topP,
            topK,
          } as any,
        })
        if (err) throw new Error((err as any).error)
      }
      onSaved()
    } catch (err: any) {
      setError(err.message || 'Failed to save agent')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Stack>
      <Button
        variant="subtle"
        leftSection={<IconArrowLeft size={16} />}
        onClick={onCancel}
        size="sm"
        w="fit-content"
      >
        Back to Agents
      </Button>

      <Text size="xl" fw={600}>
        {isEdit ? 'Edit Agent' : 'New Agent'}
      </Text>

      {error && (
        <Alert color="red" variant="light">
          {error}
        </Alert>
      )}

      <TextInput
        label="Name"
        placeholder="e.g. Code Review Agent"
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
        required
      />

      <div>
        <Select
          label="Provider"
          placeholder="Select a provider"
          data={providerSelectData}
          value={providerId}
          onChange={(val) => setProviderId(val ?? '')}
          searchable
          required
        />
        <Anchor size="xs" href="/settings" mt={4} display="block">
          Configure providers
        </Anchor>
      </div>

      <div>
        <Text size="sm" fw={500} mb={4}>
          Model <span style={{ color: 'var(--mantine-color-red-6)' }}>*</span>
        </Text>
        <Combobox
          store={modelCombobox}
          onOptionSubmit={(val) => {
            setModelId(val)
            setModelSearch(val)
            modelCombobox.closeDropdown()
          }}
        >
          <Combobox.Target>
            <InputBase
              placeholder={
                presetModels.length > 0
                  ? 'Select or type a model name'
                  : 'Type a model name'
              }
              value={modelSearch}
              onChange={(e) => {
                const val = e.currentTarget.value
                setModelSearch(val)
                setModelId(val)
                modelCombobox.openDropdown()
                modelCombobox.resetSelectedOption()
              }}
              onFocus={() => modelCombobox.openDropdown()}
              onBlur={() => modelCombobox.closeDropdown()}
              rightSection={<Combobox.Chevron />}
              rightSectionPointerEvents="none"
            />
          </Combobox.Target>
          <Combobox.Dropdown>
            <Combobox.Options>
              <ScrollArea.Autosize mah={220} type="scroll">
                {presetModels
                  .filter((m) =>
                    m.toLowerCase().includes(modelSearch.toLowerCase())
                  )
                  .map((m) => (
                    <Combobox.Option value={m} key={m}>
                      {m}
                    </Combobox.Option>
                  ))}
                {modelSearch &&
                  !presetModels.some(
                    (m) => m.toLowerCase() === modelSearch.toLowerCase()
                  ) && (
                    <Combobox.Option value={modelSearch}>
                      Use &quot;{modelSearch}&quot;
                    </Combobox.Option>
                  )}
                {!modelSearch && presetModels.length === 0 && (
                  <Combobox.Empty>Type a model name</Combobox.Empty>
                )}
              </ScrollArea.Autosize>
            </Combobox.Options>
          </Combobox.Dropdown>
        </Combobox>
      </div>

      <div>
        <Select
          label="System Prompt"
          placeholder="None (optional)"
          data={promptSelectData}
          value={systemPromptId || null}
          onChange={(val) => setSystemPromptId(val ?? '')}
          searchable
          clearable
        />
        <Anchor size="xs" href="/prompts" mt={4} display="block">
          Add a system prompt
        </Anchor>
      </div>

      <div>
        <Text size="sm" fw={500} mb={4}>
          Temperature ({temperature.toFixed(2)})
        </Text>
        <Group align="center" gap="md">
          <Slider
            value={temperature}
            onChange={setTemperature}
            min={0}
            max={2}
            step={0.05}
            style={{ flex: 1 }}
            label={(v) => v.toFixed(2)}
          />
          <NumberInput
            value={temperature}
            onChange={(val) => setTemperature(typeof val === 'number' ? val : 0.95)}
            min={0}
            max={2}
            step={0.05}
            decimalScale={2}
            w={80}
          />
        </Group>
      </div>

      <div>
        <Text size="sm" fw={500} mb={4}>
          Top P ({topP.toFixed(2)})
        </Text>
        <Group align="center" gap="md">
          <Slider
            value={topP}
            onChange={setTopP}
            min={0}
            max={1}
            step={0.05}
            style={{ flex: 1 }}
            label={(v) => v.toFixed(2)}
          />
          <NumberInput
            value={topP}
            onChange={(val) => setTopP(typeof val === 'number' ? val : 0.95)}
            min={0}
            max={1}
            step={0.05}
            decimalScale={2}
            w={80}
          />
        </Group>
      </div>

      <div>
        <Text size="sm" fw={500} mb={4}>
          Top K
        </Text>
        <NumberInput
          value={topK}
          onChange={(val) => setTopK(typeof val === 'number' ? val : 40)}
          min={1}
          max={500}
          step={1}
          w={120}
        />
      </div>

      <Group>
        <Button onClick={handleSubmit} loading={saving}>
          {isEdit ? 'Save Changes' : 'Create Agent'}
        </Button>
        <Button variant="default" onClick={onCancel}>
          Cancel
        </Button>
      </Group>
    </Stack>
  )
}
