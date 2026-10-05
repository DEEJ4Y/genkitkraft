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
