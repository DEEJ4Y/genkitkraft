namespace Api.Models;

/** A prompt (system instruction) for an agent. */
model PromptResponse {
  /** Unique prompt ID. */
  id: string;

  /** Display name for this prompt. */
  name: string;

  /** Prompt content in markdown format. */
  content: string;

  /** When this prompt was created. */
  createdAt: utcDateTime;

  /** When this prompt was last updated. */
  updatedAt: utcDateTime;
}

/** Request to create a new prompt. */
model CreatePromptRequest {
  /** Display name for this prompt. */
  name: string;

  /** Prompt content in markdown format. */
  content: string;
}

/** Request to update an existing prompt. */
model UpdatePromptRequest {
  /** Updated display name. */
  name?: string;

  /** Updated prompt content in markdown format. */
  content?: string;
}

/** Paginated list of prompts. */
model PromptListResponse {
  /** Array of prompts. */
  prompts: PromptResponse[];

  /** Total number of prompts. */
  total: int32;

  /** Number of prompts per page. */
  limit: int32;

  /** Number of prompts skipped. */
  offset: int32;
}
