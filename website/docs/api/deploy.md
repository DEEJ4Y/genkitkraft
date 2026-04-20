---
sidebar_position: 3
---

# Deploy API (Chat Completions)

GenKitKraft exposes your configured agents through an OpenAI-compatible chat completions endpoint. This allows you to integrate your agents into any application that supports the OpenAI API format.

## Endpoint

```
POST /api/v1/agents/{agentId}/deploy/chat/completions
```

- **`agentId`** — The UUID of the agent to use. You can find this in the **Deploy** tab of the agent edit screen.

## Authentication

The deploy endpoint uses API key authentication via the `Authorization` header, separate from the session-based auth used by the management UI.

### Setting Up an API Key

Set the `PUBLIC_API_KEY` environment variable before starting GenKitKraft:

```bash
export PUBLIC_API_KEY=my-secret-key
```

If `PUBLIC_API_KEY` is not set, the deploy endpoint is publicly accessible (no authentication required).

### Using the API Key

Include the key as a Bearer token in the `Authorization` header:

```bash
curl http://localhost:8080/api/v1/agents/{agentId}/deploy/chat/completions \
  -H "Authorization: Bearer my-secret-key" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

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

## Examples

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
