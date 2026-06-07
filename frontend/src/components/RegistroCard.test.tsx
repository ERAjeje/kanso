import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { RegistroCard } from './RegistroCard'
import type { RegistroDoc, RegistroWithAnalise, AnaliseNlpDoc } from '../types'

// Mock pouchdb to prevent PouchDB initialization in jsdom
vi.mock('../services/pouchdb', () => ({
  registrosDB: { get: vi.fn(), put: vi.fn() },
  sentimentosDB: { allDocs: vi.fn() },
  treinamentoDB: { put: vi.fn() },
  getUserId: () => 'test-user-123',
}))

function makeRegistro(overrides: Partial<RegistroDoc> = {}): RegistroDoc {
  return {
    _id: '1',
    _rev: '1-abc',
    type: 'registro',
    userId: 'u1',
    dataHora: '2026-05-17T14:30:00.000Z',
    sensacoes: 'Coração acelerado, mãos frias',
    sentimentoId: null,
    sentimentoNome: 'ansiedade',
    contexto: 'Reunião com o chefe',
    pensamentos: 'Será que fiz o suficiente?',
    createdAt: '2026-05-17T14:30:00.000Z',
    updatedAt: '2026-05-17T14:30:00.000Z',
    ...overrides,
  }
}

describe('RegistroCard', () => {
  it('renders sentimentoNome when present', () => {
    render(<RegistroCard registro={makeRegistro()} />)
    expect(screen.getByText('ansiedade')).toBeDefined()
  })

  it('renders fallback text when sentimentoNome is empty string', () => {
    render(<RegistroCard registro={makeRegistro({ sentimentoNome: '' })} />)
    expect(screen.getByText('Buscando sentimento')).toBeDefined()
  })

  it('renders fallback text when sentimentoNome is whitespace only', () => {
    render(<RegistroCard registro={makeRegistro({ sentimentoNome: '   ' })} />)
    expect(screen.getByText('Buscando sentimento')).toBeDefined()
  })

  it('shows date/time formatted in pt-BR', () => {
    render(<RegistroCard registro={makeRegistro({ dataHora: '2026-05-17T14:30:00.000Z' })} />)
    expect(screen.getByText(/17 de maio/i)).toBeDefined()
    expect(screen.getByText(/2026/i)).toBeDefined()
    expect(screen.getByText(/às/i)).toBeDefined()
  })

  it('shows content preview when collapsed', () => {
    render(<RegistroCard registro={makeRegistro()} />)
    const preview = screen.getByText(/Coração acelerado/)
    expect(preview).toBeDefined()
  })

  it('expands inline on click to show all fields', () => {
    render(<RegistroCard registro={makeRegistro()} />)
    fireEvent.click(screen.getByRole('button'))

    expect(screen.getByText('Sensações')).toBeDefined()
    expect(screen.getByText('Sentimento')).toBeDefined()
    expect(screen.getByText('Contexto')).toBeDefined()
    expect(screen.getByText('Pensamentos')).toBeDefined()
  })

  it('collapses on second click', () => {
    render(<RegistroCard registro={makeRegistro()} />)
    const card = screen.getByRole('button')

    fireEvent.click(card)
    expect(screen.getByText('Sensações')).toBeDefined()

    fireEvent.click(card)
    expect(screen.queryByText('Sensações')).toBeNull()
  })

  it('shows expanded fallback text when sentimentoNome is empty', () => {
    render(<RegistroCard registro={makeRegistro({ sentimentoNome: '' })} />)
    fireEvent.click(screen.getByRole('button'))
    const labels = screen.getAllByText('Buscando sentimento')
    expect(labels.length).toBeGreaterThanOrEqual(1)
  })

  it('shows SentimentoEditor when expanded and sentimentoId is null', () => {
    const onSentimentoUpdated = vi.fn()
    render(
      <RegistroCard
        registro={makeRegistro({ sentimentoId: null, sentimentoNome: '' })}
        onSentimentoUpdated={onSentimentoUpdated}
      />
    )
    // Click to expand
    fireEvent.click(screen.getByRole('button'))
    // Should show the SentimentoEditor placeholder
    expect(screen.getByPlaceholderText('Selecionar sentimento')).toBeDefined()
  })

  it('shows static Sentimento field when sentimentoId exists', () => {
    const onSentimentoUpdated = vi.fn()
    render(
      <RegistroCard
        registro={makeRegistro({ sentimentoId: 'label-ansiedade', sentimentoNome: 'ansiedade' })}
        onSentimentoUpdated={onSentimentoUpdated}
      />
    )
    // Click to expand
    fireEvent.click(screen.getByRole('button'))
    // Should still show the static Sentimento field, not the editor
    expect(screen.queryByPlaceholderText('Selecionar sentimento')).toBeNull()
    // ansiedade appears in heading and chips (multiple times) — use getAllByText
    const ansiedadeElements = screen.getAllByText('ansiedade')
    expect(ansiedadeElements.length).toBeGreaterThanOrEqual(1)
  })
})

function makeEnrichedRegistro(overrides: Partial<RegistroDoc> = {}, analiseOverrides: Partial<AnaliseNlpDoc> = {}): RegistroWithAnalise {
  return {
    ...makeRegistro(overrides),
    analise: {
      _id: 'analise:' + (overrides._id || '1'),
      type: 'analise_nlp',
      userId: 'u1',
      registroId: overrides._id || '1',
      emotionPrincipal: 'ansiedade',
      emotions: [
        { emotion: 'ansiedade', score: 0.85 },
        { emotion: 'medo', score: 0.42 },
      ],
      scores: { ansiedade: 0.85, medo: 0.42 },
      intensidade: 0.85,
      modeloVersao: 'v1.0',
      analisadoEm: '2026-05-23T10:01:00Z',
      ...analiseOverrides,
    },
  }
}

describe('emotion chips', () => {
  it('shows emotionPrincipal chip when analise exists', () => {
    const registro = makeEnrichedRegistro({ _id: '1' }, { emotionPrincipal: 'alegria' })
    render(<RegistroCard registro={registro} />)
    expect(screen.getByText('alegria')).toBeDefined()
  })

  it('shows emotionPrincipal chip with correct color class', () => {
    const registro = makeEnrichedRegistro({ _id: '1' }, { emotionPrincipal: 'alegria' })
    render(<RegistroCard registro={registro} />)
    const chip = screen.getByText('alegria')
    expect(chip.className).toContain('bg-emerald-100')
    expect(chip.className).toContain('text-emerald-700')
  })

  it('shows secondary emotion chips', () => {
    const registro = makeEnrichedRegistro({ _id: '1' })
    render(<RegistroCard registro={registro} />)
    // ansiedade appears in both heading (sentimentoNome) and chip — use getAllByText
    const ansiedadeElements = screen.getAllByText('ansiedade')
    expect(ansiedadeElements.length).toBeGreaterThanOrEqual(2)
    // medo only appears as secondary chip
    expect(screen.getByText('medo')).toBeDefined()
  })

  it('renders no chips when analise is undefined', () => {
    const registro = makeRegistro({ _id: '1', sentimentoNome: 'outro' }) as RegistroWithAnalise
    render(<RegistroCard registro={registro} />)
    expect(screen.queryByText('medo')).toBeNull()
  })

  it('shows chips in collapsed state', () => {
    const registro = makeEnrichedRegistro({ _id: '1' }, { emotionPrincipal: 'gratidão' })
    render(<RegistroCard registro={registro} />)
    // Card starts collapsed by default
    expect(screen.getByText('gratidão')).toBeDefined()
  })
})
