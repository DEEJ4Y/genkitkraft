import { ReactNode } from 'react'
import { AppShell, NavLink, Title, Text, Button } from '@mantine/core'
import { IconHome2, IconSettings, IconLogout } from '@tabler/icons-react'
import { useRouter } from 'next/router'
import { useAuth } from '../lib/auth'

export function AppLayout({ children }: { children: ReactNode }) {
  const router = useRouter()
  const { user, logout, isAuthRequired } = useAuth()

  return (
    <AppShell navbar={{ width: 240, breakpoint: 'sm' }} padding="md">
      <AppShell.Navbar p="md">
        <AppShell.Section>
          <Title order={4} mb="md">
            GenKitKraft
          </Title>
        </AppShell.Section>

        <AppShell.Section grow>
          <NavLink
            label="Dashboard"
            leftSection={<IconHome2 size={18} />}
            active={router.pathname === '/'}
            onClick={() => router.push('/')}
          />
          <NavLink
            label="Settings"
            leftSection={<IconSettings size={18} />}
            active={router.pathname === '/settings'}
            onClick={() => router.push('/settings')}
          />
        </AppShell.Section>

        {isAuthRequired && user && (
          <AppShell.Section>
            <Text size="sm" c="dimmed" mb="xs">
              {user}
            </Text>
            <Button
              variant="subtle"
              size="xs"
              leftSection={<IconLogout size={14} />}
              onClick={logout}
            >
              Sign out
            </Button>
          </AppShell.Section>
        )}
      </AppShell.Navbar>

      <AppShell.Main>{children}</AppShell.Main>
    </AppShell>
  )
}
