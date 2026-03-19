namespace Api.Models;

/** Supported LLM provider types. */
enum ProviderType {
  /** Anthropic (Claude models). */
  anthropic: "anthropic",

  /** OpenAI (GPT models). */
  openai: "openai",

  /** Google AI / Gemini. */
  googleai: "googleai",
}

/** A configured LLM provider with masked API key. */
model ProviderResponse {
  /** Unique provider ID. */
  id: string;

  /** Display name for this provider configuration. */
  name: string;

  /** The provider type. */
  providerType: ProviderType;

  /** Masked API key (last 4 characters visible). */
  apiKey: string;

  /** Optional base URL override (e.g. for OpenAI-compatible proxies). */
  baseUrl?: string;

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

  /** Full API key (stored securely, returned masked). */
  apiKey: string;

  /** Optional base URL override. */
  baseUrl?: string;
}

/** Request to update an existing provider configuration. */
model UpdateProviderRequest {
  /** Updated display name. */
  name?: string;

  /** New API key (leave blank to keep existing). */
  apiKey?: string;

  /** Updated base URL. */
  baseUrl?: string;

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
