import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string
            callback: (response: { credential: string }) => void
          }) => void
          renderButton: (element: HTMLElement, options: {
            theme: string
            size: string
            text: string
          }) => void
          prompt: (momentListener?: (moment: string) => void) => void
          cancel: () => void
        }
      }
    }
  }
}

function initGIS(handleGoogleSignIn: (response: { credential: string }) => void) {
  if (!window.google) return false
  window.google.accounts.id.initialize({
    client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
    callback: handleGoogleSignIn,
  })
  const btn = document.getElementById('google-signin-btn')
  if (btn) {
    window.google.accounts.id.renderButton(btn, {
      theme: 'outline',
      size: 'large',
      text: 'signin_with',
    })
  }
  return true
}

export function Login() {
  const { signIn, user } = useAuth()
  const navigate = useNavigate()
  const [gisReady, setGisReady] = useState(false)
  const [showFallback, setShowFallback] = useState(false)
  const pollingRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined)

  useEffect(() => {
    if (user) navigate('/register', { replace: true })
  }, [user, navigate])

  const handleGoogleSignIn = useCallback(async (response: { credential: string }) => {
    try {
      await signIn(response.credential)
      navigate('/register', { replace: true })
    } catch (err) {
      console.warn('Login failed:', err)
    }
  }, [signIn, navigate])

  useEffect(() => {
    if (initGIS(handleGoogleSignIn)) {
      setGisReady(true)
      setShowFallback(false)
      return
    }

    pollingRef.current = setInterval(() => {
      if (initGIS(handleGoogleSignIn)) {
        setGisReady(true)
        setShowFallback(false)
        clearInterval(pollingRef.current)
      }
    }, 500)

    const fallbackTimer = setTimeout(() => {
      setShowFallback(true)
      clearInterval(pollingRef.current)
    }, 8000)

    return () => {
      clearInterval(pollingRef.current)
      clearTimeout(fallbackTimer)
    }
  }, [handleGoogleSignIn])

  const handleFallbackClick = useCallback(() => {
    if (window.google?.accounts?.id) {
      window.google.accounts.id.prompt()
      return
    }
    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.onload = () => {
      if (window.google) {
        window.google.accounts.id.initialize({
          client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
          callback: handleGoogleSignIn,
        })
        const btn = document.getElementById('google-signin-btn')
        if (btn) {
          window.google.accounts.id.renderButton(btn, {
            theme: 'outline',
            size: 'large',
            text: 'signin_with',
          })
        }
        setGisReady(true)
        setShowFallback(false)
      }
    }
    document.head.appendChild(script)
  }, [handleGoogleSignIn])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/5 to-background">
      <div className="bg-white p-8 rounded-2xl shadow-lg max-w-md w-full text-center">
        <h1 className="text-3xl font-bold text-primary mb-2">Kanso</h1>
        <p className="text-gray-500 mb-8">Diário Emocional</p>
        <p className="text-sm text-gray-600 mb-6">
          Faça login com sua conta Google para começar
        </p>
        <div id="google-signin-btn" className="flex justify-center" />
        {showFallback && !gisReady && (
          <button
            onClick={handleFallbackClick}
            className="mt-4 flex items-center justify-center gap-3 w-full px-6 py-3 border border-gray-300 rounded-lg shadow-sm bg-white text-gray-700 hover:bg-gray-50 transition-colors text-sm font-medium"
          >
            <svg className="w-5 h-5" viewBox="0 0 48 48">
              <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
              <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
              <path fill="#FBBC05" d="M10.54 28.59A14.5 14.5 0 0 1 9.5 24c0-1.59.28-3.14.76-4.59l-7.98-6.19A23.99 23.99 0 0 0 0 24c0 3.77.87 7.35 2.56 10.56l7.98-5.97z" />
              <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 5.97C6.51 42.62 14.62 48 24 48z" />
            </svg>
            Entrar com Google
          </button>
        )}
      </div>
    </div>
  )
}
