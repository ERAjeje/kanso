import { useState, useEffect, useRef } from 'react'
import { FileText, Loader, AlertCircle, CheckCircle, Download } from 'lucide-react'
import { createReport, getReportStatus, getReportsList, downloadReport } from '../services/reports'
import type { ReportJob } from '../types'
import { format } from 'date-fns'

type State = 'idle' | 'generating' | 'polling' | 'completed' | 'error'

export function ReportSection() {
  const [reports, setReports] = useState<ReportJob[]>([])
  const [state, setState] = useState<State>('idle')
  const [latestReport, setLatestReport] = useState<ReportJob | null>(null)
  const [errorMsg, setErrorMsg] = useState('')
  const jobIdRef = useRef<string | null>(null)

  useEffect(() => {
    getReportsList()
      .then(setReports)
      .catch(() => {})
  }, [])

  useEffect(() => {
    if (state !== 'polling' || !jobIdRef.current) return

    const id = setInterval(async () => {
      try {
        const job = await getReportStatus(jobIdRef.current!)
        if (job.status === 'done') {
          setState('completed')
          setLatestReport(job)
          setReports((prev) => [...prev.filter((r) => r._id !== job._id), job])
          jobIdRef.current = null
        } else if (job.status === 'failed') {
          setState('error')
          setErrorMsg(job.errorMessage || 'Falha ao gerar relatório')
          jobIdRef.current = null
        }
      } catch {
        setState('error')
        setErrorMsg('Erro ao verificar status do relatório')
        jobIdRef.current = null
      }
    }, 3000)

    return () => clearInterval(id)
  }, [state])

  const handleGenerate = async () => {
    setState('generating')
    setErrorMsg('')
    try {
      const { jobId } = await createReport()
      jobIdRef.current = jobId
      setState('polling')
    } catch {
      setState('error')
      setErrorMsg('Não foi possível iniciar a geração do relatório')
      jobIdRef.current = null
    }
  }

  const handleDownload = (id: string) => {
    downloadReport(id).catch(() => {
      setErrorMsg('Erro ao baixar relatório')
      setState('error')
    })
  }

  const handleReset = () => {
    setState('idle')
    setLatestReport(null)
    setErrorMsg('')
    jobIdRef.current = null
  }

  const formatPeriod = (job: ReportJob): string => {
    if (!job.periodStart) return 'todos os registros até hoje'
    try {
      const inicio = format(new Date(job.periodStart), 'dd/MM/yyyy')
      return `desde ${inicio} até hoje`
    } catch {
      return 'todos os registros até hoje'
    }
  }

  return (
    <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
      <div className="flex items-center gap-2 mb-4">
        <FileText className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold text-gray-800">Relatórios</h2>
      </div>

      {/* State: Idle */}
      {state === 'idle' && (
        <>
          <button
            onClick={handleGenerate}
            className="w-full bg-primary text-white rounded-lg py-3 font-medium hover:brightness-110 disabled:opacity-50 transition-all"
          >
            Gerar Relatório
          </button>
          {reports.length === 0 && (
            <p className="text-gray-400 text-sm text-center mt-4">
              Nenhum relatório gerado ainda
            </p>
          )}
        </>
      )}

      {/* State: Generating */}
      {state === 'generating' && (
        <button
          disabled
          className="w-full bg-primary text-white rounded-lg py-3 font-medium opacity-50 cursor-not-allowed transition-all flex items-center justify-center gap-2"
        >
          <Loader className="h-5 w-5 animate-spin" />
          Gerando...
        </button>
      )}

      {/* State: Polling */}
      {state === 'polling' && (
        <div className="space-y-3">
          <button
            disabled
            className="w-full bg-primary text-white rounded-lg py-3 font-medium opacity-50 cursor-not-allowed transition-all flex items-center justify-center gap-2"
          >
            <Loader className="h-5 w-5 animate-spin" />
            Gerando...
          </button>
          <p className="text-gray-500 text-sm text-center flex items-center justify-center gap-2">
            <Loader className="h-4 w-4 animate-spin" />
            Relatório está sendo processado...
          </p>
        </div>
      )}

      {/* State: Completed */}
      {state === 'completed' && latestReport && (
        <div className="space-y-4">
          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-center gap-2 text-green-700 font-medium mb-1">
              <CheckCircle className="h-5 w-5" />
              Relatório pronto!
            </div>
            <p className="text-green-600 text-sm">
              {formatPeriod(latestReport)}
            </p>
          </div>

          <button
            onClick={() => handleDownload(latestReport._id)}
            className="w-full bg-primary text-white rounded-lg py-3 font-medium hover:brightness-110 transition-all flex items-center justify-center gap-2"
          >
            <Download className="h-5 w-5" />
            Baixar PDF
          </button>

          <button
            onClick={handleReset}
            className="w-full text-gray-500 text-sm hover:text-gray-700 transition-colors"
          >
            Gerar novo relatório
          </button>
        </div>
      )}

      {/* State: Error */}
      {state === 'error' && (
        <div className="space-y-4">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center gap-2 text-red-700 font-medium mb-1">
              <AlertCircle className="h-5 w-5" />
              Erro
            </div>
            <p className="text-red-600 text-sm">{errorMsg}</p>
          </div>

          <button
            onClick={handleReset}
            className="w-full bg-primary text-white rounded-lg py-3 font-medium hover:brightness-110 transition-all"
          >
            Tentar novamente
          </button>
        </div>
      )}

      {/* Previous reports list */}
      {reports.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-100">
          <h3 className="text-sm font-medium text-gray-500 mb-2">
            Relatórios anteriores
          </h3>
          <ul className="space-y-2">
            {reports
              .filter((r) => r._id !== (state === 'completed' ? latestReport?._id : undefined))
              .map((report) => (
                <li
                  key={report._id}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="text-gray-600">
                    {formatPeriod(report)}
                  </span>
                  {report.status === 'done' && (
                    <button
                      onClick={() => handleDownload(report._id)}
                      className="text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                    >
                      <Download className="h-4 w-4" />
                      Baixar
                    </button>
                  )}
                </li>
              ))}
          </ul>
        </div>
      )}
    </div>
  )
}
