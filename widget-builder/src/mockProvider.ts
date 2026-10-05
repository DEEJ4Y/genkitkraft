import type { ChatProvider, ChatProviderListSessionMessagesMessage, ChatProviderSession } from 'navigableai-chat-widget'

/** In-memory provider used by previews. No network. */
export function createMockProvider(): ChatProvider {
  const sessions: ChatProviderSession[] = [
    { id: 'demo-1', title: 'Getting started', createdAt: new Date(Date.now() - 3600_000).toISOString() },
  ]
  const messages: Record<string, ChatProviderListSessionMessagesMessage[]> = {
    'demo-1': [
      { role: 'user', content: 'How do I get started?' },
      { role: 'assistant', content: 'This is a **preview**. Your real agent answers here once your backend is connected.' },
    ],
  }
  const wait = (ms: number) => new Promise((r) => setTimeout(r, ms))
  let n = 1

  return {
    async listSessions() {
      return [...sessions]
    },
    async createSession() {
      const id = `demo-${++n}`
      sessions.unshift({ id, title: 'New chat', createdAt: new Date().toISOString() })
      messages[id] = []
      return id
    },
    async listSessionMessages({ sessionId }) {
      return [...(messages[sessionId ?? ''] ?? [])]
    },
    async sendMessage({ sessionId, content }) {
      await wait(500)
      const id = sessionId ?? 'demo-1'
      const reply: ChatProviderListSessionMessagesMessage = {
        role: 'assistant',
        content: `You said: "${content}". (Preview mode: no agent is connected.)`,
        createdAt: new Date().toISOString(),
      }
      messages[id] = [...(messages[id] ?? []), { role: 'user', content }, reply]
      return reply
    },
  }
}
