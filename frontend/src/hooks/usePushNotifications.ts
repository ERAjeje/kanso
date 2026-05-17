import { useEffect, useState } from 'react'
import { useAuth } from './useAuth'
import { pushSubscription } from '../services/push'

type Permission = NotificationPermission | 'unavailable'

export function usePushNotifications() {
  const { user } = useAuth()
  const [permission, setPermission] = useState<Permission>('default')
  const [subscribed, setSubscribed] = useState(false)

  useEffect(() => {
    if (!user) return
    if (!('Notification' in window)) {
      setPermission('unavailable')
      return
    }
    setPermission(Notification.permission)
  }, [user])

  const requestPermission = async () => {
    if (permission === 'unavailable') return
    const result = await Notification.requestPermission()
    setPermission(result)

    if (result === 'granted') {
      try {
        const reg = await navigator.serviceWorker.register('/sw.js')
        const vapidKey = import.meta.env.VITE_VAPID_PUBLIC_KEY
        if (!vapidKey) {
          console.warn('VITE_VAPID_PUBLIC_KEY not set — push subscription skipped')
          return
        }
        const sub = await reg.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(vapidKey),
        })
        const fcmToken = JSON.stringify(sub)
        const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
        await pushSubscription({ fcmToken, timezone })
        setSubscribed(true)
      } catch (err) {
        console.error('Failed to subscribe to push:', err)
      }
    }
  }

  return { permission, subscribed, requestPermission }
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}
