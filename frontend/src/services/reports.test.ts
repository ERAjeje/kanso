import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockAuthenticatedFetch = vi.fn()

vi.mock('./auth', () => ({
  authenticatedFetch: mockAuthenticatedFetch,
}))

const { createReport, getReportStatus, getReportsList, getDownloadUrl } = await import('./reports')

describe('reports API service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('createReport', () => {
    it('POSTs to /api/reports and returns the report job', async () => {
      const mockJob = { _id: '1', status: 'pending' }
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockJob),
      })

      const result = await createReport()

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://kanso.local/api/reports',
        { method: 'POST' }
      )
      expect(result).toEqual(mockJob)
    })

    it('throws on non-ok response', async () => {
      mockAuthenticatedFetch.mockResolvedValue({ ok: false })

      await expect(createReport()).rejects.toThrow('Falha ao criar relatório')
    })
  })

  describe('getReportStatus', () => {
    it('GETs /api/reports/{id} and returns the report job', async () => {
      const mockJob = { _id: 'abc-123', status: 'completed' }
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockJob),
      })

      const result = await getReportStatus('abc-123')

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://kanso.local/api/reports/abc-123'
      )
      expect(result).toEqual(mockJob)
    })

    it('encodes the ID in the URL', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({}),
      })

      await getReportStatus('special/id+?')

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://kanso.local/api/reports/special%2Fid%2B%3F'
      )
    })

    it('throws on non-ok response', async () => {
      mockAuthenticatedFetch.mockResolvedValue({ ok: false })

      await expect(getReportStatus('1')).rejects.toThrow('Falha ao buscar status do relatório')
    })
  })

  describe('getReportsList', () => {
    it('GETs /api/reports and returns the reports array', async () => {
      const mockList = [{ _id: '1' }, { _id: '2' }]
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockList),
      })

      const result = await getReportsList()

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://kanso.local/api/reports'
      )
      expect(result).toEqual(mockList)
    })

    it('throws on non-ok response', async () => {
      mockAuthenticatedFetch.mockResolvedValue({ ok: false })

      await expect(getReportsList()).rejects.toThrow('Falha ao buscar lista de relatórios')
    })
  })

  describe('getDownloadUrl', () => {
    it('returns the correct download URL for a given ID', () => {
      const url = getDownloadUrl('abc-123')
      expect(url).toBe('https://kanso.local/api/reports/abc-123/download')
    })

    it('encodes special characters in the ID', () => {
      const url = getDownloadUrl('id/with+slashes')
      expect(url).toBe('https://kanso.local/api/reports/id%2Fwith%2Bslashes/download')
    })
  })
})
