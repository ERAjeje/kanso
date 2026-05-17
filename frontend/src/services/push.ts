import { authService } from './auth'

const API = import.meta.env.VITE_API_URL || ''

export interface PushPreferences {
  enabled: boolean
  times: string[]
  timezone: string
  userSub?: string
}

export async function subscribe(fcmToken: string): Promise<void> {
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const token = authService.getStoredJWT()
  const res = await fetch(`${API}/api/push/subscribe`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ fcmToken, timezone }),
  })
  if (!res.ok) throw new Error('Failed to subscribe to push notifications')
}

export async function getPreferences(): Promise<PushPreferences> {
  const token = authService.getStoredJWT()
  const res = await fetch(`${API}/api/push/preferences`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error('Failed to get push preferences')
  return res.json()
}

export async function updatePreferences(prefs: { enabled: boolean; times: string[] }): Promise<void> {
  const token = authService.getStoredJWT()
  const res = await fetch(`${API}/api/push/preferences`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(prefs),
  })
  if (!res.ok) throw new Error('Failed to update push preferences')
}
