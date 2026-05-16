export interface RegistroDoc {
  _id: string
  _rev?: string
  type: 'registro'
  userId: string
  dataHora: string
  sensacoes: string
  sentimentoId: string | null
  sentimentoNome: string
  contexto: string
  pensamentos: string
  createdAt: string
  updatedAt: string
}

export interface SentimentoDoc {
  _id: string
  _rev?: string
  type: 'sentimento'
  userId: string
  nome: string
  criadoEm: string
}

export type SyncState = 'online' | 'offline' | 'syncing'

export interface ReportJob {
  _id: string
  _rev?: string
  type: 'relatorio'
  userId: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  requestedAt: string
  completedAt?: string
  periodoInicio: string
  periodoFim: string
  totalRegistros?: number
  downloadUrl?: string
  errorMessage?: string
}
