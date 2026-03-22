import { Alert, Button, Group, Modal, Text } from '@mantine/core'

interface SaveConfigModalProps {
  opened: boolean
  onClose: () => void
  onConfirm: () => void
  saving: boolean
  error?: string
}

export function SaveConfigModal({ opened, onClose, onConfirm, saving, error }: SaveConfigModalProps) {
  return (
    <Modal opened={opened} onClose={onClose} title="Save Configuration to Agent" centered>
      {error ? (
        <Alert color="red" variant="light" mb="lg">
          {error}
        </Alert>
      ) : (
        <>
          <Text size="sm" mb="lg">
            This will overwrite the agent&apos;s saved configuration with the current playground
            settings. This action cannot be undone.
          </Text>
          <Text size="sm" c="dimmed" mb="lg">
            The updated configuration will be used as the default for all future playground sessions
            and API calls for this agent.
          </Text>
        </>
      )}
      <Group justify="flex-end">
        <Button variant="default" onClick={onClose} disabled={saving}>
          Cancel
        </Button>
        <Button color="red" onClick={onConfirm} loading={saving} disabled={!!error}>
          Overwrite Agent Config
        </Button>
      </Group>
    </Modal>
  )
}
