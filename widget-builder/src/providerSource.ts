/**
 * Source of the reference custom ChatProvider that talks to YOUR backend (never to genkitkraft
 * directly). Emitted as TypeScript (npm) or plain JavaScript (CDN) so the docs, the builder's
 * "Provider" tab and the generated snippets always show the same code.
 */
export function providerSource(lang: 'ts' | 'js'): string {
  const ts = lang === 'ts'
  const t = (s: string) => (ts ? s : '')
  return `${
    ts
      ? `import type {
  ChatProvider,
  ChatProviderSession,
  ChatProviderListSessionMessagesMessage,
  ChatProviderSendMessageOptions,
} from "navigableai-chat-widget";

`
      : ''
  }export class GenkitkraftChatProvider${t(' implements ChatProvider')} {
  constructor(${t('private ')}options${t(': { baseUrl: string; getHeaders?: () => Record<string, string> }')}) {${
    ts ? '' : '\n    this.options = options;\n  '
  }}

  ${t('private ')}async request${t('<T>')}(path${t(': string')}, init${t(': RequestInit')} = {})${t(': Promise<T>')} {
    const res = await fetch(\`\${this.options.baseUrl}\${path}\`, {
      ...init,
      credentials: "include", // send your app's session cookie
      headers: {
        "Content-Type": "application/json",
        ...(this.options.getHeaders?.() ?? {}), // e.g. Authorization: Bearer <your JWT>
        ...init.headers,
      },
    });
    if (!res.ok) throw new Error(\`Chat request failed (\${res.status})\`);
    return res.json();
  }

  async listSessions()${t(': Promise<ChatProviderSession[]>')} {
    const sessions${t(': ChatProviderSession[]')} = await this.request("/sessions");
    // The widget treats the first session as the current one, so list the newest first.
    return [...sessions].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt));
  }

  async createSession()${t(': Promise<string>')} {
    const session = await this.request${t('<{ id: string }>')}("/sessions", { method: "POST", body: "{}" });
    return session.id;
  }

  async listSessionMessages({ sessionId }${t(': { sessionId?: string }')})${t(
    ': Promise<ChatProviderListSessionMessagesMessage[]>',
  )} {
    if (!sessionId) return [];
    return this.request(\`/sessions/\${encodeURIComponent(sessionId)}/messages\`);
  }

  async sendMessage({ sessionId, content }${t(': ChatProviderSendMessageOptions')})${t(
    ': Promise<ChatProviderListSessionMessagesMessage>',
  )} {
    const id = sessionId ?? (await this.createSession());
    return this.request(\`/sessions/\${encodeURIComponent(id)}/messages\`, {
      method: "POST",
      body: JSON.stringify({ content }),
    });
  }
}
`
}
