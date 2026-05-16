import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { ReportSection } from './ReportSection'

const mockCreateReport = vi.fn()
const mockGetReportStatus = vi.fn()
const mockGetReportsList = vi.fn()
const mockGetDownloadUrl = vi.fn()

vi.mock('../services/reports', () => ({
  createReport: (...args: unknown[]) => mockCreateReport(...args),
  getReportStatus: (...args: unknown[]) => mockGetReportStatus(...args),
  getReportsList: (...args: unknown[]) => mockGetReportsList(...args),
  getDownloadUrl: (...args: unknown[]) => mockGetDownloadUrl(...args),
}))

describe('ReportSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetReportsList.mockResolvedValue([])
    mockGetDownloadUrl.mockImplementation((id: string) => `/download/${id}`)
  })

  it('renders the heading', async () => {
    render(<ReportSection />)
    await waitFor(() => {
      expect(screen.getByText('Relatórios')).toBeDefined()
    })
  })

  describe('idle state', () => {
    it('shows the generate button and empty state text when no reports exist', async () => {
      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })
      expect(screen.getByText('Nenhum relatório gerado ainda')).toBeDefined()
    })

    it('hides empty state text when previous reports exist', async () => {
      mockGetReportsList.mockResolvedValue([
        {
          _id: 'prev-1',
          type: 'relatorio',
          status: 'completed',
          userId: 'u1',
          requestedAt: '2026-05-15T10:00:00.000Z',
          periodoInicio: '2026-01-01T00:00:00.000Z',
          periodoFim: '2026-05-15T10:00:00.000Z',
          downloadUrl: '/download/prev-1',
        },
      ])

      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })
      expect(screen.queryByText('Nenhum relatório gerado ainda')).toBeNull()
    })
  })

  describe('generating state', () => {
    it('shows loader while report is being created', async () => {
      mockCreateReport.mockImplementation(
        () => new Promise(() => {}) // never resolves
      )

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      expect(await screen.findByText('Gerando...')).toBeDefined()
    })
  })

  describe('completed state', () => {
    it('shows success card with download button when report completes immediately', async () => {
      const completedJob = {
        _id: '1',
        type: 'relatorio' as const,
        userId: 'u1',
        status: 'completed' as const,
        requestedAt: '2026-05-16T20:00:00.000Z',
        completedAt: '2026-05-16T20:01:00.000Z',
        periodoInicio: '2026-01-01T00:00:00.000Z',
        periodoFim: '2026-05-16T20:00:00.000Z',
        totalRegistros: 15,
        downloadUrl: '/download/1',
      }
      mockCreateReport.mockResolvedValue(completedJob)
      mockGetDownloadUrl.mockReturnValue('/download/1')

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      await waitFor(() => {
        expect(screen.getByText('Relatório pronto!')).toBeDefined()
      })
      expect(screen.getByText('15 registros incluídos')).toBeDefined()
      expect(screen.getByText('Baixar PDF')).toBeDefined()
    })

    it('shows singular "registro" when totalRegistros is 1', async () => {
      const job = {
        _id: '2',
        type: 'relatorio' as const,
        userId: 'u1',
        status: 'completed' as const,
        requestedAt: '2026-05-16T20:00:00.000Z',
        completedAt: '2026-05-16T20:01:00.000Z',
        periodoInicio: '2026-01-01T00:00:00.000Z',
        periodoFim: '2026-05-16T20:00:00.000Z',
        totalRegistros: 1,
        downloadUrl: '/download/2',
      }
      mockCreateReport.mockResolvedValue(job)
      mockGetDownloadUrl.mockReturnValue('/download/2')

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      await waitFor(() => {
        expect(screen.getByText('1 registro incluído')).toBeDefined()
      })
    })

    it('allows generating another report after completion', async () => {
      const completedJob = {
        _id: '1',
        type: 'relatorio' as const,
        userId: 'u1',
        status: 'completed' as const,
        requestedAt: '2026-05-16T20:00:00.000Z',
        completedAt: '2026-05-16T20:01:00.000Z',
        periodoInicio: '2026-01-01T00:00:00.000Z',
        periodoFim: '2026-05-16T20:00:00.000Z',
        totalRegistros: 5,
        downloadUrl: '/download/1',
      }
      mockCreateReport.mockResolvedValue(completedJob)
      mockGetDownloadUrl.mockReturnValue('/download/1')

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))
      await waitFor(() => {
        expect(screen.getByText('Relatório pronto!')).toBeDefined()
      })

      // Click "Gerar novo relatório" to reset
      fireEvent.click(screen.getByText('Gerar novo relatório'))
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })
    })
  })

  describe('error state', () => {
    it('shows error card when createReport throws', async () => {
      mockCreateReport.mockRejectedValue(new Error('Network error'))

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      await waitFor(() => {
        expect(screen.getByText('Erro')).toBeDefined()
      })
      expect(
        screen.getByText('Não foi possível iniciar a geração do relatório')
      ).toBeDefined()
    })

    it('shows retry button and resets to idle on click', async () => {
      mockCreateReport.mockRejectedValue(new Error('Network error'))

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))
      await waitFor(() => {
        expect(screen.getByText('Erro')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Tentar novamente'))

      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })
    })
  })

  describe('period formatting', () => {
    it('shows "desde DD/MM/AAAA até hoje" for completed reports', async () => {
      const job = {
        _id: '1',
        type: 'relatorio' as const,
        userId: 'u1',
        status: 'completed' as const,
        requestedAt: '2026-05-16T20:00:00.000Z',
        completedAt: '2026-05-16T20:01:00.000Z',
        periodoInicio: '2026-03-15T12:00:00.000Z',
        periodoFim: '2026-05-16T20:00:00.000Z',
        totalRegistros: 10,
        downloadUrl: '/download/1',
      }
      mockCreateReport.mockResolvedValue(job)
      mockGetDownloadUrl.mockReturnValue('/download/1')

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      await waitFor(() => {
        expect(screen.getByText(/desde 15\/03\/2026 até hoje/)).toBeDefined()
      })
    })

    it('falls back to "todos os registros" when periodoInicio is missing', async () => {
      const job = {
        _id: '1',
        type: 'relatorio' as const,
        userId: 'u1',
        status: 'completed' as const,
        requestedAt: '2026-05-16T20:00:00.000Z',
        completedAt: '2026-05-16T20:01:00.000Z',
        periodoInicio: '',
        periodoFim: '2026-05-16T20:00:00.000Z',
        downloadUrl: '/download/1',
      }
      mockCreateReport.mockResolvedValue(job)
      mockGetDownloadUrl.mockReturnValue('/download/1')

      render(<ReportSection />)
      await waitFor(() => {
        expect(screen.getByText('Gerar Relatório')).toBeDefined()
      })

      fireEvent.click(screen.getByText('Gerar Relatório'))

      await waitFor(() => {
        expect(screen.getByText(/todos os registros até hoje/)).toBeDefined()
      })
    })
  })

  describe('previous reports list', () => {
    it('shows previous reports when they exist', async () => {
      mockGetReportsList.mockResolvedValue([
        {
          _id: 'prev-1',
          type: 'relatorio',
          status: 'completed',
          userId: 'u1',
          requestedAt: '2026-05-10T10:00:00.000Z',
          periodoInicio: '2026-01-01T00:00:00.000Z',
          periodoFim: '2026-05-10T10:00:00.000Z',
          downloadUrl: '/download/prev-1',
        },
      ])

      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText('Relatórios anteriores')).toBeDefined()
      })
    })
  })
})
