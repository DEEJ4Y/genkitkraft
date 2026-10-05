---
sidebar_position: 7
---

# Embedding an Agent with the Chat Widget

Put a GenKitKraft agent in front of your users with the open-source [AI chat widget](https://github.com/techorionai/ai-chat-widget): a floating launcher and chat window you add to any website.

This guide covers the four pieces of an integration. Use the [**widget builder**](/widget-builder) (also available in the dashboard under **Agents → your agent → Widget**) to design the look and copy the embed code.

1. [Create your agent](#1-create-your-agent) in GenKitKraft.
2. [Expose endpoints from your backend](#2-expose-endpoints-from-your-backend) that sit between the browser and GenKitKraft.
3. [Create a custom provider](#3-create-a-custom-provider) that connects the widget to those endpoints.
4. [Integrate the widget](#4-integrate-the-widget-in-your-web-client) in your web client.

## Architecture

```
┌──────────────────┐   your auth    ┌──────────────────┐  Bearer API key  ┌──────────────┐
│  Browser         │ ─────────────► │  Your backend    │ ───────────────► │ GenKitKraft  │
│  chat widget     │ ◄───────────── │  /api/chat/...   │ ◄─────────────── │ deploy API   │
└──────────────────┘                └──────────────────┘                  └──────────────┘
```

:::danger Never connect the browser directly to GenKitKraft
The deploy API is protected by a single shared `PUBLIC_API_KEY` (or is fully public if unset). Anything shipped to a browser can be read, so a key in client code lets anyone call your agent, run up your provider bill and read any session. The deploy API also does not send CORS headers, so cross-origin browser calls fail anyway.

Always put your own backend in the middle. It authenticates your users, owns the GenKitKraft key, and checks that each user can only reach their own sessions.
:::

## 1. Create your agent

1. In GenKitKraft, [create an agent](./agents) with a provider, model and system prompt, and test it in the [Playground](./playground).
2. Copy the **agent ID** from the agent's **Deploy** tab.
3. Set `PUBLIC_API_KEY` on your GenKitKraft server so the deploy endpoints require a key. See the [Deploy API](../api/deploy#authentication).
4. Store the GenKitKraft URL, agent ID and API key as **server-side** secrets in your backend (for example `GENKITKRAFT_URL`, `GENKITKRAFT_AGENT_ID`, `GENKITKRAFT_API_KEY`).

Use the stateful [session endpoints](../api/deploy#stateful-chat-sessions): GenKitKraft keeps the conversation (and returns it on request), and you only send the new user message each turn.

## 2. Expose endpoints from your backend

Your backend exposes four endpoints, all behind **your own authentication** (session cookie, JWT, etc.):

| Method and path | What it does |
| --- | --- |
| `POST /api/chat/sessions` | Creates a GenKitKraft session (`POST …/deploy/sessions`) and records that it belongs to the signed-in user. Returns `{ "id": "…" }`. |
| `GET /api/chat/sessions` | Lists the signed-in user's sessions, **newest first**: `[{ "id", "title", "createdAt" }]`. Reads each owned session from GenKitKraft (`GET …/deploy/sessions/{id}`). |
| `GET /api/chat/sessions/:id/messages` | Returns the conversation: `[{ "role", "content", "createdAt" }]`, read from GenKitKraft (`GET …/deploy/sessions/{id}/messages`). |
| `POST /api/chat/sessions/:id/messages` | Takes `{ "content": "…" }`, forwards it to `…/deploy/sessions/{id}/chat/completions`, and returns `{ "role": "assistant", "content": "…" }`. |

:::note GenKitKraft stores the history, you store who owns what
GenKitKraft keeps every session and message, and the [deploy API](../api/deploy#list-messages-history) returns them. It does **not** know which of your end users a session belongs to, and `GET …/deploy/sessions` lists *all* of the agent's sessions (including ones from the dashboard Playground and other users). So your backend stores only a small `userId → sessionId` mapping, and uses it to decide what each user may see. Never return GenKitKraft's session list as-is.

GenKitKraft also names sessions for you: the `title` updates from the conversation, so the sessions list shows meaningful titles without any work on your side.
:::

### Example: Node.js (Express)

```ts
import express from "express";

const GENKITKRAFT = process.env.GENKITKRAFT_URL;
const AGENT_ID = process.env.GENKITKRAFT_AGENT_ID;
const API_KEY = process.env.GENKITKRAFT_API_KEY; // server-side only

const deploy = (path: string, init: RequestInit = {}) =>
  fetch(`${GENKITKRAFT}/api/v1/agents/${AGENT_ID}/deploy${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${API_KEY}`,
      ...init.headers,
    },
  });

const router = express.Router();
router.use(requireLogin); // your auth middleware, sets req.user

// db.sessions only stores { id, userId }: the ownership mapping.

router.post("/sessions", async (req, res) => {
  const r = await deploy("/sessions", { method: "POST", body: "{}" });
  if (!r.ok) return res.sendStatus(502);
  const session = await r.json();
  await db.sessions.insert({ id: session.id, userId: req.user.id });
  res.status(201).json({ id: session.id });
});

router.get("/sessions", async (req, res) => {
  const ids: string[] = await db.sessions.listIdsByUser(req.user.id); // cap this (e.g. 50 most recent)
  const sessions = await Promise.all(
    ids.map(async (id) => {
      const r = await deploy(`/sessions/${id}`);
      if (!r.ok) return null; // deleted in GenKitKraft
      const s = await r.json();
      return { id: s.id, title: s.title, createdAt: s.created_at };
    }),
  );
  res.json(
    sessions
      .filter((s) => s !== null)
      .sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt)), // newest first
  );
});

router.get("/sessions/:id/messages", async (req, res) => {
  const { id } = req.params;
  if (!(await db.sessions.owns(req.user.id, id))) return res.sendStatus(404);

  const r = await deploy(`/sessions/${id}/messages`);
  if (!r.ok) return res.sendStatus(r.status === 404 ? 404 : 502);
  const { messages } = await r.json();
  res.json(
    messages
      .filter((m) => m.status !== "streaming") // skip a reply that is still being generated
      .map((m) => ({ role: m.role, content: m.content, createdAt: m.created_at })),
  );
});

router.post("/sessions/:id/messages", async (req, res) => {
  const { id } = req.params;
  const { content } = req.body;
  if (!(await db.sessions.owns(req.user.id, id))) return res.sendStatus(404);
  if (typeof content !== "string" || !content || content.length > 4000) return res.sendStatus(400);

  const r = await deploy(`/sessions/${id}/chat/completions`, {
    method: "POST",
    body: JSON.stringify({ messages: [{ role: "user", content }], stream: false }),
  });
  if (!r.ok) return res.sendStatus(502);
  const reply = (await r.json()).choices[0].message.content;

  res.json({ role: "assistant", content: reply, createdAt: new Date().toISOString() });
});

export default router;
```

### Example: Python (FastAPI)

```python
import asyncio
import os
from datetime import datetime

import httpx
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel, Field

BASE = f"{os.environ['GENKITKRAFT_URL']}/api/v1/agents/{os.environ['GENKITKRAFT_AGENT_ID']}/deploy"
HEADERS = {"Authorization": f"Bearer {os.environ['GENKITKRAFT_API_KEY']}"}  # server-side only

router = APIRouter(prefix="/api/chat")


class Message(BaseModel):
    content: str = Field(min_length=1, max_length=4000)


# db.sessions only stores (id, user_id): the ownership mapping.


@router.post("/sessions", status_code=201)
async def create_session(user=Depends(require_login)):  # your auth dependency
    async with httpx.AsyncClient() as client:
        r = await client.post(f"{BASE}/sessions", json={}, headers=HEADERS)
    r.raise_for_status()
    session = r.json()
    await db.sessions.insert(id=session["id"], user_id=user.id)
    return {"id": session["id"]}


@router.get("/sessions")
async def list_sessions(user=Depends(require_login)):
    ids = await db.sessions.list_ids_by_user(user.id)  # cap this (e.g. 50 most recent)
    async with httpx.AsyncClient() as client:
        responses = await asyncio.gather(*(client.get(f"{BASE}/sessions/{i}", headers=HEADERS) for i in ids))
    sessions = [
        {"id": s["id"], "title": s["title"], "createdAt": s["created_at"]}
        for s in (r.json() for r in responses if r.status_code == 200)  # skip sessions deleted in GenKitKraft
    ]
    sessions.sort(key=lambda s: datetime.fromisoformat(s["createdAt"].replace("Z", "+00:00")), reverse=True)  # newest first
    return sessions


@router.get("/sessions/{session_id}/messages")
async def list_messages(session_id: str, user=Depends(require_login)):
    if not await db.sessions.owns(user.id, session_id):
        raise HTTPException(404)
    async with httpx.AsyncClient() as client:
        r = await client.get(f"{BASE}/sessions/{session_id}/messages", headers=HEADERS)
    if r.status_code == 404:
        raise HTTPException(404)
    if r.status_code != 200:
        raise HTTPException(502)
    return [
        {"role": m["role"], "content": m["content"], "createdAt": m["created_at"]}
        for m in r.json()["messages"]
        if m["status"] != "streaming"  # skip a reply that is still being generated
    ]


@router.post("/sessions/{session_id}/messages")
async def send_message(session_id: str, body: Message, user=Depends(require_login)):
    if not await db.sessions.owns(user.id, session_id):
        raise HTTPException(404)
    async with httpx.AsyncClient(timeout=120) as client:
        r = await client.post(
            f"{BASE}/sessions/{session_id}/chat/completions",
            json={"messages": [{"role": "user", "content": body.content}], "stream": False},
            headers=HEADERS,
        )
    if r.status_code != 200:
        raise HTTPException(502)
    reply = r.json()["choices"][0]["message"]["content"]
    return {"role": "assistant", "content": reply}
```

### Security checklist

- **Authenticate every request** and check that the session belongs to the caller. Session IDs are not secrets.
- **Filter every GenKitKraft list by your ownership mapping.** `GET …/deploy/sessions` returns every session of the agent, not just your caller's.
- **Keep the GenKitKraft key server-side.** Never return it, log it, or put it in front-end config.
- **Validate input**: require a string, cap the length, and reject empty messages.
- **Rate limit** per user. Every message costs LLM tokens.
- **Restrict CORS** to your own origin if the backend is on a different origin from the page, and use `credentials: "include"` only with an explicit origin allow-list.
- **Don't expose raw upstream errors.** Return a generic error and log the details.
- **Don't let clients choose the agent.** The agent ID comes from your server configuration.

## 3. Create a custom provider

The widget gets its data through a [`ChatProvider`](https://github.com/techorionai/ai-chat-widget#implementing-a-custom-chatprovider), an object with four methods. This provider maps them to the endpoints above:

```ts title="GenkitkraftChatProvider.ts"
import type {
  ChatProvider,
  ChatProviderSession,
  ChatProviderListSessionMessagesMessage,
  ChatProviderSendMessageOptions,
} from "navigableai-chat-widget";

export class GenkitkraftChatProvider implements ChatProvider {
  constructor(private options: { baseUrl: string; getHeaders?: () => Record<string, string> }) {}

  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const res = await fetch(`${this.options.baseUrl}${path}`, {
      ...init,
      credentials: "include", // send your app's session cookie
      headers: {
        "Content-Type": "application/json",
        ...(this.options.getHeaders?.() ?? {}), // e.g. Authorization: Bearer <your JWT>
        ...init.headers,
      },
    });
    if (!res.ok) throw new Error(`Chat request failed (${res.status})`);
    return res.json();
  }

  async listSessions(): Promise<ChatProviderSession[]> {
    const sessions: ChatProviderSession[] = await this.request("/sessions");
    // The widget treats the first session as the current one, so list the newest first.
    return [...sessions].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt));
  }

  async createSession(): Promise<string> {
    const session = await this.request<{ id: string }>("/sessions", { method: "POST", body: "{}" });
    return session.id;
  }

  async listSessionMessages({ sessionId }: { sessionId?: string }): Promise<ChatProviderListSessionMessagesMessage[]> {
    if (!sessionId) return [];
    return this.request(`/sessions/${encodeURIComponent(sessionId)}/messages`);
  }

  async sendMessage({ sessionId, content }: ChatProviderSendMessageOptions): Promise<ChatProviderListSessionMessagesMessage> {
    const id = sessionId ?? (await this.createSession());
    return this.request(`/sessions/${encodeURIComponent(id)}/messages`, {
      method: "POST",
      body: JSON.stringify({ content }),
    });
  }
}
```

Notes:

- **No streaming.** `sendMessage` resolves once with the complete reply, so the widget shows a loading state and then the full answer. That is why the backend calls GenKitKraft with `"stream": false`.
- **Session order matters.** The widget treats the first session in the list as the current one, so the list must be newest first. The provider above also sorts by `createdAt` to be safe.
- **Errors** are signalled by throwing, and the widget shows an error message.
- **Auth** is yours to choose. Use cookies (`credentials: "include"`) for same-site apps, or pass a token with `getHeaders`.

## 4. Integrate the widget in your web client

### With a bundler (npm)

```bash
npm install navigableai-chat-widget@0.9.17
```

Initialise it once, in client-side code only. The widget touches `window`, so it can't run during server-side rendering.

```ts
import { injectAiChatWidget } from "navigableai-chat-widget";
import { GenkitkraftChatProvider } from "./GenkitkraftChatProvider";

injectAiChatWidget({
  chatProvider: new GenkitkraftChatProvider({ baseUrl: "/api/chat" }),
  chatWindow: {
    defaults: { primaryColor: "teal" },
    header: { title: { title: "Support", showOnlineSubtitle: true } },
  },
});
```

In React or Next.js, call it from an effect and guard against running twice:

```tsx
"use client";
import { useEffect } from "react";

export function ChatWidget() {
  useEffect(() => {
    let cancelled = false;
    (async () => {
      const { injectAiChatWidget } = await import("navigableai-chat-widget");
      const { GenkitkraftChatProvider } = await import("./GenkitkraftChatProvider");
      if (cancelled || document.getElementById("chat-widget-iframe")) return;
      injectAiChatWidget({
        chatProvider: new GenkitkraftChatProvider({ baseUrl: "/api/chat" }),
      });
    })();
    return () => {
      cancelled = true;
    };
  }, []);
  return null;
}
```

Only render the widget for signed-in users if your chat endpoints require login.

### With a script tag (CDN)

```html
<script type="module" src="https://chat.techorionai.com/builds/0.9.17/main/index.js"></script>
<script type="module">
  // paste the GenkitkraftChatProvider class here, as plain JavaScript (the builder generates it)
  window.initAiChatWidget({
    chatProvider: new GenkitkraftChatProvider({ baseUrl: "/api/chat" }),
  });
</script>
```

The [widget builder](/widget-builder) generates both snippets with your styling already filled in.

### Controlling the widget

After initialisation, `window.$aiChatWidget` exposes `open()`, `close()`, `toggle()`, `toggleColorScheme("light" | "dark")` and `getColorScheme()`. For example, to open the chat from your own button:

```ts
document.querySelector("#help-button")?.addEventListener("click", () => window.$aiChatWidget.open());
```

### Customising the look

Everything in the [builder](/widget-builder) maps to the config object you pass to `injectAiChatWidget`: colors and color scheme, header and avatars, home screen cards, welcome message, footer tabs, a custom launcher button, and named actions. Export the builder's JSON config to keep your design in version control.

The widget's position and size are fixed (bottom right of the page); only the launcher button HTML can be replaced.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Widget loads but the chat is empty or errors | Open the browser network tab. A `401` or `404` from `/api/chat/...` means your auth or session ownership check rejected the request. |
| CORS error | The widget is calling a different origin than the page. Serve the chat endpoints from the same origin, or allow your origin explicitly and use `credentials: "include"`. |
| Past conversations are empty after a reload | Your backend isn't reading `GET …/deploy/sessions/{id}/messages`, or it filters out everything. Check that the session ID is the one returned by `POST …/deploy/sessions` and that the call returns `200` (a `404` means the session was deleted or belongs to another agent). |
| `502` from your backend | Your backend couldn't reach GenKitKraft, or the key is wrong. Check the `Authorization` header, `PUBLIC_API_KEY`, and the agent ID. |
| Nothing appears | Content-Security-Policy is blocking the widget. Allow `https://chat.techorionai.com` in `script-src` (CDN usage) and `frame-src`. |
| `window is not defined` | The widget was initialised during server-side rendering. Move it into an effect or client-only component. |
| The launcher appears twice | `injectAiChatWidget` ran more than once. Guard it as shown in the React example. |
| Self-hosting the widget UI | Pass `_dev: { iframeSrc, iframeOrigin }` in the config to load the widget's iframe app from your own host. |

## Next steps

- [Deploy API reference](../api/deploy): all session and chat endpoints.
- [Reverse proxy](../deployment/reverse-proxy): if GenKitKraft sits behind a proxy.
- [Widget builder](/widget-builder): design and export the embed code.
