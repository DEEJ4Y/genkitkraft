import { Box, Text } from '@mantine/core'
import { MarkdownContent } from './MarkdownContent'

interface ChatMessageProps {
  role: 'user' | 'assistant'
  content: string
}

export function ChatMessage({ role, content }: ChatMessageProps) {
  const isUser = role === 'user'

  return (
    <Box
      style={{
        display: 'flex',
        justifyContent: isUser ? 'flex-end' : 'flex-start',
        marginBottom: 8,
      }}
    >
      <Box
        style={{
          maxWidth: '75%',
          padding: '10px 14px',
          borderRadius: 12,
          backgroundColor: isUser
            ? 'var(--mantine-color-blue-6)'
            : 'var(--mantine-color-gray-1)',
          color: isUser ? 'white' : 'inherit',
          ...(isUser ? { whiteSpace: 'pre-wrap' as const } : {}),
          wordBreak: 'break-word',
        }}
      >
        <Text size="xs" fw={600} mb={2} c={isUser ? 'white' : 'dimmed'}>
          {isUser ? 'You' : 'Assistant'}
        </Text>
        {isUser ? (
          <Text size="sm" style={{ lineHeight: 1.5 }}>
            {content}
          </Text>
        ) : (
          <MarkdownContent content={content} />
        )}
      </Box>
    </Box>
  )
}
