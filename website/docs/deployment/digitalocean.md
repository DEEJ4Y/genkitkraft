---
title: DigitalOcean Droplet
description: Deploy GenKitKraft on a DigitalOcean droplet with Docker Compose, Caddy, and automatic HTTPS.
sidebar_position: 1
---

# DigitalOcean Droplet

This guide deploys GenKitKraft on one DigitalOcean droplet. The deployment uses Docker Compose, the pre-built GenKitKraft image, and Caddy for HTTPS.

The result:

- GenKitKraft is available at `https://genkitkraft.example.com`.
- Caddy gets and renews the TLS certificate automatically.
- The data stays in a Docker volume when you update or recreate the container.
- The containers start again automatically after a reboot or a crash.

## Prerequisites

- A DigitalOcean droplet with the **Docker on Ubuntu 22.04** image from the Marketplace. 1 vCPU and 2 GB of memory is sufficient.
- A domain or subdomain that you control. This guide uses `genkitkraft.example.com`. Replace it with your domain in all steps.
- SSH access to the droplet as `root`.

:::note

This guide uses the default single-node configuration: SQLite for the database and in-memory cache. These defaults need no other services. For more than one instance, see [Horizontal Scaling](./horizontal-scaling).

:::

## 1. Configure DNS

1. Find the public IPv4 address of the droplet in the DigitalOcean control panel.
2. At your DNS provider, add an `A` record:
   - **Name:** `genkitkraft` (the subdomain part of your domain)
   - **Value:** the droplet IP address
3. If you use Cloudflare, set the record to **DNS only** (grey cloud). Caddy must reach Let's Encrypt directly to get a certificate.
4. Make sure that the record resolves. Run this command on the droplet or on your computer:

```bash
dig +short genkitkraft.example.com
```

The output must show the droplet IP address. Do not continue to step 6 until it does.

## 2. Open the firewall ports

Caddy needs ports 80 and 443 to get the certificate and to serve HTTPS.

Use the **DigitalOcean Cloud Firewall** (recommended). Add inbound rules for:

- TCP 22 (SSH)
- TCP 80 (HTTP)
- TCP 443 (HTTPS)

:::warning

Docker writes its own iptables rules. Ports that Docker publishes are open even when `ufw` blocks them. A DigitalOcean Cloud Firewall operates outside the droplet, so it controls these ports correctly.

:::

If you also use `ufw`, run these commands on the droplet:

```bash
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

## 3. Create the project directory

Connect to the droplet. Then create a directory for the deployment:

```bash
ssh root@YOUR_DROPLET_IP
mkdir -p /opt/genkitkraft && cd /opt/genkitkraft
```

All later commands in this guide run from `/opt/genkitkraft`.

## 4. Create the environment file

This command generates the secrets and writes them to `.env`:

```bash
cat > .env <<EOF
ENCRYPTION_KEY=$(openssl rand -base64 32)
AUTH_CREDENTIALS=admin:$(openssl rand -hex 16)
PUBLIC_API_KEY=sk-$(openssl rand -hex 24)
EOF
chmod 600 .env
cat .env
```

Copy the output to a password manager.

| Variable           | Purpose                                                                                                 |
| ------------------ | ------------------------------------------------------------------------------------------------------- |
| `ENCRYPTION_KEY`   | Encrypts the LLM provider API keys in the database. **Required.** The server does not start without it. |
| `AUTH_CREDENTIALS` | The username and password for the UI, the management API, and the MCP endpoint.                         |
| `PUBLIC_API_KEY`   | The bearer key for the [Deploy API](../api/deploy) endpoints (chat widget and chat completions).        |

:::danger

Keep a backup of `ENCRYPTION_KEY`. If you lose it or change it, GenKitKraft cannot read the saved provider API keys. You must then create all providers again.

:::

:::warning

Always set `AUTH_CREDENTIALS` on a public server. If it is not set, authentication is disabled. Then any person who finds the URL has full access to the UI and the API.

:::

The generated password is hexadecimal. It does not contain commas or colons, because these characters separate the `username:password` pairs.

For all variables, see [Environment Variables](../configuration/environment-variables).

## 5. Create the Caddy and Docker Compose files

Create the `Caddyfile`. Replace the domain with your domain:

```bash
cat > Caddyfile <<'EOF'
genkitkraft.example.com {
    reverse_proxy genkitkraft:8080
}
EOF
```

Create `docker-compose.yml`:

```bash
cat > docker-compose.yml <<'EOF'
services:
  genkitkraft:
    image: ghcr.io/deej4y/genkitkraft:latest
    container_name: genkitkraft
    restart: unless-stopped
    volumes:
      - genkitkraft-data:/data
    environment:
      PORT: 8080
      DATABASE_PATH: /data/app.db
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AUTH_CREDENTIALS: ${AUTH_CREDENTIALS}
      PUBLIC_API_KEY: ${PUBLIC_API_KEY}

  caddy:
    image: caddy:2-alpine
    container_name: caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
    depends_on:
      - genkitkraft

