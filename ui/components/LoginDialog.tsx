import { useState, FormEvent } from 'react'
import { Center, Paper, Title, Text, TextInput, PasswordInput, Button, Stack } from '@mantine/core'
import { useAuth } from '../lib/auth'

export function LoginDialog() {
  const { login } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(username, password)
    } catch (err: any) {
      setError(err.message || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Center h="100vh" bg="gray.0">
      <Paper p="xl" w={380} shadow="md" radius="md">
        <Title order={3} ta="center" mb={4}>
          GenKitKraft
        </Title>
        <Text size="sm" c="dimmed" ta="center" mb="lg">
          Sign in to continue
        </Text>
        <form onSubmit={handleSubmit}>
          <Stack gap="sm">
            <TextInput
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.currentTarget.value)}
              autoFocus
              required
            />
            <PasswordInput
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.currentTarget.value)}
              required
            />
            {error && (
              <Text size="xs" c="red">
                {error}
              </Text>
            )}
            <Button type="submit" fullWidth loading={loading} color="dark">
              Sign in
            </Button>
          </Stack>
        </form>
      </Paper>
    </Center>
  )
}
