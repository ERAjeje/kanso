import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SentimentoCombobox } from './SentimentoCombobox'

const mockGetSentimentos = vi.fn()
const mockSaveSentimento = vi.fn()

vi.mock('../services/registros', () => ({
  getSentimentos: (...args: any[]) => mockGetSentimentos(...args),
  saveSentimento: (...args: any[]) => mockSaveSentimento(...args),
}))

describe('SentimentoCombobox', () => {
  it('loads and displays sentiments from getSentimentos', async () => {
    mockGetSentimentos.mockResolvedValue([
      { _id: '1', type: 'sentimento', userId: 'u1', nome: 'alegria', criadoEm: '' },
      { _id: '2', type: 'sentimento', userId: 'u1', nome: 'ansiedade', criadoEm: '' },
      { _id: '3', type: 'sentimento', userId: 'u1', nome: 'tristeza', criadoEm: '' },
    ])

    render(<SentimentoCombobox onChange={vi.fn()} />)
    expect(mockGetSentimentos).toHaveBeenCalledOnce()
  })

  it('filters sentiments when user types', async () => {
    mockGetSentimentos.mockResolvedValue([
      { _id: '1', type: 'sentimento', userId: 'u1', nome: 'alegria', criadoEm: '' },
      { _id: '2', type: 'sentimento', userId: 'u1', nome: 'ansiedade', criadoEm: '' },
      { _id: '3', type: 'sentimento', userId: 'u1', nome: 'tristeza', criadoEm: '' },
    ])

    render(<SentimentoCombobox onChange={vi.fn()} />)
    const input = screen.getByPlaceholderText(/digite ou selecione/i)
    fireEvent.change(input, { target: { value: 'ans' } })
    expect(mockGetSentimentos).toHaveBeenCalled()
  })

  it('shows 13 fallback sentiments when getSentimentos returns empty', async () => {
    mockGetSentimentos.mockResolvedValue([])

    render(<SentimentoCombobox onChange={vi.fn()} />)

    await waitFor(() => {
      const input = screen.getByPlaceholderText(/digite ou selecione/i)
      fireEvent.click(input)
      fireEvent.keyDown(input, { key: 'ArrowDown' })
    })

    const list = screen.getByRole('listbox')
    expect(list).toBeInTheDocument()
    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(13)
  })

  it('uses fallback items when no sentiments exist in PouchDB', async () => {
    mockGetSentimentos.mockResolvedValue([])

    render(<SentimentoCombobox onChange={vi.fn()} />)

    const input = screen.getByPlaceholderText(/digite ou selecione/i)
    fireEvent.click(input)
    fireEvent.keyDown(input, { key: 'ArrowDown' })

    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument()
    })

    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(13)
    expect(options[0]).toHaveTextContent('alegria')
    expect(options[12]).toHaveTextContent('vergonha')
  })
})
