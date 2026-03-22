import { useState } from 'react'
import { TextInput, Button, Group, Stack, Alert, Text } from '@mantine/core'
import { IconArrowLeft } from '@tabler/icons-react'
import { fetchClient } from '../lib/api/client'
import { PromptEditor } from './PromptEditor'
import type { components } from '../lib/api/schema'

type PromptResponse = components['schemas']['Models.PromptResponse']

interface PromptFormProps {
  prompt?: PromptResponse
  onSaved: () => void
  onCancel: () => void
}

export function PromptForm({ prompt, onSaved, onCancel }: PromptFormProps) {
  const isEdit = !!prompt
  const [name, setName] = useState(prompt?.name ?? '')
  const [content, setContent] = useState(prompt?.content ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit() {
    if (!name.trim()) {
      setError('Name is required')
      return
    }

    setSaving(true)
    setError('')

    try {
      if (isEdit) {
        const { error: err } = await fetchClient.PUT('/api/v1/prompts/{id}', {
          params: { path: { id: prompt.id } },
          body: { name: name.trim(), content } as any,
        })
        if (err) throw new Error((err as any).error)
      } else {
        const { error: err } = await fetchClient.POST('/api/v1/prompts', {
          body: { name: name.trim(), content } as any,
        })
        if (err) throw new Error((err as any).error)
      }
      onSaved()
    } catch (err: any) {
      setError(err.message || 'Failed to save prompt')
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
        Back to Prompts
      </Button>

      <Text size="xl" fw={600}>
        {isEdit ? 'Edit Prompt' : 'New Prompt'}
      </Text>

      {error && (
        <Alert color="red" variant="light">
          {error}
        </Alert>
      )}

      <TextInput
        label="Name"
        placeholder="e.g. Code Review Assistant"
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
        required
      />

      <div>
        <Text size="sm" fw={500} mb={4}>
          Content
        </Text>
        <PromptEditor content={content} onChange={setContent} />
      </div>

      <Group>
        <Button onClick={handleSubmit} loading={saving}>
          {isEdit ? 'Save Changes' : 'Create Prompt'}
        </Button>
        <Button variant="default" onClick={onCancel}>
          Cancel
        </Button>
      </Group>
    </Stack>
  )
}
