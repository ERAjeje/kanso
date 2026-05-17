import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Profile } from './Profile'

vi.mock('../components/ReportSection', () => ({
  ReportSection: () => <div data-testid="report-section">Report Section</div>,
}))

const mockGetPrefs = vi.fn()
const mockUpdatePrefs = vi.fn()

vi.mock('../services/push', () => ({
  getPreferences: (...args: unknown[]) => mockGetPrefs(...args),
  savePreferences: (...args: unknown[]) => mockUpdatePrefs(...args),
}))

describe('Profile', () => {
  beforeEach(() => {
    mockGetPrefs.mockResolvedValue({ enabled: true, times: ['12:00', '18:00', '23:00'], timezone: '' })
  })

  it('renders the page heading', async () => {
    render(<Profile />)
    await waitFor(() => expect(screen.getByText('Perfil')).toBeDefined())
  })

  it('renders the ReportSection component', async () => {
    render(<Profile />)
    await waitFor(() => expect(screen.getByTestId('report-section')).toBeDefined())
  })

  it('renders notification settings section', async () => {
    render(<Profile />)
    await waitFor(() => expect(screen.getByText('Lembretes')).toBeDefined())
  })

  it('shows permission status', async () => {
    render(<Profile />)
    await waitFor(() => expect(screen.getByText(/Status da permissão/)).toBeDefined())
  })

  it('shows time inputs when enabled', async () => {
    render(<Profile />)
    await waitFor(() => {
      const inputs = screen.getAllByDisplayValue('12:00')
      expect(inputs.length).toBeGreaterThanOrEqual(1)
    })
  })
})
