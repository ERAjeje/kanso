import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockAuthenticatedFetch = vi.fn()

vi.mock('./auth', () => ({
  authenticatedFetch: mockAuthenticatedFetch,
}))

const { createReport, getReportStatus, getReportsList, getDownloadUrl, downloadReport } = await import('./reports')

describe('reports API service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('createReport', () => {
    it('POSTs to /api/reports and returns the jobId', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ jobId: 'abc-123' }),
      })

      const result = await createReport()

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        '/api/reports',
        { method: 'POST' }
      )
      expect(result).toEqual({ jobId: 'abc-123' })
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
        '/api/reports/abc-123'
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
        '/api/reports/special%2Fid%2B%3F'
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
        '/api/reports'
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
      expect(url).toBe('/api/reports/abc-123/download')
    })

    it('encodes special characters in the ID', () => {
      const url = getDownloadUrl('id/with+slashes')
      expect(url).toBe('/api/reports/id%2Fwith%2Bslashes/download')
    })
  })

  describe('downloadReport', () => {
    beforeEach(() => {
      // Mock Blob and URL.createObjectURL
      globalThis.URL.createObjectURL = vi.fn(() => 'blob:test')
      globalThis.URL.revokeObjectURL = vi.fn()
    })

    it('fetches the PDF with auth and triggers download', async () => {
      const blob = new Blob(['pdf-content'], { type: 'application/pdf' })
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        blob: () => Promise.resolve(blob),
      })

      const appendChild = vi.spyOn(document.body, 'appendChild')
      const removeChild = vi.spyOn(document.body, 'removeChild')

      await downloadReport('abc-123')

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        '/api/reports/abc-123/download'
      )
      expect(appendChild).toHaveBeenCalled()
      expect(removeChild).toHaveBeenCalled()
      expect(globalThis.URL.revokeObjectURL).toHaveBeenCalledWith('blob:test')
    })

    it('throws on non-ok response', async () => {
      mockAuthenticatedFetch.mockResolvedValue({ ok: false })

      await expect(downloadReport('1')).rejects.toThrow('Falha ao baixar relatório')
    })
  })
})
