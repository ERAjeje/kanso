import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockRegistroGet = vi.fn()
const mockRegistroPut = vi.fn()
const mockSentimentoPut = vi.fn()
const mockAllDocs = vi.fn()

vi.mock('../services/pouchdb', () => ({
  registrosDB: { allDocs: mockAllDocs, put: mockRegistroPut, get: mockRegistroGet },
  sentimentosDB: { allDocs: mockAllDocs, put: mockSentimentoPut },
  getUserId: () => 'test-user-123',
}))

const { saveRegistro, saveSentimento, getSentimentos, getRegistros, updateRegistroSentimento } = await import('./registros')

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

describe('updateRegistroSentimento', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('updates sentimentoNome, sentimentoId, and updatedAt', async () => {
    const existingRegistro = {
      _id: 'reg-1',
      _rev: '1-abc',
      type: 'registro' as const,
      userId: 'u1',
      dataHora: '2026-05-17T14:30:00.000Z',
      sensacoes: 'coração acelerado',
      sentimentoId: null,
      sentimentoNome: '',
      contexto: 'reunião',
      pensamentos: 'nervoso',
      createdAt: '2026-05-17T14:30:00.000Z',
      updatedAt: '2026-05-17T14:30:00.000Z',
    }

    mockRegistroGet.mockResolvedValue(existingRegistro)
    mockRegistroPut.mockResolvedValue({ ok: true, id: 'reg-1', rev: '2-def' })

    const result = await updateRegistroSentimento(existingRegistro, 'ansiedade', 'label-ansiedade')

    expect(mockRegistroGet).toHaveBeenCalledWith('reg-1')
    expect(mockRegistroPut).toHaveBeenCalledWith(
      expect.objectContaining({
        _id: 'reg-1',
        sentimentoNome: 'ansiedade',
        sentimentoId: 'label-ansiedade',
      })
    )
    expect(result.sentimentoNome).toBe('ansiedade')
    expect(result.sentimentoId).toBe('label-ansiedade')
    expect(result.updatedAt).toBeDefined()
    expect(result.updatedAt).not.toBe(existingRegistro.updatedAt)
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

describe('getRegistros with analise merge', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('merges analise_nlp docs with matching registros', async () => {
    mockAllDocs
      .mockResolvedValueOnce({
        rows: [
          { doc: { _id: 'reg1', type: 'registro', userId: 'u1', dataHora: '2026-05-23T10:00:00Z', sensacoes: 'coração acelerado', sentimentoNome: 'ansiedade', contexto: 'reunião', pensamentos: 'nervoso', createdAt: '', updatedAt: '' } },
          { doc: { _id: 'reg2', type: 'registro', userId: 'u1', dataHora: '2026-05-23T09:00:00Z', sensacoes: 'tranquilo', sentimentoNome: 'calma', contexto: 'casa', pensamentos: 'relaxado', createdAt: '', updatedAt: '' } },
        ],
      })
      .mockResolvedValueOnce({
        rows: [
          { doc: { _id: 'analise:reg1', type: 'analise_nlp', userId: 'u1', registroId: 'reg1', emotionPrincipal: 'ansiedade', emotions: [{ emotion: 'ansiedade', score: 0.85 }], scores: { ansiedade: 0.85 }, intensidade: 0.85, modeloVersao: 'v1.0', analisadoEm: '2026-05-23T10:01:00Z' } },
        ],
      })

    const result = await getRegistros()
    expect(result).toHaveLength(2)
    const reg1 = result.find(r => r._id === 'reg1')
    expect(reg1?.analise).toBeDefined()
    expect(reg1?.analise?.emotionPrincipal).toBe('ansiedade')
    const reg2 = result.find(r => r._id === 'reg2')
    expect(reg2?.analise).toBeUndefined()
  })

  it('returns undefined analise when no analise_nlp docs exist', async () => {
    mockAllDocs
      .mockResolvedValueOnce({
        rows: [
          { doc: { _id: 'reg1', type: 'registro', userId: 'u1', dataHora: '2026-05-23T10:00:00Z', sensacoes: 'teste', sentimentoNome: 'neutro', contexto: '', pensamentos: '', createdAt: '', updatedAt: '' } },
        ],
      })
      .mockResolvedValueOnce({ rows: [] })

    const result = await getRegistros()
    expect(result).toHaveLength(1)
    expect(result[0].analise).toBeUndefined()
  })

  it('ignores analise_nlp docs that do not match any registro', async () => {
    mockAllDocs
      .mockResolvedValueOnce({
        rows: [
          { doc: { _id: 'reg1', type: 'registro', userId: 'u1', dataHora: '2026-05-23T10:00:00Z', sensacoes: 'teste', sentimentoNome: 'neutro', contexto: '', pensamentos: '', createdAt: '', updatedAt: '' } },
        ],
      })
      .mockResolvedValueOnce({
        rows: [
          { doc: { _id: 'analise:other', type: 'analise_nlp', userId: 'u1', registroId: 'other', emotionPrincipal: 'alegria', emotions: [{ emotion: 'alegria', score: 0.9 }], scores: { alegria: 0.9 }, intensidade: 0.9, modeloVersao: 'v1.0', analisadoEm: '2026-05-23T10:01:00Z' } },
        ],
      })

    const result = await getRegistros()
    expect(result).toHaveLength(1)
    expect(result[0].analise).toBeUndefined()
  })
})
