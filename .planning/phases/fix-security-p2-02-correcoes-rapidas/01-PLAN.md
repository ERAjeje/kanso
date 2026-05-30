---
phase: fix-security-p2-02-correcoes-rapidas
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - nlp-service/src/classifier.py
  - nlp-service/src/model_config.py
  - nlp-service/Dockerfile
  - frontend/src/main.tsx
  - frontend/src/pages/Login.tsx
  - frontend/src/components/RegistrationForm.tsx
  - frontend/src/services/pouchdb.ts
  - frontend/src/hooks/usePushNotifications.ts
  - backend/internal/repository/couchdb.go
  - backend/cmd/kanso-api/main.go
autonomous: false
requirements:
  - LO-01: NLP classifier não usa subprocess para git hash
  - LO-03: Service worker registration error não é silencioso
  - IN-01: console.error removido ou substituído por warn em produção
  - IN-02: Catch blocks vazios têm logging mínimo
  - IN-03: Database names movidos para constantes
must_haves:
  truths:
    - classifier.py lê MODEL_VERSION de env var, não de subprocess git
    - Dockerfile do nlp-service grava GIT_HASH em /app/version.txt e expõe como env var
    - navigator.serviceWorker.register tem catch com console.warn
    - console.error substituído por warn ou logger nos 4 arquivos do frontend
    - Catch blocks vazios têm ao menos console.warn
    - Database names (registros, sentimentos, preferencias, relatorios, usuarios) são constantes
  artifacts:
    - path: nlp-service/src/classifier.py
      provides: "MODEL_VERSION de env var"
      contains: "os.environ.get"
    - path: frontend/src/main.tsx
      provides: "SW error logging"
      contains: "console.warn"
    - path: backend/internal/repository/couchdb.go
      provides: "DB name constants"
      contains: "const ( DBRegistros )"
---

<objective>
Corrigir 5 achados de baixa severidade em um único plano: LO-01 (NLP subprocess), LO-03 (SW error), IN-01 (console.error), IN-02 (empty catches), IN-03 (hardcoded DB names).
</objective>

<execution_context>
@nlp-service/src/classifier.py
@nlp-service/src/model_config.py
@nlp-service/Dockerfile
@frontend/src/main.tsx
@frontend/src/pages/Login.tsx
@frontend/src/components/RegistrationForm.tsx
@frontend/src/services/pouchdb.ts
@frontend/src/hooks/usePushNotifications.ts
@backend/internal/repository/couchdb.go
@backend/cmd/kanso-api/main.go
</execution_context>

---

## Tasks

### Task 1: LO-01 — NLP classifier: remover subprocess git hash

**Arquivos:** `nlp-service/src/classifier.py`, `nlp-service/src/model_config.py`, `nlp-service/Dockerfile`

**Ação:**

1. Em `model_config.py`, `MODEL_VERSION` já é `"v1.0"` — manter como fallback, mas permitir override via env var:
   ```python
   MODEL_VERSION = os.environ.get("MODEL_VERSION", "v1.0")
   ```

2. Em `classifier.py`, remover a função `get_model_version()` e import de `subprocess`. Substituir usos por `MODEL_VERSION` direto.

3. No `nlp-service/Dockerfile`, durante o build, gravar o git hash e expor como env var:
   ```dockerfile
   ARG GIT_HASH
   ENV MODEL_VERSION=${GIT_HASH:-v1.0}
   ```
   E no `docker-compose.yml`, passar o build arg:
   ```yaml
   nlp:
     build:
       args:
         GIT_HASH: ${GIT_HASH:-v1.0}
   ```

**Simplificação:** Se o Docker Compose build arg for complexo, manter apenas `os.environ.get("MODEL_VERSION", "v1.0")` no model_config.py e definir `MODEL_VERSION=v1.0` no environment do docker-compose. A função `get_model_version()` é removida de qualquer forma.

**Verificação:**
- `grep -c "subprocess" classifier.py` retorna 0
- `grep -c "git rev-parse" classifier.py` retorna 0
- NLP container sobe sem erro de import
- `get_model_version()` não é mais chamada em lugar nenhum

---

### Task 2: LO-03 — Service worker error logging

**Arquivo:** `frontend/src/main.tsx`

**Ação:**

Adicionar `console.warn` no catch do service worker registration:
```typescript
navigator.serviceWorker.register('/sw.js').catch((err) => {
  console.warn('Service worker registration failed — push unavailable:', err)
})
```

**Verificação:**
- Código compila (`pnpm build` ou `pnpm typecheck`)
- Catch block não está mais vazio

---

### Task 3: IN-01 + IN-02 — console.error e catch blocks vazios

**Arquivos:**
- `frontend/src/pages/Login.tsx` (line 60)
- `frontend/src/components/RegistrationForm.tsx` (line 55)
- `frontend/src/services/pouchdb.ts` (line 50)
- `frontend/src/hooks/usePushNotifications.ts` (line 43)

**Ação:**

Substituir `console.error` por `console.warn` em todos os 4 locais. Em produção, console.error expõe detalhes no browser console que podem ser usados em ataques de engenharia social.

Nenhum catch block puramente vazio foi encontrado além do LO-03 (já tratado na Task 2). Manter vigilant.

**Verificação:**
- `grep -r "console.error" frontend/src/` retorna 0 (ou apenas falso positivo em comentário)

---

### Task 4: IN-03 — Database names como constantes

**Arquivos:** `backend/internal/repository/couchdb.go`, `backend/cmd/kanso-api/main.go`

**Ação:**

1. Em `backend/internal/repository/couchdb.go`, adicionar constantes no topo:
   ```go
   const (
       DBRegistros    = "registros"
       DBSentimentos  = "sentimentos"
       DBPreferencias = "preferencias"
       DBRelatorios   = "relatorios"
       DBUsuarios     = "usuarios"
   )
   ```

2. Nos métodos do repository, substituir strings literais pelas constantes:
   - `"usuarios"` → `DBUsuarios` (CreateOrUpdateUser, GetUser)
   - `"preferencias"` → `DBPreferencias` (SavePushPrefs, GetAllPushPrefs)
   - `"relatorios"` → `DBRelatorios` (CreateReportJob, GetReportJob, UpdateReportJobStatus, ListReportJobsByUser, GetLastCompletedReport)
   - `"registros"` → `DBRegistros` (GetCheckpoint, SaveCheckpoint, FindRegistrosByPeriod)
   - `"sentimentos"` → `DBSentimentos` (SaveAnalise, FindAnaliseByRegistroIds)

3. Em `backend/cmd/kanso-api/main.go`, substituir strings literais pelos exports (precisa referenciar via `repository.DBRegistros`, etc.).

**Verificação:**
- `go build ./...` exit 0
- `go vet ./...` exit 0
- Nenhum database name hardcoded como string literal nos arquivos modificados (exceto nas constantes)

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-LO-01-01 | Command Injection | NLP classifier | mitigate | Subprocess removido — sem execução de comandos externos |
| T-IN-01-01 | Information Disclosure | Frontend | mitigate | console.error removido — menos info leaks |
| T-IN-03-01 | Maintainability | Backend | accept | Baixo risco, mas melhora qualidade do código |

## Verification

- `go build ./...` exit 0
- `go vet ./...` exit 0
- `pnpm typecheck` passa (ou `pnpm build`)
- `grep -c "subprocess" nlp-service/src/classifier.py` = 0
- `grep -r "console.error" frontend/src/` = 0
- NLP container sobe sem erros
