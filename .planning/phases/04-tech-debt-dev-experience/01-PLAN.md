---
phase: 4
plan_id: 01
wave: 1
depends_on: []
files_modified:
  - frontend/src/services/auth.ts
  - frontend/src/services/pouchdb.ts
  - frontend/src/vite-env.d.ts
  - frontend/vite.config.ts
  - frontend/.env
  - .env.example
  - backend/cmd/kanso-api/main.go
  - backend/go.mod
  - backend/go.sum
  - infra/docker-compose.yml
  - infra/traefik/dynamic.yml
  - infra/traefik/traefik.yml
  - Makefile
  - nlp-service/README.md
autonomous: true
requirements: []
---

# Plan 01: Resolve All Tech Debt (P0 + P1 + P2)

## Objective

Enable localhost development without Docker/Traefik dependency by extracting hardcoded URLs to env vars, adding CORS middleware, configuring Vite proxy, adding Traefik to docker-compose, creating a Makefile, and documenting the deferred nlp-service.

## Tasks

### Task 1: Extract hardcoded URLs to Vite env vars

**type**: execute
**wave**: 1
**files_modified**:
  - frontend/src/services/auth.ts
  - frontend/src/services/pouchdb.ts
  - frontend/src/vite-env.d.ts
  - frontend/.env
  - .env.example

<read_first>
- frontend/src/services/auth.ts
- frontend/src/services/pouchdb.ts
- frontend/.env.example
</read_first>

<acceptance_criteria>
- frontend/src/services/auth.ts line 1 contains `const API_BASE = import.meta.env.VITE_API_URL`
- frontend/src/services/pouchdb.ts line 17 contains `const COUCHDB_URL = import.meta.env.VITE_COUCHDB_URL`
- frontend/src/vite-env.d.ts contains `interface ImportMetaEnv { VITE_API_URL: string; VITE_COUCHDB_URL: string; VITE_GOOGLE_CLIENT_ID: string; }`
- frontend/.env contains `VITE_API_URL=/api`, `VITE_COUCHDB_URL=/db`, `VITE_GOOGLE_CLIENT_ID=`
- .env.example contains `VITE_API_URL=/api`, `VITE_COUCHDB_URL=/db`, `VITE_GOOGLE_CLIENT_ID=`
- No hardcoded `kanso.local` strings remain in frontend/src/services/
</acceptance_criteria>

<action>
1. Create `frontend/.env` with:
   ```
   VITE_API_URL=/api
   VITE_COUCHDB_URL=/db
   VITE_GOOGLE_CLIENT_ID=
   ```

2. Update `.env.example` to include frontend vars alongside backend vars:
   ```
   # Backend
   COUCHDB_PASSWORD=admin123
   JWT_SECRET=dev-secret-change-in-production
   GOOGLE_CLIENT_ID=your-google-client-id

   # Frontend
   VITE_API_URL=/api
   VITE_COUCHDB_URL=/db
   VITE_GOOGLE_CLIENT_ID=your-google-client-id
   ```

3. Update `frontend/src/services/auth.ts`:
   - Replace line 1: `const API_BASE = 'https://kanso.local/api'` → `const API_BASE = import.meta.env.VITE_API_URL`

4. Update `frontend/src/services/pouchdb.ts`:
   - Replace line 17: `const COUCHDB_URL = 'https://kanso.local/db'` → `const COUCHDB_URL = import.meta.env.VITE_COUCHDB_URL`

5. Create or update `frontend/src/vite-env.d.ts`:
   ```typescript
   /// <reference types="vite/client" />
   interface ImportMetaEnv {
     readonly VITE_API_URL: string
     readonly VITE_COUCHDB_URL: string
     readonly VITE_GOOGLE_CLIENT_ID: string
   }
   interface ImportMeta {
     readonly env: ImportMetaEnv
   }
   ```
</action>

---

### Task 2: Add CORS middleware to backend

**type**: execute
**wave**: 1
**depends_on**: []
**files_modified**:
  - backend/cmd/kanso-api/main.go
  - backend/go.mod
  - backend/go.sum

<read_first>
- backend/cmd/kanso-api/main.go
- backend/go.mod
</read_first>

<acceptance_criteria>
- backend/go.mod contains `github.com/go-chi/cors v1`
- backend/cmd/kanso-api/main.go imports `github.com/go-chi/cors`
- backend/cmd/kanso-api/main.go has `r.Use(cors.Handler(...))` or equivalent before route definitions
- CORS config allows origins `http://localhost:5173` and `https://kanso.local`
- CORS config allows methods `GET, POST, PUT, DELETE, OPTIONS`
- CORS config allows headers `Authorization, Content-Type`
- CORS config sets `AllowCredentials: true`
- `cd backend && go mod tidy` succeeds
</acceptance_criteria>

