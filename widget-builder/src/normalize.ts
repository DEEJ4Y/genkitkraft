import { isValidPrimaryColor, toWidgetPrimaryColor } from './colors'
import type { BuilderConfig } from './types'

/**
 * The widget's FooterConfig requires both tabs. The builder prunes empty values, so fill the
 * missing tab with `{}` right before the config is handed to the widget / printed as a snippet.
 * An incomplete primary color (the color input reports every keystroke) is dropped: the widget
 * crashes on it, and it would otherwise be persisted and crash the preview on every reload.
 */
export function normalizeConfig(config: BuilderConfig): BuilderConfig {
  const out: BuilderConfig = structuredClone(config)
  if (out.footerConfig) {
    out.footerConfig = { ...out.footerConfig, home: out.footerConfig.home ?? {}, messages: out.footerConfig.messages ?? {} }
  }
  const defaults = out.chatWindow?.defaults
  if (defaults && defaults.primaryColor !== undefined && !isValidPrimaryColor(defaults.primaryColor)) {
    delete defaults.primaryColor
  } else if (defaults?.primaryColor) {
    defaults.primaryColor = toWidgetPrimaryColor(defaults.primaryColor)
  }
  return out
}
