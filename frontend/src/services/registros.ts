import { registrosDB, sentimentosDB, getUserId } from './pouchdb'
import type { RegistroDoc, SentimentoDoc, AnaliseNlpDoc, RegistroWithAnalise } from '../types'
import type PouchDB from 'pouchdb-browser'

type RegistroInput = Omit<RegistroDoc, '_id' | '_rev' | 'type' | 'userId' | 'createdAt' | 'updatedAt'>

export async function saveRegistro(data: RegistroInput): Promise<RegistroDoc> {
  const now = new Date().toISOString()
  const doc: RegistroDoc = {
    _id: crypto.randomUUID(),
    type: 'registro',
    userId: getUserId(),
    ...data,
    createdAt: now,
    updatedAt: now,
  }
  await registrosDB.put(doc)
  return doc
}

export async function saveSentimento(nome: string): Promise<SentimentoDoc> {
  const doc: SentimentoDoc = {
    _id: crypto.randomUUID(),
    type: 'sentimento',
    userId: getUserId(),
    nome: nome.trim().toLowerCase(),
    criadoEm: new Date().toISOString(),
  }
  await sentimentosDB.put(doc)
  return doc
}

export async function getRegistros(): Promise<RegistroWithAnalise[]> {
  const result = await registrosDB.allDocs<RegistroDoc>({ include_docs: true })
  const registros = result.rows
    .map((row: PouchDB.Core.Row<RegistroDoc>) => row.doc!)
    .filter((doc: RegistroDoc) => doc.type === 'registro')
    .sort((a: RegistroDoc, b: RegistroDoc) => new Date(b.dataHora).getTime() - new Date(a.dataHora).getTime())

  // Fetch analise_nlp docs and merge in-memory
  const analiseResult = await registrosDB.allDocs<AnaliseNlpDoc>({ include_docs: true })
  const analiseMap = new Map<string, AnaliseNlpDoc>()
  for (const row of analiseResult.rows) {
    const doc = row.doc!
    if (doc.type === 'analise_nlp') {
      const registroId = doc._id.replace('analise:', '')
      analiseMap.set(registroId, doc)
    }
  }

  return registros.map(r => ({
    ...r,
    analise: analiseMap.get(r._id),
  }))
}

export async function getSentimentos(): Promise<SentimentoDoc[]> {
  const result = await sentimentosDB.allDocs<SentimentoDoc>({ include_docs: true })
  return result.rows
      .map((row: PouchDB.Core.Row<SentimentoDoc>) => row.doc!)
    .filter((doc: SentimentoDoc) => doc.type === 'sentimento')
    .sort((a: SentimentoDoc, b: SentimentoDoc) => a.nome.localeCompare(b.nome, 'pt-BR'))
}
