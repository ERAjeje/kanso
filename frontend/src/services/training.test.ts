import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockTreinamentoPut = vi.fn()
const mockAllDocs = vi.fn()

vi.mock('../services/pouchdb', () => ({
  treinamentoDB: { allDocs: mockAllDocs, put: mockTreinamentoPut },
  getUserId: () => 'test-user-123',
}))

const { saveTrainingExample, getTotalTrainingCount } = await import('./training')

describe('saveTrainingExample', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('saves a treinamento document to PouchDB', async () => {
    mockTreinamentoPut.mockResolvedValue({ ok: true, id: 'abc-123', rev: '1-xxx' })

    const doc = await saveTrainingExample('Estou muito feliz hoje!', 'alegria')

    expect(mockTreinamentoPut).toHaveBeenCalledOnce()
    expect(mockTreinamentoPut).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'treinamento',
        texto: 'Estou muito feliz hoje!',
        label: 'alegria',
        userId: 'test-user-123',
        origem: 'usuario',
      })
    )
    expect(doc._id).toBeDefined()
    expect(doc.type).toBe('treinamento')
    expect(doc.criadoEm).toBeDefined()
  })
})

describe('getTotalTrainingCount', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns total_rows from allDocs with limit 0', async () => {
    mockAllDocs.mockResolvedValue({ total_rows: 5, rows: [] })

    const count = await getTotalTrainingCount()

    expect(mockAllDocs).toHaveBeenCalledOnce()
    expect(mockAllDocs).toHaveBeenCalledWith({ limit: 0 })
    expect(count).toBe(5)
  })

  it('returns 0 when there are no training documents', async () => {
    mockAllDocs.mockResolvedValue({ total_rows: 0, rows: [] })

    const count = await getTotalTrainingCount()

    expect(count).toBe(0)
  })
})
