import { ActionIcon, Button, Card, ColorInput, Group, MultiSelect, Select, SegmentedControl, Stack, Text, TextInput } from '@mantine/core'
import { IconPlus, IconTrash } from '@tabler/icons-react'
import { NAMED_COLORS } from './colors'
import {
  ColorField,
  ContrastWarning,
  JsonField,
  SelectField,
  SwitchField,
  TextAreaField,
  TextField,
  useBuilder,
} from './fields'
import { getIn } from './state'

const COLOR_NAMES = Object.keys(NAMED_COLORS)

export function BrandingSection() {
  const { config, set } = useBuilder()
  const avatarUrl = getIn<string>(config, 'chatWindow.header.avatars.0.url') ?? ''
  const avatarName = getIn<string>(config, 'chatWindow.header.avatars.0.name') ?? ''
  const setAvatar = (field: 'url' | 'name', value: string) => {
    set(`chatWindow.header.avatars.0.${field}`, value)
    set(`homeScreenConfig.avatars.0.${field}`, value)
  }
  return (
    <Stack>
      <TextField label="Agent name" path="chatWindow.header.title.title" placeholder="Assistant" />
      <TextInput label="Agent avatar URL" value={avatarUrl} onChange={(e) => setAvatar('url', e.currentTarget.value)} />
      <TextInput label="Avatar tooltip name" value={avatarName} onChange={(e) => setAvatar('name', e.currentTarget.value)} />
      <TextField label="Logo URL" path="homeScreenConfig.logoUrl" description="Shown on the home screen" />
      <TextField label="Logo URL (dark mode)" path="homeScreenConfig.logoUrlDark" description="Falls back to the light logo" />
    </Stack>
  )
}

function PrimaryColor() {
  const { config, set } = useBuilder()
  const value = getIn<string>(config, 'chatWindow.defaults.primaryColor') ?? ''
  return (
    <Group grow align="flex-start">
      <Select
        label="Primary color"
        data={COLOR_NAMES}
        searchable
        clearable
        value={COLOR_NAMES.includes(value) ? value : null}
        onChange={(v) => set('chatWindow.defaults.primaryColor', v ?? undefined)}
      />
      <ColorInput
        label="Custom primary color"
        placeholder="Any hex color"
        value={COLOR_NAMES.includes(value) ? '' : value}
        onChange={(v) => set('chatWindow.defaults.primaryColor', v)}
      />
    </Group>
  )
}

function MessageColors({ title, base }: { title: string; base: 'userMessage' | 'assistantMessage' }) {
  return (
    <Stack gap="xs">
      <Text fw={500} size="sm">
        {title}
      </Text>
      {(['light', 'dark'] as const).map((mode) => {
        const p = `chatWindow.defaults.${base}.${mode}`
        return (
          <Stack gap="xs" key={mode}>
            <Group grow>
              <ColorField label={`${mode === 'light' ? 'Light' : 'Dark'} background`} path={`${p}.bg`} />
              <ColorField label={`${mode === 'light' ? 'Light' : 'Dark'} text`} path={`${p}.color`} />
            </Group>
            <ContrastWarning bgPath={`${p}.bg`} fgPath={`${p}.color`} />
          </Stack>
        )
      })}
    </Stack>
  )
}

export function ColorsSection() {
  const { config, set } = useBuilder()
  const scheme = getIn<string>(config, 'chatWindow.defaults.colorScheme') ?? 'light'
  return (
    <Stack>
      <PrimaryColor />
      <div>
        <Text size="sm" fw={500} mb={4}>
          Default color scheme
        </Text>
        <SegmentedControl
          value={scheme}
          onChange={(v) => set('chatWindow.defaults.colorScheme', v)}
          data={[
            { value: 'light', label: 'Light' },
            { value: 'dark', label: 'Dark' },
          ]}
        />
      </div>
      <MessageColors title="User messages" base="userMessage" />
      <MessageColors title="Assistant messages" base="assistantMessage" />
      <Stack gap="xs">
        <Text fw={500} size="sm">
          Header
        </Text>
        <Group grow>
          <ColorField label="Header background" path="chatWindow.header.bg" />
          <ColorField label="Header text" path="chatWindow.header.color" />
        </Group>
        <ContrastWarning bgPath="chatWindow.header.bg" fgPath="chatWindow.header.color" />
      </Stack>
    </Stack>
  )
}

