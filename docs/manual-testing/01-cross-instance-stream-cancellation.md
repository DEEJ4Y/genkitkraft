# Manual smoke test: cross-instance stream cancellation

Proves that a playground "stop generation" request landing on one server instance can cancel a
generation actually running on a *different* instance, via the shared cache
(`internal/adapters/cache_stream_registry`). The automated integration test
(`internal/adapters/cache_stream_registry/registry_integration_test.go`) proves the registry's
signalling logic in isolation; this test proves it end-to-end over real HTTP, with two live
`cmd/server` processes and (optionally) a real LLM call.

Only `CACHE_PROVIDER=redis` or `valkey` exercises the cross-instance path — `memory` is
process-local by design, so this test is meaningless under the default config.

## Prerequisites

- Docker running, with a shared MySQL/Postgres and Valkey/Redis reachable from the host. The
  simplest source for both is the repo's own `docker-compose.yml`:
  ```bash
  docker compose up -d
  ```
  This publishes MySQL on `localhost:3306` and Valkey on `localhost:6379`, and boots one
  `genkitkraft` instance on `:8080` (leave it running — it's a third, unrelated instance and
  doesn't interfere with the two you'll start for this test).
- An Agent backed by a working LLM Provider, in that same database. Two ways to get one:
  - **Reuse an existing one.** Query the shared DB directly:
    ```bash
    docker exec genkitkraft-mysql-1 mysql -ugenkitkraft -pgenkitkraft genkitkraft \
      -e "SELECT id, name, provider_id, model_id FROM agents;"
    ```
    If a row exists with a real, working provider, skip straight to [Steps](#steps) and use its
    `id` as `AGENT_ID`.
  - **Create a free, zero-cost one against a fake local LLM**, if no real provider is configured
    or you don't want to spend against a real one. See [Appendix: fake provider](#appendix-a-fake-local-llm-provider-no-cost-no-network)
    below, then continue at [Steps](#steps).

## Steps

### 1. Build the binary under test

```bash
go build -o /tmp/gkk-smoke-server ./cmd/server
```

### 2. Start two instances against the shared backends

Both instances must use the **same `ENCRYPTION_KEY`** the provider's API key was encrypted with,
or the second instance can't decrypt it. If you used `docker compose up -d` above, that key is the
literal placeholder `your-encryption-key` from `docker-compose.yml`.

```bash
PORT=8081 \
DATABASE_PROVIDER=mysql \
DATABASE_URL='genkitkraft:genkitkraft@tcp(127.0.0.1:3306)/genkitkraft?parseTime=true' \
CACHE_PROVIDER=valkey \
CACHE_URL='valkey://127.0.0.1:6379' \
ENCRYPTION_KEY='your-encryption-key' \
/tmp/gkk-smoke-server > /tmp/gkk-instanceA.log 2>&1 &

PORT=8082 \
DATABASE_PROVIDER=mysql \
DATABASE_URL='genkitkraft:genkitkraft@tcp(127.0.0.1:3306)/genkitkraft?parseTime=true' \
CACHE_PROVIDER=valkey \
CACHE_URL='valkey://127.0.0.1:6379' \
ENCRYPTION_KEY='your-encryption-key' \
/tmp/gkk-smoke-server > /tmp/gkk-instanceB.log 2>&1 &
```

Verify both came up and both see the same agent (confirms shared DB, not two isolated stores):

```bash
curl -s http://localhost:8081/readyz
curl -s http://localhost:8082/readyz
curl -s http://localhost:8081/api/v1/agents/$AGENT_ID
curl -s http://localhost:8082/api/v1/agents/$AGENT_ID   # must return the identical JSON
```

### 3. Create a playground session (on instance A)

```bash
SESSION_ID=$(curl -s -X POST http://localhost:8081/api/v1/agents/$AGENT_ID/playground/sessions \
  -H "Content-Type: application/json" \
  -d '{"title": "cross-instance cancel smoke test"}' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "$SESSION_ID"
```

### 4. Start a stream on instance A, and cancel from instance B *immediately*

This is the one step worth getting exactly right. **Do not do anything else — checking other
output, reading a file, running another command — between starting the stream and sending the
cancel.** A real model can stream a "long" response (even ~1500 words) in a few seconds, and if you
context-switch to something else first (as happened the first time this test was run — see
[Gotcha](#gotcha-dont-context-switch-between-start-and-cancel)), the generation may finish
naturally before you get back to firing the cancel, silently turning your test into a no-op
control run instead of a real one. Put both commands in one shell invocation:

```bash
curl -N -s -X POST http://localhost:8081/api/v1/agents/$AGENT_ID/playground/chat \
  -H "Content-Type: application/json" \
  -d '{"sessionId": "'"$SESSION_ID"'", "content": "Write a very long, detailed 1500-word short story about a lighthouse keeper who discovers something strange washed ashore. Do not summarize or stop early - write the full story in rich detail."}' \
  > /tmp/gkk-stream-A.log 2>&1 &

sleep 0.4

curl -i -s -X POST http://localhost:8082/api/v1/agents/$AGENT_ID/playground/sessions/$SESSION_ID/stream/cancel
```

The cancel call needs **only the session ID** — it never touches or knows about instance A's
in-memory message registration. That's the whole point: instance B resolves the session's latest
message from the shared database, then writes a "please cancel" signal into the shared cache,
which instance A's registry is polling for (every 1s by default —
`cachestreamregistry.DefaultPollInterval`).

Expect `HTTP/1.1 204 No Content` from the cancel call, typically within tens of milliseconds (the
call itself only needs to write into the cache — it doesn't wait for instance A to react).

### 5. Verify the generation actually stopped

Check the persisted message (read-only DB query):

```bash
docker exec genkitkraft-mysql-1 mysql -ugenkitkraft -pgenkitkraft genkitkraft -e \
  "SELECT id, status, LENGTH(content) AS len FROM playground_messages WHERE session_id='$SESSION_ID';"
```

A successful cross-instance cancel looks like: the assistant row has `status='error'`, with
`content` far shorter than a full reply (possibly `len=0`, if the cancel landed before the first
token arrived — the ~1s cache poll interval on instance A is often faster than model
time-to-first-token).

Check instance A's stream log:

```bash
tail -c 200 /tmp/gkk-stream-A.log
```

A successful cancel ends the SSE stream with `data: [ERROR] stream interrupted` instead of the
story continuing to `data: [DONE]`.

### 6. Control run — prove the cancel, not something else, caused it

Repeat steps 3-4 with a fresh session, but skip the cancel call entirely and let it run to
completion. Confirm the message reaches `status='complete'` with a full-length `content`. This is
what tells you the truncation in step 5 was caused by your cancel, not an unrelated error (a bad
API key, a network hiccup, etc.) — without this control, a truncated/errored message alone is
ambiguous.

### 7. Clean up

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE http://localhost:8081/api/v1/agents/$AGENT_ID/playground/sessions/$SESSION_ID
pkill -f /tmp/gkk-smoke-server
```

The shared `docker-compose` stack (`mysql`, `valkey`, and its own `genkitkraft` instance) is never
touched directly by this test — only read and written through the two temporary instances you
started, which is why it's safe to leave it running across repeated test runs.

## Gotcha: don't context-switch between start and cancel

The first time this test was run, the stream was started on instance A, and then — before sending
the cancel — several other commands ran first (checking unrelated integration test output). By the
time the cancel request went out, the real model had already streamed the entire ~1500-word story
and the message had reached `status='complete'`. The run looked like a pass-through, not a
cancellation, because it accidentally *was* one. Re-running with the cancel fired in the same
shell invocation as the stream start (a fixed `sleep 0.4` in between, no other commands) reproduced
the real cross-instance cancel and gave the `status='error'`, 0-length-content result described in
step 5. If you ever see a `status='complete'` result when you expected a cancel, suspect this
before suspecting the cache-signalling logic itself.

## Appendix A: fake local LLM provider (no cost, no network)

If no real provider is configured, or you'd rather not spend against a real one, GenKitKraft has a
provider type built exactly for pointing at an arbitrary HTTP endpoint:
`providerType: "openai_compatible"`. Point it at a tiny local HTTP server that speaks the OpenAI
streaming wire format, and you get a fully deterministic, free, offline version of this test.

**Wire format your fake server must emit** (`Content-Type: text/event-stream`, flush after every
line, ignore the incoming request body and `Authorization` header entirely):

```
data: {"id":"chatcmpl-fake1","object":"chat.completion.chunk","created":1700000000,"model":"fake-model-1","choices":[{"index":0,"delta":{"content":"Once "},"finish_reason":null}]}

data: {"id":"chatcmpl-fake1","object":"chat.completion.chunk","created":1700000000,"model":"fake-model-1","choices":[{"index":0,"delta":{"content":"upon "},"finish_reason":null}]}

... (repeat every 1-2s, for 20-30s, so there's a wide cancel window) ...

data: {"id":"chatcmpl-fake1","object":"chat.completion.chunk","created":1700000000,"model":"fake-model-1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]

```

The route requested is `POST {baseUrl}/chat/completions` (note: `baseUrl` needs a trailing path
segment like `/v1` — the client resolves `chat/completions` relative to it, so
`http://localhost:9999/v1` → `http://localhost:9999/v1/chat/completions`). Have your fake handler
watch `r.Context().Done()` between chunks and stop early — this fires naturally once GenKitKraft
cancels its side of the connection, so you don't need to do anything special to make cancellation
observable on the fake-server side.

Create the provider and agent (against instance A, before [Steps](#steps) 3 onward):

```bash
PROVIDER_ID=$(curl -s -X POST http://localhost:8081/api/v1/settings/providers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fake Local LLM",
    "providerType": "openai_compatible",
    "apiKey": "sk-fake-not-checked",
    "baseUrl": "http://localhost:9999/v1"
  }' | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

AGENT_ID=$(curl -s -X POST http://localhost:8081/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "Smoke Test Agent", "providerId": "'"$PROVIDER_ID"'", "modelId": "fake-model-1"}' \
  | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
```

Then continue at [Steps](#steps), step 3, using this `AGENT_ID`.
