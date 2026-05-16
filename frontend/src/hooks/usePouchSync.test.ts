import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'

const mockOnSyncChange = vi.fn()

vi.mock('../services/pouchdb', () => ({
  onSyncChange: (...args: any[]) => mockOnSyncChange(...args),
}))

describe('usePouchSync', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(navigator, 'onLine', { value: true, configurable: true })
  })

  it('returns initial state based on navigator.onLine', async () => {
    const { usePouchSync } = await import('./usePouchSync')
    const { result } = renderHook(() => usePouchSync())
    expect(result.current).toBe('online')
  })

  it('subscribes to onSyncChange on mount', async () => {
    const { usePouchSync } = await import('./usePouchSync')
    renderHook(() => usePouchSync())
    expect(mockOnSyncChange).toHaveBeenCalledOnce()
  })

  it('unsubscribes on unmount', async () => {
    const mockUnsub = vi.fn()
    mockOnSyncChange.mockReturnValue(mockUnsub)
    const { usePouchSync } = await import('./usePouchSync')
    const { unmount } = renderHook(() => usePouchSync())
    unmount()
    expect(mockUnsub).toHaveBeenCalledOnce()
  })
})
