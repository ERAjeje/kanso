# Phase 4 Context — Technical Debt & Dev Experience

## Scope

Resolve all accumulated tech debt (P0, P1, P2) to enable localhost development without Docker/Traefik dependency.

## Decisions

### P0 — Hardcoded URLs → Env Vars
- `VITE_API_URL` and `VITE_COUCHDB_URL` go in `.env` (required, no fallback)
- Update `frontend/src/services/auth.ts` to use `import.meta.env.VITE_API_URL`
- Update `frontend/src/services/pouchdb.ts` to use `import.meta.env.VITE_COUCHDB_URL`
- Update `.env.example` with these variables
- Add `ViteEnv` type declaration in `src/vite-env.d.ts`

### P0 — CORS Middleware
- Add `github.com/go-chi/cors` to backend
- Configure in `main.go` to allow `localhost:5173` and `kanso.local` origins
- Allow methods: GET, POST, PUT, DELETE, OPTIONS
- Allow headers: Authorization, Content-Type
- Allow credentials: true

### P0 — Vite Proxy
- Add `server.proxy` to `vite.config.ts`
- Proxy `/api` → `http://localhost:8080`
- Proxy `/db` → `http://localhost:5984`
- Frontend env vars can use relative paths (`/api`, `/db`) when proxy is active

### P1 — Traefik in docker-compose
- Add Traefik v3 service to `infra/docker-compose.yml`
- Configure entrypoints: web (80), websecure (443)
- Configure file provider for routing rules
- TLS via mkcert or self-signed for local dev
- Labels on api and couchdb services for Traefik routing
- Frontend: `kanso.local` → frontend dev server or static build
- API: `kanso.local/api` → api:8080
- CouchDB: `kanso.local/db` → couchdb:5984

### P1 — Makefile
- Root-level `Makefile` with targets:
  - `make up` — docker compose up (infra only)
  - `make down` — docker compose down
  - `make dev` — start frontend dev server
  - `make test` — run all tests (frontend + backend)
  - `make build` — build frontend + backend
  - `make logs` — docker compose logs -f

### P2 — nlp-service README
- Add `nlp-service/README.md` explaining:
  - Deferred to v2
  - Will be Python + FastAPI + transformers
  - Purpose: NLP emotion analysis of diary entries

## Files to Modify
- `frontend/src/services/auth.ts`
- `frontend/src/services/pouchdb.ts`
- `frontend/src/vite-env.d.ts`
- `frontend/vite.config.ts`
- `frontend/.env` (new)
- `.env.example`
- `backend/cmd/kanso-api/main.go`
- `backend/go.mod` (add go-chi/cors)
- `infra/docker-compose.yml`
- `infra/traefik.yml` (new) or dynamic config
- `Makefile` (new)
- `nlp-service/README.md` (new)

## Constraints
- No changes to existing API contracts or frontend behavior
- All existing tests must pass after changes
- Traefik config must not break current docker-compose setup
