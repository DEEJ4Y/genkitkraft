namespace Api.Models;

/** An AI agent configuration. */
model AgentResponse {
  /** Unique agent ID. */
  id: string;

  /** Display name for this agent. */
  name: string;

  /** ID of the provider this agent uses. */
  providerId: string;

  /** Resolved display name of the provider. */
  providerName: string;

  /** Resolved type of the provider. */
  providerType: ProviderType;

  /** Model identifier (e.g. gemini-2.5-flash, gpt-4o). */
  modelId: string;

  /** ID of the system prompt, if any. */
  systemPromptId?: string;

  /** Resolved name of the system prompt, if any. */
  systemPromptName?: string;

  /** Sampling temperature. */
  temperature: float32;

  /** Nucleus sampling (top-p). */
  topP: float32;

  /** Top-k sampling. */
  topK: int32;

  /** When this agent was created. */
  createdAt: utcDateTime;

  /** When this agent was last updated. */
  updatedAt: utcDateTime;
}

/** Request to create a new agent. */
model CreateAgentRequest {
  /** Display name for this agent. */
  name: string;

  /** ID of the provider to use. */
  providerId: string;

  /** Model identifier. */
  modelId: string;

  /** ID of an existing system prompt (optional). */
  systemPromptId?: string;

  /** Sampling temperature (default 0.95). */
  temperature?: float32;

  /** Nucleus sampling top-p (default 0.95). */
  topP?: float32;

  /** Top-k sampling (default 40). */
  topK?: int32;
}

/** Request to update an existing agent. */
model UpdateAgentRequest {
  /** Updated display name. */
  name?: string;

  /** Updated provider ID. */
  providerId?: string;

  /** Updated model identifier. */
  modelId?: string;

  /** Updated system prompt ID (empty string to clear). */
  systemPromptId?: string;

  /** Updated sampling temperature. */
  temperature?: float32;

  /** Updated nucleus sampling top-p. */
  topP?: float32;

  /** Updated top-k sampling. */
  topK?: int32;
}

/** Paginated list of agents. */
model AgentListResponse {
  /** Array of agents. */
  agents: AgentResponse[];

  /** Total number of agents. */
  total: int32;

  /** Number of agents per page. */
  limit: int32;

  /** Number of agents skipped. */
  offset: int32;
}
