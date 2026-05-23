# Plan: fix-sentimentos-db

## Problem
NLP watcher saves `analise_nlp` docs in `registros` DB mixed with registro docs. The `sentimentos` DB was designed for this purpose but is never written to.

## Changes

### Backend
1. `couchdb.go:621` — `SaveAnalise()`: "registros" → "sentimentos"
2. `couchdb.go:704` — `FindAnaliseByRegistroIds()`: "registros/_find" → "sentimentos/_find"
3. `watcher.go:17` — Update doc comment

### Frontend
4. `registros.ts:41` — `getRegistros()`: `registrosDB.allDocs` → `sentimentosDB.allDocs`

### Migration
5. Copy existing `analise_nlp` docs from `registros` → `sentimentos`, then delete from `registros`

## Verification
- `go build ./...` + `go vet ./...` + `go test ./...` — pass
- `npx tsc --noEmit` + `npx vitest run` — pass
- Migration script copies data correctly
