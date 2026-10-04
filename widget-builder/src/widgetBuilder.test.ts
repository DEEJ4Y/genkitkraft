import { describe, expect, it } from 'vitest'
import { contrastRatio } from './colors'
import { applyPreset, presets } from './presets'
import { providerSource } from './providerSource'
import { backendExample, generateCdnSnippet, generateNpmSnippet } from './snippets'
import { getIn, parseConfigJson, serializeConfig, setIn, unsetIn } from './state'
import type { BuilderConfig } from './types'

describe('setIn / unsetIn', () => {
  it('sets nested values immutably', () => {
    const a: BuilderConfig = { chatWindow: { expanded: true } }
    const b = setIn(a, 'chatWindow.header.title.title', 'Hi')
    expect(getIn(b, 'chatWindow.header.title.title')).toBe('Hi')
    expect(a).toEqual({ chatWindow: { expanded: true } })
    expect(b.chatWindow?.expanded).toBe(true)
  })

  it('removes empty values and prunes empty parents', () => {
    const a: BuilderConfig = { chatWindow: { header: { title: { title: 'Hi' } } } }
    expect(setIn(a, 'chatWindow.header.title.title', '')).toEqual({})
    expect(unsetIn(a, 'chatWindow.header')).toEqual({})
  })

  it('keeps false booleans', () => {
    expect(setIn({}, 'disableCloseButton', false)).toEqual({ disableCloseButton: false })
  })

  it('supports array indexes', () => {
    const a = setIn<BuilderConfig>({}, 'chatWindow.header.avatars.0.url', 'x.png')
    expect(a.chatWindow?.header?.avatars).toEqual([{ url: 'x.png' }])
    expect(setIn(a, 'chatWindow.header.avatars.0.url', '')).toEqual({})
  })
})

describe('import / export', () => {
  it('round-trips', () => {
    const cfg: BuilderConfig = { chatWindow: { defaults: { primaryColor: '#123456' } }, debug: true }
    expect(parseConfigJson(serializeConfig(cfg))).toEqual(cfg)
  })

  it('rejects non-objects', () => {
    expect(() => parseConfigJson('[]')).toThrow()
    expect(() => parseConfigJson('null')).toThrow()
    expect(() => parseConfigJson('nope')).toThrow()
  })
})

describe('snippets', () => {
  const cfg: BuilderConfig = { chatWindow: { defaults: { primaryColor: 'teal' } } }

  it('does not mutate the config', () => {
    const copy = structuredClone(cfg)
    generateNpmSnippet(cfg)
    generateCdnSnippet(cfg)
    expect(cfg).toEqual(copy)
  })

  it('npm snippet uses the real export and the provider', () => {
    const s = generateNpmSnippet(cfg, { backendBaseUrl: '/api/support' })
    expect(s).toContain('import { injectAiChatWidget } from "navigableai-chat-widget"')
    expect(s).toContain('new GenkitkraftChatProvider({ baseUrl: "/api/support" })')
    expect(s).toContain('"primaryColor": "teal"')
    expect(s).not.toContain('injectChatWidget')
  })

  it('cdn snippet uses window.initAiChatWidget and has valid JS (no type annotations)', () => {
    const s = generateCdnSnippet(cfg)
    expect(s).toContain('window.initAiChatWidget(')
    expect(s).toContain('class GenkitkraftChatProvider')
    expect(s).not.toMatch(/: Promise</)
    expect(s).not.toContain('implements')
  })

  it('handles an empty config', () => {
    expect(generateNpmSnippet({})).toContain('injectAiChatWidget({\n  "chatProvider"')
  })
})

describe('backendExample', () => {
  it('reads history from GenKitKraft and stores only ownership', () => {
    expect(backendExample).toContain('/messages`)') // GET .../deploy/sessions/{id}/messages
    expect(backendExample).toContain('deploy(`/sessions/${id}/messages`)')
    expect(backendExample).toContain('db.sessions.insert({ id: session.id, userId: req.user.id })')
    expect(backendExample).not.toContain('db.messages')
  })

  it('is syntactically valid JavaScript module code', () => {
    const body = backendExample.replace(/^import .*$/m, '').replace('export default router;', 'return router;')
    expect(() => new Function('express', 'process', 'fetch', 'requireLogin', 'db', body)).not.toThrow()
  })
})

