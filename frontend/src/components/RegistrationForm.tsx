import { useState, type FormEvent } from 'react'
import { SentimentoCombobox } from './SentimentoCombobox'
import { saveRegistro } from '../services/registros'
import { usePouchSync } from '../hooks/usePouchSync'

interface Props {
  onSaved: () => void
  onShowToast: (message: string) => void
}

export function RegistrationForm({ onSaved, onShowToast }: Props) {
  const syncState = usePouchSync()
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const [dataHora, setDataHora] = useState(() => {
    const now = new Date()
    return now.toISOString().slice(0, 16)
  })
  const [sentimento, setSentimento] = useState({ nome: '', id: null as string | null })
  const [sensacoes, setSensacoes] = useState('')
  const [contexto, setContexto] = useState('')
  const [pensamentos, setPensamentos] = useState('')

  const inputClass = 'w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500'

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (saving) return
    setSaving(true)
    setError('')

    try {
      await saveRegistro({
        dataHora: new Date(dataHora).toISOString(),
        sensacoes,
        sentimentoId: sentimento.id,
        sentimentoNome: sentimento.nome,
        contexto,
        pensamentos,
      })

      setDataHora(new Date().toISOString().slice(0, 16))
      setSentimento({ nome: '', id: null })
      setSensacoes('')
      setContexto('')
      setPensamentos('')

      if (syncState !== 'online') {
        onShowToast('Salvo localmente — será sincronizado quando você estiver online')
      }

      onSaved()
    } catch (err) {
      console.error('Failed to save:', err)
      setError('Não foi possível salvar o registro. Verifique sua conexão e tente novamente.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <div>
        <label className="block text-sm font-normal text-gray-700 mb-1">Data e Hora</label>
        <input
          type="datetime-local"
          value={dataHora}
          onChange={e => setDataHora(e.target.value)}
          className={inputClass}
        />
      </div>

      <div>
        <label className="block text-sm font-normal text-gray-700 mb-1">Sentimento</label>
        <SentimentoCombobox
          onChange={(nome, id) => setSentimento({ nome, id })}
        />
      </div>

      <div>
        <label className="block text-sm font-normal text-gray-700 mb-1">Sensações</label>
        <textarea
          rows={3}
          value={sensacoes}
          onChange={e => setSensacoes(e.target.value)}
          placeholder="O que você está sentindo no corpo?"
          className={`${inputClass} resize-none`}
        />
      </div>

      <div>
        <label className="block text-sm font-normal text-gray-700 mb-1">Contexto</label>
        <textarea
          rows={3}
          value={contexto}
          onChange={e => setContexto(e.target.value)}
          placeholder="O que estava acontecendo?"
          className={`${inputClass} resize-none`}
        />
      </div>

      <div>
        <label className="block text-sm font-normal text-gray-700 mb-1">Pensamentos</label>
        <textarea
          rows={3}
          value={pensamentos}
          onChange={e => setPensamentos(e.target.value)}
          placeholder="O que está passando pela sua cabeça?"
          className={`${inputClass} resize-none`}
        />
      </div>

      <button
        type="submit"
        disabled={saving || !sentimento.nome.trim()}
        className="w-full bg-indigo-600 text-white rounded-lg py-3 font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {saving ? 'Salvando...' : 'Registrar'}
      </button>

      {error && (
        <div className="text-red-500 text-sm bg-red-50 p-3 rounded-lg">
          {error}
        </div>
      )}
    </form>
  )
}
