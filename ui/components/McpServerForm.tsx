import { useState } from 'react'
import {
  TextInput,
  Select,
  Button,
  Group,
  Stack,
  Alert,
  Text,
  ActionIcon,
} from '@mantine/core'
import { IconArrowLeft, IconPlus, IconTrash } from '@tabler/icons-react'
import { fetchClient } from '../lib/api/client'
import type { components } from '../lib/api/schema'

type McpServerResponse = components['schemas']['Models.McpServerResponse']
type McpTransport = components['schemas']['Models.McpTransport']

interface McpServerFormProps {
  server?: McpServerResponse
  onSaved: () => void
  onCancel: () => void
}

interface HeaderRow {
  name: string
  value: string
}

const TRANSPORT_OPTIONS = [
  { value: 'sse', label: 'SSE (Server-Sent Events)' },
  { value: 'streamableHttp', label: 'Streamable HTTP' },
]

export function McpServerForm({ server, onSaved, onCancel }: McpServerFormProps) {
  const isEdit = !!server
  const [name, setName] = useState(server?.name ?? '')
  const [transport, setTransport] = useState<McpTransport>(server?.transport ?? 'sse')
  const [url, setUrl] = useState(server?.url ?? '')
  const [headers, setHeaders] = useState<HeaderRow[]>(
    server?.headers?.map((h) => ({ name: h.name, value: h.value })) ?? []
  )
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [urlError, setUrlError] = useState('')

  function validateUrl(value: string): string {
    if (!value.trim()) return ''
    try {
      new URL(value)
      return ''
    } catch {
      return 'Invalid URL. Must include a scheme (e.g. https://)'
    }
  }

  function addHeader() {
    setHeaders([...headers, { name: '', value: '' }])
  }

  function removeHeader(index: number) {
    setHeaders(headers.filter((_, i) => i !== index))
  }

  function updateHeader(index: number, field: 'name' | 'value', value: string) {
    const updated = [...headers]
    updated[index] = { ...updated[index], [field]: value }
    setHeaders(updated)
  }

  async function handleSubmit() {
    if (!name.trim()) {
      setError('Name is required')
      return
    }
    if (!url.trim()) {
      setError('URL is required')
      return
    }
    const urlErr = validateUrl(url)
    if (urlErr) {
      setUrlError(urlErr)
      return
    }

    setSaving(true)
    setError('')

    const validHeaders = headers.filter((h) => h.name.trim())

    try {
      if (isEdit) {
        const { error: err } = await fetchClient.PUT('/api/v1/mcp-servers/{id}', {
          params: { path: { id: server.id } },
          body: {
            name: name.trim(),
            transport,
            url: url.trim(),
            headers: validHeaders,
          },
        })
        if (err) throw new Error((err as any).error)
      } else {
        const { error: err } = await fetchClient.POST('/api/v1/mcp-servers', {
          body: {
            name: name.trim(),
            transport,
            url: url.trim(),
            headers: validHeaders.length > 0 ? validHeaders : undefined,
          },
        })
        if (err) throw new Error((err as any).error)
      }
      onSaved()
    } catch (err: any) {
      setError(err.message || 'Failed to save MCP server')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Stack>
      <Button
        variant="subtle"
        leftSection={<IconArrowLeft size={16} />}
        onClick={onCancel}
        size="sm"
        w="fit-content"
      >
        Back to MCP Servers
      </Button>

      <Text size="xl" fw={600}>
        {isEdit ? 'Edit MCP Server' : 'New MCP Server'}
      </Text>

      {error && (
        <Alert color="red" variant="light">
          {error}
        </Alert>
      )}

      <TextInput
        label="Name"
        placeholder="My MCP Server"
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
        required
      />

      <Select
        label="Transport"
        data={TRANSPORT_OPTIONS}
        value={transport}
        onChange={(val) => setTransport((val as McpTransport) ?? 'sse')}
        required
      />

      <TextInput
        label="URL"
        placeholder={transport === 'sse' ? 'https://mcp.example.com/sse' : 'https://mcp.example.com/mcp'}
        value={url}
        onChange={(e) => {
          setUrl(e.currentTarget.value)
          setUrlError('')
        }}
        error={urlError}
        required
      />

      <div>
        <Group justify="space-between" mb="xs">
          <Text size="sm" fw={500}>
            Headers
          </Text>
          <Button
            variant="light"
            size="xs"
            leftSection={<IconPlus size={14} />}
            onClick={addHeader}
          >
            Add Header
          </Button>
        </Group>

        {headers.length === 0 && (
          <Text size="sm" c="dimmed">
            No custom headers. Click &quot;Add Header&quot; to add authentication or other headers.
          </Text>
        )}

        <Stack gap="xs">
          {headers.map((header, index) => (
            <Group key={index} gap="xs" align="flex-end">
              <TextInput
                placeholder="Header name"
                value={header.name}
                onChange={(e) => updateHeader(index, 'name', e.currentTarget.value)}
                style={{ flex: 1 }}
              />
              <TextInput
                placeholder="Header value"
                value={header.value}
                onChange={(e) => updateHeader(index, 'value', e.currentTarget.value)}
                style={{ flex: 2 }}
              />
              <ActionIcon color="red" variant="subtle" onClick={() => removeHeader(index)}>
                <IconTrash size={16} />
              </ActionIcon>
            </Group>
          ))}
        </Stack>
      </div>

      <Group justify="flex-end" mt="md">
        <Button variant="default" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={handleSubmit} loading={saving}>
          {isEdit ? 'Update' : 'Create'} MCP Server
        </Button>
      </Group>
    </Stack>
  )
}
