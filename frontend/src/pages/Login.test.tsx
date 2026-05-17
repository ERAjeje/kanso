import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { Login } from './Login'

const mockSignIn = vi.fn()
const mockNavigate = vi.fn()

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ signIn: mockSignIn, user: null }),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

beforeEach(() => {
  vi.clearAllMocks()
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

function renderLogin() {
  return render(
    <BrowserRouter>
      <Login />
    </BrowserRouter>
  )
}

describe('Login', () => {
  it('renders the Kanso heading', () => {
    renderLogin()
    expect(screen.getByText('Kanso')).toBeDefined()
  })

  it('renders the subtitle Diário Emocional', () => {
    renderLogin()
    expect(screen.getByText('Diário Emocional')).toBeDefined()
  })

  it('renders the prompt text about Google login', () => {
    renderLogin()
    expect(screen.getByText('Faça login com sua conta Google para começar')).toBeDefined()
  })

  it('renders the Google sign-in button container', () => {
    renderLogin()
    expect(document.getElementById('google-signin-btn')).toBeDefined()
  })

  it('shows a fallback Google sign-in button after GIS timeout', () => {
    renderLogin()
    act(() => { vi.advanceTimersByTime(8500) })
    const fallbackBtn = screen.getByRole('button', { name: /entrar com google/i })
    expect(fallbackBtn).toBeDefined()
  })
})
