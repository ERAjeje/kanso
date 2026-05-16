import { useCallback, useEffect } from 'react'
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
        }
      }
    }
  }
}

export function Login() {
  const { signIn, user } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (user) navigate('/register', { replace: true })
  }, [user, navigate])

  const handleGoogleSignIn = useCallback(async (response: { credential: string }) => {
    try {
      await signIn(response.credential)
      navigate('/register', { replace: true })
    } catch (err) {
      console.error('Login failed:', err)
    }
  }, [signIn, navigate])

  useEffect(() => {
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
    }
  }, [handleGoogleSignIn])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 to-white">
      <div className="bg-white p-8 rounded-2xl shadow-lg max-w-md w-full text-center">
        <h1 className="text-3xl font-bold text-indigo-600 mb-2">Kanso</h1>
        <p className="text-gray-500 mb-8">Diário Emocional</p>
        <p className="text-sm text-gray-600 mb-6">
          Faça login com sua conta Google para começar
        </p>
        <div id="google-signin-btn" className="flex justify-center" />
      </div>
    </div>
  )
}
