import { useCallback, useEffect, useRef, useState } from 'react'
import type { BuilderConfig } from './types'

export const defaultConfig: BuilderConfig = {
  chatWindow: {
    defaults: { primaryColor: 'blue', colorScheme: 'light' },
    header: { title: { title: 'Assistant', showOnlineSubtitle: true } },
  },
}

type Path = string | Array<string | number>

const toPath = (path: Path): Array<string | number> =>
  Array.isArray(path) ? path : path.split('.').map((p) => (/^\d+$/.test(p) ? Number(p) : p))

export function getIn<T = unknown>(obj: unknown, path: Path): T | undefined {
  let cur: any = obj
  for (const key of toPath(path)) {
    if (cur == null) return undefined
    cur = cur[key]
  }
  return cur as T | undefined
}

const isEmpty = (v: unknown): boolean =>
  v === undefined ||
  v === null ||
  v === '' ||
  (typeof v === 'object' && v !== null && Object.keys(v).length === 0)

/**
 * Immutably set a value at a path. Empty values (undefined, '', {}) remove the key and prune
 * parents that become empty, so toggling a field off leaves no residue in the generated config.
 */
export function setIn<T>(obj: T, path: Path, value: unknown): T {
  const keys = toPath(path)
  const walk = (cur: any, i: number): any => {
    const key = keys[i]
    const isArr = Array.isArray(cur) || (cur == null && typeof key === 'number')
    const base: any = isArr ? [...(cur ?? [])] : { ...(cur ?? {}) }
    const next = i === keys.length - 1 ? value : walk(base[key], i + 1)
    if (isEmpty(next) && !isArr) {
      delete base[key]
    } else if (isEmpty(next) && isArr) {
      base.splice(key as number, 1)
    } else {
      base[key] = next
    }
    return base
  }
  const result = walk(obj, 0)
  return (isEmpty(result) ? {} : result) as T
}

export const unsetIn = <T>(obj: T, path: Path): T => setIn(obj, path, undefined)

export function parseConfigJson(text: string): BuilderConfig {
  const parsed = JSON.parse(text)
  if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('Config must be a JSON object')
  }
  return parsed as BuilderConfig
}

export const serializeConfig = (config: BuilderConfig) => JSON.stringify(config, null, 2)

const STORAGE_KEY = 'genkitkraft-widget-builder-config'

function readStored(): BuilderConfig | null {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    return raw ? parseConfigJson(raw) : null
  } catch {
    return null
  }
}

function readHash(): BuilderConfig | null {
  try {
    const match = window.location.hash.match(/config=([^&]+)/)
    return match ? parseConfigJson(decodeURIComponent(atob(match[1]))) : null
  } catch {
    return null
  }
}

export function shareUrl(config: BuilderConfig): string {
  const encoded = btoa(encodeURIComponent(serializeConfig(config)))
  return `${window.location.origin}${window.location.pathname}#config=${encoded}`
}

export interface UseWidgetConfig {
  config: BuilderConfig
  set: (path: Path, value: unknown) => void
  replace: (config: BuilderConfig) => void
  reset: () => void
}

/** Builder state. `persist` stores the config in localStorage (and restores a `#config=` share link). */
export function useWidgetConfig(initial: BuilderConfig = defaultConfig, persist = true): UseWidgetConfig {
  const [config, setConfig] = useState<BuilderConfig>(initial)
  const hydrated = useRef(false)

  useEffect(() => {
    if (persist) {
      const restored = readHash() ?? readStored()
      if (restored) setConfig(restored)
    }
    hydrated.current = true
  }, [persist])

  useEffect(() => {
    if (!persist || !hydrated.current) return
    try {
      window.localStorage.setItem(STORAGE_KEY, serializeConfig(config))
    } catch {
      /* storage unavailable */
    }
  }, [config, persist])

  const set = useCallback((path: Path, value: unknown) => setConfig((prev) => setIn(prev, path, value)), [])
  const reset = useCallback(() => setConfig(initial), [initial])
  return { config, set, replace: setConfig, reset }
}