volumes:
  genkitkraft-data:
  caddy-data:
EOF
```

The `genkitkraft` service does not publish port 8080. Only Caddy can connect to it, through the internal Docker network.

Caddy sends Server-Sent Events (SSE) without buffering. The playground streaming works with no more configuration.

## 6. Start GenKitKraft

```bash
docker compose pull
docker compose up -d
docker compose logs -f
```

In the logs, make sure that:

- GenKitKraft completes the database migrations and starts. If a migration fails, the server does not start, and the log shows the cause.
- Caddy gets the certificate for your domain.

Press `Ctrl+C` to stop the log output. The containers continue to run.

## 7. Open the UI

1. Go to `https://genkitkraft.example.com`.
2. Enter the username `admin` and the password from `.env`.
3. Continue with [First Steps](../getting-started/first-steps) to add an LLM provider.

## Start on boot

You do not need a systemd service for GenKitKraft. Docker is a systemd service, and the `restart: unless-stopped` policy starts the containers again when Docker starts.

Make sure that Docker starts on boot:

```bash
systemctl is-enabled docker
```

The output must be `enabled`. If it is not, run:

```bash
systemctl enable docker
```

To test, reboot the droplet. Then connect again and run `docker ps`. Both containers must show the status `Up`.

:::note

If you stop a container manually with `docker compose stop`, it stays stopped after a reboot. Run `docker compose up -d` to start it again.

:::

## Connect Claude Desktop through MCP

GenKitKraft has an MCP server at `https://genkitkraft.example.com/mcp`. It uses the Streamable HTTP transport. When `AUTH_CREDENTIALS` is set, the endpoint requires HTTP Basic Auth with the same credentials. For the list of tools, see [MCP Quickstart](../guides/mcp-quickstart).

### 1. Get the Basic Auth value

Run this command on the droplet:

```bash
cd /opt/genkitkraft && grep '^AUTH_CREDENTIALS=' .env | cut -d= -f2 | tr -d '\n' | base64 -w0; echo
```

The command assumes that `AUTH_CREDENTIALS` contains one `username:password` pair. Copy the output.

:::warning

Base64 is an encoding, not encryption. Keep this value secret. It gives full admin access.

:::

### 2. Edit the Claude Desktop configuration

In Claude Desktop, go to **Settings > Developer > Edit Config**. Or open the file directly:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

Add the `genkitkraft` entry under `mcpServers`. Replace `PASTE_BASE64_HERE` with the value from step 1. If the file already contains other servers, keep them.

This configuration uses the `mcp-remote` bridge. It requires Node.js on your computer:

```json
{
  "mcpServers": {
    "genkitkraft": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote",
        "https://genkitkraft.example.com/mcp",
        "--header",
        "Authorization:${AUTH_HEADER}"
      ],
      "env": {
        "AUTH_HEADER": "Basic PASTE_BASE64_HERE"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

Quit Claude Desktop fully from the menu bar or the system tray. Closing the window is not sufficient. Then open it again.

To test the connection, send this message in a new chat:

```text
Use genkitkraft to list the available provider types.
```

### Troubleshooting

| Problem                    | Cause and fix                                                                                   |
| -------------------------- | ----------------------------------------------------------------------------------------------- |
| `401 Unauthorized`         | The Base64 value is incorrect. Generate it again with the command in step 1.                    |
| Connection error           | DNS or HTTPS is not ready. Make sure that `https://genkitkraft.example.com` opens in a browser. |
| No tools in Claude Desktop | Node.js is not installed, or Claude Desktop did not restart fully.                              |

## Add or rotate a public API key

GenKitKraft reads its configuration only from environment variables. To change `PUBLIC_API_KEY`, edit `.env` and recreate the container.

`PUBLIC_API_KEY` accepts more than one key, separated by commas. Use this to rotate a key with no downtime for clients.

### Add a key

These commands add a new key and keep the current keys:

```bash
cd /opt/genkitkraft
NEW=sk-$(openssl rand -hex 24)
echo "$NEW"
sed -i "s|^PUBLIC_API_KEY=.*|&,$NEW|" .env
grep '^PUBLIC_API_KEY=' .env
docker compose up -d
```