describe('providerSource', () => {
  it('implements the four ChatProvider methods in both languages', () => {
    for (const lang of ['ts', 'js'] as const) {
      const s = providerSource(lang)
      for (const m of ['listSessions', 'createSession', 'listSessionMessages', 'sendMessage']) {
        expect(s).toContain(`async ${m}(`)
      }
    }
  })

  it('js output parses as JavaScript', () => {
    const src = providerSource('js').replace(/^export /m, '')
    expect(() => new Function(src)).not.toThrow()
  })
})

describe('generated provider behaviour', () => {
  it('lists sessions newest first', async () => {
    const src = providerSource('js').replace(/^export /m, '')
    const rows = [
      { id: 'old', title: 'a', createdAt: '2026-01-01T00:00:00Z' },
      { id: 'new', title: 'b', createdAt: '2026-03-01T00:00:00Z' },
      { id: 'mid', title: 'c', createdAt: '2026-02-01T00:00:00Z' },
    ]
    const Provider = new Function('fetch', `${src}; return GenkitkraftChatProvider;`)(async () => ({ ok: true, json: async () => rows }))
    const out = await new Provider({ baseUrl: '/x' }).listSessions()
    expect(out.map((s: any) => s.id)).toEqual(['new', 'mid', 'old'])
  })

  it('js provider calls the backend endpoints with credentials and returns parsed data', async () => {
    const src = providerSource('js').replace(/^export /m, '')
    const calls: Array<{ url: string; init: any }> = []
    const fakeFetch = async (url: string, init: any) => {
      calls.push({ url, init })
      const body = url.endsWith('/sessions') && init.method === 'POST' ? { id: 's1' } : url.endsWith('/messages') && init.method === 'POST' ? { role: 'assistant', content: 'hi' } : []
      return { ok: true, status: 200, json: async () => body }
    }
    const Provider = new Function('fetch', `${src}; return GenkitkraftChatProvider;`)(fakeFetch)
    const p = new Provider({ baseUrl: '/api/chat', getHeaders: () => ({ 'X-Test': '1' }) })

    expect(await p.listSessions()).toEqual([])
    expect(await p.createSession()).toBe('s1')
    expect(await p.listSessionMessages({ sessionId: 's1' })).toEqual([])
    expect(await p.listSessionMessages({})).toEqual([])
    expect(await p.sendMessage({ content: 'yo' })).toEqual({ role: 'assistant', content: 'hi' }) // creates a session first

    expect(calls.map((c) => c.url)).toEqual([
      '/api/chat/sessions',
      '/api/chat/sessions',
      '/api/chat/sessions/s1/messages',
      '/api/chat/sessions',
      '/api/chat/sessions/s1/messages',
    ])
    expect(calls[0].init.credentials).toBe('include')
    expect(calls[0].init.headers['X-Test']).toBe('1')
    expect(JSON.parse(calls[4].init.body)).toEqual({ content: 'yo' })
  })

  it('js provider rejects on non-2xx', async () => {
    const src = providerSource('js').replace(/^export /m, '')
    const Provider = new Function('fetch', `${src}; return GenkitkraftChatProvider;`)(async () => ({ ok: false, status: 401 }))
    await expect(new Provider({ baseUrl: '/x' }).listSessions()).rejects.toThrow('401')
  })
})

describe('applyPreset', () => {
  it('keeps typed-in branding and does not mutate the preset', () => {
    const current: BuilderConfig = {
      chatWindow: { header: { title: { title: 'My Bot' }, avatars: [{ url: 'a.png' }] } },
      homeScreenConfig: { logoUrl: 'logo.png' },
    }
    const dark = presets.find((p) => p.id === 'dark')!
    const before = structuredClone(dark.config)
    const out = applyPreset(dark, current)
    expect(out.chatWindow?.defaults?.colorScheme).toBe('dark')
    expect(out.chatWindow?.header?.title?.title).toBe('My Bot')
    expect(out.chatWindow?.header?.avatars).toEqual([{ url: 'a.png' }])
    expect(out.homeScreenConfig?.logoUrl).toBe('logo.png')
    expect(dark.config).toEqual(before)
  })
})

describe('contrastRatio', () => {
  it('computes WCAG ratios', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 0)
    expect(contrastRatio('#fff', '#fff')).toBeCloseTo(1, 5)
    expect(contrastRatio('red', '#fff')).toBeNull()
  })
})
