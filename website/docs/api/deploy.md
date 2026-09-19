---
sidebar_position: 3
---

# Deploy API (Chat Completions)

GenKitKraft exposes your configured agents through an OpenAI-compatible chat completions endpoint. You can use the **stateless** endpoint (provide full message history each request) or the **stateful** session-based endpoint (server manages conversation history).

## Authentication

The deploy endpoints use API key authentication via the `Authorization` header, separate from the session-based auth used by the management UI.

### Setting Up an API Key

Set the `PUBLIC_API_KEY` environment variable before starting GenKitKraft:

```bash
export PUBLIC_API_KEY=my-secret-key
```

If `PUBLIC_API_KEY` is not set, all deploy endpoints are publicly accessible (no authentication required).

### Using the API Key

Include the key as a Bearer token in the `Authorization` header:

```bash
curl http://localhost:8080/api/v1/agents/{agentId}/deploy/chat/completions \
  -H "Authorization: Bearer my-secret-key" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

---

## Stateless Chat Completions

The caller provides the full message history on every request.

### Endpoint

```
POST /api/v1/agents/{agentId}/deploy/chat/completions
```

- **`agentId`** — The UUID of the agent to use. You can find this in the **Deploy** tab of the agent edit screen.

## Request Format

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Hello, what can you do?"
    }
  ],
  "stream": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `messages` | array | Yes | Array of message objects. At least one message is required. |
| `messages[].role` | string | Yes | Message role: `"user"` or `"assistant"`. System messages are **not supported** and will return a 400 error. |
| `messages[].content` | string | Yes | The message content. |
| `stream` | boolean | No | Whether to stream the response via SSE. Defaults to `false`. |

:::note
The agent's system prompt is configured in GenKitKraft and automatically prepended — you don't need to include a system message.
:::

## Non-Streaming Response

When `stream` is `false` (default), the response is a single JSON object:

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1700000000,
  "model": "my-agent",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! I'm an AI assistant. I can help you with..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 0,
    "completion_tokens": 0,
    "total_tokens": 0
  }
}
```

## Streaming Response (SSE)

When `stream` is `true`, the response is a stream of Server-Sent Events:

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"my-agent","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"my-agent","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"my-agent","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"my-agent","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

## Stateless Examples

### curl (non-streaming)

```bash
curl -X POST http://localhost:8080/api/v1/agents/{agentId}/deploy/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-key" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello, what can you do?"}
    ],
    "stream": false
  }'
```

### curl (streaming)

```bash
curl -X POST http://localhost:8080/api/v1/agents/{agentId}/deploy/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-key" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": true
  }'
```

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/api/v1/agents/{agentId}/deploy",
    api_key="my-secret-key",
)

# Non-streaming
response = client.chat.completions.create(
    model="any",  # model is determined by the agent config
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)

# Streaming
stream = client.chat.completions.create(
    model="any",
    messages=[{"role": "user", "content": "Hello!"}],
    stream=True,
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### Node.js (OpenAI SDK)

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "http://localhost:8080/api/v1/agents/{agentId}/deploy",
  apiKey: "my-secret-key",
});

// Non-streaming
const response = await client.chat.completions.create({
  model: "any", // model is determined by the agent config
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(response.choices[0].message.content);

// Streaming
const stream = await client.chat.completions.create({
  model: "any",
  messages: [{ role: "user", content: "Hello!" }],
  stream: true,
});
for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || "");
}
```

## Error Responses

Errors follow the OpenAI error format:

| Status | Reason |
|---|---|
| 400 | Invalid request (empty messages, system messages, malformed JSON) |
| 401 | Missing or invalid API key (when `PUBLIC_API_KEY` is set) |
| 404 | Agent not found |
| 500 | Internal server error |

Example error response:

```json
{
  "error": {
    "message": "messages must not be empty",
    "type": "invalid_request_error",
    "code": "invalid_request"
  }
}
```

---

## Stateful Chat (Sessions)

The stateful API manages conversation history server-side. You create a session once, then send only the new user message on each turn — the server loads and persists history automatically.

### Session Lifecycle

#### Create a Session

```
POST /api/v1/agents/{agentId}/deploy/sessions
```

Request body (title is optional):

```json
{
  "title": "My conversation"
}
```

Response (201):

```json
{
  "id": "session-uuid",
  "agent_id": "agent-uuid",
  "title": "My conversation",
  "created_at": "2026-04-20T12:00:00Z"
}
```

#### Get a Session

```
GET /api/v1/agents/{agentId}/deploy/sessions/{sessionId}
```

Response (200):

```json
{
  "id": "session-uuid",
  "agent_id": "agent-uuid",
  "title": "My conversation",
  "created_at": "2026-04-20T12:00:00Z"
}
```

#### Delete a Session

```
DELETE /api/v1/agents/{agentId}/deploy/sessions/{sessionId}
```

Response: `204 No Content`. Deletes the session and all its messages.

### Stateful Chat Completions

```
POST /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions
```

The request and response formats are identical to the [stateless endpoint](#request-format). The key difference:

- Only the **last user message** in the `messages` array is used. Full conversation history is loaded from the session automatically.
- The last message **must** have `role: "user"`.
- The user message and assistant response are both persisted to the session.

```json
{
  "messages": [
    { "role": "user", "content": "Tell me a joke" }
  ],
  "stream": false
}
```

:::tip
You only need to send a single message per request. The server handles the full history.
:::

### Stream Resilience (Reconnect and Cancel)

Streaming generation for a session runs independently of any single HTTP request — a dropped connection does not stop it, and the assistant reply is persisted to the session as it's generated. Two additional endpoints let a client recover from a dropped connection or stop generation early.

#### Reconnect

```
GET /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/stream
```

Each frame in a streaming response carries an SSE `id:` line with a monotonically increasing sequence number, in addition to the `data:` payload:

```
id: 1
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk", ...}

