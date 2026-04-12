---
sidebar_position: 2
---

# Reverse Proxy

Running GenKitKraft behind a reverse proxy is recommended for production deployments to enable TLS, custom domains, and additional security.

## Caddy (Recommended)

Caddy automatically handles TLS certificates via Let's Encrypt.

```
genkitkraft.example.com {
    reverse_proxy localhost:8080
}
```

That's it — Caddy handles TLS automatically.

## Nginx

```nginx
server {
    listen 443 ssl;
    server_name genkitkraft.example.com;

    ssl_certificate /etc/ssl/certs/your-cert.pem;
    ssl_certificate_key /etc/ssl/private/your-key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE support for playground streaming
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400s;
    }
}
```

:::caution SSE Support
The playground uses Server-Sent Events (SSE) for streaming responses. Make sure your reverse proxy configuration disables buffering for SSE to work correctly. The Nginx config above includes the necessary `proxy_buffering off` directive.
:::

## Docker Compose with Caddy

```yaml
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    volumes:
      - genkitkraft-data:/data
    environment:
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
    restart: unless-stopped

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
    restart: unless-stopped

volumes:
  genkitkraft-data:
  caddy-data:
```

With a `Caddyfile`:

```
genkitkraft.example.com {
    reverse_proxy genkitkraft:8080
}
```
