import { describe, expect, it } from 'vitest'
import type { BuilderConfig } from './types'
import { buildWidgetConfig } from './widgetConfig'

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
