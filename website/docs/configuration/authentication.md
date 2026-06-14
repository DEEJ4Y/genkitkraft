---
sidebar_position: 2
---

# Authentication

## Overview

GenKitKraft supports optional basic authentication. When enabled, users must log in before accessing the UI or API.

## Enabling Authentication

Set the `AUTH_CREDENTIALS` environment variable:

```bash
AUTH_CREDENTIALS=admin:strongpassword,user2:anotherpassword
```

Multiple users can be configured by separating credential pairs with commas.

:::note

`AUTH_CREDENTIALS` is read once at startup. If you add, change, or remove credentials, you must **restart the server** for the changes to take effect.

:::

## Login Flow

1. When auth is enabled, users see a login dialog on first visit
2. Enter username and password
3. On successful login, a session cookie is set (HttpOnly for security)
4. The session persists until the user logs out or the cookie expires

## API Authentication

When auth is enabled, API requests also require authentication. The login endpoint:

```
POST /api/auth/login
Content-Type: application/json

{"username": "admin", "password": "strongpassword"}
```

The response sets a session cookie that must be included in subsequent requests.

### Auth Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/api/auth/status` | GET | Check if authentication is required |
| `/api/auth/login` | POST | Log in (returns session cookie) |
| `/api/auth/logout` | POST | Log out (clears session cookie) |
| `/api/auth/me` | GET | Get current authenticated user |

## Disabling Authentication

Simply don't set (or remove) the `AUTH_CREDENTIALS` environment variable. When unset, all access is open without login.

## Security Considerations

- Use strong, unique passwords
- Always enable auth when exposing GenKitKraft to the internet
- Consider placing behind a reverse proxy with TLS (see [Reverse Proxy](../deployment/reverse-proxy))
- Rate limiting is applied to the login endpoint to prevent brute-force attacks
