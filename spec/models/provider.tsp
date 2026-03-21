namespace Api.Models;

/** Supported LLM provider types. */
enum ProviderType {
  /** Google AI (Gemini). */
  google_ai: "google_ai",

  /** Vertex AI. */
  vertex_ai: "vertex_ai",

  /** OpenAI (GPT models). */
  openai: "openai",

  /** Anthropic (Claude models). */
  anthropic: "anthropic",

  /** Ollama (local models). */
  ollama: "ollama",

  /** xAI (Grok). */
  xai: "xai",

  /** DeepSeek. */
  deepseek: "deepseek",

  /** Azure OpenAI. */
  azure_openai: "azure_openai",

  /** AWS Bedrock. */
  bedrock: "bedrock",

  /** Azure AI Foundry. */
  azure_ai_foundry: "azure_ai_foundry",

  /** OpenAI-compatible API (custom endpoint). */
  openai_compatible: "openai_compatible",
}

/** A configured LLM provider with masked API key. */
model ProviderResponse {
  /** Unique provider ID. */
  id: string;

  /** Display name for this provider configuration. */
  name: string;

  /** The provider type. */
  providerType: ProviderType;

  /** Masked API key (last 4 characters visible). Null if provider does not use an API key. */
  apiKey?: string;

  /** Optional base URL override (e.g. for OpenAI-compatible proxies). */
  baseUrl?: string;

  /** Provider-specific configuration as a JSON object. */
  config?: Record<string>;

  /** Whether this provider is enabled. */
  enabled: boolean;

  /** When this provider was created. */
  createdAt: utcDateTime;

  /** When this provider was last updated. */
  updatedAt: utcDateTime;
}

/** Request to create a new provider configuration. */
model CreateProviderRequest {
  /** Display name for this provider. */
  name: string;

  /** The provider type. */
  providerType: ProviderType;

  /** Full API key (stored securely, returned masked). Optional for providers that don't use API keys. */
  apiKey?: string;

  /** Optional base URL override. */
  baseUrl?: string;

  /** Provider-specific configuration. */
  config?: Record<string>;
}

/** Request to update an existing provider configuration. */
model UpdateProviderRequest {
  /** Updated display name. */
  name?: string;

  /** New API key (leave blank to keep existing). */
  apiKey?: string;

  /** Updated base URL. */
  baseUrl?: string;

  /** Provider-specific configuration (replaces existing). */
  config?: Record<string>;

  /** Enable or disable this provider. */
  enabled?: boolean;
}

/** Result of testing provider connectivity. */
model TestProviderResponse {
  /** Whether the connection test succeeded. */
  success: boolean;

  /** Human-readable result message. */
  message: string;
}

/** List of configured providers. */
model ProviderListResponse {
  /** Array of provider configurations. */
  providers: ProviderResponse[];
}

/** Google AI provider config (no extra fields needed). */
model GoogleAIProviderConfig {}

/** Vertex AI provider config. */
model VertexAIProviderConfig {
  /** GCP project ID. */
  project: string;

  /** GCP region (e.g. "us-central1"). */
  location: string;
}

/** OpenAI provider config (no extra fields needed). */
model OpenAIProviderConfig {}

/** Anthropic provider config (no extra fields needed). */
model AnthropicProviderConfig {}

/** Ollama provider config (no extra fields needed, uses base_url). */
model OllamaProviderConfig {}

/** xAI provider config (no extra fields needed). */
model XAIProviderConfig {}

/** DeepSeek provider config (no extra fields needed). */
model DeepSeekProviderConfig {}

/** Azure OpenAI provider config. */
model AzureOpenAIProviderConfig {
  /** Azure deployment name. */
  deploymentName: string;

  /** Azure API version (e.g. "2024-10-21"). */
  apiVersion: string;
}

/** AWS Bedrock provider config. */
model BedrockProviderConfig {
  /** AWS region (e.g. "us-east-1"). */
  region: string;

  /** AWS access key ID. */
  accessKeyId: string;

  /** AWS secret access key. */
  secretAccessKey: string;

  /** AWS session token (optional). */
  sessionToken?: string;
}

/** Azure AI Foundry provider config (no extra fields needed, uses base_url for endpoint and api_key). */
model AzureAIFoundryProviderConfig {}

/** OpenAI-compatible provider config (no extra fields needed, uses base_url and api_key). */
model OpenAICompatibleProviderConfig {}

/** Describes a config field for a provider type. */
model ConfigFieldInfo {
  /** Field name (key in config object). */
  name: string;

  /** Human-readable label. */
  label: string;

  /** Whether this field is required. */
  required: boolean;

  /** Placeholder text for UI. */
  placeholder?: string;

  /** Whether the field contains sensitive data. */
  sensitive?: boolean;
}

/** Metadata about a provider type for dynamic UI rendering. */
model ProviderTypeInfo {
  /** The provider type identifier. */
  type: ProviderType;

  /** Human-readable display name. */
  displayName: string;

  /** Whether this provider type requires an API key. */
  requiresApiKey: boolean;

  /** Whether this provider type requires a base URL. */
  requiresBaseUrl: boolean;

  /** Config fields specific to this provider type. */
  configFields: ConfigFieldInfo[];

  /** Hint about which environment variable to set. */
  envVarHint: string;

  /** Model name prefix used by Genkit. */
  modelPrefix: string;

  /** Default base URL if applicable. */
  baseUrlDefault?: string;

  /** True if this provider is not yet fully supported. */
  comingSoon?: boolean;
}

/** List of provider type metadata. */
model ProviderTypeListResponse {
  /** Array of provider type information. */
  providerTypes: ProviderTypeInfo[];
}
