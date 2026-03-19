import type { AppProps } from 'next/app'
import { MantineProvider } from '@mantine/core'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '../lib/queryClient'
import { AuthProvider, AuthGate } from '../lib/auth'
import { AppLayout } from '../components/AppLayout'
import '@mantine/core/styles.css'

export default function App({ Component, pageProps }: AppProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MantineProvider>
          <AuthGate>
            <AppLayout>
              <Component {...pageProps} />
            </AppLayout>
          </AuthGate>
        </MantineProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}
