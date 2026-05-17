import { useState, useEffect } from 'react'
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Card,
  Center,
  Checkbox,
  Collapse,
  Group,
  Loader,
  Modal,
  Select,
  Stack,
  Text,
} from '@mantine/core'
import { IconPlus, IconTools, IconTrash } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { fetchClient } from '../../lib/api/client'
import type { components } from '../../lib/api/schema'

type AgentMcpServerToolConfig = components['schemas']['Models.AgentMcpServerToolConfig']
type HttpToolResponse = components['schemas']['Models.HttpToolResponse']
type McpServerResponse = components['schemas']['Models.McpServerResponse']
type McpServerToolResponse = components['schemas']['Models.McpServerToolResponse']
type BuiltInToolResponse = components['schemas']['Models.BuiltInToolResponse']

export interface PlaygroundToolConfig {
  httpToolIds: string[]
  mcpServers: AgentMcpServerToolConfig[]
  builtInToolIds: string[]
}

interface PlaygroundToolsPanelProps {
  agentId: string
  toolConfig: PlaygroundToolConfig
  onChange: (config: PlaygroundToolConfig) => void
  hasOverrides: boolean
}

export function PlaygroundToolsPanel({
  agentId,
  toolConfig,
  onChange,
  hasOverrides,
}: PlaygroundToolsPanelProps) {
  const [opened, setOpened] = useState(false)
  const [addHttpModalOpen, setAddHttpModalOpen] = useState(false)
  const [addMcpModalOpen, setAddMcpModalOpen] = useState(false)

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

  const builtInToolsQuery = useQuery({
    queryKey: ['get', '/api/v1/built-in-tools'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/v1/built-in-tools')
      if (error) throw new Error('Failed to fetch built-in tools')
      return data
    },
  })

  const allHttpTools: HttpToolResponse[] = httpToolsQuery.data?.httpTools ?? []
  const allMcpServers: McpServerResponse[] = mcpServersQuery.data?.mcpServers ?? []
  const allBuiltInTools: BuiltInToolResponse[] = builtInToolsQuery.data?.builtInTools ?? []

  const availableHttpTools = allHttpTools.filter((t) => !toolConfig.httpToolIds.includes(t.id))
  const availableMcpServers = allMcpServers.filter(
    (s) => !toolConfig.mcpServers.find((ms) => ms.mcpServerId === s.id)
  )

  function addHttpTool(toolId: string) {
    if (!toolConfig.httpToolIds.includes(toolId)) {
      onChange({ ...toolConfig, httpToolIds: [...toolConfig.httpToolIds, toolId] })
    }
    setAddHttpModalOpen(false)
  }

  function removeHttpTool(toolId: string) {
    onChange({ ...toolConfig, httpToolIds: toolConfig.httpToolIds.filter((id) => id !== toolId) })
  }

  function addMcpServer(serverId: string) {
    if (!toolConfig.mcpServers.find((s) => s.mcpServerId === serverId)) {
      onChange({
        ...toolConfig,
        mcpServers: [...toolConfig.mcpServers, { mcpServerId: serverId, selectAll: true, toolNames: [] }],
      })
    }
    setAddMcpModalOpen(false)
  }

  function removeMcpServer(serverId: string) {
    onChange({
      ...toolConfig,
      mcpServers: toolConfig.mcpServers.filter((s) => s.mcpServerId !== serverId),
    })
  }

  function updateMcpServerSelectAll(serverId: string, selectAll: boolean) {
    onChange({
      ...toolConfig,
      mcpServers: toolConfig.mcpServers.map((s) =>
        s.mcpServerId === serverId ? { ...s, selectAll, toolNames: selectAll ? [] : s.toolNames } : s
      ),
    })
  }

  function toggleMcpServerTool(serverId: string, toolName: string, checked: boolean) {
    onChange({
      ...toolConfig,
      mcpServers: toolConfig.mcpServers.map((s) => {
        if (s.mcpServerId !== serverId) return s
        const toolNames = checked
          ? [...s.toolNames, toolName]
          : s.toolNames.filter((n) => n !== toolName)
        return { ...s, toolNames }
      }),
    })
  }

  function toggleBuiltInTool(toolId: string, checked: boolean) {
    if (checked) {
      onChange({ ...toolConfig, builtInToolIds: [...toolConfig.builtInToolIds, toolId] })
    } else {
      onChange({ ...toolConfig, builtInToolIds: toolConfig.builtInToolIds.filter((id) => id !== toolId) })
    }
  }

  const toolCount = toolConfig.httpToolIds.length + toolConfig.mcpServers.length + toolConfig.builtInToolIds.length

  return (
    <Box style={{ borderBottom: '1px solid var(--mantine-color-gray-3)' }}>
      <Group justify="space-between" p="xs" px="md">
        <Group gap="xs">
          <ActionIcon
            variant={opened ? 'filled' : 'subtle'}
            size="sm"
            onClick={() => setOpened(!opened)}
          >
            <IconTools size={16} />
          </ActionIcon>
          <Text size="xs" c="dimmed">
            {toolCount > 0 ? `${toolCount} tool source${toolCount !== 1 ? 's' : ''}` : 'No tools'}
            {hasOverrides && ' (modified)'}
          </Text>
        </Group>
      </Group>

      <Collapse in={opened}>
        <Box p="md" pt={0}>
          <Stack gap="sm">
            {/* Built-in Tools */}
            <Text size="xs" fw={600}>Built-in Tools</Text>
            {allBuiltInTools.length === 0 ? (
              <Text size="xs" c="dimmed">None available</Text>
            ) : (
              allBuiltInTools.map((tool) => (
                <Checkbox
                  key={tool.id}
                  size="xs"
                  label={tool.name}
                  description={tool.description}
                  checked={toolConfig.builtInToolIds.includes(tool.id)}
                  onChange={(e) => toggleBuiltInTool(tool.id, e.currentTarget.checked)}
                />
              ))
            )}

            {/* HTTP Tools */}
            <Group justify="space-between" align="center" mt="xs">
              <Text size="xs" fw={600}>HTTP Tools</Text>
              <Button
                size="compact-xs"
                variant="light"
                leftSection={<IconPlus size={12} />}
                onClick={() => setAddHttpModalOpen(true)}
                disabled={availableHttpTools.length === 0}
              >
                Add
              </Button>
            </Group>

            {toolConfig.httpToolIds.length === 0 && (
              <Text size="xs" c="dimmed">None</Text>
            )}

            {toolConfig.httpToolIds.map((toolId) => {
              const tool = allHttpTools.find((t) => t.id === toolId)
              return (
                <Group key={toolId} gap="xs" justify="space-between">
                  <Group gap="xs">
                    <Badge size="xs" variant="light">{tool?.method ?? '?'}</Badge>
                    <Text size="xs">{tool?.name ?? toolId}</Text>
                  </Group>
                  <ActionIcon variant="subtle" color="red" size="xs" onClick={() => removeHttpTool(toolId)}>
                    <IconTrash size={12} />
                  </ActionIcon>
                </Group>
              )
            })}

            {/* MCP Servers */}
            <Group justify="space-between" align="center" mt="xs">
              <Text size="xs" fw={600}>MCP Servers</Text>
              <Button
                size="compact-xs"
                variant="light"
                leftSection={<IconPlus size={12} />}
                onClick={() => setAddMcpModalOpen(true)}
                disabled={availableMcpServers.length === 0}
              >
                Add
              </Button>
            </Group>

            {toolConfig.mcpServers.length === 0 && (
              <Text size="xs" c="dimmed">None</Text>
            )}

            {toolConfig.mcpServers.map((mc) => (
              <McpToolSection
                key={mc.mcpServerId}
                config={mc}
                server={allMcpServers.find((s) => s.id === mc.mcpServerId)}
                onRemove={() => removeMcpServer(mc.mcpServerId)}
                onToggleSelectAll={(val) => updateMcpServerSelectAll(mc.mcpServerId, val)}
                onToggleTool={(name, checked) => toggleMcpServerTool(mc.mcpServerId, name, checked)}
              />
            ))}
          </Stack>
        </Box>
      </Collapse>

      <Modal opened={addHttpModalOpen} onClose={() => setAddHttpModalOpen(false)} title="Add HTTP Tool" size="sm">
        <Select
          label="Select an HTTP tool"
          placeholder="Choose a tool"
          data={availableHttpTools.map((t) => ({ value: t.id, label: `${t.method} ${t.name}` }))}
          onChange={(val) => val && addHttpTool(val)}
          searchable
        />
      </Modal>

      <Modal opened={addMcpModalOpen} onClose={() => setAddMcpModalOpen(false)} title="Add MCP Server" size="sm">
        <Select
          label="Select an MCP server"
          placeholder="Choose a server"
          data={availableMcpServers.map((s) => ({ value: s.id, label: s.name }))}
          onChange={(val) => val && addMcpServer(val)}
          searchable
        />
      </Modal>
    </Box>
  )
}

