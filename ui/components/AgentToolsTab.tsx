import { useState, useEffect } from 'react'
import {
  Button,
  Stack,
  Group,
  Text,
  Alert,
  Loader,
  Center,
  Card,
  Checkbox,
  ActionIcon,
  Modal,
  Select,
  Badge,
} from '@mantine/core'
import { IconPlus, IconTrash } from '@tabler/icons-react'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type AgentToolConfigResponse = components['schemas']['Models.AgentToolConfigResponse']
type AgentMcpServerToolConfig = components['schemas']['Models.AgentMcpServerToolConfig']
type HttpToolResponse = components['schemas']['Models.HttpToolResponse']
type McpServerResponse = components['schemas']['Models.McpServerResponse']
type McpServerToolResponse = components['schemas']['Models.McpServerToolResponse']

interface AgentToolsTabProps {
  agentId: string
}

export function AgentToolsTab({ agentId }: AgentToolsTabProps) {
  const queryClient = useQueryClient()
  const [httpToolIds, setHttpToolIds] = useState<string[]>([])
  const [mcpServers, setMcpServers] = useState<AgentMcpServerToolConfig[]>([])
  const [dirty, setDirty] = useState(false)
  const [addHttpModalOpen, setAddHttpModalOpen] = useState(false)
  const [addMcpModalOpen, setAddMcpModalOpen] = useState(false)

  // Load current agent tool config
  const configQuery = useQuery({
    queryKey: ['get', `/api/v1/agents/${agentId}/tools`],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/agents/{agentId}/tools', {
        params: { path: { agentId } },
      })
      if (error) throw new Error('Failed to fetch agent tools')
      return data as AgentToolConfigResponse
    },
  })

  // Load all HTTP tools for the picker
  const httpToolsQuery = useQuery({
    queryKey: ['get', '/api/v1/http-tools', { limit: 100, offset: 0 }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/http-tools', {
        params: { query: { limit: 100, offset: 0 } },
      })
      if (error) throw new Error('Failed to fetch HTTP tools')
      return data
    },
  })

  // Load all MCP servers for the picker
  const mcpServersQuery = useQuery({
    queryKey: ['get', '/api/v1/mcp-servers', { limit: 100, offset: 0 }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/mcp-servers', {
        params: { query: { limit: 100, offset: 0 } },
      })
      if (error) throw new Error('Failed to fetch MCP servers')
      return data
    },
  })

  // Sync state from server
  useEffect(() => {
    if (configQuery.data) {
      setHttpToolIds(configQuery.data.httpToolIds ?? [])
      setMcpServers(configQuery.data.mcpServers ?? [])
      setDirty(false)
    }
  }, [configQuery.data])

  const saveMutation = useMutation({
    mutationFn: async () => {
      const { error } = await fetchClient.PUT('/api/v1/agents/{agentId}/tools', {
        params: { path: { agentId } },
        body: { httpToolIds, mcpServers },
      })
      if (error) throw new Error('Failed to save agent tools')
    },
    onSuccess: () => {
      setDirty(false)
      queryClient.invalidateQueries({ queryKey: ['get', `/api/v1/agents/${agentId}/tools`] })
    },
  })

  function addHttpTool(toolId: string) {
    if (!httpToolIds.includes(toolId)) {
      setHttpToolIds([...httpToolIds, toolId])
      setDirty(true)
    }
    setAddHttpModalOpen(false)
  }

  function removeHttpTool(toolId: string) {
    setHttpToolIds(httpToolIds.filter((id) => id !== toolId))
    setDirty(true)
  }

  function addMcpServer(serverId: string) {
    if (!mcpServers.find((s) => s.mcpServerId === serverId)) {
      setMcpServers([...mcpServers, { mcpServerId: serverId, selectAll: true, toolNames: [] }])
      setDirty(true)
    }
    setAddMcpModalOpen(false)
  }

  function removeMcpServer(serverId: string) {
    setMcpServers(mcpServers.filter((s) => s.mcpServerId !== serverId))
    setDirty(true)
  }

  function updateMcpServerSelectAll(serverId: string, selectAll: boolean) {
    setMcpServers(
      mcpServers.map((s) =>
        s.mcpServerId === serverId ? { ...s, selectAll, toolNames: selectAll ? [] : s.toolNames } : s
      )
    )
    setDirty(true)
  }

  function toggleMcpServerTool(serverId: string, toolName: string, checked: boolean) {
    setMcpServers(
      mcpServers.map((s) => {
        if (s.mcpServerId !== serverId) return s
        const toolNames = checked
          ? [...s.toolNames, toolName]
          : s.toolNames.filter((n) => n !== toolName)
        return { ...s, toolNames }
      })
    )
    setDirty(true)
  }

  if (configQuery.isPending) {
    return (
      <Center py="xl">
        <Loader />
      </Center>
    )
  }

  if (configQuery.error) {
    return (
      <Alert color="red" variant="light">
        Failed to load agent tool configuration.
      </Alert>
    )
  }

  const allHttpTools: HttpToolResponse[] = httpToolsQuery.data?.httpTools ?? []
  const allMcpServers: McpServerResponse[] = mcpServersQuery.data?.mcpServers ?? []

  // Filter out already-added items for pickers
  const availableHttpTools = allHttpTools.filter((t) => !httpToolIds.includes(t.id))
  const availableMcpServers = allMcpServers.filter(
    (s) => !mcpServers.find((ms) => ms.mcpServerId === s.id)
  )

  return (
    <Stack>
      {saveMutation.error && (
        <Alert color="red" variant="light">
          {(saveMutation.error as Error).message}
        </Alert>
      )}

      {/* HTTP Tools Section */}
      <Group justify="space-between" align="center">
        <Text fw={600} size="md">
          HTTP Tools
        </Text>
        <Button
          size="xs"
          variant="light"
          leftSection={<IconPlus size={14} />}
          onClick={() => setAddHttpModalOpen(true)}
          disabled={availableHttpTools.length === 0}
        >
          Add HTTP Tool
        </Button>
      </Group>

      {httpToolIds.length === 0 ? (
        <Text size="sm" c="dimmed">
          No HTTP tools assigned.
        </Text>
      ) : (
        <Stack gap="xs">
          {httpToolIds.map((toolId) => {
            const tool = allHttpTools.find((t) => t.id === toolId)
            return (
              <Card key={toolId} padding="xs" withBorder>
                <Group justify="space-between" wrap="nowrap">
                  <Group gap="sm">
                    <Badge size="sm" variant="light">
                      {tool?.method ?? '?'}
                    </Badge>
                    <Text size="sm">{tool?.name ?? toolId}</Text>
                  </Group>
                  <ActionIcon variant="subtle" color="red" size="sm" onClick={() => removeHttpTool(toolId)}>
                    <IconTrash size={14} />
                  </ActionIcon>
                </Group>
              </Card>
            )
          })}
        </Stack>
      )}

      {/* MCP Servers Section */}
      <Group justify="space-between" align="center" mt="md">
        <Text fw={600} size="md">
          MCP Servers
        </Text>
        <Button
          size="xs"
          variant="light"
          leftSection={<IconPlus size={14} />}
          onClick={() => setAddMcpModalOpen(true)}
          disabled={availableMcpServers.length === 0}
        >
          Add MCP Server
        </Button>
      </Group>

      {mcpServers.length === 0 ? (
        <Text size="sm" c="dimmed">
          No MCP servers assigned.
        </Text>
      ) : (
        <Stack gap="sm">
          {mcpServers.map((mc) => (
            <McpServerToolSection
              key={mc.mcpServerId}
              config={mc}
              server={allMcpServers.find((s) => s.id === mc.mcpServerId)}
              onRemove={() => removeMcpServer(mc.mcpServerId)}
              onToggleSelectAll={(val) => updateMcpServerSelectAll(mc.mcpServerId, val)}
              onToggleTool={(name, checked) => toggleMcpServerTool(mc.mcpServerId, name, checked)}
            />
          ))}
        </Stack>
      )}

      {/* Save Button */}
      <Group justify="flex-end" mt="md">
        <Button onClick={() => saveMutation.mutate()} loading={saveMutation.isPending} disabled={!dirty}>
          Save Tools
        </Button>
      </Group>

      {/* Add HTTP Tool Modal */}
      <Modal opened={addHttpModalOpen} onClose={() => setAddHttpModalOpen(false)} title="Add HTTP Tool">
        <Select
          label="Select an HTTP tool"
          placeholder="Choose a tool"
          data={availableHttpTools.map((t) => ({ value: t.id, label: `${t.method} ${t.name}` }))}
          onChange={(val) => val && addHttpTool(val)}
          searchable
        />
      </Modal>

      {/* Add MCP Server Modal */}
      <Modal opened={addMcpModalOpen} onClose={() => setAddMcpModalOpen(false)} title="Add MCP Server">
        <Select
          label="Select an MCP server"
          placeholder="Choose a server"
          data={availableMcpServers.map((s) => ({ value: s.id, label: s.name }))}
          onChange={(val) => val && addMcpServer(val)}
          searchable
        />
      </Modal>
    </Stack>
  )
}

