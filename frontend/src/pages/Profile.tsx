import { useEffect, useState } from 'react'
import { ReportSection } from '../components/ReportSection'
import { getPreferences, updatePreferences, type PushPreferences } from '../services/push'

const DEFAULT_TIMES = ['12:00', '18:00', '23:00']

export function Profile() {
  const [prefs, setPrefs] = useState<PushPreferences | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    getPreferences()
      .then(setPrefs)
      .catch(() => setPrefs({ enabled: true, times: DEFAULT_TIMES, timezone: '' }))
      .finally(() => setLoading(false))
  }, [])

  const toggleEnabled = async () => {
    if (!prefs) return
    const next = { enabled: !prefs.enabled, times: prefs.times }
    setSaving(true)
    try {
      await updatePreferences(next)
      setPrefs(prev => prev ? { ...prev, enabled: next.enabled } : null)
    } catch {
      // revert on error
    }
    setSaving(false)
  }

  const updateTime = async (index: number, value: string) => {
    if (!prefs) return
    const times = [...prefs.times]
    times[index] = value
    const next = { enabled: prefs.enabled, times }
    setSaving(true)
    try {
      await updatePreferences(next)
      setPrefs(prev => prev ? { ...prev, times } : null)
    } catch {
      // revert on error
    }
    setSaving(false)
  }

  const permissionStatus = 'Notification' in window
    ? Notification.permission === 'granted'
      ? 'Permitido'
      : Notification.permission === 'denied'
        ? 'Negado'
        : 'Não solicitado'
    : 'Indisponível'

  return (
    <div className="p-8 max-w-lg mx-auto space-y-8">
      <h1 className="text-2xl font-bold text-gray-800 mb-2">Perfil</h1>
      <ReportSection />

      <section>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Lembretes</h2>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100 space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-700">Notificações push</span>
            {loading ? (
              <div className="w-10 h-5 bg-gray-200 rounded-full animate-pulse" />
            ) : (
              <button
                onClick={toggleEnabled}
                disabled={saving}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  prefs?.enabled ? 'bg-indigo-600' : 'bg-gray-200'
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    prefs?.enabled ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
            )}
          </div>

          <div className="text-xs text-gray-400">
            Status da permissão: {permissionStatus}
          </div>

          {prefs?.enabled && (
            <div className="space-y-2">
              <label className="text-sm text-gray-600">Horários dos lembretes</label>
              {(prefs.times.length > 0 ? prefs.times : DEFAULT_TIMES).map((time, index) => (
                <div key={index} className="flex items-center gap-3">
                  <span className="text-xs text-gray-400 w-12 text-right">
                    {['Manhã', 'Tarde', 'Noite'][index] || `Horário ${index + 1}`}
                  </span>
                  <input
                    type="time"
                    value={time}
                    onChange={e => updateTime(index, e.target.value)}
                    disabled={saving}
                    className="flex-1 rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  )
}
