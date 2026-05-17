import { useState } from 'react'
import {
  ActionIcon,
  Box,
  Button,
  Collapse,
  Combobox,
  Group,
  InputBase,
  NumberInput,
  ScrollArea,
  Select,
  Slider,
  Switch,
  Text,
  useCombobox,
} from '@mantine/core'
import { IconAdjustments, IconDeviceFloppy } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { fetchClient } from '../../lib/api/client'
import { MODEL_OPTIONS } from '../../lib/model-options'
import type { components } from '../../lib/api/schema'

type AgentResponse = components['schemas']['Models.AgentResponse']
type ProviderResponse = components['schemas']['Models.ProviderResponse']

const DEFAULT_TEMPERATURE = 0.95
const DEFAULT_TOP_P = 0.95
const DEFAULT_TOP_K = 40
const DEFAULT_MAX_TOOL_CALLS = 10

export interface PlaygroundConfig {
  providerId: string
  modelId: string
  systemPromptId: string
  temperatureEnabled: boolean
  temperature: number
  topPEnabled: boolean
  topP: number
  topKEnabled: boolean
  topK: number
  maxToolCalls: number
  streaming: boolean
}

interface PlaygroundConfigBarProps {
  agent: AgentResponse
  config: PlaygroundConfig
  onChange: (config: PlaygroundConfig) => void
  onSaveToAgent: () => void
  hasToolOverrides?: boolean
}

