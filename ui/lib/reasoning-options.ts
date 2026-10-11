// Reasoning effort (thinking) options per provider type.
//
// The value is free text that is sent to the provider as is, so providers can
// add new values without a change here: the UI lets users type a custom value.
// The presets below are the values that the provider SDKs document today.
// Keep this list in step with buildConfig in
// internal/adapters/genkit_chat_provider.

export interface ReasoningEffortPreset {
  value: string
  label: string
}

export interface ReasoningEffortInfo {
  presets: ReasoningEffortPreset[]
  // Value that is filled in when the option is switched on.
  defaultValue: string
  // Short help text that is shown under the field.
  hint: string
}

const OPENAI_STYLE: ReasoningEffortInfo = {
  presets: [
    { value: 'minimal', label: 'minimal' },
    { value: 'low', label: 'low' },
    { value: 'medium', label: 'medium' },
    { value: 'high', label: 'high' },
  ],
  defaultValue: 'medium',
  hint:
    'Sent as reasoning_effort. Only reasoning models accept it. Type another value, ' +
    'for example none or xhigh, if your model supports it.',
}

const GEMINI: ReasoningEffortInfo = {
  presets: [
    { value: 'minimal', label: 'minimal' },
    { value: 'low', label: 'low' },
    { value: 'medium', label: 'medium' },
    { value: 'high', label: 'high' },
    { value: '-1', label: 'dynamic (-1 token budget)' },
  ],
  defaultValue: 'medium',
  hint:
    'A level (minimal, low, medium, high) or a thinking token budget. Gemini 3 models use ' +
    'a level. Gemini 2.5 models use a token budget, for example 2048, 0 (off) or -1 (dynamic).',
}

const ANTHROPIC: ReasoningEffortInfo = {
  presets: [
    { value: '1024', label: '1024 tokens' },
    { value: '2048', label: '2048 tokens' },
    { value: '4096', label: '4096 tokens' },
    { value: '8192', label: '8192 tokens' },
    { value: '16384', label: '16384 tokens' },
  ],
  defaultValue: '2048',
  hint:
    'Thinking token budget (minimum 1024). Temperature, Top P and Top K are not sent ' +
    'to the model while this option is on.',
}

const OLLAMA: ReasoningEffortInfo = {
  presets: [
    { value: 'true', label: 'true (think)' },
    { value: 'false', label: 'false (do not think)' },
    { value: 'low', label: 'low' },
    { value: 'medium', label: 'medium' },
    { value: 'high', label: 'high' },
  ],
  defaultValue: 'true',
  hint:
    'Use low, medium or high for GPT-OSS models. Use true or false for other thinking models.',
}

export const REASONING_EFFORT_OPTIONS: Record<string, ReasoningEffortInfo> = {
  google_ai: GEMINI,
  vertex_ai: GEMINI,
  anthropic: ANTHROPIC,
  ollama: OLLAMA,
  openai: OPENAI_STYLE,
  azure_openai: OPENAI_STYLE,
  xai: OPENAI_STYLE,
  deepseek: OPENAI_STYLE,
  openai_compatible: OPENAI_STYLE,
  // azure_ai_foundry and bedrock are left out on purpose: their plugins ignore
  // the option, so the toggle is hidden for them.
}

export function supportsReasoningEffort(providerType: string): boolean {
  return providerType in REASONING_EFFORT_OPTIONS
}

export function defaultReasoningEffort(providerType: string): string {
  return REASONING_EFFORT_OPTIONS[providerType]?.defaultValue ?? ''
}

// Returns an error message, or null when the value can be sent.
export function validateReasoningEffort(
  providerType: string,
  enabled: boolean,
  value: string
): string | null {
  if (!enabled || !supportsReasoningEffort(providerType)) return null
  const v = value.trim()
  if (!v) return 'Reasoning effort value is required when reasoning effort is on'
  if (v.length > 64) return 'Reasoning effort value must be at most 64 characters'
  if (providerType === 'anthropic' && !(/^\d+$/.test(v) && Number(v) >= 1024)) {
    return 'For Anthropic, reasoning effort must be a whole number of tokens, at least 1024'
  }
  return null
}
