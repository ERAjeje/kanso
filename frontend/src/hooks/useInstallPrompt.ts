import { useState, useEffect, useCallback } from 'react'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

const DISMISSED_KEY = 'pwa_install_dismissed'

export function useInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [canInstall, setCanInstall] = useState(false)

  useEffect(() => {
    if (sessionStorage.getItem(DISMISSED_KEY) === 'true') return

    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
      setCanInstall(true)
    }

    window.addEventListener('beforeinstallprompt', handler)
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const triggerInstall = useCallback(async () => {
    if (!deferredPrompt) return

    deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    setDeferredPrompt(null)
    setCanInstall(false)

    if (outcome === 'accepted') {
      sessionStorage.removeItem(DISMISSED_KEY)
    }
  }, [deferredPrompt])

  const userDismissed = useCallback(() => {
    sessionStorage.setItem(DISMISSED_KEY, 'true')
    setDeferredPrompt(null)
    setCanInstall(false)
  }, [])

  return { canInstall, triggerInstall, userDismissed }
}
