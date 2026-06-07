import { treinamentoDB, getUserId } from './pouchdb'
import type { TreinamentoDoc } from '../types'

export async function saveTrainingExample(texto: string, label: string): Promise<TreinamentoDoc> {
  const doc: TreinamentoDoc = {
    _id: crypto.randomUUID(),
    type: 'treinamento',
    texto,
    label,
    userId: getUserId(),
    origem: 'usuario',
    criadoEm: new Date().toISOString(),
  }
  await treinamentoDB.put(doc)
  return doc
}

export async function getTotalTrainingCount(): Promise<number> {
  const result = await treinamentoDB.allDocs({ limit: 0 })
  return result.total_rows
}
