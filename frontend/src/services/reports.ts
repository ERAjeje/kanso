import { authenticatedFetch } from './auth'
import type { ReportJob } from '../types'

const API_BASE = 'https://kanso.local/api'

export async function createReport(): Promise<ReportJob> {
  const res = await authenticatedFetch(`${API_BASE}/reports`, {
    method: 'POST',
  })
  if (!res.ok) throw new Error('Falha ao criar relatório')
  return res.json()
}

export async function getReportStatus(id: string): Promise<ReportJob> {
  const res = await authenticatedFetch(`${API_BASE}/reports/${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Falha ao buscar status do relatório')
  return res.json()
}

export async function getReportsList(): Promise<ReportJob[]> {
  const res = await authenticatedFetch(`${API_BASE}/reports`)
  if (!res.ok) throw new Error('Falha ao buscar lista de relatórios')
  return res.json()
}

export function getDownloadUrl(id: string): string {
  return `${API_BASE}/reports/${encodeURIComponent(id)}/download`
}
