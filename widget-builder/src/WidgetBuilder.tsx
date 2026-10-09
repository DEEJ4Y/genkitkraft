import { Accordion, Button, FileButton, Group, Menu, Stack, Text } from '@mantine/core'
import { IconChevronDown, IconDownload, IconLink, IconRestore, IconUpload } from '@tabler/icons-react'
import { useState } from 'react'
import { Embed } from './Embed'
import { BuilderProvider } from './fields'
import { Preview } from './Preview'
import { applyPreset, presets } from './presets'
import {
  AdvancedSection,
  BehaviourSection,
  BrandingSection,
  ChatSection,
  ColorsSection,
  FooterSection,
  HomeSection,
  SessionsSection,
} from './sections'
import { parseConfigJson, serializeConfig, shareUrl, useWidgetConfig, defaultConfig } from './state'
import type { BuilderConfig } from './types'

export interface WidgetBuilderProps {
  /** Agent name used to pre-fill the header title. */
  agentName?: string
  /** Pre-filled backend base URL for the generated snippets. */
  backendBaseUrl?: string
  /** Link to the integration guide. */
  docsUrl?: string
  /** Text shown above the snippets. */
  embedNote?: string
  /** Persist the config in localStorage and honour `#config=` share links. Default true. */
  persist?: boolean
  /** Render the live preview (injects the widget into the page). Default true. */
  preview?: boolean
}

const SECTIONS = [
  { value: 'branding', label: 'Branding', Component: BrandingSection },
  { value: 'colors', label: 'Colors', Component: ColorsSection },
  { value: 'home', label: 'Home screen', Component: HomeSection },
  { value: 'sessions', label: 'Sessions list', Component: SessionsSection },
  { value: 'chat', label: 'Chat screen', Component: ChatSection },
  { value: 'footer', label: 'Footer tabs', Component: FooterSection },
  { value: 'behaviour', label: 'Behaviour', Component: BehaviourSection },
  { value: 'advanced', label: 'Advanced', Component: AdvancedSection },
]

export function WidgetBuilder({ agentName, backendBaseUrl, docsUrl, embedNote, persist = true, preview = true }: WidgetBuilderProps) {
  const initial: BuilderConfig = agentName
    ? { ...defaultConfig, chatWindow: { ...defaultConfig.chatWindow, header: { title: { title: agentName, showOnlineSubtitle: true } } } }
    : defaultConfig
  const builder = useWidgetConfig(initial, persist)
  const [status, setStatus] = useState('')
  const [importError, setImportError] = useState<string | null>(null)

  const exportJson = () => {
    const url = URL.createObjectURL(new Blob([serializeConfig(builder.config)], { type: 'application/json' }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'chat-widget-config.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  const copyShareLink = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl(builder.config))
      setStatus('Share link copied')
    } catch {
      setStatus('Could not copy the share link')
    }
  }

  return (
    <BuilderProvider value={builder}>
      {preview && <Preview config={builder.config} />}
      <Stack gap="lg">
        <Group justify="space-between" align="center">
          <Group gap="xs">
            <Menu withinPortal>
              <Menu.Target>
                <Button variant="default" rightSection={<IconChevronDown size={14} />}>
                  Presets
                </Button>
              </Menu.Target>
              <Menu.Dropdown>
                {presets.map((p) => (
                  <Menu.Item key={p.id} onClick={() => builder.replace(applyPreset(p, builder.config))}>
                    {p.label}
                  </Menu.Item>
                ))}
              </Menu.Dropdown>
            </Menu>
            <Button variant="default" leftSection={<IconRestore size={14} />} onClick={builder.reset}>
              Reset
            </Button>
          </Group>
          <Group gap="xs">
            <FileButton
              accept="application/json"
              onChange={async (file) => {
                if (!file) return
                try {
                  builder.replace(parseConfigJson(await file.text()))
                  setImportError(null)
                  setStatus('Config imported')
                } catch (e) {
                  setImportError(e instanceof Error ? e.message : 'Invalid config file')
                }
              }}
            >
              {(props) => (
                <Button {...props} variant="default" leftSection={<IconUpload size={14} />}>
                  Import
                </Button>
              )}
            </FileButton>
            <Button variant="default" leftSection={<IconDownload size={14} />} onClick={exportJson}>
              Export
            </Button>
            {persist && (
              <Button variant="default" leftSection={<IconLink size={14} />} onClick={copyShareLink}>
                Share link
              </Button>
            )}
          </Group>
        </Group>
        <Text size="sm" c={importError ? 'red' : 'dimmed'} role="status" aria-live="polite" mih={20}>
          {importError ?? status}
        </Text>

        <Accordion multiple defaultValue={['branding', 'colors']} variant="separated">
          {SECTIONS.map(({ value, label, Component }) => (
            <Accordion.Item key={value} value={value}>
              <Accordion.Control>{label}</Accordion.Control>
              <Accordion.Panel>
                <Component />
              </Accordion.Panel>
            </Accordion.Item>
          ))}
        </Accordion>

        <Text fw={600} size="lg">
          Embed code
        </Text>
        <Embed config={builder.config} backendBaseUrl={backendBaseUrl} note={embedNote} docsUrl={docsUrl} />
      </Stack>
    </BuilderProvider>
  )
}
