import { useState } from 'react'
import { Combobox } from '@headlessui/react'
import { ChevronsUpDown } from 'lucide-react'

const EMOTIONS = [
  'alegria', 'amor', 'ansiedade', 'culpa', 'gratidão',
  'medo', 'neutro', 'nojo', 'raiva', 'saudade',
  'surpresa', 'tristeza', 'vergonha',
]

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

interface SentimentoEditorProps {
  currentValue: string
  disabled: boolean
  onSave: (label: string) => Promise<void>
}

export function SentimentoEditor({ currentValue, disabled, onSave }: SentimentoEditorProps) {
  const [query, setQuery] = useState('')

  if (disabled) {
    const colors = getChipColors(currentValue)
    return (
      <div>
        <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Sentimento</span>
        {currentValue ? (
          <span className={`inline-block mt-1 text-sm font-semibold px-3 py-1 rounded-full ${colors.bg} ${colors.text}`}>
            {currentValue}
          </span>
        ) : (
          <p className="text-sm text-gray-400 mt-0.5">—</p>
        )}
      </div>
    )
  }

  const sortedEmotions = [...EMOTIONS].sort((a, b) => a.localeCompare(b, 'pt-BR'))
  const filtered = query === ''
    ? sortedEmotions
    : sortedEmotions.filter(e => e.toLowerCase().includes(query.toLowerCase()))

  return (
    <div>
      <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Sentimento</span>
      <Combobox
        value={currentValue}
        onChange={(value) => {
          if (value) {
            onSave(value)
          }
        }}
        nullable
      >
        <div className="relative mt-1">
          <Combobox.Input
            className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="Selecionar sentimento"
            onChange={e => setQuery(e.target.value)}
          />
          <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-3">
            <ChevronsUpDown className="h-4 w-4 text-gray-400" />
          </Combobox.Button>
          <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg text-sm">
            {filtered.length > 0 ? filtered.map(emotion => (
              <Combobox.Option
                key={emotion}
                value={emotion}
                className={({ active }) =>
                  `cursor-pointer px-4 py-2 ${active ? 'bg-indigo-50 text-indigo-700' : 'text-gray-900'}`
                }
              >
                {emotion}
              </Combobox.Option>
            )) : (
              <div className="px-4 py-2 text-gray-500">
                Nenhuma emoção encontrada
              </div>
            )}
          </Combobox.Options>
        </div>
      </Combobox>
    </div>
  )
}
