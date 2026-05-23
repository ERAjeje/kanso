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
  userSub: string
  status: 'pending' | 'processing' | 'done' | 'failed'
  createdAt?: string
  completedAt?: string
  periodStart?: string
  periodEnd?: string
  fileName?: string
  errorMessage?: string
}

export interface EmotionScore {
  emotion: string
  score: number
}

export interface AnaliseNlpDoc {
  _id: string
  _rev?: string
  type: 'analise_nlp'
  userId: string
  registroId: string
  emotionPrincipal: string
  emotions: EmotionScore[]
  scores: Record<string, number>
  intensidade: number
  modeloVersao: string
  analisadoEm: string
}

export interface RegistroWithAnalise extends RegistroDoc {
  analise?: AnaliseNlpDoc
}
