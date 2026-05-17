# Debug: scheduler-go-sum

## Symptom
```
$ make up
cd infra && docker compose up -d
ERROR [scheduler build 3/6] COPY go.mod go.sum ./
=> "/go.sum": not found
```

## Capture
- `make up` falha ao buildar o container `scheduler`
- Erro no `scheduler/Dockerfile:3` — `COPY go.mod go.sum ./`
- `scheduler/go.sum` não existe

## Root Cause
O módulo `scheduler/` não possui dependências externas:
```
// scheduler/go.mod
module github.com/edson/kanso-scheduler
go 1.26.2
```
Como não há dependências, `go.sum` nunca foi gerado. O Dockerfile tenta copiá-lo e quebra.

## Fix
`scheduler/go.mod` não tem dependências, então `go mod tidy` não gera `go.sum`.
Alterado `scheduler/Dockerfile:3` de:
```
COPY go.mod go.sum ./
```
Para:
```
COPY go.* ./
```
Assim o Docker copia apenas `go.mod` se `go.sum` não existir.

## Verification
- `docker compose build scheduler` ✅ conclui com sucesso
- Layer de COPY + `go mod download` funciona sem `go.sum`
- Build final da imagem: `sha256:2f280086f0cb698704f65f9f7e22c0d3958384eb6f6dd4cb13bb6d641b111674`
