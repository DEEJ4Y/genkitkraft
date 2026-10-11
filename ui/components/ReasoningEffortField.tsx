import { useState } from 'react'
import {
  Combobox,
  Group,
  InputBase,
  ScrollArea,
  Switch,
  Text,
  useCombobox,
} from '@mantine/core'
import {
  REASONING_EFFORT_OPTIONS,
  defaultReasoningEffort,
  supportsReasoningEffort,
} from '../lib/reasoning-options'

interface ReasoningEffortFieldProps {
  providerType: string
  enabled: boolean
  value: string
  onChange: (enabled: boolean, value: string) => void
  size?: 'xs' | 'sm'
}

// Toggle plus a value field for the reasoning effort (thinking) option.
// The dropdown lists the values that the selected provider type documents.
// Any other text that the user types is kept as a custom value.
export function ReasoningEffortField({
  providerType,
  enabled,
  value,
  onChange,
  size = 'sm',
}: ReasoningEffortFieldProps) {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  })
  const [focused, setFocused] = useState(false)

  if (!supportsReasoningEffort(providerType)) return null

  const info = REASONING_EFFORT_OPTIONS[providerType]
  const search = value.trim()
  const isPreset = info.presets.some((p) => p.value.toLowerCase() === search.toLowerCase())

  function handleToggle(checked: boolean) {
    // Fill in the provider default when the option is switched on with no value.
    const nextValue = checked && !value.trim() ? defaultReasoningEffort(providerType) : value
    onChange(checked, nextValue)
  }

  return (
    <div>
      <Group gap="xs" mb={4}>
        <Switch
          size={size}
          checked={enabled}
          onChange={(e) => handleToggle(e.currentTarget.checked)}
        />
        <Text size={size} fw={500} c={enabled ? undefined : 'dimmed'}>
          Reasoning Effort {enabled ? (value.trim() ? `(${value.trim()})` : '') : '(off)'}
        </Text>
      </Group>
      <Combobox
        store={combobox}
        onOptionSubmit={(val) => {
          onChange(enabled, val)
          combobox.closeDropdown()
        }}
      >
        <Combobox.Target>
          <InputBase
            size={size}
            placeholder="Select or type a value"
            value={value}
            disabled={!enabled}
            onChange={(e) => {
              onChange(enabled, e.currentTarget.value)
              combobox.openDropdown()
              combobox.resetSelectedOption()
            }}
            onFocus={() => {
              setFocused(true)
              combobox.openDropdown()
            }}
            onBlur={() => {
              setFocused(false)
              combobox.closeDropdown()
            }}
            rightSection={<Combobox.Chevron />}
            rightSectionPointerEvents="none"
            aria-label="Reasoning effort value"
          />
        </Combobox.Target>
        <Combobox.Dropdown>
          <Combobox.Options>
            <ScrollArea.Autosize mah={220} type="scroll">
              {info.presets.map((p) => (
                <Combobox.Option value={p.value} key={p.value} active={p.value === search}>
                  {p.label}
                </Combobox.Option>
              ))}
              {search && !isPreset && (
                <Combobox.Option value={search}>Use &quot;{search}&quot;</Combobox.Option>
              )}
            </ScrollArea.Autosize>
          </Combobox.Options>
        </Combobox.Dropdown>
      </Combobox>
      {(enabled || focused) && (
        <Text size="xs" c="dimmed" mt={4}>
          {info.hint}
        </Text>
      )}
    </div>
  )
}
