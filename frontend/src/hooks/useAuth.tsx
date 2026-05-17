import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { authService, type User } from '../services/auth'
import { registrosDB, sentimentosDB } from '../services/pouchdb'
import { pushSubscription } from '../services/push'
import { Toast } from '../components/Toast'

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
  const [toast, setToast] = useState('')

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

    if ('Notification' in window && Notification.permission === 'default') {
      try {
        const reg = await navigator.serviceWorker.register('/sw.js')
        const vapidKey = import.meta.env.VITE_VAPID_PUBLIC_KEY
        if (!vapidKey) {
          setToast('Não foi possível atualizar. Tente novamente mais tarde.')
          return
        }
        const perm = await Notification.requestPermission()
        if (perm === 'granted') {
          const sub = await reg.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array(vapidKey),
          })
          const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
          const fcmToken = JSON.stringify(sub)
          await pushSubscription({ fcmToken, timezone })
        }
      } catch {
        setToast('Não foi possível ativar notificações push')
      }
    }
  }, [])

  const signOut = useCallback(async () => {
    await authService.logout()
    await destroyPouchDB()
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, signIn, signOut }}>
      <Toast message={toast} visible={!!toast} onClose={() => setToast('')} />
      {children}
    </AuthContext.Provider>
  )
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; i++) outputArray[i] = rawData.charCodeAt(i)
  return outputArray
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
