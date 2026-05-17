import PouchDB from 'pouchdb-browser'
import { authService } from './auth'
import type { SyncState } from '../types'

type SyncListener = (state: SyncState) => void
const syncListeners = new Set<SyncListener>()

export function onSyncChange(listener: SyncListener): () => void {
  syncListeners.add(listener)
  return () => syncListeners.delete(listener)
}

function notifySyncState(state: SyncState): void {
  syncListeners.forEach(fn => fn(state))
}

const COUCHDB_URL = import.meta.env.VITE_COUCHDB_URL

function createSyncedDB(dbName: string): { local: PouchDB.Database; remote: PouchDB.Database } {
  const local = new PouchDB(`kanso_${dbName}`)

  const remote = new PouchDB(`${COUCHDB_URL}/${dbName}`, {
    fetch: (url: RequestInfo | URL, opts: RequestInit = {}) => {
      const jwt = authService.getStoredJWT()
      const headers = new Headers(opts.headers)
      if (jwt) headers.set('Authorization', `Bearer ${jwt}`)
      return PouchDB.fetch(url, { ...opts, headers })
    },
  })

  local.sync(remote, { live: true, retry: true })
    .on('change', () => notifySyncState('syncing'))
    .on('paused', () => notifySyncState(navigator.onLine ? 'online' : 'offline'))
    .on('active', () => notifySyncState('syncing'))
    .on('error', (err: unknown) => {
      console.error(`PouchDB sync error (${dbName}):`, err)
      notifySyncState('offline')
    })

  return { local, remote }
}

export const { local: registrosDB } = createSyncedDB('registros')
export const { local: sentimentosDB } = createSyncedDB('sentimentos')
export const { local: preferenciasDB } = createSyncedDB('preferencias')

export function getUserId(): string {
  const jwt = authService.getStoredJWT()
  if (!jwt) return ''
  try {
    const payload = JSON.parse(atob(jwt.split('.')[1]))
    return payload.sub || ''
  } catch {
    return ''
  }
}
