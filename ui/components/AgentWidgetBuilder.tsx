import { Alert, Text } from '@mantine/core'
import { WidgetBuilder } from 'genkitkraft-widget-builder'
import { DOCS_BASE_URL } from '../lib/constants'

interface AgentWidgetBuilderProps {
  agentId: string
  agentName: string
}

/**
 * Chat widget builder for an agent. The preview injects the real widget into this page with a
 * mock provider (no agent call, no API key). Render this only while its tab is active so the
 * preview is removed when the user switches tabs.
 */
export function AgentWidgetBuilder({ agentId, agentName }: AgentWidgetBuilderProps) {
  const docsUrl = `${DOCS_BASE_URL}/docs/guides/chat-widget`
  return (
    <>
      <Alert variant="light" mb="md" title="Embed this agent in your app">
        <Text size="sm">
          Design the widget, then follow the{' '}
          <a href={docsUrl} target="_blank" rel="noreferrer">
            integration guide
          </a>{' '}
          to connect it through your own backend. The preview below uses sample replies, not this agent.
        </Text>
      </Alert>
      <WidgetBuilder
        agentName={agentName}
        docsUrl={docsUrl}
        persist={false}
        embedNote={`Agent ID for your backend: ${agentId}`}
      />
    </>
  )
}
