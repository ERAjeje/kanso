import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { History } from './History'
import type { RegistroDoc } from '../types'

const mockGetRegistros = vi.fn()

vi.mock('../services/pouchdb', () => ({
  registrosDB: { get: vi.fn(), put: vi.fn() },
  sentimentosDB: { allDocs: vi.fn() },
  treinamentoDB: { put: vi.fn() },
  getUserId: () => 'test-user-123',
}))

vi.mock('../services/registros', () => ({
  getRegistros: (...args: any[]) => mockGetRegistros(...args),
  saveRegistro: vi.fn(),
  getSentimentos: vi.fn(),
  saveSentimento: vi.fn(),
}))

function makeRegistro(id: string, overrides: Partial<RegistroDoc> = {}): RegistroDoc {
  return {
    _id: id,
    _rev: '1-abc',
    type: 'registro',
    userId: 'u1',
    dataHora: '2026-05-17T14:30:00.000Z',
    sensacoes: 'Coração acelerado',
    sentimentoId: null,
    sentimentoNome: 'ansiedade',
    contexto: 'Reunião com o chefe',
    pensamentos: 'Será que fiz o suficiente?',
    createdAt: '2026-05-17T14:30:00.000Z',
    updatedAt: '2026-05-17T14:30:00.000Z',
    ...overrides,
  }
}

function renderHistory() {
  return render(
    <MemoryRouter>
      <History />
    </MemoryRouter>
  )
}

describe('History', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading state initially', () => {
    mockGetRegistros.mockReturnValue(new Promise(() => {}))
    renderHistory()
    expect(screen.getByText('Carregando...')).toBeDefined()
  })

  it('renders list of RegistroCards after data loads', async () => {
    mockGetRegistros.mockResolvedValue([
      makeRegistro('1', { sentimentoNome: 'alegria', dataHora: '2026-05-17T10:00:00.000Z' }),
      makeRegistro('2', { sentimentoNome: 'ansiedade', dataHora: '2026-05-16T14:00:00.000Z' }),
    ])

    renderHistory()

    await waitFor(() => {
      expect(screen.getByText('alegria')).toBeDefined()
    })
    expect(screen.getByText('ansiedade')).toBeDefined()
  })

  it('renders empty state when no registros', async () => {
    mockGetRegistros.mockResolvedValue([])

    renderHistory()

    await waitFor(() => {
      expect(screen.getByText('Nenhum registro ainda')).toBeDefined()
    })
  })

  it('renders error state when getRegistros fails', async () => {
    mockGetRegistros.mockRejectedValue(new Error('Falha ao carregar'))

    renderHistory()

    await waitFor(() => {
      expect(screen.getByText('Erro ao carregar registros')).toBeDefined()
    })
  })

  it('has a heading', () => {
    mockGetRegistros.mockReturnValue(new Promise(() => {}))
    renderHistory()
    expect(screen.getByText('Histórico')).toBeDefined()
  })
})
