import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { InstallBanner } from './InstallBanner'

describe('InstallBanner', () => {
  it('renders nothing when canInstall is false', () => {
    const { container } = render(
      <InstallBanner canInstall={false} triggerInstall={vi.fn()} userDismissed={vi.fn()} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders banner with install text when canInstall is true', () => {
    render(
      <InstallBanner canInstall={true} triggerInstall={vi.fn()} userDismissed={vi.fn()} />
    )
    expect(screen.getByText('Instale o Kanso para uma experiência melhor')).toBeDefined()
    expect(screen.getByText('Instalar App')).toBeDefined()
    expect(screen.getByText('Agora não')).toBeDefined()
  })

  it('calls triggerInstall when Install button is clicked', () => {
    const triggerInstall = vi.fn()
    render(
      <InstallBanner canInstall={true} triggerInstall={triggerInstall} userDismissed={vi.fn()} />
    )
    fireEvent.click(screen.getByText('Instalar App'))
    expect(triggerInstall).toHaveBeenCalledOnce()
  })

  it('calls userDismissed when Agora não button is clicked', () => {
    const userDismissed = vi.fn()
    render(
      <InstallBanner canInstall={true} triggerInstall={vi.fn()} userDismissed={userDismissed} />
    )
    fireEvent.click(screen.getByText('Agora não'))
    expect(userDismissed).toHaveBeenCalledOnce()
  })
})