export function PlaygroundConfigBar({ agent, config, onChange, onSaveToAgent, hasToolOverrides }: PlaygroundConfigBarProps) {
  const [opened, setOpened] = useState(false)
  const [modelSearch, setModelSearch] = useState(config.modelId ?? '')
  const modelCombobox = useCombobox({
    onDropdownClose: () => modelCombobox.resetSelectedOption(),
  })

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
    (p: ProviderResponse) => p.id === config.providerId
  )
  const selectedProviderType = selectedProvider?.providerType ?? ''

  const presetModels = MODEL_OPTIONS[selectedProviderType] ?? []

  const providerSelectData = enabledProviders.map((p: ProviderResponse) => ({
    value: p.id,
    label: `${p.name} (${p.providerType})`,
  }))

  const promptSelectData = (promptsQuery.data?.prompts ?? []).map((p) => ({
    value: p.id,
    label: p.name,
  }))

  const hasOverrides =
    config.providerId !== agent.providerId ||
    config.modelId !== agent.modelId ||
    config.systemPromptId !== (agent.systemPromptId ?? '') ||
    config.temperatureEnabled !== agent.temperatureEnabled ||
    config.temperature !== agent.temperature ||
    config.topPEnabled !== agent.topPEnabled ||
    config.topP !== agent.topP ||
    config.topKEnabled !== agent.topKEnabled ||
    config.topK !== agent.topK ||
    !config.streaming ||
    hasToolOverrides

  function handleProviderChange(val: string | null) {
    setModelSearch('')
    onChange({ ...config, providerId: val ?? '', modelId: '' })
  }

  return (
    <Box style={{ borderBottom: '1px solid var(--mantine-color-gray-3)' }}>
      <Group justify="space-between" p="xs" px="md">
        <Group gap="xs">
          <ActionIcon
            variant={opened ? 'filled' : 'subtle'}
            size="sm"
            onClick={() => setOpened(!opened)}
          >
            <IconAdjustments size={16} />
          </ActionIcon>
          <Text size="xs" c="dimmed">
            {config.modelId || 'No model selected'}
            {hasOverrides && ' (modified)'}
          </Text>
        </Group>
        {hasOverrides && (
          <Button
            variant="light"
            size="compact-xs"
            leftSection={<IconDeviceFloppy size={14} />}
            onClick={onSaveToAgent}
          >
            Save to Agent
          </Button>
        )}
      </Group>

      <Collapse in={opened}>
        <Box p="md" pt={0}>
          <Group grow gap="md" align="flex-start">
            <Select
              label="Provider"
              size="xs"
              data={providerSelectData}
              value={config.providerId}
              onChange={handleProviderChange}
              searchable
            />
            <div style={{ flex: 1 }}>
              <Text size="xs" fw={500} mb={4}>
                Model
              </Text>
              <Combobox
                store={modelCombobox}
                onOptionSubmit={(val) => {
                  setModelSearch(val)
                  onChange({ ...config, modelId: val })
                  modelCombobox.closeDropdown()
                }}
              >
                <Combobox.Target>
                  <InputBase
                    size="xs"
                    placeholder={
                      !config.providerId
                        ? 'Select a provider first'
                        : presetModels.length > 0
                          ? 'Select or type a model name'
                          : 'Type a model name'
                    }
                    value={modelSearch}
                    onChange={(e) => {
                      const val = e.currentTarget.value
                      setModelSearch(val)
                      onChange({ ...config, modelId: val })
                      modelCombobox.openDropdown()
                      modelCombobox.resetSelectedOption()
                    }}
                    onFocus={() => modelCombobox.openDropdown()}
                    onBlur={() => modelCombobox.closeDropdown()}
                    rightSection={<Combobox.Chevron />}
                    rightSectionPointerEvents="none"
                    disabled={!config.providerId}
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
          </Group>
          <Group grow gap="md" mt="sm" align="flex-start">
            <Select
              label="System Prompt"
              size="xs"
              data={promptSelectData}
              value={config.systemPromptId || null}
              onChange={(val) => onChange({ ...config, systemPromptId: val ?? '' })}
              searchable
              clearable
              placeholder="None"
            />
          </Group>
          <Group grow gap="md" mt="sm" align="flex-start">
            <div>
              <Group gap="xs" mb={2}>
                <Switch
                  size="xs"
                  checked={config.temperatureEnabled}
                  onChange={(e) => {
                    const enabled = e.currentTarget.checked
                    onChange({
                      ...config,
                      temperatureEnabled: enabled,
                      temperature: enabled && !config.temperatureEnabled ? DEFAULT_TEMPERATURE : config.temperature,
                    })
                  }}
                />
                <Text size="xs" fw={500} c={config.temperatureEnabled ? undefined : 'dimmed'}>
                  Temperature {config.temperatureEnabled ? `(${config.temperature.toFixed(2)})` : '(off)'}
                </Text>
              </Group>
              <Slider
                value={config.temperature}
                onChange={(val) => onChange({ ...config, temperature: val })}
                min={0}
                max={2}
                step={0.05}
                size="xs"
                label={(v) => v.toFixed(2)}
                disabled={!config.temperatureEnabled}
              />
            </div>
            <div>
              <Group gap="xs" mb={2}>
                <Switch
                  size="xs"
                  checked={config.topPEnabled}
                  onChange={(e) => {
                    const enabled = e.currentTarget.checked
                    onChange({
                      ...config,
                      topPEnabled: enabled,
                      topP: enabled && !config.topPEnabled ? DEFAULT_TOP_P : config.topP,
                    })
                  }}
                />
                <Text size="xs" fw={500} c={config.topPEnabled ? undefined : 'dimmed'}>
                  Top P {config.topPEnabled ? `(${config.topP.toFixed(2)})` : '(off)'}
                </Text>
              </Group>
              <Slider
                value={config.topP}
                onChange={(val) => onChange({ ...config, topP: val })}
                min={0}
                max={1}
                step={0.05}
                size="xs"
                label={(v) => v.toFixed(2)}
                disabled={!config.topPEnabled}
              />
            </div>
          </Group>
          <Group grow gap="md" mt="sm" align="flex-start">
            <div>
              <Group gap="xs" mb={2}>
                <Switch
                  size="xs"
                  checked={config.topKEnabled}
                  onChange={(e) => {
                    const enabled = e.currentTarget.checked
                    onChange({
                      ...config,
                      topKEnabled: enabled,
                      topK: enabled && !config.topKEnabled ? DEFAULT_TOP_K : config.topK,
                    })
                  }}
                />
                <Text size="xs" fw={500} c={config.topKEnabled ? undefined : 'dimmed'}>
                  Top K {config.topKEnabled ? '' : '(off)'}
                </Text>
              </Group>
              <NumberInput
                value={config.topK}
                onChange={(val) => onChange({ ...config, topK: typeof val === 'number' ? val : DEFAULT_TOP_K })}
                min={1}
                max={500}
                size="xs"
                disabled={!config.topKEnabled}
              />
            </div>
            <div>
              <Text size="xs" fw={500} mb={2}>
                Max Tool Calls
              </Text>
              <NumberInput
                value={config.maxToolCalls}
                onChange={(val) => onChange({ ...config, maxToolCalls: typeof val === 'number' ? val : DEFAULT_MAX_TOOL_CALLS })}
                min={1}
                max={100}
                size="xs"
              />
            </div>
          </Group>
          <Group grow gap="md" mt="sm" align="flex-start">
            <div>
              <Group gap="xs" mb={2}>
                <Switch
                  size="xs"
                  checked={config.streaming}
                  onChange={(e) => onChange({ ...config, streaming: e.currentTarget.checked })}
                />
                <Text size="xs" fw={500} c={config.streaming ? undefined : 'dimmed'}>
                  Streaming {config.streaming ? '(on)' : '(off)'}
                </Text>
              </Group>
              <Text size="xs" c="dimmed">
                Disable for providers that don&apos;t support streaming.
              </Text>
            </div>
          </Group>
        </Box>
      </Collapse>
    </Box>
  )
}
