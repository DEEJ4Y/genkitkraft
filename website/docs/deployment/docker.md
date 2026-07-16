---
sidebar_position: 1
---

# Docker Deployment

## Quick Start

```bash
docker run -d \
  --name genkitkraft \
  -p 8080:8080 \
  -v genkitkraft-data:/data \
  -e ENCRYPTION_KEY=$(openssl rand -base64 32) \
  -e AUTH_CREDENTIALS=admin:changeme \
  --restart unless-stopped \
  ghcr.io/deej4y/genkitkraft:latest
```

## Docker Compose (Recommended)

For production, use Docker Compose:

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    ports:
      - "8080:8080"
    volumes:
      - genkitkraft-data:/data
    environment:
      PORT: 8080
      DATABASE_PATH: /data/app.db
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
    restart: unless-stopped

volumes:
  genkitkraft-data:
```

Use a `.env` file for secrets:

```bash
ENCRYPTION_KEY=your-generated-key-here
AUTH_CREDENTIALS=admin:strongpassword
```

Then: `docker compose up -d`

## Persistent Storage

By default, GenKitKraft stores all data in a SQLite database at `/data/app.db`. Mount a Docker volume or bind mount to `/data` to persist data across container restarts.

Multi-instance deployments need two shared backends, not one:

- **A shared database** — PostgreSQL, MySQL, or MariaDB via `DATABASE_PROVIDER` and `DATABASE_URL`.
- **A shared cache** — Redis or Valkey via `CACHE_PROVIDER` and `CACHE_URL`. See [Shared Cache](#shared-cache) below.

Both are required. A shared database alone still leaves sessions process-local, so logins will appear to fail at random as requests land on different instances.

See [Environment Variables](/docs/configuration/environment-variables) for the full reference.

### PostgreSQL

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    ports:
      - "8080:8080"
    environment:
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
      DATABASE_PROVIDER: postgres
      DATABASE_URL: postgres://genkitkraft:${DB_PASSWORD}@db:5432/genkitkraft?sslmode=disable
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: genkitkraft
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: genkitkraft
    volumes:
      - db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U genkitkraft"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  db-data:
```

### MySQL / MariaDB

Replace the `db` service image with `mysql:8` or `mariadb:11` and set `DATABASE_PROVIDER` accordingly:

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    ports:
      - "8080:8080"
    environment:
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
      DATABASE_PROVIDER: mysql   # or mariadb
      DATABASE_URL: genkitkraft:${DB_PASSWORD}@tcp(db:3306)/genkitkraft?parseTime=true
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: mysql:8
    environment:
      MYSQL_USER: genkitkraft
      MYSQL_PASSWORD: ${DB_PASSWORD}
      MYSQL_DATABASE: genkitkraft
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
    volumes:
      - db-data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 5s
      retries: 10

volumes:
  db-data:
```

:::note parseTime=true
The `parseTime=true` parameter is required in the MySQL/MariaDB DSN for correct timestamp handling.
:::

## Shared Cache

Session tokens, login rate-limit counters, and web-fetch results are cached in-process by default. Running more than one instance that way breaks authentication — a login handled by one instance is unknown to the others, so subsequent requests return `401`.

Point every instance at one Redis or Valkey server to fix that. Add the service alongside your database:

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    ports:
      - "8080:8080"
    environment:
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
      DATABASE_PROVIDER: postgres
      DATABASE_URL: postgres://genkitkraft:${DB_PASSWORD}@db:5432/genkitkraft?sslmode=disable
      CACHE_PROVIDER: valkey   # or redis
      CACHE_URL: valkey://cache:6379
    depends_on:
      db:
        condition: service_healthy
      cache:
        condition: service_healthy
    restart: unless-stopped

  cache:
    image: valkey/valkey:8-alpine
    command: ["valkey-server", "--maxmemory-policy", "volatile-ttl"]
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 10
    restart: unless-stopped
```

Swap the image for `redis:7-alpine` (and the healthcheck for `redis-cli ping`) to use Redis instead — set `CACHE_PROVIDER: redis` and a `redis://` URL. The two are interchangeable.

No volume is mounted: the cache holds no durable data, and losing it only logs users out. Do configure an eviction policy of `volatile-ttl` or `noeviction` so valid session tokens are not evicted under memory pressure.

The connection is checked at startup, so a misconfigured `CACHE_URL` fails immediately rather than after traffic arrives.

## Health Checks

GenKitKraft exposes health check endpoints:

| Endpoint | Description |
|---|---|
| `GET /livez` | Returns 200 if the server is running |
| `GET /readyz` | Returns 200 if the server is ready, 503 if not |

Example Docker Compose health check:

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/readyz"]
  interval: 30s
  timeout: 5s
  retries: 3
  start_period: 10s
```

## Updating

```bash
docker compose pull
docker compose up -d
```

Your data is preserved in the named volume.

## Building from Source

If you prefer to build the Docker image yourself:

```bash
git clone https://github.com/DEEJ4Y/genkitkraft.git
cd genkitkraft
docker build -t genkitkraft .
```

The Dockerfile uses a multi-stage build (Node.js for UI → Go for server → Alpine runtime).