Copy the new key from the `echo` output. The `grep` output shows all active keys.

### Rotate a key

1. Add a new key with the procedure above.
2. Change your clients to send `Authorization: Bearer <new key>`.
3. When no client uses the old key, remove it from `.env`. Replace `NEW_KEY` with the key to keep:

```bash
cd /opt/genkitkraft
sed -i "s|^PUBLIC_API_KEY=.*|PUBLIC_API_KEY=NEW_KEY|" .env
docker compose up -d
```

To revoke one leaked key immediately, do step 3 only. Write all the keys to keep, separated by commas, with no spaces.

:::warning

Use `docker compose up -d` after you edit `.env`. The command `docker compose restart` does not read `.env` again, so the old keys stay active.

:::

Recreating the container causes a few seconds of downtime. The cache is in memory, so all UI users must sign in again. The data in the volume does not change.

You can change `AUTH_CREDENTIALS` with the same procedure. After you change it, generate the Claude Desktop Base64 value again.

Do not change `ENCRYPTION_KEY` with this procedure. A new key makes the saved provider API keys unreadable.

## Upgrade GenKitKraft

The database is in the named Docker volume `genkitkraft-data`, mounted at `/data`. An upgrade replaces only the container. The volume and its data do not change. The Caddy certificates are in a separate volume, `caddy-data`.

GenKitKraft runs database migrations automatically when it starts. Make a backup before each upgrade.

### 1. Make a backup

Find the full volume name. Docker Compose adds the project directory name as a prefix:

```bash
docker volume ls | grep genkitkraft
```

The examples below use `genkitkraft_genkitkraft-data`. If your volume name is different, use your name.

Stop GenKitKraft so that the SQLite file is consistent. Then make the backup:

```bash
cd /opt/genkitkraft
docker compose stop genkitkraft
mkdir -p /root/backups
docker run --rm -v genkitkraft_genkitkraft-data:/data -v /root/backups:/backup alpine tar czf /backup/genkitkraft-$(date +%F-%H%M).tar.gz -C /data .
ls -lh /root/backups
```

### 2. Pull the new image and recreate the container

```bash
docker compose pull genkitkraft
docker compose up -d
docker compose logs -f genkitkraft
```

In the logs, make sure that the migrations complete and the server starts.

### 3. Remove old images

When the new version operates correctly, remove the unused images:

```bash
docker image prune -f
```

### Use a specific version

The `latest` tag moves to each new release. To control upgrades, set a specific tag in `docker-compose.yml`:

```yaml
image: ghcr.io/deej4y/genkitkraft:<version>
```

The available tags are on the [GHCR page for GenKitKraft](https://github.com/DEEJ4Y/genkitkraft/pkgs/container/genkitkraft). A specific tag also makes it easier to go back to a previous version.

### Go back to a previous version

A migration can change the database schema. An older version possibly cannot read a newer schema. Thus, you must restore the backup together with the previous image tag.

1. Stop GenKitKraft and restore the backup. Replace `BACKUP_FILE` with the backup file name:

```bash
cd /opt/genkitkraft
docker compose stop genkitkraft
docker run --rm -v genkitkraft_genkitkraft-data:/data -v /root/backups:/backup alpine sh -c "rm -rf /data/* && tar xzf /backup/BACKUP_FILE.tar.gz -C /data"
```

2. Set the previous version tag in `docker-compose.yml`.
3. Start GenKitKraft:

```bash
docker compose up -d
```

## Protect your data

:::danger

These commands delete the GenKitKraft data volume:

- `docker compose down -v`
- `docker volume rm genkitkraft_genkitkraft-data`
- `docker system prune --volumes`

The command `docker compose down` without `-v` is safe.

:::

The volume is on the droplet disk. If you lose the droplet, you lose the volume. Copy the backups to a different location regularly, for example with `scp`. You can also enable DigitalOcean droplet backups.

## Useful commands

Run these commands from `/opt/genkitkraft`:

| Task                        | Command                                       |
| --------------------------- | --------------------------------------------- |
| Show the container status   | `docker compose ps`                           |
| Show the GenKitKraft logs   | `docker compose logs -f genkitkraft`          |
| Show the Caddy logs         | `docker compose logs -f caddy`                |
| Restart GenKitKraft         | `docker compose restart genkitkraft`          |
| Apply changes to `.env`     | `docker compose up -d`                        |
| Upgrade to the latest image | `docker compose pull && docker compose up -d` |
