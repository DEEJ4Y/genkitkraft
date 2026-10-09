---
title: App Platform
description: Deploy GenKitKraft on DigitalOcean App Platform with one click. This guide is for trials. It includes the steps to replace the public default credentials.
sidebar_position: 1
---

# DigitalOcean App Platform

This guide deploys GenKitKraft on DigitalOcean App Platform with one click. The deployment uses the pre-built GenKitKraft image from GHCR.

[![Deploy to DO](https://www.deploytodo.com/do-btn-blue.svg)](https://cloud.digitalocean.com/apps/new?repo=https://github.com/DEEJ4Y/genkitkraft/tree/main)

:::warning

**Trial defaults:** this deployment uses public default credentials (`admin` / `demo-change-me`) and a public default `ENCRYPTION_KEY`. Anyone who knows them can sign in to your app. Update the three environment variables before you add a real LLM provider key. See [Update the three environment variables](#update-the-three-environment-variables-after-you-deploy).

:::

Which deployment guide to use:

- **App Platform (this page):** use it to try GenKitKraft. You do not manage a server.
- **[DigitalOcean Droplet](./digitalocean):** use it when you want your own server, a domain, and HTTPS from Caddy.
- **[Docker](./docker):** use it when you run GenKitKraft on your own host or on your computer.

## What the button creates

The button creates one app with three components:

- The `genkitkraft` service. It runs the public GHCR image `ghcr.io/deej4y/genkitkraft:latest`.
- A development PostgreSQL database named `db`.
- An internal Valkey service named `valkey`. It is the shared cache.

You pay DigitalOcean for these components. For the current prices, see [App Platform pricing](https://www.digitalocean.com/pricing/app-platform).

## Deploy

1. Click the **Deploy to DO** button.
2. Sign in to DigitalOcean.
3. Choose a region.
4. Click **Create**.
5. Wait for the build and the deployment to finish.

The template sets default values for `ENCRYPTION_KEY`, `AUTH_CREDENTIALS`, and `PUBLIC_API_KEY`. The app starts with no input from you.

## First sign-in

1. Open the app URL that DigitalOcean shows for the `genkitkraft` service.
2. Sign in with these credentials:
   - **Username:** `admin`
   - **Password:** `demo-change-me`

:::danger

These credentials are public. Anyone who reads the repository knows them. Do not add a real provider key until you replace them.

:::

## Update the three environment variables after you deploy

:::danger

This step is required. Do it before you add a real LLM provider key.

:::

1. Open the app in DigitalOcean.
2. Go to **Settings**, then **Components**, then **genkitkraft**, then **Environment Variables**, then **Edit**.
3. Replace `ENCRYPTION_KEY`, `AUTH_CREDENTIALS`, and `PUBLIC_API_KEY` with new values.
4. Select **Encrypt** for each variable.
5. Save. DigitalOcean starts a new deployment.

This command generates the values:

```bash
echo "ENCRYPTION_KEY=$(openssl rand -base64 32)"
echo "AUTH_CREDENTIALS=admin:$(openssl rand -hex 16)"
echo "PUBLIC_API_KEY=sk-$(openssl rand -hex 24)"
```

| Variable           | Purpose                                                                                                 |
| ------------------ | ------------------------------------------------------------------------------------------------------- |
| `ENCRYPTION_KEY`   | Encrypts the LLM provider API keys in the database. **Required.** The server does not start without it. |
| `AUTH_CREDENTIALS` | The username and password for the UI, the management API, and the MCP endpoint.                         |
| `PUBLIC_API_KEY`   | The bearer key for the [Deploy API](../api/deploy) endpoints (chat widget and chat completions).        |

:::warning

Change `ENCRYPTION_KEY` before you add providers. A changed key makes saved provider API keys unreadable.

:::

Save `ENCRYPTION_KEY` in a password manager.

After the new deployment finishes, sign in with the new `AUTH_CREDENTIALS` value. For all variables, see [Environment Variables](../configuration/environment-variables).

## Custom domain

To use your own domain, follow the DigitalOcean documentation: [How to Manage Domains in App Platform](https://docs.digitalocean.com/products/app-platform/how-to/manage-domains/).

## Scale up

You can increase the instance count of the `genkitkraft` service. For the requirements, see [Horizontal Scaling](./horizontal-scaling).

Keep the `valkey` service at one instance. Each extra `valkey` instance has its own cache. This breaks sign-in.

## Limits and notes

- The database is a development database. It has a low connection limit. The app uses a pool of up to 25 connections. For production, use a managed PostgreSQL cluster and set `DATABASE_URL`.
- The Valkey cache is not durable. A restart signs users out. No application data is lost.
- If the first deployment fails because Valkey is not ready, deploy again.
- The template uses the `latest` image tag. For production, pin a version tag. The available tags are on the [GHCR page for GenKitKraft](https://github.com/DEEJ4Y/genkitkraft/pkgs/container/genkitkraft).

## Upgrade

GenKitKraft runs database migrations automatically when it starts.

To deploy a new image version, pin a version tag:

1. Open the app in DigitalOcean.
2. Change the image tag of the `genkitkraft` component to the new version. For the steps, see [How to Update an App's Spec](https://docs.digitalocean.com/products/app-platform/how-to/update-app-spec/).
3. Deploy the app.

To start a deployment from the control panel, open the **Actions** menu of the app and select **Deploy**. For more information, see [How to Manage Deployments in App Platform](https://docs.digitalocean.com/products/app-platform/how-to/manage-deployments/).

:::note

The DigitalOcean documentation does not state that a deployment pulls a newer image for the `latest` tag. Pin a version tag to be sure which version runs.

:::

## Troubleshooting

| Symptom                                                                | Cause                                                       | Fix                                                                                          |
| ---------------------------------------------------------------------- | ----------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Error `ENCRYPTION_KEY environment variable is required`                | The variable is empty.                                      | Set a value for `ENCRYPTION_KEY`.                                                            |
| Sign-in returns 503.                                                   | The app cannot reach the cache.                             | Check that the `valkey` service runs. Check the value of `CACHE_URL`.                     |
| Sign-in returns 401 after a successful login.                          | More than one instance uses the in-memory cache.            | Set `CACHE_PROVIDER` and `CACHE_URL` the same way on every instance.                         |
