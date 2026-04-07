namespace Api.Models;

/** A playground chat session for testing an agent. */
model PlaygroundSessionResponse {
  /** Unique session ID. */
  id: string;

  /** ID of the agent this session belongs to. */
  agentId: string;

  /** Session title (auto-generated from first message). */
  title: string;

  /** When this session was created. */
  createdAt: utcDateTime;

  /** When this session was last updated. */
  updatedAt: utcDateTime;
}

/** List of playground sessions. */
model PlaygroundSessionListResponse {
  /** Array of sessions. */
  sessions: PlaygroundSessionResponse[];
}

/** Request to create a new playground session. */
model CreatePlaygroundSessionRequest {
  /** Optional title for the session. */
  title?: string;
}

/** A single message in a playground session. */
model PlaygroundMessageResponse {
  /** Unique message ID. */
  id: string;

  /** Session this message belongs to. */
  sessionId: string;

  /** Message role. */
  role: "user" | "assistant";

  /** Message text content. */
  content: string;

  /** When this message was created. */
  createdAt: utcDateTime;
}

/** List of messages in a session. */
model PlaygroundMessageListResponse {
  /** Array of messages. */
  messages: PlaygroundMessageResponse[];
}

/** Request to send a chat message in the playground. */
model PlaygroundChatRequest {
  /** The session to send the message in. */
  sessionId: string;

  /** The user message content. */
  content: string;

  /** Optional provider override for testing. */
  providerId?: string;

  /** Optional model override for testing. */
  modelId?: string;

  /** Optional system prompt override for testing. */
  systemPromptId?: string;

  /** Optional temperature override for testing. */
  temperature?: float32;

  /** Optional top-p override for testing. */
  topP?: float32;

  /** Optional top-k override for testing. */
  topK?: int32;
}
