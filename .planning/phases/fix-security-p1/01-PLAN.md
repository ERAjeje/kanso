---
phase: fix-security-p1
plan: 01
type: execute
wave: 1
depends_on:
  - fix-security-p0
files_modified:
  - infra/traefik/dynamic.yml
  - infra/traefik/traefik.yml
  - infra/docker-compose.yml
  - backend/internal/middleware/auth.go
  - backend/go.mod
  - backend/go.sum
  - backend/Dockerfile
  - scheduler/Dockerfile
  - nlp-service/Dockerfile
  - backend/cmd/kanso-api/main.go
autonomous: false
requirements:
  - P1-SEC-01: HTTP Security Headers (CSP, HSTS, XFO, XCTO) via Traefik
  - P1-SEC-02: Rate limiting on auth endpoints
  - P1-SEC-03: CouchDB hardening — remove port 5984 host exposure
  - P1-SEC-04: JWT algorithm validation in JWTRequired middleware
  - P1-SEC-05: Update golang.org/x/net to latest
  - P1-SEC-06: Security event logging for auth failures
  - P1-SEC-07: Non-root user in Dockerfiles
must_haves:
  truths:
    - Traefik edge aplica security headers em todas as respostas
    - Rate limiting via Traefik nos endpoints /api/auth/*
    - CouchDB continua acessível internamente via docker network, mas não exposto no host
    - JWTRequired middleware valida algoritmo de assinatura (HS256)
    - golang.org/x/net atualizado sem dependências quebradas
    - Auth failures logados com IP e timestamp via slog
    - Containers rodam com USER não-root
  artifacts:
    - path: infra/traefik/dynamic.yml
      provides: "Security headers + rate limit middleware"
      contains: "middlewares.headers"
      contains: "middlewares.ratelimit"
    - path: backend/internal/middleware/auth.go
      provides: "JWT algorithm validation"
      contains: "jwt.WithValidMethods"
    - path: backend/cmd/kanso-api/main.go
      provides: "Security event logging"
      contains: "slog"
    - path: infra/docker-compose.yml
      provides: "CouchDB port removed from host"
      contains: "couchdb ports remove 5984"
---

<objective>
Implementar todos os P1s de segurança identificados no security audit pós-fix-P0.
Foco em hardening de borda (Traefik), isolamento CouchDB, validação JWT, logging,
e containers não-root.
</objective>

<execution_context>
@.planning/phases/fix-security-p0/01-PLAN.md
@infra/traefik/dynamic.yml
@infra/traefik/traefik.yml
@infra/docker-compose.yml
@backend/internal/middleware/auth.go
@backend/cmd/kanso-api/main.go
@backend/Dockerfile
@scheduler/Dockerfile
@nlp-service/Dockerfile
</execution_context>

---

## Tasks

### Task 1: HTTP Security Headers via Traefik

**Arquivos:** `infra/traefik/dynamic.yml`

**Ação:**

Adicionar middleware `secHeaders` no `http.middlewares` do `dynamic.yml`:

```yaml
http:
  middlewares:
    secHeaders:
      headers:
        contentSecurityPolicy: "default-src 'self'; script-src 'self' https://accounts.google.com https://accounts.google.com/gsi/ https://www.googleapis.com; style-src 'self' 'unsafe-inline' https://accounts.google.com; frame-src https://accounts.google.com; connect-src 'self' https://accounts.google.com; img-src 'self' data:; font-src 'self' data:"
        strictTransportSecurity: "max-age=31536000; includeSubDomains; preload"
        contentTypeNosniff: true
        browserXssFilter: true
        referrerPolicy: "strict-origin-when-cross-origin"
        customFrameOptionsValue: "DENY"
        permissionsPolicy: "camera=(), microphone=(), geolocation=()"
```

Aplicar o middleware `secHeaders` em todos os routers (api, couchdb).

**Verificação:**
- `curl -sI https://kanso.local/api/health | grep -i "content-security-policy"` retorna header
- `curl -sI https://kanso.local/api/health | grep -i "strict-transport-security"` retorna header
- Docker compose reload sem erros

---

### Task 2: Rate Limiting em Auth Endpoints

**Arquivos:** `infra/traefik/dynamic.yml`, `infra/traefik/traefik.yml`

**Ação:**

Adicionar middleware `auth-ratelimit` no `dynamic.yml`:

```yaml
    auth-ratelimit:
      rateLimit:
        average: 10
        burst: 20
        period: 1m
        sourceCriterion:
          ipStrategy:
            depth: 1
```

Adicionar label no router do CouchDB + API auth para usar rate limit.

**Como aplicar no Traefik:** Adicionar o middleware `auth-ratelimit` nos routers relevantes via labels no `docker-compose.yml` ou via `dynamic.yml` com router específico.

Abordagem: Criar um router separado no `dynamic.yml` para os endpoints de auth, ou aplicar via Traefik file-based router.

**Alternativa mais simples:** Adicionar o middleware aos routers existentes via `docker-compose.yml` labels.

Vamos aplicar o rate limit apenas nos routers via `docker-compose.yml`:

No service `api`, adicionar label:
```yaml
- "traefik.http.routers.api.middlewares=auth-ratelimit,secHeaders"
```

Isso aplica rate limit + security headers em TODAS as chamadas ao API. Para granularidade fina (só auth endpoints), precisaríamos separar routers — mas o P1 pode ser coberto com rate limit geral.

**Verificação:**
- Requisições simultâneas acima de 20/min retornam 429
- Docker compose reload sem erros

---

### Task 3: CouchDB Hardening — Remover Exposição Host

**Arquivos:** `infra/docker-compose.yml`

**Ação:**

1. Remover `ports: "5984:5984"` do service `couchdb` — CouchDB continua acessível internamente via Docker network.
2. Adicionar comment documenting the change.

**Por que isso é seguro:**
- API Go acessa CouchDB internamente via `couchdb:5984` (Docker DNS)
- PouchDB no frontend acessa via Traefik proxy (`/db` → CouchDB)
- validate_doc_update já protege escritas
- Auditoria futura pode adicionar proxy Go para controle de leitura

**Impacto no dev local:**
- Dev não consegue mais acessar `localhost:5984` diretamente
- Precisa acessar via Traefik em `https://kanso.local/db`
- Ou via `docker exec -it kanso-couchdb curl http://localhost:5984`

**Verificação:**
- `docker compose ps` mostra CouchDB sem porta mapeada
- API ainda consegue conectar em `couchdb:5984`
- Frontend ainda sincroniza via Traefik `/db`

---

### Task 4: JWT Algorithm Validation

**Arquivos:** `backend/internal/middleware/auth.go`

**Ação:**

Adicionar validação de algoritmo HMAC no `JWTRequired`:

```go
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return secret, nil
})
```

Adicionar `"fmt"` ao import.

**Verificação:**
- `go build ./cmd/kanso-api/` exit 0
- `go vet ./cmd/kanso-api/` exit 0
- Token com algoritmo diferente (ex: `none`) é rejeitado

---

### Task 5: Update golang.org/x/net

**Arquivos:** `backend/go.mod`, `backend/go.sum`

**Ação:**

```bash
cd backend && go get golang.org/x/net@latest && go mod tidy
```

**Verificação:**
- `grep 'golang.org/x/net' go.mod` mostra versão >= 0.38.0
- `go build ./...` exit 0
- `go vet ./...` exit 0

---

### Task 6: Security Event Logging

**Arquivos:** `backend/cmd/kanso-api/main.go`, `backend/internal/middleware/auth.go`

**Ação:**

Adicionar logging estruturado com `log/slog` nos pontos:
1. `JWTRequired` middleware: log token inválido/expirado com IP e path
2. `pushAuthMiddleware`: log tentativas sem credenciais
3. `HandleGoogleLogin`: log falhas de autenticação Google
4. `HandleRefresh`: log tentativas com refresh token inválido

**Código exemplo no middleware:**

```go
import "log/slog"

func JWTRequired(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... existing code ...
            if err != nil || !token.Valid {
                slog.Warn("auth: invalid token",
                    "path", r.URL.Path,
                    "ip", r.RemoteAddr,
                    "error", err,
                )
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }
            // ...
        })
    }
}
```

**Verificação:**
- `go build ./...` exit 0
- Logs aparecem com nível WARN em auth failures
- IP do cliente e path são incluídos

---

### Task 7: Non-Root Containers

**Arquivos:** `backend/Dockerfile`, `scheduler/Dockerfile`, `nlp-service/Dockerfile`

**Ação:**

**backend/Dockerfile (runtime stage):**
Adicionar antes do ENTRYPOINT:
```dockerfile
RUN adduser --disabled-password --gecos '' appuser
USER appuser
```

**scheduler/Dockerfile:**
```dockerfile
RUN adduser --disabled-password --gecos '' appuser
USER appuser
```

**nlp-service/Dockerfile (runtime stage):**
```dockerfile
RUN adduser --disabled-password --gecos '' appuser
USER appuser
```

**Considerações:**
- Backend precisa escrever em `/tmp/kanso-pdf` — volume montado precisa ser writable pelo appuser
- Nginx/similar: N/A
- chromedp inside backend container: headless-shell roda como root atualmente — mudar pode quebrar. Verificar se o usuário `appuser` consegue executar headless-shell. Se não, manter como root para o backend mas documentar.

**Verificação:**
- `docker compose build` sem erros
- Containers sobem e healthcheck passa
- `docker compose exec api whoami` retorna `appuser`

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-P1-01 | Information Disclosure | Traefik | mitigate | Security headers bloqueiam clickjacking, MIME sniffing, XSS |
| T-P1-02 | Denial of Service | Auth endpoints | mitigate | Rate limit de 10 req/min por IP via Traefik |
| T-P1-03 | Network Exposure | CouchDB | mitigate | Porta 5984 removida do host, apenas rede interna Docker |
| T-P1-04 | Spoofing | JWT | mitigate | Algorithm validation impede token com alg=none ou diferente |
| T-P1-05 | Dependency Vuln | golang.org/x/net | mitigate | Atualizado para versão sem CVEs conhecidos |
| T-P1-06 | Auditing | Auth | mitigate | Slog registra todas as falhas de autenticação com IP |
| T-P1-07 | Privilege Escalation | Containers | mitigate | USER não-root reduz impacto de RCE |

## Verification

- `go build ./...` compila sem erros
- `go vet ./...` passa
- `go mod tidy` consistente
- Docker compose build sem erros
- `curl -I https://kanso.local/api/health` retorna security headers
- Rate limit retorna 429 após exceder limite
- CouchDB não acessível via `localhost:5984`

## Success Criteria

1. Traefik aplica security headers em todas as respostas (CSP, HSTS, XFO, XCTO, Referrer-Policy)
2. Rate limit de 10 req/min aplicado nos endpoints auth via Traefik
3. CouchDB não está mais exposto na porta do host (apenas rede interna)
4. JWTRequired middleware rejeita tokens com algoritmo diferente de HS256
5. golang.org/x/net atualizado para versão >= 0.38.0
6. Auth failures são logados com IP e path via slog com nível WARN
7. Containers backend, scheduler e nlp rodam como usuário não-root
8. Build e vet passam
