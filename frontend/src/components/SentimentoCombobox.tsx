import { useState, useEffect, useRef } from 'react'
import { Combobox } from '@headlessui/react'
import { ChevronsUpDown } from 'lucide-react'
import { getSentimentos, saveSentimento } from '../services/registros'
import type { SentimentoDoc } from '../types'

interface Props {
  onChange: (nome: string, id: string | null) => void
}

export function SentimentoCombobox({  onChange }: Props) {
  const [sentimentos, setSentimentos] = useState<SentimentoDoc[]>([])
  const [query, setQuery] = useState('')
  const justSelected = useRef(false)

  useEffect(() => {
    getSentimentos().then(setSentimentos)
  }, [])

  const filtered = query === ''
    ? sentimentos
    : sentimentos.filter(s =>
        s.nome.toLowerCase().includes(query.toLowerCase())
      )

  const handleSelect = (nome: string | null) => {
    if (!nome) return
    justSelected.current = true
    setTimeout(() => { justSelected.current = false }, 0)
    const match = sentimentos.find(s => s.nome === nome)
    onChange(nome, match?._id ?? null)
  }

  const handleBlur = () => {
    if (justSelected.current) return
    const trimmed = query.trim().toLowerCase()
    if (trimmed && !sentimentos.some(s => s.nome === trimmed)) {
      saveSentimento(trimmed).then(doc => {
        setSentimentos(prev => [...prev, doc].sort((a, b) => a.nome.localeCompare(b.nome, 'pt-BR')))
        onChange(doc.nome, doc._id)
      })
    }
  }

  return (
    <Combobox value={query} onChange={handleSelect} nullable>
      <div className="relative">
        <Combobox.Input
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Digite ou selecione um sentimento..."
          onChange={e => setQuery(e.target.value)}
          onBlur={handleBlur}
        />
        <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-3">
          <ChevronsUpDown className="h-4 w-4 text-gray-400" />
        </Combobox.Button>
        <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg text-sm">
          {filtered.length > 0 ? filtered.map(s => (
            <Combobox.Option
              key={s._id}
              value={s.nome}
              className={({ active }) =>
                `cursor-pointer px-4 py-2 ${active ? 'bg-indigo-50 text-indigo-700' : 'text-gray-900'}`
              }
            >
              {s.nome}
            </Combobox.Option>
          )) : query ? (
            <div className="px-4 py-2 text-gray-500">
              Pressione Enter para criar &ldquo;{query}&rdquo;
            </div>
          ) : null}
        </Combobox.Options>
      </div>
    </Combobox>
  )
}
