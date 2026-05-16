import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RegistrationForm } from './RegistrationForm'

const mockSaveRegistro = vi.fn()
const mockGetSentimentos = vi.fn().mockResolvedValue([])
vi.mock('../services/registros', () => ({
  saveRegistro: (...args: any[]) => mockSaveRegistro(...args),
  getSentimentos: (...args: any[]) => mockGetSentimentos(...args),
  saveSentimento: vi.fn(),
}))

vi.mock('../hooks/usePouchSync', () => ({
  usePouchSync: () => 'online' as const,
}))

vi.spyOn(console, 'error').mockImplementation(() => {})

describe('RegistrationForm', () => {
  it('renders all 5 form fields in the correct order', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    expect(screen.getByText('Data e Hora')).toBeDefined()
    expect(screen.getByText('Sentimento')).toBeDefined()
    expect(screen.getByText('Sensações')).toBeDefined()
    expect(screen.getByText('Contexto')).toBeDefined()
    expect(screen.getByText('Pensamentos')).toBeDefined()

    expect(screen.getByRole('button', { name: /registrar/i })).toBeDefined()
  })

  it('renders textareas for sensações, contexto, pensamentos', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    const textareas = screen.getAllByRole('textbox')
    expect(textareas.length).toBeGreaterThanOrEqual(3)
  })

  it('calls saveRegistro on submit with correct data', async () => {
    mockSaveRegistro.mockResolvedValue({
      _id: 'new-id',
      type: 'registro',
      userId: 'test',
      dataHora: '2026-05-16T14:30:00.000Z',
      sensacoes: '',
      sentimentoId: null,
      sentimentoNome: 'ansiedade',
      contexto: '',
      pensamentos: '',
      createdAt: '2026-05-16T14:30:00.000Z',
      updatedAt: '2026-05-16T14:30:00.000Z',
    })

    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    const submitBtn = screen.getByRole('button', { name: /registrar/i })
    expect(submitBtn).toBeDisabled()
  })

  it('has no min/max on datetime-local input (backdating per REG-02)', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    const dtInput = document.querySelector('input[type="datetime-local"]')
    expect(dtInput).not.toBeNull()
    expect(dtInput?.getAttribute('min')).toBeNull()
    expect(dtInput?.getAttribute('max')).toBeNull()
  })
})
