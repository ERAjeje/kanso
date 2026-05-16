import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
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
})
