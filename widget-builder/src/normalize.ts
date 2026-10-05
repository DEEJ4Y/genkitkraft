import type { BuilderConfig } from './types'

/**
 * The widget's FooterConfig requires both tabs. The builder prunes empty values, so fill the
 * missing tab with `{}` right before the config is handed to the widget / printed as a snippet.
 */
export function normalizeConfig(config: BuilderConfig): BuilderConfig {
  const out: BuilderConfig = structuredClone(config)
  if (out.footerConfig) {
    out.footerConfig = { ...out.footerConfig, home: out.footerConfig.home ?? {}, messages: out.footerConfig.messages ?? {} }
  }
  return out
}
