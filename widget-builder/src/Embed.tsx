import { Alert, Button, Code, CopyButton, ScrollArea, Stack, Tabs, Text, TextInput } from '@mantine/core'
import { IconCheck, IconCopy } from '@tabler/icons-react'
import { useState } from 'react'
import { WIDGET_VERSION } from './constants'
import { backendExample, generateCdnSnippet, generateNpmSnippet, generateProviderSnippet, installCommand } from './snippets'
import type { BuilderConfig } from './types'

function CodeBlock({ code, label }: { code: string; label: string }) {
  return (
    <div style={{ position: 'relative' }}>
      <CopyButton value={code} timeout={2000}>
        {({ copied, copy }) => (
          <Button
            size="compact-xs"
            variant="light"
            onClick={copy}
            leftSection={copied ? <IconCheck size={12} /> : <IconCopy size={12} />}
            aria-label={`Copy ${label}`}
            style={{ position: 'absolute', top: 8, right: 8, zIndex: 1 }}
          >
            {copied ? 'Copied' : 'Copy'}
          </Button>
        )}
      </CopyButton>
      <ScrollArea.Autosize mah={420}>
        <Code block tabIndex={0} aria-label={label} style={{ paddingTop: 36 }}>
          {code}
        </Code>
      </ScrollArea.Autosize>
    </div>
  )
}

export interface EmbedProps {
  config: BuilderConfig
  /** Pre-filled backend base URL (e.g. "/api/chat"). */
  backendBaseUrl?: string
  /** Optional hint shown above the snippets (e.g. the agent this widget will talk to). */
  note?: string
  docsUrl?: string
}

export function Embed({ config, backendBaseUrl: initialBase = '/api/chat', note, docsUrl }: EmbedProps) {
  const [baseUrl, setBaseUrl] = useState(initialBase)
  const opts = { backendBaseUrl: baseUrl }

  return (
    <Stack>
      {note && <Text size="sm">{note}</Text>}
      <Alert color="yellow" variant="light" title="Do not call genkitkraft from the browser">
        The widget talks to <b>your backend</b>, which holds the genkitkraft API key and forwards messages. Never put{' '}
        <Code>PUBLIC_API_KEY</Code> in client code.{' '}
        {docsUrl && (
          <a href={docsUrl} target="_blank" rel="noreferrer">
            Read the integration guide
          </a>
        )}
      </Alert>
      <TextInput
        label="Your backend chat endpoint"
        description="Base URL the provider calls: /sessions, /sessions/:id/messages"
        value={baseUrl}
        onChange={(e) => setBaseUrl(e.currentTarget.value)}
      />
      <Tabs defaultValue="npm" keepMounted={false}>
        <Tabs.List>
          <Tabs.Tab value="npm">npm</Tabs.Tab>
          <Tabs.Tab value="cdn">CDN</Tabs.Tab>
          <Tabs.Tab value="provider">Provider</Tabs.Tab>
          <Tabs.Tab value="backend">Backend example</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="npm" pt="sm">
          <Stack gap="xs">
            <Text size="sm">1. Install (widget v{WIDGET_VERSION})</Text>
            <CodeBlock code={installCommand} label="install command" />
            <Text size="sm">
              2. Add the provider (see the Provider tab), then initialise the widget in client-side code:
            </Text>
            <CodeBlock code={generateNpmSnippet(config, opts)} label="npm snippet" />
          </Stack>
        </Tabs.Panel>
        <Tabs.Panel value="cdn" pt="sm">
          <Stack gap="xs">
            <Text size="sm">Paste before the closing &lt;/body&gt; tag:</Text>
            <CodeBlock code={generateCdnSnippet(config, opts)} label="CDN snippet" />
          </Stack>
        </Tabs.Panel>
        <Tabs.Panel value="provider" pt="sm">
          <Stack gap="xs">
            <Text size="sm">
              Save as <Code>GenkitkraftChatProvider.ts</Code>. It implements the widget&apos;s <Code>ChatProvider</Code> against your
              backend endpoints.
            </Text>
            <CodeBlock code={generateProviderSnippet()} label="provider source" />
          </Stack>
        </Tabs.Panel>
        <Tabs.Panel value="backend" pt="sm">
          <Stack gap="xs">
            <Text size="sm">
              Minimal Express backend implementing the four endpoints. Replace <Code>requireLogin</Code> and <Code>db</Code> with your
              auth and storage.
            </Text>
            <CodeBlock code={backendExample} label="backend example" />
          </Stack>
        </Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
