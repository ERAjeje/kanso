# Bug: Relatório — 400 invalid request body

## Symptoms
- Geração de relatório retorna HTTP 400 com "invalid request body"
- Provável mismatch de contrato: frontend envia body que o backend não espera

## Hypothesis
O contrato do backend para `POST /api/reports` espera um formato de body diferente do que o frontend está enviando.

## To Investigate
- Verificar o handler do backend (`POST /api/reports`)
- Verificar o service do frontend (`createReport` em `reports.ts`)
- Comparar o body enviado vs esperado
