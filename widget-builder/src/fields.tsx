import { Alert, ColorInput, Select, Switch, Textarea, TextInput } from '@mantine/core'
import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { contrastRatio, NAMED_COLORS } from './colors'
import { getIn, type UseWidgetConfig } from './state'

const BuilderContext = createContext<UseWidgetConfig | null>(null)

export function BuilderProvider({ value, children }: { value: UseWidgetConfig; children: ReactNode }) {
  return <BuilderContext.Provider value={value}>{children}</BuilderContext.Provider>
}

export function useBuilder(): UseWidgetConfig {
  const ctx = useContext(BuilderContext)
  if (!ctx) throw new Error('useBuilder must be used inside <BuilderProvider>')
  return ctx
}

interface BaseProps {
  label: string
  path: string
  description?: string
}

export function TextField({ label, path, description, placeholder }: BaseProps & { placeholder?: string }) {
  const { config, set } = useBuilder()
  return (
    <TextInput
      label={label}
      description={description}
      placeholder={placeholder}
      value={getIn<string>(config, path) ?? ''}
      onChange={(e) => set(path, e.currentTarget.value)}
    />
  )
}

export function TextAreaField({ label, path, description, placeholder }: BaseProps & { placeholder?: string }) {
  const { config, set } = useBuilder()
  return (
    <Textarea
      label={label}
      description={description}
      placeholder={placeholder}
      autosize
      minRows={2}
      value={getIn<string>(config, path) ?? ''}
      onChange={(e) => set(path, e.currentTarget.value)}
    />
  )
}

export function SwitchField({ label, path, description, defaultChecked = false }: BaseProps & { defaultChecked?: boolean }) {
  const { config, set } = useBuilder()
  return (
    <Switch
      label={label}
      description={description}
      checked={getIn<boolean>(config, path) ?? defaultChecked}
      onChange={(e) => set(path, e.currentTarget.checked)}
    />
  )
}

export function SelectField({
  label,
  path,
  description,
  data,
  placeholder,
}: BaseProps & { data: Array<string | { value: string; label: string }>; placeholder?: string }) {
  const { config, set } = useBuilder()
  return (
    <Select
      label={label}
      description={description}
      placeholder={placeholder}
      data={data}
      clearable
      value={getIn<string>(config, path) ?? null}
      onChange={(v) => set(path, v ?? undefined)}
    />
  )
}

export function ColorField({ label, path, description }: BaseProps) {
  const { config, set } = useBuilder()
  return (
    <ColorInput
      label={label}
      description={description}
      value={getIn<string>(config, path) ?? ''}
      onChange={(v) => set(path, v)}
      placeholder="Not set"
      aria-label={label}
    />
  )
}

/** Resolve a stored color (named or CSS) to something the contrast checker understands. */
const resolveColor = (c: string | undefined) => (c ? NAMED_COLORS[c] ?? c : undefined)

/** Warn when a chosen bg/text pair is below WCAG AA (4.5:1). */
export function ContrastWarning({ bgPath, fgPath }: { bgPath: string; fgPath: string }) {
  const { config } = useBuilder()
  const bg = resolveColor(getIn<string>(config, bgPath))
  const fg = resolveColor(getIn<string>(config, fgPath))
  if (!bg || !fg) return null
  const ratio = contrastRatio(bg, fg)
  if (ratio === null || ratio >= 4.5) return null
  return (
    <Alert color="yellow" variant="light" p="xs" role="status">
      Low contrast ({ratio.toFixed(1)}:1). Text should have at least 4.5:1 against its background.
    </Alert>
  )
}

/** Text input for JSON values; keeps the draft locally so partial edits don't get rejected. */
export function JsonField({ label, path, description }: BaseProps) {
  const { config, set } = useBuilder()
  const stored = getIn(config, path)
  const [text, setText] = useState(stored === undefined ? '' : JSON.stringify(stored, null, 2))
  const [error, setError] = useState<string | null>(null)

  // Re-sync when the config changes from outside (preset, import, reset).
  useEffect(() => {
    setText((prev) => {
      try {
        if (JSON.stringify(prev.trim() ? JSON.parse(prev) : undefined) === JSON.stringify(stored)) return prev
      } catch {
        /* keep invalid draft */
      }
      return stored === undefined ? '' : JSON.stringify(stored, null, 2)
    })
  }, [stored])

  return (
    <Textarea
      label={label}
      description={description}
      autosize
      minRows={3}
      styles={{ input: { fontFamily: 'monospace' } }}
      value={text}
      error={error}
      onChange={(e) => {
        const next = e.currentTarget.value
        setText(next)
        if (!next.trim()) {
          setError(null)
          set(path, undefined)
          return
        }
        try {
          const parsed = JSON.parse(next)
          if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('Must be a JSON object')
          setError(null)
          set(path, parsed)
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Invalid JSON')
        }
      }}
    />
  )
}
