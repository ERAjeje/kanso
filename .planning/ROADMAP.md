# Roadmap: Kanso

## Overview

Kanso is an offline-first therapeutic emotion diary PWA that helps users name and record emotions as they happen, with cloud sync and PDF report generation for sharing with a therapist. This roadmap covers the v1 MVP: authentication, emotion registration with offline-first sync, and PDF report generation.

## Phases

- [x] **Phase 1: Foundation & Authentication** — Infrastructure setup, Google OAuth, JWT auth, CouchDB isolation
- [ ] **Phase 2: Core Diary — Registro & Sync** - Emotion registration form, offline-first PouchDB storage, cloud sync
- [x] **Phase 3: Reports** - PDF report generation with async job infrastructure
- [x] **Phase 4: Technical Debt & Dev Experience** - Env vars, CORS, Vite proxy, Traefik, Makefile, nlp-service docs

## Phase Details

### Phase 1: Foundation & Authentication
**Goal**: Infrastructure is operational and users can securely access their accounts
**Mode**: mvp
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, SYNC-03
**Success Criteria** (what must be TRUE):
  1. User can sign in with their Google account via OAuth button on the login screen
  2. User stays authenticated across page refreshes (JWT persisted in storage)
  3. User can sign out from any page, which clears their session
  4. CouchDB is configured with per-user data isolation — users can only access their own documents
  5. The full development environment starts with a single `docker compose up` command
**Plans**: 3 (Plans 01-03)
  **Waves**:
  - **Wave 1** *(Plans 01, 02)* — Monorepo scaffold + Docker infra + Google OAuth/JWT auth flow
  - **Wave 2** *(blocked on Wave 1 completion)* *(Plan 03)* — CouchDB per-user isolation, PouchDB sync, sign-out
  **Cross-cutting constraints**:
  - All routes enforce HTTPS (Traefik TLS termination)
  - JWT tokens signed server-side with same secret shared via env
  - CouchDB proxy auth via Traefik (X-Auth-CouchDB-UserName header) — not native JWT
  - All 4 databases created in Phase 1 with validate_doc_update enforcing userId === userCtx.name
**UI hint**: yes

### Phase 2: Core Diary — Registro & Sync
**Goal**: Users can register emotions with structured data that persists locally and syncs to the cloud
**Mode**: mvp
**Depends on**: Phase 1
**Requirements**: REG-01, REG-02, REG-03, SYNC-01, SYNC-02
**Success Criteria** (what must be TRUE):
  1. User can submit a new registration with all 4 fields — sensações, sentimento, contexto, pensamentos — plus date/time
  2. User can backdate a registration to a past moment using a date/time picker
  3. Sentimento combobox shows previously used sentiments and allows typing new ones on the fly
  4. Registrations survive browser refresh and are visible when offline (stored in PouchDB)
  5. Registrations automatically sync to CouchDB when internet connectivity becomes available
**Plans**: 3 plans in 2 waves
**Waves**:
- **Wave 1** *(Plan 01)* — Infrastructure & Data Layer: types, pouchdb sync events, registros service, validation, usePouchSync hook, test setup
- **Wave 2** *(blocked on Wave 1 completion)* *(Plans 02 & 03, parallel)* — Registration Form + Sentimento Combobox + Toast / TabBar layout + SyncStatus enhancement + placeholder pages + routing restructure
  **Cross-cutting constraints**:
  - Tailwind v4 tokens from UI-SPEC.md (color: indigo-600 accent, spacing: 4-point scale, typography: text-2xl/text-sm/text-xs)
  - Form validation via Zod before PouchDB write
  - No Go backend changes — PouchDB syncs directly to CouchDB
  - @headlessui/react v2 Combobox for sentimento, lucide-react for icons
  - Brazilian Portuguese (pt-BR) for all user-facing copy
**UI hint**: yes
**Completed**: 2026-05-16

### Phase 3: Reports
**Goal**: Users can generate and download PDF reports of their emotional history
**Mode**: mvp
**Depends on**: Phase 2
**Requirements**: REL-01, REL-02, REL-03
**Success Criteria** (what must be TRUE):
  1. User can trigger PDF report generation from the app interface
  2. Report covers the period from the last report generation date to the current date
  3. Report includes all registered emotions in the period with date/time, sensações, sentimento, contexto, and pensamentos
  4. Report generation runs asynchronously — user can continue using the app while it processes
  5. User can check report status and download the PDF when ready
**Plans**: 2 plans in 2 waves
**Waves**:
- **Wave 1** *(Plan 01)* — Backend: report generation handler, chromedp PDF rendering, async job queue with status polling
- **Wave 2** *(blocked on Wave 1 completion)* *(Plan 02)* — Frontend: ReportJob types, reports API client, ReportSection component, Profile integration, tests
**UI hint**: yes
**Completed**: 2026-05-16

### Phase 4: Technical Debt & Dev Experience
**Goal**: Enable localhost development without Docker/Traefik dependency and resolve accumulated tech debt
**Mode**: standard
**Depends on**: Phase 1, Phase 2, Phase 3
**Requirements**: None (tech debt resolution)
**Success Criteria** (what must be TRUE):
  1. Frontend uses `VITE_API_URL` and `VITE_COUCHDB_URL` env vars instead of hardcoded URLs
  2. Backend has CORS middleware configured for `localhost:5173` and `kanso.local`
  3. Vite dev server proxies `/api` and `/db` to backend and CouchDB
  4. Traefik service added to docker-compose with TLS routing
  5. Makefile provides unified commands (`up`, `down`, `dev`, `test`, `build`, `logs`)
  6. `nlp-service/README.md` documents v2 deferred feature
  7. All existing tests pass after changes
**Plans**: 1 plan (single wave)
**Waves**:
- **Wave 1** *(Plan 01)* — All tech debt items: env vars → CORS → Vite proxy → Traefik → Makefile → nlp-service docs
**UI hint**: no

### Phase 5: Histórico de Registros
**Goal**: Users can browse their emotional history chronologically, with inline detail expansion
**Mode**: mvp
**Depends on**: Phase 2
**Requirements**: REG-01, REG-03
**Success Criteria** (what must be TRUE):
  1. History page shows registrations in reverse chronological order
  2. Each card displays sentimentoNome (or friendly placeholder "Buscando sentimento"), date/time, and a content preview
  3. Tapping a card expands it inline to show all fields — sensações, sentimento, contexto, pensamentos
  4. Friendly fallback text when sentimentoNome is null: "Buscando sentimento"
  5. Loading, empty, and error states are handled gracefully
  6. All new code has passing tests
**Plans**: 1 plan (single wave)
**Waves**:
- **Wave 1** *(Plan 01)* — Service method + RegistroCard component + History page + tests
**UI hint**: yes
**Completed**: 2026-05-17

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Authentication | 3/3 | Complete | 2026-05-16 |
| 2. Core Diary — Registro & Sync | 3/3 | Complete | 2026-05-16 |
| 3. Reports | 2/2 | Complete | 2026-05-16 |
| 4. Technical Debt & Dev Experience | 1/1 | Complete | 2026-05-17 |
| 5. Histórico de Registros | 1/1 | Complete | 2026-05-17 |
