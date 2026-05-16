import { z } from 'zod'

export const registroSchema = z.object({
  dataHora: z.string().min(1, 'Data e hora são obrigatórias'),
  sensacoes: z.string().max(2000, 'Máximo de 2000 caracteres').default(''),
  sentimentoNome: z.string().min(1, 'Sentimento é obrigatório').max(100, 'Máximo de 100 caracteres'),
  sentimentoId: z.string().nullable().default(null),
  contexto: z.string().max(2000, 'Máximo de 2000 caracteres').default(''),
  pensamentos: z.string().max(2000, 'Máximo de 2000 caracteres').default(''),
})

export type RegistroFormData = z.infer<typeof registroSchema>
