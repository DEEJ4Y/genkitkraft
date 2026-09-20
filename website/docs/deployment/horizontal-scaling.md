---
sidebar_position: 3
---

# Horizontal Scaling

GenKitKraft can run as multiple stateless instances behind a normal load balancer — no session affinity ("sticky sessions") required. This page is the checklist for going from one instance to N; it links out to the pages that cover each piece in detail rather than repeating them.

## Requirements

All three of these are required together. Meeting only one or two still breaks in ways described below.

### 1. A shared database

Point every instance at the same PostgreSQL, MySQL, or MariaDB database via [`DATABASE_PROVIDER` and `DATABASE_URL`](/docs/configuration/environment-variables#database_provider). SQLite is single-node by design — it doesn't support this.

### 2. A shared cache

Point every instance at the same Redis or Valkey server via [`CACHE_PROVIDER` and `CACHE_URL`](/docs/configuration/environment-variables#cache_provider). This is what makes login sessions, rate-limit counters, and cross-instance "stop generation" work no matter which instance a request lands on. See [Shared Cache](/docs/deployment/docker#shared-cache) for a Docker Compose example.

### 3. An identical `ENCRYPTION_KEY` on every instance

This one doesn't appear anywhere else in the docs, and it's easy to miss until you add a second instance: **every instance must be configured with the exact same `ENCRYPTION_KEY`.**

LLM provider API keys are encrypted with this key before being written to the shared database. If instance A creates a provider, its encrypted credentials are only readable by an instance holding the same key. An instance with a different key gets a decryption failure — loud, per-request, not silent corruption — the first time it tries to use that provider.

Generate the key once and distribute it to every replica through your secrets manager or orchestrator (a single Kubernetes `Secret`, a shared `.env` file, etc.) — don't let each instance generate its own or receive a different value:

```bash
openssl rand -base64 32
```

See [`ENCRYPTION_KEY`](/docs/configuration/environment-variables#encryption_key) for the rest of its behavior (e.g. what happens if it's lost).

## What already works once the above is in place

With a shared DB, a shared cache, and a matching `ENCRYPTION_KEY`, the following need no further configuration:

- **Concurrent startup is safe.** If multiple instances start at once against a database with pending migrations, they don't race. PostgreSQL uses a session-level advisory lock; MySQL/MariaDB use `GET_LOCK()`/`RELEASE_LOCK()`. Losing instances block until the migrating instance finishes, then proceed — none crash.
- **"Stop generation" reaches the right instance.** A playground or [Deploy API cancel](/docs/api/deploy#cancel) request is relayed through the shared cache to whichever instance is actually running that stream, even if the request lands on a different one.
- **Interrupted streams resume from any instance.** The [`Last-Event-ID` reconnect mechanism](/docs/api/deploy#reconnect) replays already-generated content from what's persisted in the shared database — it does not depend on reconnecting to the same instance that started the stream. This is the reason session affinity isn't needed at the load balancer even for long-running streamed replies.
- **Health checks are instance-local.** Point your load balancer at `GET /livez` and `GET /readyz` on each instance (see [Health Checks](/docs/deployment/docker#health-checks)).

## What to watch for

- Running any instance with `CACHE_PROVIDER=memory` (the default) while others use a shared cache defeats the point — that instance's sessions, rate limits, and cancel signals stay local to it. Set `CACHE_PROVIDER` the same way on every instance.
- The built-in `web_fetch` tool's result cache is intentionally **not** shared across instances, even with a shared cache configured — see the note in [`CACHE_PROVIDER`](/docs/configuration/environment-variables#cache_provider). This only means a cold instance re-fetches a URL another instance already fetched; it's not a correctness issue.
