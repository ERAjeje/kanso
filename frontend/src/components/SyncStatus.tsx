import { useState, useEffect } from 'react'

type SyncState = 'online' | 'offline' | 'syncing'

export function SyncStatus() {
  const [state, setState] = useState<SyncState>('online')

  useEffect(() => {
    const handleOnline = () => setState('online')
    const handleOffline = () => setState('offline')

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    setState(navigator.onLine ? 'online' : 'offline')

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  const colors = {
    online: 'bg-green-500',
    offline: 'bg-red-500',
    syncing: 'bg-yellow-500',
  }

  return (
    <div className="flex items-center gap-2 text-sm text-gray-500">
      <span className={`w-2 h-2 rounded-full ${colors[state]}`} />
      {state === 'online' && 'Online'}
      {state === 'offline' && 'Offline'}
      {state === 'syncing' && 'Sincronizando...'}
    </div>
  )
}
