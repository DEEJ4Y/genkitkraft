import { createContext, useContext, useCallback, ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Center, Loader } from '@mantine/core'
import { fetchClient } from './api/client'
import { LoginDialog } from '../components/LoginDialog'

interface AuthContextType {
  isAuthRequired: boolean | undefined
  user: string | null
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()

  const statusQuery = useQuery({
    queryKey: ['get', '/api/auth/status'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/auth/status')
      if (error) throw new Error('Failed to fetch auth status')
      return data
    },
    staleTime: Infinity,
  })

  const meQuery = useQuery({
    queryKey: ['get', '/api/auth/me'],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET('/api/auth/me')
      if (error) throw new Error(error.error)
      return data
    },
    enabled: statusQuery.data?.required === true,
    retry: false,
  })

  const isAuthRequired = statusQuery.data?.required
  const user = meQuery.data?.username ?? null
  const isLoading =
    statusQuery.isPending || (isAuthRequired === true && meQuery.isPending)

  const login = useCallback(
    async (username: string, password: string) => {
      const { error } = await fetchClient.POST('/api/auth/login', {
        body: { username, password },
      })
      if (error) throw new Error(error.error)
      queryClient.invalidateQueries({ queryKey: ['get', '/api/auth/me'] })
    },
    [queryClient],
  )

  const logout = useCallback(async () => {
    await fetchClient.POST('/api/auth/logout')
    queryClient.setQueryData(['get', '/api/auth/me'], null)
  }, [queryClient])

  return (
    <AuthContext.Provider value={{ isAuthRequired, user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function AuthGate({ children }: { children: ReactNode }) {
  const { isAuthRequired, user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <Center h="100vh">
        <Loader />
      </Center>
    )
  }

  if (isAuthRequired === false) {
    return <>{children}</>
  }

  if (!user) {
    return <LoginDialog />
  }

  return <>{children}</>
}