interface McpToolSectionProps {
  config: AgentMcpServerToolConfig
  server?: McpServerResponse
  onRemove: () => void
  onToggleSelectAll: (selectAll: boolean) => void
  onToggleTool: (toolName: string, checked: boolean) => void
}

function McpToolSection({ config, server, onRemove, onToggleSelectAll, onToggleTool }: McpToolSectionProps) {
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
    <Card withBorder padding="xs">
      <Group justify="space-between" mb={4}>
        <Group gap="xs">
          <Text size="xs" fw={600}>{server?.name ?? config.mcpServerId}</Text>
          {server && <Badge size="xs" variant="light" color="teal">{server.transport}</Badge>}
        </Group>
        <ActionIcon variant="subtle" color="red" size="xs" onClick={onRemove}>
          <IconTrash size={12} />
        </ActionIcon>
      </Group>
      <Checkbox
        size="xs"
        label="Select all tools"
        checked={config.selectAll}
        onChange={(e) => onToggleSelectAll(e.currentTarget.checked)}
        mb={4}
      />
      {!config.selectAll && (
        <Stack gap={2}>
          {toolsQuery.isPending && <Center py="xs"><Loader size="xs" /></Center>}
          {tools.map((tool) => (
            <Checkbox
              key={tool.name}
              size="xs"
              label={<Text size="xs">{tool.name}</Text>}
              checked={config.toolNames.includes(tool.name)}
              onChange={(e) => onToggleTool(tool.name, e.currentTarget.checked)}
            />
          ))}
        </Stack>
      )}
    </Card>
  )
}
