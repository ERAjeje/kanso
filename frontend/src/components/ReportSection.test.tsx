import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { ReportSection } from './ReportSection'

const mockCreateReport = vi.fn()
const mockGetReportStatus = vi.fn()
const mockGetReportsList = vi.fn()
const mockDownloadReport = vi.fn()

vi.mock('../services/reports', () => ({
  createReport: (...args: unknown[]) => mockCreateReport(...args),
  getReportStatus: (...args: unknown[]) => mockGetReportStatus(...args),
  getReportsList: (...args: unknown[]) => mockGetReportsList(...args),
  downloadReport: (...args: unknown[]) => mockDownloadReport(...args),
}))

describe('ReportSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetReportsList.mockResolvedValue([])
    mockDownloadReport.mockResolvedValue(undefined)
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
          userSub: 'u1',
          status: 'done',
          createdAt: '2026-05-15T10:00:00.000Z',
          periodStart: '2026-01-01T00:00:00.000Z',
          periodEnd: '2026-05-15T10:00:00.000Z',
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
    it('shows "desde DD/MM/AAAA até hoje" for reports with periodStart', async () => {
      mockGetReportsList.mockResolvedValue([
        {
          _id: '1',
          type: 'relatorio',
          userSub: 'u1',
          status: 'done',
          createdAt: '2026-05-16T20:00:00.000Z',
          completedAt: '2026-05-16T20:01:00.000Z',
          periodStart: '2026-01-01T00:00:00.000Z',
          periodEnd: '2026-05-16T20:00:00.000Z',
        },
      ])

      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText(/desde 31\/12\/2025 até hoje/)).toBeDefined()
      })
    })

    it('falls back to "todos os registros" when periodStart is missing', async () => {
      mockGetReportsList.mockResolvedValue([
        {
          _id: '1',
          type: 'relatorio',
          userSub: 'u1',
          status: 'done',
          createdAt: '2026-05-16T20:00:00.000Z',
          completedAt: '2026-05-16T20:01:00.000Z',
          periodStart: '',
          periodEnd: '2026-05-16T20:00:00.000Z',
        },
      ])

      render(<ReportSection />)

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
          userSub: 'u1',
          status: 'done',
          createdAt: '2026-05-10T10:00:00.000Z',
          periodStart: '2026-01-01T00:00:00.000Z',
          periodEnd: '2026-05-10T10:00:00.000Z',
        },
      ])

      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText('Relatórios anteriores')).toBeDefined()
      })
    })

    it('shows download button for done reports', async () => {
      mockGetReportsList.mockResolvedValue([
        {
          _id: 'prev-1',
          type: 'relatorio',
          userSub: 'u1',
          status: 'done',
          createdAt: '2026-05-10T10:00:00.000Z',
          periodStart: '2026-01-01T00:00:00.000Z',
          periodEnd: '2026-05-10T10:00:00.000Z',
        },
      ])
      render(<ReportSection />)

      await waitFor(() => {
        expect(screen.getByText('Baixar')).toBeDefined()
      })
    })
  })
})