<action>
1. Add CORS dependency: `cd backend && go get github.com/go-chi/cors`

2. Update `backend/cmd/kanso-api/main.go`:
   - Add import: `github.com/go-chi/cors`
   - After `r := chi.NewRouter()` and existing middleware, add:
   ```go
   r.Use(cors.Handler(cors.Options{
       AllowedOrigins:   []string{"http://localhost:5173", "https://kanso.local"},
       AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
       AllowedHeaders:   []string{"Authorization", "Content-Type"},
       AllowCredentials: true,
       MaxAge:           300,
   }))
   ```
   - Place CORS middleware after `Logger` and `Recoverer`, before `Timeout`

3. Run `cd backend && go mod tidy`
</action>

---

### Task 3: Configure Vite proxy for dev

**type**: execute
**wave**: 1
**depends_on**: []
**files_modified**:
  - frontend/vite.config.ts

<read_first>
- frontend/vite.config.ts
</read_first>

<acceptance_criteria>
- frontend/vite.config.ts `server` block contains `proxy` configuration
- Proxy `/api` targets `http://localhost:8080` with `changeOrigin: true`
- Proxy `/db` targets `http://localhost:5984` with `changeOrigin: true`
- Proxy `/db` has `rewrite` rule to strip `/db` prefix: `path.replace(/^\/db/, '')`
- Vite config remains valid TypeScript (no syntax errors)
</acceptance_criteria>

<action>
1. Update `frontend/vite.config.ts` server block:
   ```typescript
   server: {
     host: true,
     port: 5173,
     proxy: {
       '/api': {
         target: 'http://localhost:8080',
         changeOrigin: true,
       },
       '/db': {
         target: 'http://localhost:5984',
         changeOrigin: true,
         rewrite: (path) => path.replace(/^\/db/, ''),
       },
     },
   },
   ```
</action>

---

### Task 4: Add Traefik to docker-compose

**type**: execute
**wave**: 1
**depends_on**: []
**files_modified**:
  - infra/docker-compose.yml
  - infra/traefik/traefik.yml
  - infra/traefik/dynamic.yml

<read_first>
- infra/docker-compose.yml
</read_first>

<acceptance_criteria>
- infra/docker-compose.yml contains a `traefik` service
- Traefik service uses `traefik:v3` image
- Traefik exposes ports 80 and 443
- Traefik mounts docker socket: `/var/run/docker.sock:/var/run/docker.sock:ro`
- Traefik mounts config directory: `./traefik:/etc/traefik`
- api service has Traefik labels for routing `kanso.local/api` → api:8080
- couchdb service has Traefik labels for routing `kanso.local/db` → couchdb:5984
- infra/traefik/traefik.yml exists with entrypoints (web:80, websecure:443) and providers config
- infra/traefik/dynamic.yml exists with TLS cert resolver or self-signed config
- docker-compose.yml syntax is valid (docker compose config succeeds)
</acceptance_criteria>

<action>
1. Create `infra/traefik/traefik.yml`:
   ```yaml
   entryPoints:
     web:
       address: ":80"
     websecure:
       address: ":443"

   providers:
     docker:
       exposedByDefault: false
     file:
       filename: /etc/traefik/dynamic.yml

   log:
     level: INFO

   api:
     dashboard: false
   ```

2. Create `infra/traefik/dynamic.yml`:
   ```yaml
   tls:
     certificates:
       - certFile: /etc/traefik/certs/kanso.local.pem
         keyFile: /etc/traefik/certs/kanso.local-key.pem

   http:
     middlewares:
       redirect-https:
         redirectScheme:
           scheme: https
           permanent: true
   ```

3. Update `infra/docker-compose.yml`:
   - Add `traefik` service:
   ```yaml
   traefik:
     image: traefik:v3
     container_name: kanso-traefik
     ports:
       - "80:80"
       - "443:443"
     volumes:
       - /var/run/docker.sock:/var/run/docker.sock:ro
       - ./traefik:/etc/traefik
       - ./traefik/certs:/etc/traefik/certs
     restart: unless-stopped
   ```

   - Add Traefik labels to `api` service:
   ```yaml
   labels:
     - "traefik.enable=true"
     - "traefik.http.routers.api.rule=Host(`kanso.local`) && PathPrefix(`/api`)"
     - "traefik.http.routers.api.entrypoints=websecure"
     - "traefik.http.routers.api.tls=true"
     - "traefik.http.services.api.loadbalancer.server.port=8080"
   ```

   - Add Traefik labels to `couchdb` service:
   ```yaml
   labels:
     - "traefik.enable=true"
     - "traefik.http.routers.couchdb.rule=Host(`kanso.local`) && PathPrefix(`/db`)"
     - "traefik.http.routers.couchdb.entrypoints=websecure"
     - "traefik.http.routers.couchdb.tls=true"
     - "traefik.http.services.couchdb.loadbalancer.server.port=5984"
     - "traefik.http.routers.couchdb.middlewares=couchdb-strip"
     - "traefik.http.middlewares.couchdb-strip.stripprefix.prefixes=/db"
   ```