type CardType = 'button' | 'image' | 'link'

const newCard = (type: CardType) => ({
  type,
  config:
    type === 'button'
      ? { title: '', description: 'Describe this card', buttonText: 'Learn more', action: 'https://example.com' }
      : type === 'image'
        ? { imageUrl: 'https://example.com/image.png', description: 'Describe this card', action: 'https://example.com' }
        : { description: 'Describe this card', action: 'https://example.com' },
})

function CardsEditor() {
  const { config, set } = useBuilder()
  const cards = getIn<Array<{ type: CardType }>>(config, 'homeScreenConfig.additionalCards') ?? []
  return (
    <Stack>
      <Text fw={500} size="sm">
        Home screen cards
      </Text>
      {cards.map((card, i) => {
        const p = `homeScreenConfig.additionalCards.${i}`
        return (
          <Card key={i} withBorder padding="sm">
            <Stack gap="xs">
              <Group justify="space-between">
                <Select
                  aria-label={`Card ${i + 1} type`}
                  size="xs"
                  allowDeselect={false}
                  data={[
                    { value: 'button', label: 'Button card' },
                    { value: 'image', label: 'Image card' },
                    { value: 'link', label: 'Link card' },
                  ]}
                  value={card.type}
                  onChange={(v) => v && set(p, newCard(v as CardType))}
                />
                <ActionIcon variant="subtle" color="red" aria-label={`Remove card ${i + 1}`} onClick={() => set(p, undefined)}>
                  <IconTrash size={16} />
                </ActionIcon>
              </Group>
              <TextField label="Title" path={`${p}.config.title`} />
              {card.type === 'image' && <TextField label="Image URL" path={`${p}.config.imageUrl`} />}
              <TextAreaField label="Description" path={`${p}.config.description`} />
              {card.type === 'button' && <TextField label="Button text" path={`${p}.config.buttonText`} />}
              <TextField label="Link URL" path={`${p}.config.action`} />
            </Stack>
          </Card>
        )
      })}
      <Group>
        {(['button', 'image', 'link'] as const).map((type) => (
          <Button
            key={type}
            size="xs"
            variant="light"
            leftSection={<IconPlus size={14} />}
            onClick={() => set(`homeScreenConfig.additionalCards.${cards.length}`, newCard(type))}
          >
            {type} card
          </Button>
        ))}
      </Group>
    </Stack>
  )
}

export function HomeSection() {
  const { config, set } = useBuilder()
  const bgType = getIn<string>(config, 'homeScreenConfig.bgColor.type') ?? 'default'
  return (
    <Stack>
      <TextField label="Heading" path="homeScreenConfig.heading" />
      <TextField label="Second heading" path="homeScreenConfig.heading2" />
      <div>
        <Text size="sm" fw={500} mb={4}>
          Background
        </Text>
        <SegmentedControl
          value={bgType}
          onChange={(v) => set('homeScreenConfig.bgColor.type', v)}
          data={[
            { value: 'default', label: 'Gradient' },
            { value: 'plain', label: 'Plain' },
            { value: 'custom', label: 'Custom' },
          ]}
        />
      </div>
      {bgType === 'custom' && (
        <TextField label="Background (CSS)" path="homeScreenConfig.bgColor.background" placeholder="#eef or linear-gradient(...)" />
      )}
      <SwitchField label={'Hide "Send us a message"'} path="homeScreenConfig.sendUsAMessageConfig.hidden" />
      <TextField label={'"Send us a message" title'} path="homeScreenConfig.sendUsAMessageConfig.title" />
      <TextField label={'"Send us a message" description'} path="homeScreenConfig.sendUsAMessageConfig.description" />
      <CardsEditor />
    </Stack>
  )
}

export function SessionsSection() {
  return (
    <Stack>
      <TextField label="Sessions list title" path="sessionsListConfig.title" />
      <TextField label="New session button text" path="sessionsListConfig.newSessionButton.text" />
    </Stack>
  )
}

