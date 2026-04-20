namespace Api.Models;

/** A message in an OpenAI-compatible chat completions request. */
model DeployChatMessage {
  /** The role of the message author. */
  role: "user" | "assistant" | "system";

  /** The content of the message. */
  content: string;
}

/** OpenAI-compatible chat completions request body. */
model DeployChatCompletionsRequest {
  /** Model identifier. Accepted but ignored — the agent's configured model is used. */
  `model`?: string;

  /** The conversation messages. The agent's system prompt is injected server-side; callers should not send a system message. */
  messages: DeployChatMessage[];

  /** Whether to stream the response as SSE. Defaults to false. */
  stream?: boolean;
}

/** A choice in an OpenAI chat completion response. */
model DeployChatCompletionChoice {
  /** The index of this choice in the choices array. */
  index: int32;

  /** The assistant's response message. */
  message: DeployChatMessage;

  /** The reason the model stopped generating. */
  finish_reason: string;
}

/** Token usage statistics. */
model DeployChatCompletionUsage {
  /** Number of tokens in the prompt. */
  prompt_tokens: int32;

  /** Number of tokens in the completion. */
  completion_tokens: int32;

  /** Total number of tokens used. */
  total_tokens: int32;
}

/** OpenAI-compatible chat completion response. */
model DeployChatCompletionResponse {
  /** Unique completion identifier. */
  id: string;

  /** Object type, always "chat.completion". */
  object: string;

  /** Unix timestamp of when the completion was created. */
  created: int64;

  /** The model used for the completion. */
  `model`: string;

  /** The list of completion choices. */
  choices: DeployChatCompletionChoice[];

  /** Token usage statistics. May be zero if the provider does not report usage. */
  usage: DeployChatCompletionUsage;
}

/** Request to create a new deploy session. */
model CreateDeploySessionRequest {
  /** Optional title for the session. */
  title?: string;
}

/** A deploy session for stateful chat with an agent. */
model DeploySessionResponse {
  /** Unique session ID. */
  id: string;

  /** ID of the agent this session belongs to. */
  agent_id: string;

  /** Session title. */
  title: string;

  /** When this session was created (ISO 8601). */
  created_at: utcDateTime;
}
