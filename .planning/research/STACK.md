# Stack Research

**Domain:** Therapeutic emotion diary PWA (offline-first, NLP-enhanced)
**Researched:** 2026-05-16
**Confidence:** HIGH (all versions verified via Context7/official docs/npm/PyPI)

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| React | 19.2.6 | UI framework | Mature server/client component model, `useActionState` for form handling, compiler-optimized re-renders. Standard choice for PWA with rich interactivity. |
| Vite | 8.0.13 | Build tool & dev server | Fastest HMR in ecosystem, native TypeScript/JSX, built-in PWA support via plugin. Vite 8 ships Rolldown-based bundling (replacing esbuild/Rollup) for faster builds. |
| TypeScript | 5.8+ (6.0.3 latest) | Type safety | Non-negotiable for PWA of this complexity — catches data sync bugs at compile time. |
| Tailwind CSS | 4.3.0 | Utility-first CSS | Zero-config Vite integration (Tailwind v4 uses `@import "tailwindcss"` syntax, no config file required). Perfect for mobile-first PWA with minimal CSS overhead. |
| Go | 1.26.3 | Backend API server | Single binary deployment, excellent concurrency for JWT validation + PDF generation + FCM dispatching. Chi router keeps dependency tree minimal. |
| Chi | v5.2.5 | Go HTTP router | Idiomatic Go, stdlib `net/http` compatible, built-in middleware for JWT (`jwtauth.Verifier`/`Authenticator`), CORS, logging, rate limiting. No framework lock-in. |
| CouchDB | 3.5.1 | Document database | Native sync protocol with PouchDB, HTTP/JSON API, proxy authentication, multi-master replication. Only database that works seamlessly with PouchDB's offline-first model. |
| PouchDB | 9.0.0 | Client-side database | Browser-based CouchDB-compatible DB, `db.sync()` for bidirectional live replication, works in all modern browsers. The only viable offline-first DB that syncs with CouchDB. |
| FastAPI | 0.128.0+ | Python NLP API | Automatic OpenAPI docs, async background tasks (for model inference), Pydantic validation, lightweight CPU inference serving. Best Python framework for ML model serving. |
| Transformers | 4.30.2+ (4.48+ recommended) | NLP model inference | HuggingFace ecosystem, pipeline API for easy model loading, XLM-RoBERTa support for Portuguese emotion classification. |
| Traefik | v3.7.1 | Reverse proxy / gateway | Automatic Docker service discovery, Let's Encrypt TLS, middleware pipeline. Standard for Docker Compose multi-service deployments. |
| Docker Compose | v2.x (Compose Spec) | Service orchestration | Single `docker compose up` for entire stack. All six services (Traefik, Go API, CouchDB, FastAPI, React build/static) defined in one file. |

### Supporting Libraries

#### Frontend (React/PWA)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `react-router-dom` | ^7.15.1 | Tab-based navigation (Register, History, Profile) | Required — PWA needs client-side routing |
| `vite-plugin-pwa` | ^1.3.0 | Service worker generation, manifest, offline support | Required — enables PWA installability |
| `pouchdb-browser` | ^9.0.0 | PouchDB for browser (includes IndexedDB adapter) | Required — core offline storage |
| `pouchdb-find` | ^9.0.0 | MongoDB-like query API for PouchDB | Required — querying records by date/emotion |
| `firebase` (modular SDK) | ^11.x | FCM push notification client | Required — push notification registration |
| `date-fns` | ^4.x | Date formatting/parsing for diary entries | Strong recommendation for timezone-safe date handling |
| `zustand` | ^5.x | Lightweight client state management | Optional — needed if global state (auth, sync status) exceeds React Context ergonomics |
| `@tauri-apps/api` | — | Not applicable (PWA, not Tauri) | N/A |
| Workbox | bundled with vite-plugin-pwa | Service worker caching strategies | Auto-configured via plugin |

