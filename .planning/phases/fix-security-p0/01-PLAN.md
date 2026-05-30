---
phase: fix-security-p0
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .env
  - infra/couchdb/local.ini
  - .gitignore
  - backend/cmd/kanso-api/main.go
  - backend/internal/handler/push.go
  - backend/internal/middleware/auth.go
  - infra/docker-compose.yml
autonomous: false
requirements:
  - SEC-01 (CR-02+CR-03): JWT secret forte + isolamento CouchDB por validate_doc_update
  - SEC-02 (CR-01): OAuth secret protegido (env var, .gitignore)
  - SEC-03 (CR-04): Push endpoint autenticado (JWT admin + API key)
user_setup:
  - Revogar OAuth client ID no Google Cloud Console e gerar novo
must_haves:
  truths:
    - JWT secret e criptograficamente forte (256-bit, gerado via openssl)
    - Cada banco CouchDB tem validate_doc_update que verifica userCtx.name === doc.userSub
    - Nenhum usuario consegue ler documentos de outro usuario via CouchDB
    - OAuth client secret removido do disco (usar env var)
    - POST /api/push/send exige JWT admin OU header X-API-Key
    - Scheduler envia X-API-Key nas requisicoes
    - infra/couchdb/local.ini removido do git tracking
  artifacts:
    - path: backend/cmd/kanso-api/main.go
      provides: "validate_doc_update deployment on startup"
      contains: "ensureValidateDocUpdate"
      contains: "validate_doc_update"
    - path: backend/internal/handler/push.go
      provides: "Push endpoint with JWT admin + API key auth"
      contains: "apiKey"
      contains: "admin"
    - path: .gitignore
      provides: "Secrets directory and local.ini gitignored"
      contains: ".secrets/"
      contains: "local.ini"
  key_links:
    - from: "main.go startup"
      to: "CouchDB databases"
      via: "ensureValidateDocUpdate -> PUT _design/security"
    - from: "push handler"
      to: "auth middleware"
      via: "JWTRequired + apiKey middleware"
    - from: "scheduler"
      to: "push endpoint"
      via: "X-API-Key header"
---

<objective>
Corrigir todos os issues P0 do SECURITY-AUDIT.md: CR-01, CR-02, CR-03, CR-04.

**Purpose:** Eliminar as brechas de seguranca que podem expor dados emocionais dos usuarios. Implementar isolamento real entre usuarios no CouchDB, proteger endpoints, e rotacionar secrets comprometidos.
</objective>

<execution_context>
@SECURITY-AUDIT.md
</execution_context>

---

## Tasks

### Task 1: CR-02 — Gerar novo JWT secret e atualizar configs

**Arquivos:** `.env`, `infra/couchdb/local.ini`, `.gitignore`

**Ação:**

1. Gerar novo JWT secret:
   ```bash
   openssl rand -base64 32
   ```

2. Atualizar `.env`:
   ```
   JWT_SECRET=<novo-valor>
   ```

3. Codificar em base64 para o CouchDB:
   ```bash
   echo -n '<novo-jwt-secret>' | base64
   ```

4. Atualizar `infra/couchdb/local.ini` com o base64.

5. Adicionar `infra/couchdb/local.ini` ao `.gitignore`:
   ```
   infra/couchdb/local.ini
   ```

6. Remover do tracking do git:
   ```bash
   git rm --cached infra/couchdb/local.ini
   ```

7. Criar `infra/couchdb/local.ini.example` como template (sem secrets reais).

**Verificação:**
- `grep -c "dev-secret-change-in-production" .env` retorna 0
- `grep "local.ini" .gitignore` retorna 1
- `git ls-files infra/couchdb/local.ini` vazio

---

### Task 2: CR-03 — validate_doc_update para isolamento por usuario

**Arquivos:** `backend/cmd/kanso-api/main.go`, `backend/internal/repository/couchdb.go`

**Ação:**

Adicionar no startup do backend (`main.go`) uma funcao `ensureValidateDocUpdate` que cria/atualiza `_design/security` em cada banco com uma `validate_doc_update` function JavaScript.

