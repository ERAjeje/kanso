import { useState, useEffect } from 'react'
import { getRegistros } from '../services/registros'
import { RegistroCard } from '../components/RegistroCard'
import { SyncStatus } from '../components/SyncStatus'
import type { RegistroWithAnalise } from '../types'

export function History() {
  const [registros, setRegistros] = useState<RegistroWithAnalise[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const fetchRegistros = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await getRegistros()
      setRegistros(data)
    } catch {
      setError('Erro ao carregar registros')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchRegistros()
  }, [])

  return (
    <div className="p-8 max-w-lg mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Histórico</h1>
        <SyncStatus />
      </div>

      {loading && (
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <p className="text-gray-400 text-center py-12">Carregando...</p>
        </div>
      )}

      {!loading && error && (
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <div className="text-center py-12">
            <p className="text-red-500 mb-4">{error}</p>
            <button
              onClick={fetchRegistros}
              className="bg-primary text-white rounded-lg px-4 py-2 text-sm font-medium hover:brightness-110 transition-all"
            >
              Tentar novamente
            </button>
          </div>
        </div>
      )}

      {!loading && !error && registros.length === 0 && (
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <p className="text-gray-400 text-center py-12">Nenhum registro ainda</p>
        </div>
      )}

      {!loading && !error && registros.length > 0 && (
        <div className="space-y-4">
          {registros.map(r => (
            <RegistroCard
              key={r._id}
              registro={r}
              onSentimentoUpdated={fetchRegistros}
            />
          ))}
        </div>
      )}
    </div>
  )
}