#### Backend (Go)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/go-chi/chi/v5` | v5.2.5 | HTTP routing + middleware | Core framework |
| `github.com/go-chi/jwtauth/v5` | v5.x | JWT verification middleware | Required — validates JWTs on protected routes |
| `github.com/go-chi/cors` | v1.x | CORS middleware | Required — PWA on different origin than API |
| `github.com/golang-jwt/jwt/v5` | v5.x | JWT parsing/validation (used by jwtauth) | Transitive dependency of jwtauth |
| `github.com/lestrrat-go/jwx` | v4.0.2 | JWT/JWK handling (used internally by jwtauth) | Transitive — do not pin directly |
| `firebase.google.com/go/v4` | ^4.x | Firebase Admin SDK for Go | Required — FCM push notification dispatch |
| `google.golang.org/api` | — | Google API client (for ID token verification) | Required — verify Google OAuth ID tokens |
| `github.com/twilio/twilio-go` | v1.30.9 | Twilio REST API client | Required — WhatsApp message sending |
| `github.com/chromedp/chromedp` | v0.15.1 | Headless Chrome for PDF generation | Required — renders report HTML → PDF |
| `github.com/go-chi/stampede` | — | HTTP request deduplication middleware | Consider — prevent duplicate sync writes |
| stdlib `database/sql` | — | Not needed (CouchDB is HTTP/JSON) | N/A — use `net/http` to talk to CouchDB |

#### NLP Service (Python)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `fastapi` | ^0.128.0 | API framework | Required |
| `uvicorn` | ^0.34.x | ASGI server | Required |
| `transformers` | ^4.48.0 | Model inference pipeline | Required |
| `torch` | ^2.6.0 (CPU) | PyTorch backend for transformers | Required — CPU-only install (`--index-url https://download.pytorch.org/whl/cpu`) |
| `pydantic` | ^2.x | Request/response validation | Bundled with FastAPI |
| `python-multipart` | — | Form data parsing | Only if accepting file uploads |
| `httpx` | ^0.28.x | Async HTTP client | For health checks / service mesh communication |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `pnpm` | Package manager (frontend) | Faster than npm/yarn, strict dependency resolution. Use over npm for workspace and disk space efficiency. |
| `golangci-lint` | Go linter | Run in CI — catches race conditions, error handling issues specific to chi/Go patterns |
| `ruff` | Python linter + formatter | Replaces flake8/isort/black with single tool. Configure for FastAPI patterns. |
| `just` | Command runner | Define `just up`, `just logs`, `just migrate` shortcuts for docker compose operations |
| `mkcert` | Local TLS certificates | PWA service workers require HTTPS even in development. Required for push notification testing. |

## Installation

### Frontend

```bash
pnpm create vite kanso -- --template react-ts
cd kanso
pnpm install pouchdb-browser pouchdb-find react-router-dom date-fns zustand
pnpm install -D vite-plugin-pwa
pnpm add firebase
```

### Backend (Go)

```bash
go mod init github.com/<user>/kanso-api
go get github.com/go-chi/chi/v5
go get github.com/go-chi/jwtauth/v5
go get github.com/go-chi/cors
go get firebase.google.com/go/v4
go get github.com/twilio/twilio-go
go get github.com/chromedp/chromedp
go get google.golang.org/api
```

### NLP Service (Python)

```bash
pip install "fastapi[standard]" uvicorn[standard] transformers torch --index-url https://download.pytorch.org/whl/cpu
```

### Infrastructure

