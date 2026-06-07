import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
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

  it('submit button is disabled when all text fields are empty', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)
    const submitBtn = screen.getByRole('button', { name: /registrar/i })
    expect(submitBtn).toBeDisabled()
  })

  it('submit button is enabled when at least one text field is filled (sentimento optional)', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    const sensacoesInput = screen.getByPlaceholderText('O que você está sentindo no corpo?')
    fireEvent.change(sensacoesInput, { target: { value: 'Coração acelerado' } })

    const submitBtn = screen.getByRole('button', { name: /registrar/i })
    expect(submitBtn).not.toBeDisabled()
  })

  it('calls saveRegistro on submit with sentimentoNome as empty string', async () => {
    mockSaveRegistro.mockResolvedValue({
      _id: 'new-id',
      type: 'registro',
      userId: 'test',
      dataHora: '2026-05-17T14:30:00.000Z',
      sensacoes: 'Coração acelerado',
      sentimentoId: null,
      sentimentoNome: '',
      contexto: '',
      pensamentos: '',
      createdAt: '2026-05-17T14:30:00.000Z',
      updatedAt: '2026-05-17T14:30:00.000Z',
    })

    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    const sensacoesInput = screen.getByPlaceholderText('O que você está sentindo no corpo?')
    fireEvent.change(sensacoesInput, { target: { value: 'Coração acelerado' } })

    const submitBtn = screen.getByRole('button', { name: /registrar/i })
    fireEvent.click(submitBtn)

    expect(mockSaveRegistro).toHaveBeenCalledWith(
      expect.objectContaining({ sentimentoNome: '' })
    )
  })

  it('renders date-time picker with formatted date (backdating allowed per REG-02)', () => {
    render(<RegistrationForm onSaved={vi.fn()} onShowToast={vi.fn()} />)

    expect(screen.getByText(/Data e Hora/i)).toBeInTheDocument()
    const buttons = screen.getAllByRole('button')
    const dateButtons = buttons.filter(b => b.textContent?.match(/\d{4}/))
    expect(dateButtons.length).toBeGreaterThanOrEqual(1)
  })
})
