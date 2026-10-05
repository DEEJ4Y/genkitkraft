import { normalizeConfig } from './normalize'
import type { BuilderConfig, ChatWidgetConfig } from './types'

/**
 * Builds the config handed to the widget, for both the initial injection and live updates.
 *
 * `injectAiChatWidget` and the widget's `init` handler always give the iframe an `actionsMap`
 * object and drop welcome actions that have no matching entry. The builder prunes empty values,
 * so apply the same guarantees here; otherwise an `override_config` after an edit carries a
 * shape the iframe never sees on first load.
 */
export function buildWidgetConfig(config: BuilderConfig, chatProvider: ChatWidgetConfig['chatProvider']): ChatWidgetConfig {
  const out = normalizeConfig(config)

  const actionsMap: Record<string, unknown> = {}
  for (const [name, value] of Object.entries(out.actionsMap ?? {})) {
    if (typeof value === 'string' || typeof value === 'function') actionsMap[name] = value
  }
  out.actionsMap = actionsMap as BuilderConfig['actionsMap']

  const welcome = out.chatWindow?.welcomeMessage
  if (welcome?.actions) {
    welcome.actions = welcome.actions.filter((action) => Object.prototype.hasOwnProperty.call(actionsMap, action))
  }

  return { ...out, chatProvider } as ChatWidgetConfig
}

/** The color scheme the config asks for. The widget treats an unset or unknown value as light. */
export function getColorScheme(config: BuilderConfig): 'light' | 'dark' {
  return config.chatWindow?.defaults?.colorScheme === 'dark' ? 'dark' : 'light'
}