A function JS verifica:
```javascript
function(newDoc, oldDoc, userCtx, secObj) {
    if (userCtx.roles.indexOf('_admin') !== -1) {
        return true; // admin sempre pode
    }
    if (newDoc.type === 'config') {
        return true; // documentos de config sao publicos no banco
    }
    if (newDoc.userSub && userCtx.name) {
        if (newDoc.userSub === userCtx.name) {
            return true; // dono do documento
        }
    }
    throw({ forbidden: 'Voce nao tem permissao para acessar este documento.' });
}
```

**Bancos que recebem a validate_doc_update:**
- `registros` — campo `userSub`
- `sentimentos` — campo `userSub`
- `preferencias` — campo `userSub`
- `relatorios` — campo `userSub`
- `usuarios` — campo `_id` (formato `user:{sub}`)

**Padrao:** Seguir o mesmo padrao de `ensureCouchDBDatabases` e `ensureCouchDBIndexes` — HTTP PUT com admin credentials.

**Verificação:**
- `go build ./cmd/kanso-api/` exit 0
- `go vet ./cmd/kanso-api/` exit 0
- `grep -c "ensureValidateDocUpdate" backend/cmd/kanso-api/main.go` retorna >= 2

---

### Task 3: CR-01 — OAuth secret cleanup

**Arquivos:** `.gitignore`

**Ação:**

1. Verificar se `.secrets/` ja esta no `.gitignore`. Se nao, adicionar.
2. Verificar no codigo se o OAuth client secret ja vem de env var (deve estar em `GOOGLE_CLIENT_SECRET` ou similar). Se nao, adicionar suporte.
3. Usuario revoga o client ID no Google Cloud Console manualmente.

**Verificação:**
- `grep ".secrets/" .gitignore` retorna 1
- Codigo le o secret de env var, nao de arquivo

---

### Task 4: CR-04 — Push endpoint com JWT admin + API key

**Arquivos:** `backend/cmd/kanso-api/main.go`, `backend/internal/handler/push.go`, `backend/internal/service/push.go`, `infra/docker-compose.yml`, `scheduler/Dockerfile`

**Ação:**

1. Adicionar middleware de API key em `main.go`:
   ```go
   func apiKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               key := r.Header.Get("X-API-Key")
               if key != expectedKey {
                   http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

2. Mover `/api/push/send` para dentro do grupo JWT, com verificacao de admin role:
   ```go
   r.Group(func(r chi.Router) {
       r.Use(jwtMiddleware.JWTRequired)
       r.Post("/api/push/send", pushHandler.HandleSend)
   })
   ```

3. Adicionar `SCHEDULER_API_KEY` ao `.env` e `docker-compose.yml`.

4. Atualizar scheduler para enviar `X-API-Key` header.

5. No handler `HandleSend`, verificar se o caller e admin OU se tem API key (para scheduler).

**Verificação:**
- `go build ./cmd/kanso-api/` exit 0
- `go vet ./cmd/kanso-api/` exit 0
- `grep -c "apiKeyMiddleware" backend/cmd/kanso-api/main.go` retorna >= 2
- `grep -c "X-API-Key" backend/cmd/kanso-api/main.go` retorna 1

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-P0-01 | Information Disclosure | CouchDB | mitigate | validate_doc_update impede leitura跨 de dados de outros usuarios |
| T-P0-02 | Spoofing | JWT | mitigate | Secret forte de 256 bits substitui o default conhecido |
| T-P0-03 | Spoofing | Push endpoint | mitigate | JWT admin + API key — dois fatores de autenticacao |
| T-P0-04 | Information Disclosure | OAuth secret | mitigate | Secret removido do disco, usado via env var |
| T-P0-05 | Elevation of Privilege | validate_doc_update | accept | Admin CouchDB ainda tem acesso total. Necessario para operacao. |

## Verification

- `go build ./...` compila sem erros
- `go vet ./...` passa
- Docker compose sobe sem erros
- `git ls-files infra/couchdb/local.ini` vazio (removido do tracking)

## Success Criteria

1. JWT secret forte gerado e configurado, default conhecido removido
2. validate_doc_update implantado em todos os 5 bancos CouchDB
3. Usuario A nao consegue ler documentos do usuario B via CouchDB
4. .secrets/ e local.ini no .gitignore
5. POST /api/push/send bloqueado sem JWT admin ou API key
6. Build e vet passam
