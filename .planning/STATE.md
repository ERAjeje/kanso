# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-16)

**Core value:** O usuário consegue nomear e registrar suas emoções no momento em que as sente, criando um histórico que torna o processo terapêutico mais concreto e orientado a dados.
**Current focus:** All v1 MVP phases complete — tagging v0.1.0

## Current Position

Phase: 4 of 4 (Technical Debt & Dev Experience) — Complete
Phase 4: Technical Debt & Dev Experience — 1 plan executed
Plans: 1 of 1 (complete)
Status: All tech debt resolved, verified, committed
Last activity: 2026-05-17 — Phase 4 complete

Progress: [████████████] Phase 3 complete (2/2 plans done)
Progress: [████████████] Phase 4 complete (1/1 plans done)

## Performance Metrics

**Velocity:**
- Total plans completed: 6
- Average duration: ~10 min
- Total execution time: ~62 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 3 | 3 | ~10 min |
| 3 | 2 | 2 | ~10 min |

**Recent Trend:**
- feat(1-01): scaffold monorepo
- feat(1-02): Google OAuth login flow
- feat(1-03): CouchDB isolation + PouchDB sync
- feat(3-02): ReportJob types, reports API, ReportSection component, tests
- test(3-01): types, config, test contracts, stubs for PDF report backend
- feat(3-01): handler, service, repository, templates, wire routes
- feat(3-01): chromedp PDF generator, multi-stage Docker, docker-compose
- Trend: ✓ All verification criteria passing

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- **chi** chosen over Gin for Go backend
- **localStorage** for JWT, httpOnly cookie for refresh token
- **1h JWT** with auto-refresh
- **All 4 CouchDB databases** created in Phase 1 with proxy auth via Traefik
- **Base + override** Docker Compose pattern
- **mkcert + Traefik** for local HTTPS
- **Dedicated /login page** → redirects to /register
- **Headless UI v2 Combobox** for sentimento autocomplete
- **Tailwind CSS v4** utility classes (no shadcn) — 4/2 spacing/weight scales
- **indigo-600** accent color for CTA, active tab, combobox highlight, focus ring
- **Bottom tab bar** with lucide-react (Pencil, Clock, User) + text labels
- **Ref-based job ID** for polling — prevents re-triggering polling effect on job object updates
- **Reports API** at /api/reports with 4 endpoints (create, status, list, download)
- **5-state machine** for ReportSection (idle → generating → polling → completed | error)
- **Previous reports always visible** even during generation (not hidden from user)
- **chromedp** for HTML-to-PDF conversion (chromedp/headless-shell Docker image)
- **sync.Mutex** in ReportService for single-concurrent PDF generation
- **filepath.Base** path traversal protection in GetPDF
- **Mock CouchDB via httptest.Server** for handler tests (no interface mocking)

### Pending Todos

| # | Item | Files | Priority | 
|---|------|-------|----------|
| 1 | Extrair URLs hardcoded (`kanso.local`) para env vars do Vite (`VITE_API_URL`, `VITE_COUCHDB_URL`) | `frontend/src/services/auth.ts`, `pouchdb.ts` | 🔴 P0 |
| 2 | Adicionar middleware CORS no backend (`go-chi/cors`) | `backend/cmd/kanso-api/main.go` | 🔴 P0 |
| 3 | Configurar `server.proxy` no Vite para dev sem Traefik | `frontend/vite.config.ts` | 🔴 P0 |
| 4 | Adicionar Traefik ao docker-compose (TLS + roteamento) | `infra/docker-compose.yml`, novos arquivos | 🟡 P1 |
| 5 | Criar Makefile ou justfile com comandos unificados | — | 🟡 P1 |
| 6 | Preencher ou remover `nlp-service/` (diretório vazio) | `nlp-service/` | 🟢 P2 |

### Blockers/Concerns

None yet.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Tech Debt | Traefik no docker-compose (TLS + roteamento) | Pending | 2026-05-16 |
| Tech Debt | Makefile/justfile para comandos unificados | Pending | 2026-05-16 |
| Tech Debt | nlp-service stub (diretório vazio) | Pending | 2026-05-16 |

## Session Continuity

Last session: 2026-05-16 — Phase 3 completed
Stopped at: Phase 4 queued (Tech Debt & Dev Experience)
Resume with: Resolver P0 items (URLs hardcoded → env vars, CORS, proxy Vite)
See README.md → Dívida Técnica for full list
