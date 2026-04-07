import { useState } from 'react'
import {
  ActionIcon,
  Box,
  Button,
  Collapse,
  Group,
  NumberInput,
  Select,
  Slider,
  Text,
} from '@mantine/core'
import { IconAdjustments, IconDeviceFloppy } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { fetchClient } from '../../lib/api/client'
import { MODEL_OPTIONS } from '../../lib/model-options'
import type { components } from '../../lib/api/schema'

type AgentResponse = components['schemas']['Models.AgentResponse']
type ProviderResponse = components['schemas']['Models.ProviderResponse']

export interface PlaygroundConfig {
  providerId: string
  modelId: string
  systemPromptId: string
  temperature: number
  topP: number
  topK: number
}

interface PlaygroundConfigBarProps {
  agent: AgentResponse
  config: PlaygroundConfig
  onChange: (config: PlaygroundConfig) => void
  onSaveToAgent: () => void
}

export function PlaygroundConfigBar({ agent, config, onChange, onSaveToAgent }: PlaygroundConfigBarProps) {
  const [opened, setOpened] = useState(false)

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
  const modelSelectData = presetModels.map((m) => ({ value: m, label: m }))

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
    config.temperature !== agent.temperature ||
    config.topP !== agent.topP ||
    config.topK !== agent.topK

  function handleProviderChange(val: string | null) {
    // Reset model when provider changes
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
            <Select
              label="Model"
              size="xs"
              data={modelSelectData}
              value={config.modelId || null}
              onChange={(val) => onChange({ ...config, modelId: val ?? '' })}
              searchable
              placeholder={presetModels.length > 0 ? 'Select a model' : 'Select a provider first'}
              disabled={!config.providerId}
            />
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
              <Text size="xs" fw={500} mb={2}>Temperature ({config.temperature.toFixed(2)})</Text>
              <Slider
                value={config.temperature}
                onChange={(val) => onChange({ ...config, temperature: val })}
                min={0}
                max={2}
                step={0.05}
                size="xs"
                label={(v) => v.toFixed(2)}
              />
            </div>
            <div>
              <Text size="xs" fw={500} mb={2}>Top P ({config.topP.toFixed(2)})</Text>
              <Slider
                value={config.topP}
                onChange={(val) => onChange({ ...config, topP: val })}
                min={0}
                max={1}
                step={0.05}
                size="xs"
                label={(v) => v.toFixed(2)}
              />
            </div>
            <div>
              <Text size="xs" fw={500} mb={2}>Top K</Text>
              <NumberInput
                value={config.topK}
                onChange={(val) => onChange({ ...config, topK: typeof val === 'number' ? val : 40 })}
                min={1}
                max={500}
                size="xs"
              />
            </div>
          </Group>
        </Box>
      </Collapse>
    </Box>
  )
}
