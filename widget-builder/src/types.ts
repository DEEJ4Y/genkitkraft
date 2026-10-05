import type { ChatWidgetConfig } from 'navigableai-chat-widget'

/** The config the builder edits: a JSON-serializable subset of ChatWidgetConfig (no functions, no provider). */
export type BuilderConfig = Omit<ChatWidgetConfig, 'chatProvider' | 'functionsMap'>

export type { ChatWidgetConfig }
