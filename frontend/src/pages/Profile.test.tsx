import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Profile } from './Profile'

// Mock ReportSection since it imports the reports service internally
vi.mock('../components/ReportSection', () => ({
  ReportSection: () => <div data-testid="report-section">Report Section</div>,
}))

describe('Profile', () => {
  it('renders the page heading', () => {
    render(<Profile />)
    expect(screen.getByText('Perfil')).toBeDefined()
  })

  it('renders the ReportSection component', () => {
    render(<Profile />)
    expect(screen.getByTestId('report-section')).toBeDefined()
  })
})
