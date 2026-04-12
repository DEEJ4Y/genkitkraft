import React, { useMemo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Box, Code, Table, Anchor } from '@mantine/core'
import type { Components } from 'react-markdown'
import styles from './MarkdownContent.module.css'

interface MarkdownContentProps {
  content: string
}

const components: Components = {
  code({ className, children, ...props }) {
    const match = /language-(\w+)/.exec(className || '')
    const codeString = String(children).replace(/\n$/, '')

    if (match) {
      return (
        <Box
          style={{
            borderRadius: 8,
            overflow: 'hidden',
            margin: '8px 0',
          }}
        >
          <SyntaxHighlighter
            style={oneLight}
            language={match[1]}
            PreTag="div"
            customStyle={{
              margin: 0,
              borderRadius: 8,
              fontSize: '0.8rem',
            }}
          >
            {codeString}
          </SyntaxHighlighter>
        </Box>
      )
    }

    return (
      <Code {...props} className={className}>
        {children}
      </Code>
    )
  },

  a({ href, children }) {
    return (
      <Anchor
        href={href}
        target="_blank"
        rel="noopener noreferrer"
      >
        {children}
      </Anchor>
    )
  },

  table({ children }) {
    return (
      <Box style={{ overflowX: 'auto', margin: '8px 0' }}>
        <Table striped highlightOnHover withTableBorder withColumnBorders>
          {children}
        </Table>
      </Box>
    )
  },

  thead({ children }) {
    return <Table.Thead>{children}</Table.Thead>
  },

  tbody({ children }) {
    return <Table.Tbody>{children}</Table.Tbody>
  },

  tr({ children }) {
    return <Table.Tr>{children}</Table.Tr>
  },

  th({ children }) {
    return <Table.Th>{children}</Table.Th>
  },

  td({ children }) {
    return <Table.Td>{children}</Table.Td>
  },

  blockquote({ children }) {
    return (
      <Box
        component="blockquote"
        style={{
          margin: '8px 0',
          padding: '4px 12px',
          borderLeft: '3px solid var(--mantine-color-gray-4)',
          color: 'var(--mantine-color-gray-7)',
        }}
      >
        {children}
      </Box>
    )
  },

  img({ src, alt }) {
    return (
      <img
        src={src}
        alt={alt}
        style={{ maxWidth: '100%', borderRadius: 6 }}
      />
    )
  },
}

const remarkPlugins = [remarkGfm]

function MarkdownContentInner({ content }: MarkdownContentProps) {
  const memoizedContent = useMemo(
    () => (
      <ReactMarkdown remarkPlugins={remarkPlugins} components={components}>
        {content}
      </ReactMarkdown>
    ),
    [content]
  )

  return (
    <Box className={styles.markdownContent} style={{ fontSize: '0.875rem', lineHeight: 1.6 }}>
      {memoizedContent}
    </Box>
  )
}

export const MarkdownContent = React.memo(MarkdownContentInner)
