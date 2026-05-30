# Phase 8: V3 — Integração & Qualidade — Context

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Corrigir vulnerabilidades de segurança identificadas na auditoria (SECURITY-AUDIT.md), refinar experiência pós-NLP (emotion chips) e integrar WhatsApp para envio automático de relatórios.

**This CONTEXT.md covers only the security vulnerability fixes.** Emotion chips e WhatsApp serão discutidos separadamente.

**Issues addressed:** CR-01, CR-02, CR-03, CR-04, CR-05, HI-07

</domain>

<decisions>
## Implementation Decisions

### Isolamento de Dados CouchDB (CR-03)
- **D-01:** **Documento separado `config:active_users`** no CouchDB contendo lista de emails liberados (allowlist)
- **D-02:** **Começar do zero** — nenhum usuário tem acesso até aprovação manual
- **D-03:** Backend Go verifica email na allowlist **antes de gerar o JWT** no fluxo de login Google OAuth
- **D-04:** Se email não estiver na lista → retornar **403 Forbidden** com mensagem "Conta não autorizada"
- **D-05:** Usuários ativos manualmente pelo desenvolvedor via inserção no documento `config:active_users` no CouchDB (Fauxton ou API)

### Gerenciamento de Secrets (CR-01, CR-02)
- **D-06:** Revogar OAuth client secret atual no Google Cloud Console (pendente — usuário fará manualmente)
- **D-07:** Gerar **novo JWT_SECRET forte** (256-bit random via `openssl rand -base64 32`)
- **D-08:** Gerar **nova COUCHDB_PASSWORD forte**
- **D-09:** Secrets armazenados em **`.env` (gitignored)** — não Docker secrets
- **D-10:** `infra/couchdb/local.ini` mantido no git com placeholder. Substituição do valor real via script de entrypoint ou montagem de arquivo separado com o secret real

### CSP e Headers de Segurança (HI-07)
- **D-11:** Adicionar middleware de security headers no Go backend:
  - `Content-Security-Policy`: `default-src 'self'; script-src 'self' https://accounts.google.com; frame-src https://accounts.google.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; connect-src 'self' https://fcm.googleapis.com; img-src 'self' data:`
  - `Strict-Transport-Security`: `max-age=31536000; includeSubDomains`
  - `X-Content-Type-Options`: `nosniff`
  - `X-Frame-Options`: `DENY`
  - `Referrer-Policy`: `strict-origin-when-cross-origin`

### Chromedp PDF Generator (CR-05)
- **D-12:** Remover `disable-web-security` e `allow-file-access-from-files` do `generator.go:46-47`
- Flags não são necessárias — HTML usa `data:text/html` + Go `html/template` (auto-escaping)

### Proteção do Push Endpoint (CR-04)
- **D-13:** Restringir `/api/push/send` para **rede interna Docker** via middleware Go
- Middleware verifica se IP do request pertence à sub-rede Docker (172.x.x.x, 10.x.x.x)
- Scheduler (`kanso-scheduler`) roda na rede interna — acesso mantido

### Pendentes (não discutidos nesta sessão)
Os itens abaixo foram identificados no SECURITY-AUDIT.md mas **não foram discutidos** — o planner/researcher pode incluir ou o usuário decide depois:
- Atualizar dependências (`golang.org/x/net`, `golang.org/x/crypto`) — HI-06
- Containers como root (scheduler, nlp, api) — HI-03, HI-04, HI-05
- gRPC sem TLS (HI-02) — manter insecure em rede interna por ora
- JWT em localStorage (ME-05) — manter localStorage por ora
- CouchDB porta 5984 exposta (ME-02) — remover port mapping
- Proxy `/db` do Vite (ME-01) — remover
- Validação de algoritmo JWT (ME-03) — adicionar validação explícita de `SigningMethodHMAC`

### the agent's Discretion
- Nome exato do documento CouchDB para allowlist (sugestão: `config:active_users`)
- Formato do documento allowlist (array de emails vs objeto com campos)
- Implementação do middleware de rede interna (check de IP range)
- Estrutura do middleware de security headers

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Security Audit
- `SECURITY-AUDIT.md` — Auditoria completa de segurança com todos os 24 achados, severidades e recomendações

### Código Afetado
- `backend/cmd/kanso-api/main.go` — Rotas, middlewares, criação de DBs CouchDB com \_security
- `backend/internal/handler/auth.go` — Fluxo de login Google OAuth (onde adicionar verificação allowlist)
- `backend/internal/handler/push.go` — HandleSend sem auth + HandleSubscribe com JWT
- `backend/internal/pdf/generator.go` — Chromedp flags inseguras (CR-05)
- `backend/internal/config/config.go` — Carregamento de env vars
- `backend/internal/middleware/auth.go` — Middleware JWT existente
- `backend/internal/service/auth.go` — Lógica de autenticação
- `frontend/src/services/pouchdb.ts` — PouchDB sync direto com CouchDB
- `frontend/src/services/auth.ts` — Armazenamento de JWT em localStorage
- `infra/couchdb/local.ini` — Config JWT auth do CouchDB (placeholders commitados)
- `infra/docker-compose.yml` — Secrets expostos, porta CouchDB exposta, docker socket
- `frontend/index.html` — Google GSI script tag, sem CSP

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `backend/internal/middleware/auth.go` — Middleware JWT existente. Seguir mesmo padrão para novos middlewares (security headers, IP restriction)
- `backend/internal/config/config.go` — Padrão de carregamento de env vars via `getEnv()`. Adicionar novas configs aqui

### Established Patterns
- Handlers seguem padrão: handler → service → repository. Middleware de allowlist pode ser adicionado como step no auth handler ou como middleware separado
- Middlewares chi: `r.Use()` para globais, `r.Group(func(r chi.Router) { r.Use(...) })` para grupos de rotas

### Integration Points
- Login flow: `POST /api/auth/google` → `authHandler.HandleGoogleLogin` → `authSvc.Authenticate` → verificar allowlist → gerar JWT
- Push: `POST /api/push/send` → sem middleware atualmente. Adicionar IP restriction middleware ou mover para grupo interno
- Security headers: middleware global `r.Use()` antes de qualquer rota

</code_context>

<specifics>
## Specific Ideas

- Documento `config:active_users` no CouchDB: `{"_id": "config:active_users", "type": "config", "activeEmails": ["email1@gmail.com"]}`
- Middleware de rede interna: verificar `r.RemoteAddr` contra CIDR `172.16.0.0/12` e `10.0.0.0/8`
- Gerar novos secrets via: `openssl rand -base64 32` (JWT_SECRET) e `openssl rand -base64 12` (COUCHDB_PASSWORD)

</specifics>

<deferred>
## Deferred Ideas

- **Refatorar emotion chips** (pós-NLP) — pertence a outra sub-atividade da Fase 8
- **WhatsApp automático** — envio de relatório via Twilio — pertence a outra sub-atividade da Fase 8
- **Migrar JWT para HttpOnly cookie** (ME-05) — ficou para depois
- **gRPC com TLS** (HI-02) — manter insecure em rede interna por ora
- **Docker secrets** em vez de .env — usuário preferiu manter .env

</deferred>

---

*Phase: 8-V3-Integracao-Qualidade*
*Context gathered: 2026-05-23*
