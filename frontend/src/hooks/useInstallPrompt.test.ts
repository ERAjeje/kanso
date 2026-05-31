import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useInstallPrompt } from './useInstallPrompt'

type UserChoice = { outcome: 'accepted' | 'dismissed' }

function createBeforeInstallPromptEvent(outcome: UserChoice['outcome']) {
  const evt = new Event('beforeinstallprompt') as any
  evt.preventDevault = vi.fn()
  evt.prompt = vi.fn().mockResolvedValue(undefined)
  evt.userChoice = Promise.resolve<UserChoice>({ outcome })
  return evt
}

describe('useInstallPrompt', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('starts with canInstall false and no deferredPrompt', () => {
    const { result } = renderHook(() => useInstallPrompt())
    expect(result.current.canInstall).toBe(false)
  })

  it('sets canInstall true when beforeinstallprompt fires', () => {
    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(true)
  })

  it('calls preventDefault on the event', () => {
    renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')
    const spy = vi.spyOn(evt, 'preventDefault')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(spy).toHaveBeenCalled()
  })

  it('triggerInstall calls prompt() on deferred event', async () => {
    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')
    const promptSpy = vi.spyOn(evt, 'prompt')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(true)

    await act(async () => {
      await result.current.triggerInstall()
    })

    expect(promptSpy).toHaveBeenCalled()
  })

  it('sets canInstall false after install is accepted', async () => {
    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(true)

    await act(async () => {
      await result.current.triggerInstall()
    })

    expect(result.current.canInstall).toBe(false)
  })

  it('sets canInstall false after install is dismissed', async () => {
    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('dismissed')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(true)

    await act(async () => {
      await result.current.triggerInstall()
    })

    expect(result.current.canInstall).toBe(false)
  })

  it('sets dismissed to true when userDismiss called', () => {
    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(true)

    act(() => {
      result.current.userDismissed()
    })

    expect(result.current.canInstall).toBe(false)
    expect(sessionStorage.getItem('pwa_install_dismissed')).toBe('true')
  })

  it('returns canInstall false if previously dismissed in session', () => {
    sessionStorage.setItem('pwa_install_dismissed', 'true')

    const { result } = renderHook(() => useInstallPrompt())
    const evt = createBeforeInstallPromptEvent('accepted')

    act(() => {
      window.dispatchEvent(evt)
    })

    expect(result.current.canInstall).toBe(false)
  })

  it('handles missing beforeinstallprompt gracefully', () => {
    const { result } = renderHook(() => useInstallPrompt())
    expect(result.current.canInstall).toBe(false)
    expect(result.current.triggerInstall).toBeDefined()
    expect(result.current.userDismissed).toBeDefined()
  })
})