id: 2
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk", ...}
```

If a streaming request is interrupted (network error, timeout, client restart), reconnect by calling this endpoint with a `Last-Event-ID` header set to the last `id:` value you received. The response resumes from the next token — already-generated content is replayed from what was persisted, and generation is **not** re-run. If the reply had already finished (or failed) before you reconnect, you'll immediately get the remaining content followed by `finish_reason`/`[DONE]` (or an error chunk).

```bash
curl -N http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/stream \
  -H "Authorization: Bearer my-secret-key" \
  -H "Last-Event-ID: 12"
```

Omit `Last-Event-ID` (or send `0`) to replay the reply from the beginning.

#### Cancel

```
POST /api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/cancel
```

Stops the assistant reply currently generating for the session, if any. Returns `204 No Content` whether or not a stream was actually in progress — this is a best-effort call, not something to poll for confirmation.

```bash
curl -X POST http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions/cancel \
  -H "Authorization: Bearer my-secret-key"
```

### Stateful Examples

#### curl — Full session flow

```bash
# 1. Create a session
SESSION=$(curl -s -X POST \
  http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions \
  -H "Authorization: Bearer my-secret-key" \
  -H "Content-Type: application/json" \
  -d '{}' | jq -r '.id')

# 2. Chat (first turn)
curl -X POST \
  http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions/$SESSION/chat/completions \
  -H "Authorization: Bearer my-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "Hello! What can you do?"}],
    "stream": false
  }'

# 3. Chat (second turn — history is automatic)
curl -X POST \
  http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions/$SESSION/chat/completions \
  -H "Authorization: Bearer my-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "Tell me more about the first thing you mentioned."}],
    "stream": false
  }'

# 4. Delete when done
curl -X DELETE \
  http://localhost:8080/api/v1/agents/{agentId}/deploy/sessions/$SESSION \
  -H "Authorization: Bearer my-secret-key"
```

#### Python (OpenAI SDK + sessions)

```python
import requests
from openai import OpenAI

BASE = "http://localhost:8080/api/v1/agents/{agentId}/deploy"
HEADERS = {"Authorization": "Bearer my-secret-key"}

# Create session
session = requests.post(f"{BASE}/sessions", headers=HEADERS, json={}).json()
session_id = session["id"]

# Use OpenAI SDK for chat
client = OpenAI(
    base_url=f"{BASE}/sessions/{session_id}",
    api_key="my-secret-key",
)

# Each call only needs the new message — history is managed server-side
response = client.chat.completions.create(
    model="any",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)

# Follow-up (server remembers previous turns)
response = client.chat.completions.create(
    model="any",
    messages=[{"role": "user", "content": "Can you elaborate?"}],
)
print(response.choices[0].message.content)
```

#### Node.js (OpenAI SDK + sessions)

```javascript
import OpenAI from "openai";

const BASE = "http://localhost:8080/api/v1/agents/{agentId}/deploy";
const API_KEY = "my-secret-key";

// Create session
const session = await fetch(`${BASE}/sessions`, {
  method: "POST",
  headers: { Authorization: `Bearer ${API_KEY}`, "Content-Type": "application/json" },
  body: JSON.stringify({}),
}).then((r) => r.json());

// Use OpenAI SDK for chat
const client = new OpenAI({
  baseURL: `${BASE}/sessions/${session.id}`,
  apiKey: API_KEY,
});

const response = await client.chat.completions.create({
  model: "any",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(response.choices[0].message.content);

// Follow-up (server remembers previous turns)
const followUp = await client.chat.completions.create({
  model: "any",
  messages: [{ role: "user", content: "Can you elaborate?" }],
});
console.log(followUp.choices[0].message.content);
```
