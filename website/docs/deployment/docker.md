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

GenKitKraft stores all data in a SQLite database. The default path is `/data/app.db`. Mount a Docker volume or bind mount to `/data` to persist data across container restarts.

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
