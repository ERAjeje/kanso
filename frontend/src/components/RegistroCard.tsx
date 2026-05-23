import { useState } from 'react'
import type { RegistroDoc, RegistroWithAnalise } from '../types'

interface Props {
  registro: RegistroWithAnalise
}

const EMOTION_CHIP_COLORS: Record<string, { bg: string; text: string }> = {
  alegria: { bg: 'bg-emerald-100', text: 'text-emerald-700' },
  tristeza: { bg: 'bg-blue-100', text: 'text-blue-700' },
  raiva: { bg: 'bg-red-100', text: 'text-red-700' },
  medo: { bg: 'bg-purple-100', text: 'text-purple-700' },
  nojo: { bg: 'bg-amber-100', text: 'text-amber-700' },
  surpresa: { bg: 'bg-orange-100', text: 'text-orange-700' },
  ansiedade: { bg: 'bg-yellow-100', text: 'text-yellow-700' },
  vergonha: { bg: 'bg-pink-100', text: 'text-pink-700' },
  culpa: { bg: 'bg-rose-100', text: 'text-rose-700' },
  saudade: { bg: 'bg-violet-100', text: 'text-violet-700' },
  amor: { bg: 'bg-pink-200', text: 'text-pink-800' },
  gratidão: { bg: 'bg-teal-100', text: 'text-teal-700' },
  neutro: { bg: 'bg-gray-100', text: 'text-gray-600' },
}

function getChipColors(emotion: string): { bg: string; text: string } {
  return EMOTION_CHIP_COLORS[emotion] ?? { bg: 'bg-gray-100', text: 'text-gray-600' }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function previewText(r: RegistroDoc): string {
  return r.sensacoes || r.contexto || r.pensamentos || ''
}

function truncate(text: string, max = 80): string {
  if (text.length <= max) return text
  return text.slice(0, max).trimEnd() + '…'
}

export function RegistroCard({ registro }: Props) {
  const [expanded, setExpanded] = useState(false)
  const hasSentimento = registro.sentimentoNome && registro.sentimentoNome.trim().length > 0
  const preview = previewText(registro)
  const analise = registro.analise

  return (
    <div
      className="bg-white rounded-xl p-6 shadow-sm border border-gray-100 cursor-pointer transition-shadow hover:shadow-md"
      onClick={() => setExpanded(prev => !prev)}
      role="button"
      tabIndex={0}
      aria-expanded={expanded}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setExpanded(prev => !prev) } }}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <h3 className={`text-lg font-semibold ${hasSentimento ? 'text-gray-800' : 'text-gray-400 italic'}`}>
            {hasSentimento ? registro.sentimentoNome : 'Buscando sentimento'}
          </h3>
          <p className="text-sm text-gray-500 mt-0.5">{formatDate(registro.dataHora)}</p>
          {analise && (
            <div className="flex flex-wrap gap-2 mt-1">
              <span className={`text-xs font-semibold px-2 py-1 rounded-full ${getChipColors(analise.emotionPrincipal).bg} ${getChipColors(analise.emotionPrincipal).text}`}>
                {analise.emotionPrincipal}
              </span>
              {analise.emotions
                .filter(e => e.emotion !== analise.emotionPrincipal)
                .sort((a, b) => b.score - a.score)
                .slice(0, 3)
                .map(e => (
                  <span key={e.emotion} className={`text-xs font-semibold px-2 py-1 rounded-full ${getChipColors(e.emotion).bg} ${getChipColors(e.emotion).text}`}>
                    {e.emotion}
                  </span>
                ))
              }
            </div>
          )}
        </div>
        <span className="text-gray-400 ml-2 mt-1 transition-transform duration-200" style={{ transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)' }}>
          ▼
        </span>
      </div>

      {!expanded && preview && (
        <p className="text-sm text-gray-600 mt-3 line-clamp-2">{truncate(preview)}</p>
      )}

      {expanded && (
        <div className="mt-4 space-y-3 pt-3 border-t border-gray-100">
          <Field label="Sensações" value={registro.sensacoes} />
          <Field label="Sentimento" value={hasSentimento ? registro.sentimentoNome : 'Buscando sentimento'} />
          <Field label="Contexto" value={registro.contexto} />
          <Field label="Pensamentos" value={registro.pensamentos} />
        </div>
      )}
    </div>
  )
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">{label}</span>
      <p className="text-sm text-gray-700 mt-0.5 whitespace-pre-wrap">{value || '—'}</p>
    </div>
  )
}
