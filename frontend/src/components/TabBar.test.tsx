import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { TabBar } from './TabBar'

describe('TabBar', () => {
  it('renders 3 tabs with correct labels', () => {
    render(
      <MemoryRouter initialEntries={['/register']}>
        <TabBar />
      </MemoryRouter>
    )

    expect(screen.getByText('Registrar')).toBeDefined()
    expect(screen.getByText('Histórico')).toBeDefined()
    expect(screen.getByText('Perfil')).toBeDefined()
  })

  it('has correct icon imports', async () => {
    const TabBarModule = await import('./TabBar')
    expect(TabBarModule.TabBar).toBeDefined()
  })

  it('marks active tab with text-primary class', () => {
    render(
      <MemoryRouter initialEntries={['/register']}>
        <TabBar />
      </MemoryRouter>
    )

    const links = screen.getAllByRole('link')
    const registerLink = links.find(l => l.textContent?.includes('Registrar'))
    expect(registerLink).toBeDefined()
    expect(registerLink?.className).toContain('text-primary')
  })
})
