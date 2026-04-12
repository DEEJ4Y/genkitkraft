import { useEffect, useRef, useState } from 'react'
import { ActionIcon, Box, ScrollArea, Stack, Text, Textarea } from '@mantine/core'
import { IconPlayerStop, IconSend } from '@tabler/icons-react'
import { ChatMessage } from './ChatMessage'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
}

interface ChatViewProps {
  messages: Message[]
  streamingContent: string
  isStreaming: boolean
  error: string | null
  onSend: (content: string) => void
  onStop: () => void
  hasSession: boolean
}

export function ChatView({ messages, streamingContent, isStreaming, error, onSend, onStop, hasSession }: ChatViewProps) {
  const [input, setInput] = useState('')
  const viewportRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom on new messages or streaming
  useEffect(() => {
    if (viewportRef.current) {
      viewportRef.current.scrollTo({ top: viewportRef.current.scrollHeight, behavior: 'smooth' })
    }
  }, [messages, streamingContent])

  function handleSend() {
    const trimmed = input.trim()
    if (!trimmed || isStreaming) return
    onSend(trimmed)
    setInput('')
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  if (!hasSession) {
    return (
      <Box
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Text c="dimmed" size="sm">
          Create or select a session to start chatting.
        </Text>
      </Box>
    )
  }

  return (
    <Stack gap={0} style={{ flex: 1, minHeight: 0 }}>
      <ScrollArea style={{ flex: 1, minHeight: 0 }} p="md" viewportRef={viewportRef}>
        {messages.length === 0 && !isStreaming && (
          <Text c="dimmed" size="sm" ta="center" py="xl">
            Send a message to start the conversation.
          </Text>
        )}
        {messages.map((msg) => (
          <ChatMessage key={msg.id} role={msg.role} content={msg.content} />
        ))}
        {isStreaming && streamingContent && (
          <ChatMessage role="assistant" content={streamingContent} />
        )}
        {isStreaming && !streamingContent && (
          <Box style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: 8 }}>
            <Box
              style={{
                padding: '10px 14px',
                borderRadius: 12,
                backgroundColor: 'var(--mantine-color-gray-1)',
              }}
            >
              <Text size="sm" c="dimmed">Thinking...</Text>
            </Box>
          </Box>
        )}
        {error && (
          <Text c="red" size="sm" ta="center" py="sm">
            {error}
          </Text>
        )}
      </ScrollArea>

      <Box p="md" style={{ borderTop: '1px solid var(--mantine-color-gray-3)' }}>
        <Box style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
          <Textarea
            placeholder="Type a message... (Shift+Enter for new line)"
            value={input}
            onChange={(e) => setInput(e.currentTarget.value)}
            onKeyDown={handleKeyDown}
            autosize
            minRows={1}
            maxRows={5}
            style={{ flex: 1 }}
            disabled={isStreaming}
          />
          {isStreaming ? (
            <ActionIcon variant="filled" color="red" size="lg" onClick={onStop}>
              <IconPlayerStop size={18} />
            </ActionIcon>
          ) : (
            <ActionIcon
              variant="filled"
              size="lg"
              onClick={handleSend}
              disabled={!input.trim()}
            >
              <IconSend size={18} />
            </ActionIcon>
          )}
        </Box>
      </Box>
    </Stack>
  )
}
