import {
  Stack,
  Text,
  Paper,
  Code,
  CopyButton,
  ActionIcon,
  Tooltip,
  Anchor,
  Title,
} from '@mantine/core'
import { IconCopy, IconCheck, IconExternalLink } from '@tabler/icons-react'
import { DOCS_BASE_URL } from '../lib/constants'

interface AgentDeployInfoProps {
  agentId: string
}

function CopyableField({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <Text size="sm" fw={500} mb={4}>
        {label}
      </Text>
      <Paper
        withBorder
        p="xs"
        radius="sm"
        style={{ display: 'flex', alignItems: 'center', gap: 8 }}
      >
        <Code
          style={{
            flex: 1,
            backgroundColor: 'transparent',
            fontSize: '0.85rem',
            wordBreak: 'break-all',
          }}
        >
          {value}
        </Code>
        <CopyButton value={value}>
          {({ copied, copy }) => (
            <Tooltip label={copied ? 'Copied' : 'Copy'} withArrow>
              <ActionIcon
                variant="subtle"
                color={copied ? 'teal' : 'gray'}
                onClick={copy}
                size="sm"
              >
                {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
              </ActionIcon>
            </Tooltip>
          )}
        </CopyButton>
      </Paper>
    </div>
  )
}

export function AgentDeployInfo({ agentId }: AgentDeployInfoProps) {
  const baseUrl = typeof window !== 'undefined' ? window.location.origin : ''
  const endpoint = `${baseUrl}/api/v1/agents/${agentId}/deploy/chat/completions`

  const curlExample = `curl -X POST ${endpoint} \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }'`

  return (
    <Stack gap="lg">
      <div>
        <Title order={4} mb="xs">
          Deploy Your Agent
        </Title>
        <Text size="sm" c="dimmed">
          Use the OpenAI-compatible chat completions API to integrate this agent
          into your applications.
        </Text>
      </div>

      <CopyableField label="Agent ID" value={agentId} />

      <CopyableField label="Endpoint URL" value={endpoint} />

      <div>
        <Text size="sm" fw={500} mb={4}>
          Example Request
        </Text>
        <Paper withBorder p="sm" radius="sm" pos="relative">
          <CopyButton value={curlExample}>
            {({ copied, copy }) => (
              <Tooltip label={copied ? 'Copied' : 'Copy'} withArrow>
                <ActionIcon
                  variant="subtle"
                  color={copied ? 'teal' : 'gray'}
                  onClick={copy}
                  size="sm"
                  pos="absolute"
                  top={8}
                  right={8}
                >
                  {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                </ActionIcon>
              </Tooltip>
            )}
          </CopyButton>
          <Code block style={{ fontSize: '0.8rem', whiteSpace: 'pre-wrap' }}>
            {curlExample}
          </Code>
        </Paper>
      </div>

      <Paper withBorder p="sm" radius="sm" bg="var(--mantine-color-blue-light)">
        <Text size="sm">
          Set the <Code>PUBLIC_API_KEY</Code> environment variable to enable API
          key authentication. Without it, the deploy endpoint is publicly
          accessible.
        </Text>
      </Paper>

      <Anchor
        href={`${DOCS_BASE_URL}/docs/api/deploy`}
        target="_blank"
        size="sm"
        style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}
      >
        <IconExternalLink size={14} />
        View full deploy documentation
      </Anchor>
    </Stack>
  )
}
