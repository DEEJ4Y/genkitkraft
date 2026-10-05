import { useEffect, useRef } from 'react'
import { createMockProvider } from './mockProvider'
import { normalizeConfig } from './normalize'
import { buildWidgetConfig } from './widgetConfig'
import type { BuilderConfig } from './types'

const IFRAME_ID = 'chat-widget-iframe'
const NORMAL_SIZE = { width: '400px', height: '704px' }
const EXPANDED_SIZE = { width: '688px', height: 'calc(100vh - 48px)' }
let defaultButtonHtml = ''

/**
 * Live preview: injects the real widget into the current page with an in-memory provider and
 * pushes config changes to it without reloading. Renders nothing.
 *
 * The widget registers a window message listener the first time it creates its iframe, so the
 * iframe is intentionally kept (hidden) across unmounts; only the launcher button is removed.
 */
export function Preview({ config }: { config: BuilderConfig }) {
  const configRef = useRef(config)
  configRef.current = config
  const buttonRef = useRef<HTMLElement | null>(null)
  const readyRef = useRef(false)
  // The iframe app ignores events until it has loaded; before that `initialConfig` is what it gets via `set_config`.
  const loadedRef = useRef(false)

  // Mount: inject once, remove launcher + hide on unmount.
  useEffect(() => {
    let cancelled = false
    const provider = createMockProvider()

    import('navigableai-chat-widget').then(({ injectAiChatWidget, WIDGET_BUTTON }) => {
      defaultButtonHtml = WIDGET_BUTTON
      if (cancelled) return
      const cfg = buildWidgetConfig(configRef.current, provider)
      const before = document.body.lastElementChild
      injectAiChatWidget(cfg)
      const added = document.body.lastElementChild
      buttonRef.current = added && added !== before ? (added as HTMLElement) : null
      if (window.$aiChatWidget) {
        window.$aiChatWidget.initialConfig = cfg
        window.$aiChatWidget.chatProvider = provider
        window.$aiChatWidget.open()
      }
      readyRef.current = true
      const iframe = document.getElementById(IFRAME_ID) as HTMLIFrameElement | null
      iframe?.addEventListener(
        'load',
        () => {
          if (cancelled) return
          loadedRef.current = true
          push(configRef.current, buttonRef.current, true)
        },
        { once: true },
      )
      // apply size and launcher for the initial config; the iframe gets its config via `set_config`
      push(configRef.current, buttonRef.current, false)
    })

    return () => {
      cancelled = true
      readyRef.current = false
      loadedRef.current = false
      try {
        window.$aiChatWidget?.close()
      } catch {
        /* widget not injected */
      }
      buttonRef.current?.remove()
      buttonRef.current = null
    }
  }, [])

  // Updates: keep the page widget in sync with the builder.
  useEffect(() => {
    if (readyRef.current) push(config, buttonRef.current, loadedRef.current)
  }, [config])

  return null
}

function push(config: BuilderConfig, button: HTMLElement | null, sendToIframe: boolean) {
  const w = window.$aiChatWidget
  if (!w) return
  const normalized = normalizeConfig(config)
  const provider = w.chatProvider
  // Same shape as the initial config (actionsMap defaulted, welcome actions sanitized). The iframe
  // checks that `chatProvider` is present; its functions are dropped on serialization but the key stays.
  const widgetConfig = buildWidgetConfig(config, provider)
  // clone everything but the provider: it holds functions, which structuredClone rejects
  w.initialConfig = { ...structuredClone({ ...widgetConfig, chatProvider: undefined }), chatProvider: provider }
  if (sendToIframe) w.sendEvent('override_config', widgetConfig)

  // Size and launcher are only applied at injection time by the widget; mirror them here so
  // toggling "Start expanded" or editing the launcher HTML takes effect immediately.
  const iframe = document.getElementById(IFRAME_ID) as HTMLIFrameElement | null
  if (iframe) {
    const size = normalized.chatWindow?.expanded ? EXPANDED_SIZE : NORMAL_SIZE
    iframe.style.width = size.width
    iframe.style.height = size.height
  }
  if (button) button.innerHTML = normalized.widgetButton || defaultButtonHtml
}
