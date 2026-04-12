---
sidebar_position: 1
---

# Installation

There are three ways to run GenKitKraft: using the pre-built Docker image, Docker Compose, or building from source.

## Using the Pre-built Docker Image (Recommended)

```bash
docker pull ghcr.io/deej4y/genkitkraft:latest

docker run -d \
  -p 8080:8080 \
  -v genkitkraft-data:/data \
  -e ENCRYPTION_KEY=$(openssl rand -base64 32) \
  ghcr.io/deej4y/genkitkraft:latest
```

:::note
Visit the [GHCR page for GenKitKraft](https://github.com/DEEJ4Y/genkitkraft/pkgs/container/genkitkraft) to see the list of available tags, including older versions.
:::

## Using Docker Compose

Create a `docker-compose.yml` (or use the one from the repository):

```yaml
services:
  genkitkraft:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - genkitkraft-data:/data
    environment:
      PORT: 8080
      DATABASE_PATH: /data/app.db
      ENCRYPTION_KEY: your-encryption-key
      # AUTH_CREDENTIALS: admin:changeme

volumes:
  genkitkraft-data:
```

Then start it:

```bash
docker compose up -d
```

## From Source

**Prerequisites:** Go 1.26+, Node.js 22+

```bash
# Build the UI
cd ui && npm ci && npm run build && cd ..

# Build and run the server
go build ./cmd/server/...
ENCRYPTION_KEY=my-secret-key PORT=8080 ./server
```

---

Once running, open [http://localhost:8080](http://localhost:8080) in your browser. See [First Steps](./first-steps) to continue setup.
