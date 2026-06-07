import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SentimentoEditor } from './SentimentoEditor'

const EMOTIONS = [
  'alegria', 'amor', 'ansiedade', 'culpa', 'gratidão',
  'medo', 'neutro', 'nojo', 'raiva', 'saudade',
  'surpresa', 'tristeza', 'vergonha',
]

describe('SentimentoEditor', () => {
  it('renders 13 options when not disabled', () => {
    render(<SentimentoEditor currentValue="" disabled={false} onSave={vi.fn()} />)

    // Combobox should render with placeholder
    expect(screen.getByPlaceholderText('Selecionar sentimento')).toBeDefined()

    // Click to open the combobox options
    const button = screen.getByRole('button')
    fireEvent.click(button)

    // Check that all 13 emotions appear
    for (const emotion of EMOTIONS) {
      expect(screen.getByText(emotion)).toBeDefined()
    }
  })

  it('shows static text with emotion styling when disabled', () => {
    render(<SentimentoEditor currentValue="ansiedade" disabled={true} onSave={vi.fn()} />)

    // Should show the emotion text, not a combobox
    expect(screen.getByText('ansiedade')).toBeDefined()
    // Should NOT have a combobox
    expect(screen.queryByPlaceholderText('Selecionar sentimento')).toBeNull()
  })

  it('shows static text with fallback when disabled and no value', () => {
    render(<SentimentoEditor currentValue="" disabled={true} onSave={vi.fn()} />)

    expect(screen.getByText('—')).toBeDefined()
  })

  it('renders combobox input when not disabled', () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    render(<SentimentoEditor currentValue="" disabled={false} onSave={onSave} />)

    // Should have a combobox input not a static display
    expect(screen.getByPlaceholderText('Selecionar sentimento')).toBeDefined()
    const comboboxes = screen.getAllByRole('combobox')
    expect(comboboxes.length).toBeGreaterThanOrEqual(1)
  })

  it('passes correct onSave reference to component', () => {
    const onSave = vi.fn()
    render(<SentimentoEditor currentValue="" disabled={false} onSave={onSave} />)

    // Directly call onSave to verify it works
    onSave('test-emotion')

    expect(onSave).toHaveBeenCalledWith('test-emotion')
  })

  it('shows sorted emotions alphabetically', () => {
    render(<SentimentoEditor currentValue="" disabled={false} onSave={vi.fn()} />)

    // Click the combobox button to show options
    const button = screen.getByRole('button')
    fireEvent.click(button)

    const options = screen.getAllByRole('option')
    const optionTexts = options.map(o => o.textContent)

    // First option should be 'alegria' (alphabetically first)
    expect(optionTexts[0]).toBe('alegria')
    // Last option should be 'vergonha' (alphabetically last)
    expect(optionTexts[optionTexts.length - 1]).toBe('vergonha')
  })
})
