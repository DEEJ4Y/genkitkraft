import { Button, Stack, Tabs, Text } from '@mantine/core'
import { IconArrowLeft } from '@tabler/icons-react'
import type { components } from '../lib/api/schema'
import { AgentDeployInfo } from './AgentDeployInfo'
import { AgentForm } from './AgentForm'
import { AgentToolsTab } from './AgentToolsTab'
import { AgentPlayground } from './playground/AgentPlayground'

type AgentResponse = components['schemas']['Models.AgentResponse']

interface AgentEditViewProps {
  agent: AgentResponse
  onSaved: () => void
  onCancel: () => void
}

export function AgentEditView({ agent, onSaved, onCancel }: AgentEditViewProps) {
  return (
    <Stack>
      <Button
        variant="subtle"
        leftSection={<IconArrowLeft size={16} />}
        onClick={onCancel}
        size="sm"
        w="fit-content"
      >
        Back to Agents
      </Button>

      <Text size="xl" fw={600}>
        {agent.name}
      </Text>

      <Tabs defaultValue="config">
        <Tabs.List mb="md">
          <Tabs.Tab value="config">Configuration</Tabs.Tab>
          <Tabs.Tab value="tools">Tools</Tabs.Tab>
          <Tabs.Tab value="playground">Playground</Tabs.Tab>
          <Tabs.Tab value="deploy">Deploy</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="config">
          <AgentForm agent={agent} onSaved={onSaved} onCancel={onCancel} />
        </Tabs.Panel>

        <Tabs.Panel value="tools">
          <AgentToolsTab agentId={agent.id} />
        </Tabs.Panel>

        <Tabs.Panel value="playground">
          <AgentPlayground agent={agent} />
        </Tabs.Panel>

        <Tabs.Panel value="deploy">
          <AgentDeployInfo agentId={agent.id} />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