// Sub-component for an MCP server's tool selection
interface McpServerToolSectionProps {
  config: AgentMcpServerToolConfig
  server?: McpServerResponse
  onRemove: () => void
  onToggleSelectAll: (selectAll: boolean) => void
  onToggleTool: (toolName: string, checked: boolean) => void
}

function McpServerToolSection({
  config,
  server,
  onRemove,
  onToggleSelectAll,
  onToggleTool,
}: McpServerToolSectionProps) {
  // Fetch tools from this MCP server
  const toolsQuery = useQuery({
    queryKey: ['get', `/api/v1/mcp-servers/${config.mcpServerId}/tools`],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/mcp-servers/{id}/tools', {
        params: { path: { id: config.mcpServerId } },
      })
      if (error) throw new Error('Failed to fetch MCP server tools')
      return data
    },
  })

  const tools: McpServerToolResponse[] = toolsQuery.data?.tools ?? []

  return (
    <Card withBorder padding="sm">
      <Group justify="space-between" mb="xs">
        <Group gap="sm">
          <Text fw={600} size="sm">
            {server?.name ?? config.mcpServerId}
          </Text>
          {server && (
            <Badge size="xs" variant="light" color="teal">
              {server.transport}
            </Badge>
          )}
        </Group>
        <ActionIcon variant="subtle" color="red" size="sm" onClick={onRemove}>
          <IconTrash size={14} />
        </ActionIcon>
      </Group>

      <Checkbox
        label="Select all tools"
        checked={config.selectAll}
        onChange={(e) => onToggleSelectAll(e.currentTarget.checked)}
        mb="xs"
      />

      {!config.selectAll && (
        <>
          {toolsQuery.isPending && (
            <Center py="xs">
              <Loader size="sm" />
            </Center>
          )}
          {toolsQuery.error && (
            <Text size="xs" c="red">
              Could not load tools from this server.
            </Text>
          )}
          {tools.length === 0 && !toolsQuery.isPending && (
            <Text size="xs" c="dimmed">
              No tools discovered from this server.
            </Text>
          )}
          <Stack gap={4}>
            {tools.map((tool) => (
              <Checkbox
                key={tool.name}
                label={
                  <Text size="xs">
                    <strong>{tool.name}</strong>
                    {tool.description ? ` — ${tool.description}` : ''}
                  </Text>
                }
                checked={config.toolNames.includes(tool.name)}
                onChange={(e) => onToggleTool(tool.name, e.currentTarget.checked)}
              />
            ))}
          </Stack>
        </>
      )}
    </Card>
  )
}
