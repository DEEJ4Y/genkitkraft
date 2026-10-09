import { WIDGET_CDN_URL, WIDGET_NPM_PACKAGE } from './constants'
import { normalizeConfig } from './normalize'
import { providerSource } from './providerSource'
import type { BuilderConfig } from './types'

export interface SnippetOptions {
  /** Base URL of YOUR backend chat endpoints, e.g. "/api/chat". */
  backendBaseUrl?: string
}

const DEFAULT_BASE_URL = '/api/chat'

function indent(text: string, spaces: number): string {
  const pad = ' '.repeat(spaces)
  return text
    .split('\n')
    .map((l, i) => (i === 0 || !l ? l : pad + l))
    .join('\n')
}

/** Config literal with the provider first. Never mutates `config`. */
function configLiteral(config: BuilderConfig, baseUrl: string): string {
  const json = JSON.stringify(normalizeConfig(config), null, 2)
  const body = json === '{}' ? '' : json.slice(1, -1).replace(/^\n/, '').replace(/\n$/, '')
  const provider = `  "chatProvider": new GenkitkraftChatProvider({ baseUrl: ${JSON.stringify(baseUrl)} })`
  return `{\n${provider}${body ? ',\n' + body : ''}\n}`
}

export function generateNpmSnippet(config: BuilderConfig, opts: SnippetOptions = {}): string {
  const baseUrl = opts.backendBaseUrl || DEFAULT_BASE_URL
  return `import { injectAiChatWidget } from "${WIDGET_NPM_PACKAGE}";
import { GenkitkraftChatProvider } from "./GenkitkraftChatProvider";

injectAiChatWidget(${configLiteral(config, baseUrl)});
`
}

export function generateCdnSnippet(config: BuilderConfig, opts: SnippetOptions = {}): string {
  const baseUrl = opts.backendBaseUrl || DEFAULT_BASE_URL
  const provider = providerSource('js').replace(/^export /m, '')
  return `<script type="module" src="${WIDGET_CDN_URL}"></script>
<script type="module">
${provider}
window.initAiChatWidget(${configLiteral(config, baseUrl)});
</script>
`
}

export function generateProviderSnippet(): string {
  return providerSource('ts')
}

export const installCommand = `npm install ${WIDGET_NPM_PACKAGE}`

/**
 * Express reference backend for the four endpoints the provider calls. GenKitKraft stores the
 * history; the backend only stores which user owns which session. Keep in sync with the Node
 * example in website/docs/guides/chat-widget.md.
 */
export const backendExample = `import express from "express";

const GENKITKRAFT = process.env.GENKITKRAFT_URL;      // e.g. https://genkitkraft.internal
const AGENT_ID = process.env.GENKITKRAFT_AGENT_ID;
const API_KEY = process.env.GENKITKRAFT_API_KEY;      // server-side only, never sent to the browser

const deploy = (path, init = {}) =>
  fetch(\`\${GENKITKRAFT}/api/v1/agents/\${AGENT_ID}/deploy\${path}\`, {
    ...init,
    headers: { "Content-Type": "application/json", Authorization: \`Bearer \${API_KEY}\`, ...init.headers },
  });

const router = express.Router();
router.use(requireLogin); // YOUR auth middleware: sets req.user

// db.sessions only stores { id, userId }: who owns which session. GenKitKraft stores the rest.

router.post("/sessions", async (req, res) => {
  const r = await deploy("/sessions", { method: "POST", body: "{}" });
  if (!r.ok) return res.sendStatus(502);
  const session = await r.json();
  await db.sessions.insert({ id: session.id, userId: req.user.id });
  res.status(201).json({ id: session.id });
});

router.get("/sessions", async (req, res) => {
  // GET .../deploy/sessions lists EVERY session of the agent, so read only the sessions this user owns.
  const ids = await db.sessions.listIdsByUser(req.user.id); // cap this, e.g. the 50 most recent
  const sessions = await Promise.all(
    ids.map(async (id) => {
      const r = await deploy(\`/sessions/\${id}\`);
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

  const r = await deploy(\`/sessions/\${id}/messages\`);
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

  const r = await deploy(\`/sessions/\${id}/chat/completions\`, {
    method: "POST",
    body: JSON.stringify({ messages: [{ role: "user", content }], stream: false }),
  });
  if (!r.ok) return res.sendStatus(502);
  const reply = (await r.json()).choices[0].message.content;

  res.json({ role: "assistant", content: reply, createdAt: new Date().toISOString() });
});

export default router;
`
