import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { authService, type User } from '../services/auth'
import { registrosDB, sentimentosDB } from '../services/pouchdb'

interface AuthContextType {
  user: User | null
  loading: boolean
  signIn: (idToken: string) => Promise<void>
  signOut: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

async function destroyPouchDB() {
  try {
    await registrosDB.destroy()
    await sentimentosDB.destroy()
  } catch {
    // Ignore — DB may already be destroyed
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    authService.getCurrentUser().then(u => {
      setUser(u)
      setLoading(false)
    })
  }, [])

  const signIn = useCallback(async (idToken: string) => {
    const result = await authService.exchangeGoogleToken(idToken)
    authService.storeJWT(result.jwt)
    const u = await authService.getCurrentUser()
    setUser(u)
  }, [])

  const signOut = useCallback(async () => {
    await authService.logout()
    await destroyPouchDB()
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
