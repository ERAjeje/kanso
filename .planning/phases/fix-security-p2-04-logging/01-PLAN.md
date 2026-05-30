---
phase: fix-security-p2-04-logging
plan: 01
type: execute
wave: 2
depends_on:
  - fix-security-p1-01
files_modified:
  - backend/internal/service/watcher.go
  - backend/internal/service/report.go
  - backend/internal/handler/push.go
  - backend/internal/handler/report.go
  - backend/internal/service/auth.go
  - backend/cmd/kanso-api/main.go
autonomous: false
requirements:
  - LO-04: Logging de backend não expõe PII e usa níveis estruturados
must_haves:
  truths:
    - Todos os log.Printf no backend foram revisados
    - Logs de erro não contêm dados do usuário (userSub, email, token)
    - Logs usam log/slog com níveis (INFO, WARN, ERROR) em vez de log.Printf
    - userSub aparece apenas em logs de warn/error para rastreabilidade, não em info
  artifacts:
    - path: backend/internal/service/report.go
      provides: "Structured error logging"
      contains: "slog"
    - path: backend/internal/service/watcher.go
      provides: "Structured watcher logging"
      contains: "slog"
---

<objective>
Revisar e sanitizar todo logging do backend — substituir `log.Printf` por `slog` estruturado, remover PII dos logs, padronizar níveis.
</objective>

<execution_context>
@backend/internal/service/watcher.go
@backend/internal/service/report.go
@backend/internal/handler/push.go
@backend/internal/handler/report.go
@backend/internal/service/auth.go
@backend/cmd/kanso-api/main.go
</execution_context>

---

## Tasks

### Task 1: Migrar main.go startup logs para slog

**Arquivo:** `backend/cmd/kanso-api/main.go`

**Ação:**

Substituir `log.Printf` por `slog.Info` nos logs de startup (criação de DB, índices, security):
```go
slog.Info("database created", "db", db)
slog.Warn("could not create database", "db", db, "error", err)
```

**Verificação:**
- `go build ./...` exit 0
- Startup logs aparecem com formato estruturado

### Task 2: Revisar watcher.go — logs sem PII

**Arquivo:** `backend/internal/service/watcher.go`

**Ação:**

Substituir `log.Printf` por `slog`:
- Event loop started → `slog.Info`
- Erro ao pegar checkpoint → `slog.Warn` ou `slog.Error`
- Erro ao analisar registro → `slog.Error` com `result.ID` (ID do doc, não conteúdo do usuário)
- Falha após 3 retries → `slog.Error` com registroId (sem dados emocionais)
- Checkpoint save error → `slog.Error`

⚠️ **Cuidado:** `result.ID` é o ID do documento CouchDB, não userSub — seguro para log. Não logar `sensacoes`, `contexto`, `pensamentos`.

### Task 3: Revisar report.go — logs sem PII

**Arquivo:** `backend/internal/service/report.go`

**Ação:**

Substituir `log.Printf` por `slog` com nível adequado:
- Erros de get report → `slog.Error` com `jobID` (sem userSub no message)
- PDF generation failed → `slog.Error` com `jobID`
- Job update error → `slog.Error`

**Verificação:**
- `slog.Error` usado para falhas reais
- `slog.Warn` para problemas recuperáveis
- `slog.Info` para eventos normais

### Task 4: Revisar handlers — push.go + report.go + auth.go

**Arquivos:**
- `backend/internal/handler/push.go`
- `backend/internal/handler/report.go`
- `backend/internal/service/auth.go`

**Ação:**

Substituir `log.Printf` por `slog`:
- Push subscribe error → `slog.Warn`
- Push send error → `slog.Warn`
- Report request/list/get/PDF errors → `slog.Warn` ou `slog.Error`
- Auth save user warning → `slog.Warn`

### Task 5: Configurar slog para formato JSON (opcional)

**Arquivo:** `backend/cmd/kanso-api/main.go`

**Ação (opcional, boa prática):**

Adicionar no início do `main()`:
```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
    AddSource: true,
})))
```

Isso produz logs JSON estruturados que podem ser ingeridos por sistemas de log aggregation.

**Verificação:**
- `go build ./...` exit 0
- Logs aparecem como JSON no stderr

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-LO-04-01 | Information Disclosure | Backend logs | mitigate | Logs revisados — sem PII, sem dados emocionais, sem tokens |
| T-LO-04-02 | Auditing | Backend logs | mitigate | Slog estruturado com níveis — facilita auditoria |

## Verification

- `go build ./...` exit 0
- `go vet ./...` exit 0
- Nenhum `log.Printf` nos arquivos modificados (apenas `slog`)
- Logs de erro não contêm dados sensíveis do usuário

## Success Criteria

1. Todos os `log.Printf` substituídos por `slog.Info`, `slog.Warn`, ou `slog.Error`
2. Nenhum dado pessoal (email, userSub, token) em message strings de log
3. Níveis de log semanticamente corretos (erro real → Error, recuperável → Warn, normal → Info)
4. Build e vet passam