export function ChatSection() {
  const { config } = useBuilder()
  const actionNames = Object.keys(getIn<Record<string, string>>(config, 'actionsMap') ?? {})
  const { set } = useBuilder()
  return (
    <Stack>
      <SelectField
        label="Message corner radius"
        path="chatWindow.defaults.messageRadius"
        data={['xs', 'sm', 'md', 'lg', 'xl']}
        placeholder="Default"
      />
      <SwitchField label="Show online status" path="chatWindow.header.title.showOnlineSubtitle" defaultChecked />
      <SwitchField label="Hide assistant avatar in messages" path="chatWindow.hideAssistantMessageAvatar" />
      <SwitchField label="Hide user avatar in messages" path="chatWindow.hideUserMessageAvatar" />
      <TextAreaField label="Welcome message" path="chatWindow.welcomeMessage.message" />
      <TextField label="Info text" path="chatWindow.welcomeMessage.infoText" description="Shown at the top of the chat" />
      <MultiSelect
        label="Welcome message actions"
        description="Choose from the named actions defined under Advanced"
        data={actionNames}
        value={getIn<string[]>(config, 'chatWindow.welcomeMessage.actions') ?? []}
        onChange={(v) => set('chatWindow.welcomeMessage.actions', v)}
        placeholder={actionNames.length ? 'Pick actions' : 'No actions defined yet'}
      />
    </Stack>
  )
}

export function FooterSection() {
  return (
    <Stack>
      {(['home', 'messages'] as const).map((tab) => (
        <Stack key={tab} gap="xs">
          <Text fw={500} size="sm" tt="capitalize">
            {tab} tab
          </Text>
          <TextField label="Text" path={`footerConfig.${tab}.text`} />
          <TextField
            label="Icon"
            path={`footerConfig.${tab}.altIcon`}
            description="A Tabler icon path from the widget icon directory, or an SVG URL"
          />
        </Stack>
      ))}
    </Stack>
  )
}

export function BehaviourSection() {
  return (
    <Stack>
      <SwitchField label="Start expanded" path="chatWindow.expanded" />
      <SwitchField label="Disallow expand / collapse" path="chatWindow.disallowExpand" />
      <SwitchField label="Hide close button" path="disableCloseButton" />
      <SwitchField label="Debug logging" path="debug" description="Logs widget events to the browser console" />
    </Stack>
  )
}

function ActionsMapEditor() {
  const { config, set } = useBuilder()
  const map = getIn<Record<string, string>>(config, 'actionsMap') ?? {}
  const entries = Object.entries(map)
  const write = (next: Array<[string, string]>) => set('actionsMap', Object.fromEntries(next))
  return (
    <Stack gap="xs">
      <Text fw={500} size="sm">
        Named actions
      </Text>
      <Text size="xs" c="dimmed">
        Name to URL. The agent (or the welcome message) can offer these as buttons. Function actions need code and are not supported here.
      </Text>
      {entries.map(([name, url], i) => (
        <Group key={i} gap="xs" align="flex-end" wrap="nowrap">
          <TextInput
            label={i === 0 ? 'Name' : undefined}
            aria-label={`Action ${i + 1} name`}
            value={name}
            onChange={(e) => write(entries.map((en, j) => (j === i ? [e.currentTarget.value, en[1]] : en)))}
          />
          <TextInput
            style={{ flex: 1 }}
            label={i === 0 ? 'URL' : undefined}
            aria-label={`Action ${i + 1} URL`}
            value={url}
            onChange={(e) => write(entries.map((en, j) => (j === i ? [en[0], e.currentTarget.value] : en)))}
          />
          <ActionIcon variant="subtle" color="red" aria-label={`Remove action ${i + 1}`} onClick={() => write(entries.filter((_, j) => j !== i))}>
            <IconTrash size={16} />
          </ActionIcon>
        </Group>
      ))}
      <Button
        size="xs"
        variant="light"
        w="fit-content"
        leftSection={<IconPlus size={14} />}
        onClick={() => write([...entries, [`action_${entries.length + 1}`, 'https://example.com']])}
      >
        Add action
      </Button>
    </Stack>
  )
}

export function AdvancedSection() {
  return (
    <Stack>
      <TextAreaField
        label="Custom launcher button (HTML)"
        path="widgetButton"
        description="Replaces the default round button. Give it its own aria-label."
        placeholder={'<button aria-label="Open assistant">Chat</button>'}
      />
      <ActionsMapEditor />
      <JsonField
        label="Mantine theme override (JSON)"
        path="chatWindow.defaults.mantineThemeOverride"
        description="Advanced: see mantine.dev/theming/theme-object"
      />
    </Stack>
  )
}
