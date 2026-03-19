import type { AppProps } from 'next/app'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '../lib/queryClient'
import { AuthProvider, AuthGate } from '../lib/auth'

export default function App({ Component, pageProps }: AppProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <AuthGate>
          <Component {...pageProps} />
        </AuthGate>
      </AuthProvider>
    </QueryClientProvider>
  )
}
