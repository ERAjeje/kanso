import { preferenciasDB, getUserId } from './pouchdb'

function pushPrefsId(): string {
  return `push_prefs:${getUserId()}`
}

export interface PushPreferences {
  enabled: boolean
  times: string[]
  timezone: string
  fcmSubscription?: string
}

interface PushDoc extends PushPreferences {
  _id: string
  _rev?: string
  type: 'push_prefs'
  userId: string
  createdAt: string
  updatedAt: string
}

export async function getPreferences(): Promise<PushPreferences> {
  const id = pushPrefsId()
  try {
    const doc = await preferenciasDB.get<PushDoc>(id)
    return {
      enabled: doc.enabled,
      times: doc.times,
      timezone: doc.timezone,
    }
  } catch {
    return { enabled: true, times: ['12:00', '18:00', '23:00'], timezone: '' }
  }
}

export async function savePreferences(prefs: Partial<PushPreferences>): Promise<void> {
  const userId = getUserId()
  const now = new Date().toISOString()

  const id = pushPrefsId()
  try {
    const existing = await preferenciasDB.get<PushDoc>(id)
    await preferenciasDB.put({
      ...existing,
      ...prefs,
      updatedAt: now,
    })
  } catch {
    await preferenciasDB.put({
      _id: id,
      type: 'push_prefs',
      userId,
      enabled: true,
      times: ['12:00', '18:00', '23:00'],
      timezone: '',
      ...prefs,
      createdAt: now,
      updatedAt: now,
    })
  }
}

export async function pushSubscription(doc: { fcmToken: string; timezone: string }): Promise<void> {
  await savePreferences({ fcmSubscription: doc.fcmToken, timezone: doc.timezone })
}
