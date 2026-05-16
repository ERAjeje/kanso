import { useState, useEffect } from 'react'
import { onSyncChange } from '../services/pouchdb'
import type { SyncState } from '../types'

export function usePouchSync(): SyncState {
  const [state, setState] = useState<SyncState>(
    navigator.onLine ? 'online' : 'offline'
  )

  useEffect(() => {
    const unsub = onSyncChange(setState)
    return unsub
  }, [])

  return state
}
