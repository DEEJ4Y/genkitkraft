import { describe, expect, it } from 'vitest'
import type { BuilderConfig } from './types'
import { buildWidgetConfig, getColorScheme } from './widgetConfig'

const provider = {} as never

describe('buildWidgetConfig', () => {
  it('always provides an actionsMap object', () => {
    expect(buildWidgetConfig({}, provider).actionsMap).toEqual({})
  })

  it('attaches the provider', () => {
    expect(buildWidgetConfig({}, provider).chatProvider).toBe(provider)
  })

  it('drops welcome actions without a matching action', () => {
    const cfg = {
      actionsMap: { a: 'run()' },
      chatWindow: { welcomeMessage: { actions: ['a', 'missing'] } },
    } as unknown as BuilderConfig
    expect(buildWidgetConfig(cfg, provider).chatWindow?.welcomeMessage?.actions).toEqual(['a'])
    const noMap = { chatWindow: { welcomeMessage: { actions: ['a'] } } } as unknown as BuilderConfig
    expect(buildWidgetConfig(noMap, provider).chatWindow?.welcomeMessage?.actions).toEqual([])
  })

  it('does not mutate its input', () => {
    const cfg = {
      actionsMap: { a: 'run()' },
      chatWindow: { welcomeMessage: { actions: ['a', 'missing'] } },
    } as unknown as BuilderConfig
    const before = structuredClone(cfg)
    buildWidgetConfig(cfg, provider)
    expect(cfg).toEqual(before)
  })

  it('survives the serialization used by sendEvent', () => {
    const out = buildWidgetConfig({ debug: true }, provider)
    expect(JSON.parse(JSON.stringify(out)).actionsMap).toEqual({})
  })
})

describe('invalid primary color', () => {
  it.each(['#', '#ff', 'abc', '#ff000', 'rgb(1,2'])('drops %s', (color) => {
    const cfg = { chatWindow: { defaults: { primaryColor: color, colorScheme: 'light' } } } as unknown as BuilderConfig
    expect(buildWidgetConfig(cfg, provider).chatWindow?.defaults?.primaryColor).toBeUndefined()
    expect(cfg.chatWindow?.defaults?.primaryColor).toBe(color)
  })

  it.each(['blue', '#fff', '#ff0000', 'rgb(1, 2, 3)', 'hsl(10, 50%, 50%)'])('keeps %s', (color) => {
    const cfg = { chatWindow: { defaults: { primaryColor: color } } } as unknown as BuilderConfig
    expect(buildWidgetConfig(cfg, provider).chatWindow?.defaults?.primaryColor).toBe(color)
  })
})

describe('primary colors the hosted widget cannot parse', () => {
  const cfgWith = (color: string) => ({ chatWindow: { defaults: { primaryColor: color } } }) as unknown as BuilderConfig

  it.each([
    ['grape', '#be4bdb'],
    ['dark', '#2e2e2e'],
  ])('sends %s as %s', (name, hex) => {
    const cfg = cfgWith(name)
    expect(buildWidgetConfig(cfg, provider).chatWindow?.defaults?.primaryColor).toBe(hex)
    expect(cfg.chatWindow?.defaults?.primaryColor).toBe(name)
  })

  it.each(['red', 'pink', 'violet', 'indigo', 'blue', 'cyan', 'green', 'lime', 'yellow', 'orange', 'teal', 'gray'])(
    'leaves %s alone',
    (name) => {
      expect(buildWidgetConfig(cfgWith(name), provider).chatWindow?.defaults?.primaryColor).toBe(name)
    },
  )
})

describe('getColorScheme', () => {
  const cfg = (colorScheme?: unknown) => ({ chatWindow: { defaults: { colorScheme } } }) as unknown as BuilderConfig

  it('returns the configured scheme', () => {
    expect(getColorScheme(cfg('dark'))).toBe('dark')
    expect(getColorScheme(cfg('light'))).toBe('light')
  })

  it('falls back to light when unset or invalid', () => {
    expect(getColorScheme({})).toBe('light')
    expect(getColorScheme(cfg(undefined))).toBe('light')
    expect(getColorScheme(cfg('auto'))).toBe('light')
  })
})