</action>

---

### Task 5: Create Makefile

**type**: execute
**wave**: 1
**depends_on**: []
**files_modified**:
  - Makefile

<read_first>
- infra/docker-compose.yml
- frontend/package.json
- backend/go.mod
</read_first>

<acceptance_criteria>
- Root Makefile exists
- `make up` runs `cd infra && docker compose up -d`
- `make down` runs `cd infra && docker compose down`
- `make dev` runs `cd frontend && pnpm dev`
- `make test` runs frontend tests (`cd frontend && pnpm test`) and backend tests (`cd backend && go test ./...`)
- `make build` builds frontend (`cd frontend && pnpm build`) and backend (`cd backend && go build ./cmd/kanso-api`)
- `make logs` runs `cd infra && docker compose logs -f`
- Makefile has `.PHONY` declaration for all targets
- Makefile has a `help` target that lists all commands
</acceptance_criteria>

<action>
1. Create root `Makefile`:
   ```makefile
   .PHONY: up down dev test build logs help

   help: ## Show this help
   	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

   up: ## Start infrastructure (CouchDB, API, Traefik)
   	cd infra && docker compose up -d

   down: ## Stop infrastructure
   	cd infra && docker compose down

   dev: ## Start frontend dev server
   	cd frontend && pnpm dev

   test: ## Run all tests
   	cd frontend && pnpm test
   	cd backend && go test ./...

   build: ## Build frontend and backend
   	cd frontend && pnpm build
   	cd backend && go build -o bin/kanso-api ./cmd/kanso-api

   logs: ## Follow infrastructure logs
   	cd infra && docker compose logs -f
   ```
</action>

---

### Task 6: Document nlp-service as v2 deferred feature

**type**: execute
**wave**: 1
**depends_on**: []
**files_modified**:
  - nlp-service/README.md

<read_first>
- .planning/PROJECT.md
</read_first>

<acceptance_criteria>
- nlp-service/README.md exists
- README states this is a placeholder for v2
- README describes planned stack: Python + FastAPI + transformers
- README describes purpose: NLP emotion analysis of diary entries in Portuguese
- README links to PROJECT.md or ROADMAP.md for context
</acceptance_criteria>

<action>
1. Create `nlp-service/README.md`:
   ```markdown
   # NLP Service (v2 — Deferred)

   This directory is reserved for the NLP analysis service, planned for v2.

   ## Purpose

   Analyze diary entries using natural language processing to detect emotions
   in Portuguese (pt-BR) and enrich registrations with detected emotion tags.

   ## Planned Stack

   - **Python 3.12**
   - **FastAPI** — REST API
   - **transformers** — Pre-trained emotion classification model
   - **Portuguese language model** — e.g., BERTimbau or similar

   ## Architecture

   The NLP service will be an internal service (not exposed to the frontend).
   The Go backend will call it asynchronously during report generation or
   as a background job when new entries are synced.

   ## Status

   Deferred to v2. The MVP focuses on manual emotion registration with
   user-defined sentiment fields. Automated emotion detection via NLP
   is a future enhancement.

   See: [PROJECT.md](../.planning/PROJECT.md) | [ROADMAP.md](../.planning/ROADMAP.md)
   ```
</action>

---

## Verification

After all tasks complete:

1. **Frontend env vars**: `grep -r "kanso.local" frontend/src/services/` returns nothing
2. **CORS configured**: `grep "go-chi/cors" backend/cmd/kanso-api/main.go` finds import
3. **Vite proxy**: `grep "proxy" frontend/vite.config.ts` finds proxy config
4. **Traefik service**: `grep "traefik" infra/docker-compose.yml` finds traefik service
5. **Makefile targets**: `make help` shows all targets
6. **nlp-service docs**: `test -f nlp-service/README.md` passes
7. **Backend compiles**: `cd backend && go build ./...` succeeds
8. **Frontend type-checks**: `cd frontend && npx tsc --noEmit` succeeds
9. **Tests pass**: `cd frontend && pnpm test` and `cd backend && go test ./...` both succeed