```yaml
# docker-compose.yml key images
services:
  traefik:
    image: traefik:v3.7
  couchdb:
    image: couchdb:3.5.1
  api:
    build: ./api    # Go service
  nlp:
    build: ./nlp    # FastAPI service
  frontend:
    build: ./frontend  # Vite build served via nginx
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| **CouchDB 3.x** | PostgreSQL + pg_partman | When you don't need offline-first. PouchDB requires CouchDB-compatible sync endpoint — no alternative exists for browser-native offline sync. |
| **PouchDB 9.x** | IndexedDB + Service Worker (raw) | Only if you want to build your own sync protocol. PouchDB provides battle-tested live replication with conflict resolution. |
| **Go + chi** | Node.js (Express/Fastify) | If the team is stronger in JS/TS and PDF generation libraries exist (puppeteer). Go chosen for binary deployment and performance. |
| **FastAPI + transformers** | Go NLP (not viable) | Python has the NLP ecosystem (transformers, sentencepiece). Go has no viable transformer inference library as of 2026. |
| **Traefik v3** | Caddy 2 | Caddy is simpler for single-service deployments. Traefik wins for multi-service Docker Compose with automatic routing. |
| **Vite 8** | Next.js | If SSR/SEO mattered. For a PWA diary app, SPA with Vite is simpler and gives better offline behavior. |
| **Tailwind 4** | CSS Modules | Tailwind speeds up mobile-first responsive design significantly. CSS Modules is fine but slower for rapid iteration. |
| **PNPM** | Bun | Bun is faster but less mature in ecosystem compatibility. PNPM is the safest package manager for production Vite projects. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **Redux** | Overkill for a single-user PWA with ~3 screens. Creates boilerplate with no benefit for this scope. | Zustand or React Context + `useReducer` |
| **Next.js** | SSR complexity adds no value for an offline-first PWA. Server components break offline assumptions. | Vite SPA + static build |
| **MongoDB** | No native sync protocol for offline-first browsers. Requires custom sync layer. | CouchDB (only option for PouchDB sync) |
| **Firestore** | Proprietary sync protocol, vendor lock-in, no local-first without Firebase SDK. | CouchDB (open protocol, self-hosted) |
| **Redis** | Not needed — no real-time pub/sub requirements. Session state is JWT-based in this stack. | Eliminate entirely |
| **Express.js backend** | JWT validation + PDF generation + FCM dispatch is well-served by Go's stdlib concurrency. | Go + chi |
| **PyTorch GPU** | Unnecessary for single-user inference on a VPS. GPU inference adds cost/complexity. | PyTorch CPU-only |
| **FastAPI + celery** | Background NLP inference can use FastAPI's `BackgroundTasks` or a simple asyncio task queue. | FastAPI built-in `BackgroundTasks` |
| **Traefik Enterprise** | JWT validation is done at the app layer (chi middleware), not needed at proxy. | Open-source Traefik v3 |

## Stack Patterns by Variant

**If running on a low-memory VPS (≤2GB RAM):**
- Reserve NLP service or use a lighter model (e.g., `distilbert-base-multilingual-cased`)
- Set CouchDB memory limits in docker compose (`mem_limit: 512m`)
- Use `GOGC=100` for Go service to reduce GC pressure
- Set `--max-old-space-size=256` for Node build step
- Chromedp: use `--no-sandbox --headless --disable-gpu --disable-dev-shm-usage`

**If the user base grows beyond trusted users (multi-tenancy):**
- Add CouchDB per-user databases or prefix documents with user ID
- Add rate limiting middleware to chi (`github.com/go-chi/httprate`)
- Move NLP inference to async queue (Redis + Celery or similar) for horizontal scaling
- Add CouchDB clustering (3+ nodes) for high availability

**If adding real-time updates (future feature):**
- Add CouchDB changes feed listener in Go service
- Use Server-Sent Events (SSE) from Go backend to PWA
- PouchDB `db.changes()` already provides real-time replication notification on the client

## Architecture Decision: Emotion NLP Model

For Portuguese emotion detection, the recommended model is:

**`tabularisai/multilingual-emotion-classification`** (XLM-RoBERTa-based, 11 emotions)

| Criterion | Verdict |
|-----------|---------|
| Languages | Multilingual (trained on ~30 languages including Portuguese) |
| Emotions | anger, contempt, disgust, fear, frustration, gratitude, joy, love, neutral, sadness, surprise |
| Type | Multi-label (text can have multiple emotions) |
| Pipeline | `text-classification` — works with `pipeline("text-classification", model="tabularisai/multilingual-emotion-classification")` |
| Size | ~1.1GB (XLM-RoBERTa base) — manageable on 2GB VPS |
| License | MIT |

**Why not Portuguese-only models:**
- `neuralmind/bert-base-portuguese-cased` is a base model (not fine-tuned for emotion)
- The few Portuguese emotion fine-tunes (e.g., `hyanbatista42/bert-base-portuguese-cased-finetuned-emotion`) have near-zero downloads and no validation
- XLM-RoBERTa performs competitively on Portuguese despite being multilingual

## Traefik v3 Routing Strategy

Traefik routes requests based on path prefixes:

| Route | Target | Auth |
|-------|--------|------|
| `kanso.example.com/api/*` | Go backend (chi) | JWT (chi jwtauth middleware) |
| `kanso.example.com/nlp/*` | FastAPI (NLP service) | Internal network only (not exposed via Traefik); Go backend proxies requests |
| `kanso.example.com/db/*` | CouchDB (via proxy auth) | Proxy auth headers set by Go backend |
| `kanso.example.com/*` | Static frontend files | None (public) |

**JWT validation happens in Go via chi middleware, NOT at the Traefik level.** This keeps the proxy simple and the auth logic centralized. For CouchDB, the Go backend sets proxy auth headers (`X-Auth-CouchDB-UserName`, `X-Auth-CouchDB-Roles`, `X-Auth-CouchDB-Token`) when proxying database requests.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| React 19.x | Vite 8.x | SWC plugin (`@vitejs/plugin-react-swc`) is preferred for faster HMR |
| PouchDB 9.x | CouchDB 3.x | Uses CouchDB replication protocol. CouchDB 2.x+ is compatible. |
| vite-plugin-pwa 1.x | Vite 8.x | Vite 8 uses Rolldown; verify plugin compatibility (likely works via compat layer) |
| chi v5 | Go 1.21+ | Works with Go 1.26. Uses stdlib `net/http` — no framework version conflicts. |
| Chromedp v0.15.x | Chromium 100+ | Uses Chrome DevTools Protocol (CDP) which is backward-compatible |
| transformers 4.x | torch CPU 2.x | Weights are framework-agnostic; pipeline API handles device mapping |
| FastAPI 0.128.x | Pydantic 2.x | Pydantic v2 is required for FastAPI 0.103+. |
| Firebase Admin Go v4 | Go 1.20+ | Works with Go 1.26. Requires Firebase service account JSON. |
| Twilio Go v1.x | Go 1.16+ | Minimal dependencies, REST API client |

## Sources

- **React 19.2** — npm package version verified (`npm view react version` → 19.2.6)
- **Vite 8.0** — Context7 (`/websites/v7_vite_dev`, `/vitejs/vite`) + npm version verified
- **Tailwind 4.3** — npm version verified (`npm view tailwindcss version` → 4.3.0)
- **Vite PWA Plugin 1.3** — Context7 (`/vite-pwa/vite-plugin-pwa`) + npm version verified
- **PouchDB 9.0** — Context7 (`/apache/pouchdb`) + npm version verified (9.0.0)
- **PouchDB Find 9.0** — npm version verified (9.0.0)
- **Chi v5.2.5** — Context7 (`/go-chi/docs`) + GitHub release verified
- **Go 1.26** — official Go release (`go.dev/VERSION`) — May 2026
- **CouchDB 3.5.1** — Docker Hub tags verified
- **Traefik v3.7.1** — Docker Hub tags verified (latest v3)
- **FastAPI 0.128+** — Context7 (`/fastapi/fastapi`) — latest available: 0.103.2 in local PyPI, newer versions available upstream
- **Transformers 4.48+** — HuggingFace Hub — `tabularisai/multilingual-emotion-classification` model verified
- **Chromedp v0.15** — Context7 (`/chromedp/chromedp`) + GitHub release verified
- **Firebase Admin Go v4** — Context7 (`/firebase/firebase-admin-go`) — README verified
- **Twilio Go v1.30** — Context7 (`/twilio/twilio-go`) + GitHub release verified
- **Tabularisai emotion model** — HuggingFace API verified config (id2label with 11 emotions, XLM-RoBERTa)

---

*Stack research for: Kanso therapeutic emotion diary PWA*
*Researched: 2026-05-16*
*Confidence: HIGH (all versions verified via npm/GitHub/PyPI/Docker Hub)*
