---
sidebar_position: 1
---

# API Overview

GenKitKraft provides a REST API for managing agents, prompts, and providers programmatically. The API also includes an OpenAI-compatible chat completions endpoint.

## Base URL

All API endpoints are served on the same port as the UI (default: `8080`).

```
http://localhost:8080
```

## Authentication

When `AUTH_CREDENTIALS` is set, all API requests require authentication. First, obtain a session by calling the login endpoint:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "yourpassword"}' \
  -c cookies.txt

# Use the session cookie in subsequent requests
curl http://localhost:8080/api/v1/agents -b cookies.txt
```

When auth is disabled, no authentication is needed.

## OpenAI-Compatible API (Work In Progress)

GenKitKraft exposes configured agents through an OpenAI-compatible chat completions endpoint. This allows you to use GenKitKraft as a drop-in replacement in applications that support the OpenAI API format.

The agent name is used as the model identifier in OpenAI-compatible requests.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "my-agent-name",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

This works with any OpenAI-compatible client library (Python openai, Node.js openai, etc.):

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed"  # unless auth is enabled
)

response = client.chat.completions.create(
    model="my-agent-name",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## Response Format

All API responses use JSON. Successful responses return the resource directly. Errors return:

```json
{
  "message": "Description of the error"
}
```

## Pagination

List endpoints (agents, prompts, sessions) support pagination with `limit` and `offset` query parameters:

```bash
GET /api/v1/agents?limit=10&offset=20
```

- `limit` — Number of items to return (default: 20)
- `offset` — Number of items to skip (default: 0)
