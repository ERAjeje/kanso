import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockRegistroPut = vi.fn()
const mockSentimentoPut = vi.fn()
const mockAllDocs = vi.fn()

vi.mock('../services/pouchdb', () => ({
  registrosDB: { put: mockRegistroPut },
  sentimentosDB: { allDocs: mockAllDocs, put: mockSentimentoPut },
  getUserId: () => 'test-user-123',
}))

const { saveRegistro, saveSentimento, getSentimentos } = await import('./registros')

describe('saveRegistro', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('saves a registro document to PouchDB', async () => {
    mockRegistroPut.mockResolvedValue({ ok: true, id: 'abc-123', rev: '1-xxx' })
    const input = {
      dataHora: '2026-05-16T14:30:00.000Z',
      sensacoes: 'coração acelerado',
      sentimentoId: null,
      sentimentoNome: 'ansiedade',
      contexto: 'antes da reunião',
      pensamentos: 'será que vai dar certo?',
    }

    const doc = await saveRegistro(input)

    expect(mockRegistroPut).toHaveBeenCalledOnce()
    expect(mockRegistroPut).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'registro',
        userId: 'test-user-123',
        dataHora: '2026-05-16T14:30:00.000Z',
        sensacoes: 'coração acelerado',
        sentimentoNome: 'ansiedade',
      })
    )
    expect(doc._id).toBeDefined()
    expect(doc.type).toBe('registro')
    expect(doc.createdAt).toBeDefined()
    expect(doc.updatedAt).toBeDefined()
  })
})

describe('saveSentimento', () => {
  it('saves a normalized sentimento document', async () => {
    mockSentimentoPut.mockResolvedValue({ ok: true })
    await saveSentimento('  Ansiedade  ')
    expect(mockSentimentoPut).toHaveBeenCalledOnce()
    expect(mockSentimentoPut).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'sentimento',
        nome: 'ansiedade',
        userId: 'test-user-123',
      })
    )
  })
})

describe('getSentimentos', () => {
  it('returns sorted sentimentos from allDocs', async () => {
    mockAllDocs.mockResolvedValue({
      rows: [
        { doc: { _id: '1', type: 'sentimento', userId: 'u1', nome: 'tristeza', criadoEm: '' } },
        { doc: { _id: '2', type: 'sentimento', userId: 'u1', nome: 'alegria', criadoEm: '' } },
        { doc: { _id: '3', type: 'sentimento', userId: 'u1', nome: 'ansiedade', criadoEm: '' } },
        { doc: { _id: '4', type: 'other', userId: 'u1', nome: 'ignored', criadoEm: '' } },
      ],
    })

    const result = await getSentimentos()
    expect(result).toHaveLength(3)
    expect(result[0].nome).toBe('alegria')
    expect(result[1].nome).toBe('ansiedade')
    expect(result[2].nome).toBe('tristeza')
  })
})
