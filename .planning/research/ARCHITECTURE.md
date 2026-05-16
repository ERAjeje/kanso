# Architecture Patterns: Offline-first PWA Emotion Diary

**Project:** Kanso  
**Researched:** 2026-05-16  
**Overall confidence:** HIGH

## System Overview

Kanso is an offline-first PWA emotion diary with CouchDB sync, Go backend API, Python NLP service, and Traefik ingress. The architecture centers on **CouchDB's multi-master replication** as the backbone: PouchDB in the browser syncs directly with CouchDB through Traefik, while the Go backend orchestrates side-effects (PDF generation, WhatsApp sending, NLP analysis) as a stateless API layer.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                    Browser (PWA)                     │
│  ┌──────────┐    ┌────────────────────────────────┐ │
│  │  React   │◄──►│  PouchDB (IndexedDB)           │ │
│  │  (Vite)  │    │  - Local registros DB           │ │
│  │          │    │  - Live sync {live:true,        │ │
│  │          │    │              retry:true}         │ │
│  └────┬─────┘    └──────────┬─────────────────────┘ │
│       │                     │                       │
│       │ JWT in              │ JWT in Authorization  │
│       │ Authorization       │ (PouchDB sync URL)    │
└───────┼─────────────────────┼───────────────────────┘
        │                     │
        ▼                     ▼
┌─────────────────────────────────────────────────────┐
│                   Traefik Gateway                     │
│                                                      │
│  /api/* ──────► ForwardAuth ──► Go Backend            │
│  /db/*  ──────► JWT validate ──► CouchDB              │
│  /nlp/* ──────► [INTERNAL ONLY - no public route]     │
│                                                      │
│  Middleware chain per router:                         │
│  - /api/* : CORS, RateLimit, ForwardAuth(JWT)        │
│  - /db/*  : CORS, ForwardAuth(JWT→inject headers)    │
└──────────┬──────────────────────────────────────────┘
           │                    │              │
           ▼                    ▼              │
┌──────────────────┐  ┌──────────────────┐    │
│   CouchDB         │  │   Go Backend     │    │
│   (database)      │  │   (chi router)   │    │
│                   │  │                  │    │
│  Databases:       │  │  Handlers:       │    │
│  - registros      │  │  - /auth/google  │    │
│  - sentimentos    │  │  - /usuario      │    │
│  - usuarios       │  │  - /relatorios   │    │
│  - relatorios     │  │  - /whatsapp     │    │
│                   │  │  - /push         │    │
│  Auth: JWT native │  │  - /analise      │    │
│  or proxy headers │  │                  │    │
└──────────────────┘  └────────┬─────────┘    │
                               │              │
                               │ HTTP POST     │
                               │ /analyze      │
                               │              │
                               ▼              ▼
                    ┌──────────────────┐  ┌──────────────────┐
                    │  Python NLP      │  │  Headless Chrome │
                    │  (FastAPI)       │  │  (chromedp)      │
                    │                  │  │                  │
                    │  POST /analyze   │  │  Renders HTML    │
                    │  text→emotions   │  │  → PDF           │
                    │  Model: bert/pt   │  └──────────────────┘
                    │  (CPU inference) │
                    └──────────────────┘
```

### Network Segmentation

```
                    Public Internet
                           │
                    ┌──────┴──────┐
                    │   Traefik    │  Port 443 (HTTPS)
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         CouchDB       Go API      [NLP Service]
          :5984        :8080         :8000
                                    ╔══════════════╗
                                    ║  Internal    ║
                                    ║  network     ║
                                    ║  only        ║
                                    ╚══════════════╝
```

**Key design decision:** NLP service is NOT exposed through Traefik. It lives on an internal Docker network and is only reachable by the Go backend via `http://nlp-service:8000`. This eliminates a public attack surface and keeps the single-tenant NLP model accessible only to authorized server-side code.

---

## Component Boundaries

### 1. Frontend PWA (React + Vite + PouchDB + Tailwind)

| Responsibility | NOT responsible for |
|----------------|---------------------|
| Rendering UI (3-tab layout: Register, History, Profile) | Business logic for PDF, WhatsApp, NLP |
| Local-first data with PouchDB (IndexedDB) | Session management beyond JWT storage |
| JWT management (store, attach to requests, refresh) | Direct CouchDB admin operations |
| PouchDB ↔ CouchDB live sync with retry | Conflict resolution beyond simple last-write-wins |
| Offline queue (entries created offline sync when online) | Database schema migrations |
| Service Worker for PWA installability | |

**Key interaction pattern:** PouchDB syncs directly to CouchDB via `db.sync(remoteCouchUrl, {live: true, retry: true})`. This is the heart of offline-first — no Go backend involvement for CRUD. The Go API is only for authentication, side-effects (PDF, WhatsApp), and NLP triggering.

### 2. CouchDB

| Responsibility | NOT responsible for |
|----------------|---------------------|
| Document storage (registros, sentimentos, usuarios, relatorios) | Application logic |
| Replication target for PouchDB sync | User authentication (JWT validation at gateway/CouchDB level) |
| Per-database security (members/admins) | Serving the frontend |
| Changes feed for reactive processing | Session management |
| Mango indexes for query performance | |
| Validate-doc-update functions for data integrity | |

**Important CouchDB 3.x+ native JWT auth:** CouchDB 3.x supports validating JWT tokens natively via the `jwt_authentication_handler`. This means PouchDB can sync to CouchDB through Traefik with the JWT in `Authorization: Bearer <token>`, and CouchDB validates it directly — no need for proxy auth header injection. See [`[jwt_keys]` config](https://docs.couchdb.org/en/stable/api/server/authn.html#jwt-authentication). This is the **recommended approach** for Kanso (simplest, no header manipulation needed).

### 3. Go Backend (chi)

| Responsibility | NOT responsible for |
|----------------|---------------------|
| Google OAuth token validation | CRUD operations on registros (handled by PouchDB↔CouchDB) |
| JWT issuance (sign tokens on successful Google auth) | Serving static files |
| Report PDF generation (chromedp) | User registration beyond initial creation |
| WhatsApp sending via Twilio | |
| Push notification registration (FCM) | |
| NLP orchestration (read from CouchDB → call NLP → write back) | |
| JWT validation endpoint (for Traefik ForwardAuth) | |
| User preferences (CRUD via API, stored in CouchDB) | |

### 4. Python NLP Service

| Responsibility | NOT responsible for |
|----------------|---------------------|
| Emotion analysis from text | Direct database access |
| Model loading and inference | User management |
| Serving POST /analyze on internal port 8000 | Serving public endpoints |
| Returning structured emotion probabilities | PDF generation |

### 5. Traefik

| Responsibility | NOT responsible for |
|----------------|---------------------|
| TLS termination (Let's Encrypt via ACME) | Business logic |
| Path-based routing (/api/* → Go, /db/* → CouchDB) | User authentication (delegates to ForwardAuth or CouchDB JWT) |
| CORS headers | |
| Rate limiting | |
| ForwardAuth middleware for JWT validation | |
| Stripping path prefixes (/db/ when proxying to CouchDB) | |

---

## Data Flow

### Flow 1: User Registration / Login

```
1. User clicks "Sign in with Google"
2. PWA gets Google ID token (via OAuth flow)
3. PWA sends POST /api/auth/google { idToken } to Go backend
4. Go validates token with Google's API
5. Go extracts sub, email, name
6. Go creates/updates document in CouchDB `usuarios` database
7. Go signs a JWT with { sub, email, name, exp }
8. Go returns JWT to PWA
9. PWA stores JWT in memory/localStorage
10. PWA initializes PouchDB sync with CouchDB URL using JWT

POST /api/auth/google
Auth: none (public endpoint)
Rate limit: yes (prevent brute force)
```

### Flow 2: Offline-First Entry Creation

```
1. User opens PWA, navigates to "Register" tab
2. User fills in: sensações, sentimento, contexto, pensamentos
3. PWA saves to PouchDB local IndexedDB immediately (synchronous feel)
4. PouchDB live sync picks up the new doc
5. Sync replicates to CouchDB via Traefik (JWT in sync URL)
6. CouchDB validates JWT natively (or via proxy auth)
7. If offline: doc stays in PouchDB queue
8. When online: sync resumes automatically (retry:true)

Read path (History tab):
1. PWA queries PouchDB (local-first) using Mango queries
2. If online, PouchDB may also fetch from CouchDB
3. UI renders from local data (no loading spinner for cached data)

Document structure (registros database):
{
  "_id": "<uuid>",                          // UUID to avoid conflicts
  "type": "registro",
  "userId": "<google-sub>",                 // Owner
  "dataHora": "2026-05-16T14:30:00Z",
  "sensacoes": "aperto no peito",
  "sentimentoId": "<uuid-or-null>",
  "sentimentoNome": "ansiedade",
  "contexto": "pensando na reunião",
  "pensamentos": "será que vou conseguir?",
  "analiseEmocoes": {                       // Added by NLP pipeline
    "dataAnalise": "2026-05-16T15:00:00Z",
    "modelo": "j-hartmann/emotion-english-distilroberta-base",
    "resultados": [
      {"emocao": "ansiedade", "probabilidade": 0.85}
    ]
  },
  "createdAt": "2026-05-16T14:35:00Z",
  "updatedAt": "2026-05-16T15:00:00Z"
}
```

### Flow 3: NLP Analysis Pipeline

```
Option A: On-demand (MVP)
1. User creates entry → syncs to CouchDB
2. User (or background job) calls POST /api/analise/registro/{registroId}
3. Go fetches doc from CouchDB (as admin, via internal network)
4. Go builds text: sensações + " " + contexto + " " + pensamentos
5. Go calls POST http://nlp-service:8000/analyze { texto, idioma: "pt" }
6. NLP returns { emocoes: [{nome, prob}] }
7. Go updates doc.analiseEmocoes in CouchDB
8. Updated doc syncs to PouchDB via changes feed
9. UI shows emotion analysis on next render

Option B: Changes-feed driven (post-MVP)
1. Go backend watches CouchDB _changes feed on registros db
2. New doc detected → auto-triggers analysis (same flow as steps 3-7)
3. No user action needed
```

### Flow 4: Report PDF Generation

```
1. User navigates to Profile tab → "Generate Report"
2. User sees period (auto-calculated: last report date → today)
3. User confirms → POST /api/relatorios { dataInicio, dataFim }
4. Go creates job document in CouchDB relatorios db (status: "pendente")
5. Go returns { jobId } immediately (async pattern)
6. Go service:
   a. Queries CouchDB registros for userId + date range
   b. Groups registros with emotion analysis data
   c. Builds HTML template (Go html/template)
   d. Generates SVG bar chart for emotion frequencies
   e. Launches chromedp (headless Chrome) to render HTML → PDF
   f. Stores PDF bytes (option: return in response, or upload somewhere)
   g. Updates job status to "concluido"
7. PWA polls GET /api/relatorios/{jobId}/status
8. When "concluido", offers download or send-to-WhatsApp

chromedp resource considerations:
- Use chromedp/headless-shell Docker image (~150MB)
- Allocate 512MB memory per Chrome instance
- Single concurrent PDF generation for MVP (mutex per userId)
- Context timeout: 30s max per PDF
```

### Flow 5: WhatsApp Delivery

```
1. Report generated (Flow 4) with job status "concluido"
2. User clicks "Send to therapist" → POST /api/whatsapp/enviar { jobId }
3. Go fetches report job + registros data
4. Go rebuilds PDF (or reuses cached)
5. Go calls Twilio API:
   - Media URL points to PDF (temporary signed URL)
   - To: therapist's WhatsApp number (from user settings)
   - Body: "Relatório de {dataInicio} a {dataFim}"
6. Twilio sends WhatsApp message with PDF attachment
7. Go returns { status: "enviado" }
```

---

## Patterns to Follow

### Pattern 1: Offline-First with PouchDB Live Sync

**What:** PouchDB syncs bidirectionally with CouchDB using live replication with auto-retry.

**When:** Any data that must be available offline and sync when connectivity returns.

**Implementation:**
```javascript
// PWA setup
const localDB = new PouchDB('kanso_registros');
const remoteDB = new PouchDB(`https://api.kanso.app/db/registros`, {
  fetch: (url, opts) => {
    opts.headers.set('Authorization', `Bearer ${getJWT()}`);
    return PouchDB.fetch(url, opts);
  }
});

// Live sync with auto-retry
localDB.sync(remoteDB, {
  live: true,    // continuous replication
  retry: true    // auto-reconnect on connection loss
}).on('change', handleChange)
  .on('paused', () => console.log('offline'))
  .on('active', () => console.log('online'))
  .on('error', handleError);

// Save locally (works offline)
async function saveRegistro(data) {
  const doc = {
    _id: uuidv4(),
    type: 'registro',
    userId: getUserId(),
    dataHora: data.dataHora,
    sensacoes: data.sensacoes,
    // ...
    createdAt: new Date().toISOString()
  };
  return await localDB.put(doc);  // Immediate local save
}
```

**Key considerations:**
- Use `_id: uuidv4()` not sequential IDs — prevents conflicts from concurrent offline writes
- The `retry: true` option handles network flakiness automatically (mobile connectivity)
- Cancel sync on logout: `syncHandler.cancel()`
- Each database type gets its own PouchDB instance (registros, sentimentos)

### Pattern 2: JWT-Native CouchDB Authentication

**What:** CouchDB 3.x validates JWT tokens natively, eliminating the need for proxy auth header injection.

**When:** Using CouchDB 3.x+ with direct PouchDB-to-CouchDB sync through a gateway.

**Configuration:**
```ini
# CouchDB local.ini
[chttpd]
authentication_handlers = {chttpd_auth, jwt_authentication_handler}, {chttpd_auth, default_authentication_handler}

[jwt_auth]
; The "sub" claim in JWT becomes the CouchDB username
required_claims = exp,iat

[jwt_keys]
; HMAC-SHA256 key (base64-encoded shared secret)
; This MUST match the signing key used by Go backend
hmac:_default = c2VjdXJlLWtleS1mb3Itam93LXNpZ25pbmc=
```

**How it works:**
1. Go backend signs JWT with `{ sub: google-sub, email, name, exp, iat }`
2. PWA includes JWT in PouchDB sync URL: `Authorization: Bearer <jwt>`
3. Traefik proxies `/db/*` to CouchDB (with or without StripPrefix)
4. CouchDB validates JWT using key from `[jwt_keys]`
5. CouchDB's `sub` claim becomes the authenticated user name
6. Database `_security` uses this name for access control

**Database security (per-database):**
```json
// PUT /db/registros/_security
{
  "members": {
    "names": ["user1", "user2"],
    "roles": []
  },
  "admins": {
    "names": ["couchdb-admin"],
    "roles": ["_admin"]
  }
}
```

**Why this over proxy auth:** No custom header injection, no ForwardAuth complexity for CouchDB traffic, CouchDB validates cryptography directly — simpler and more secure.

### Pattern 3: Go chi Layered Architecture

**What:** Standard Go project layout with handler → service → repository layers, plus middleware.

**When:** Any Go HTTP service with business logic and external dependencies.

**Structure:**
```
backend/
├── cmd/
│   └── kanso-api/
│       └── main.go                    # Entry point: config init, DI setup, router mount
├── internal/
│   ├── config/
│   │   └── config.go                  # Env-based config (envconfig or viper)
│   ├── handler/
│   │   ├── auth.go                    # Auth handlers (Google token validation)
│   │   ├── usuario.go                 # User preferences CRUD handlers
│   │   ├── relatorio.go              # Report generation handlers
│   │   ├── whatsapp.go                # WhatsApp sending handlers
│   │   ├── push.go                    # Push token registration handlers
│   │   └── analise.go                 # NLP analysis handlers
│   ├── service/
│   │   ├── auth.go                    # Google token verification, JWT signing
│   │   ├── usuario.go                 # User preferences business logic
│   │   ├── relatorio.go              # PDF generation orchestration
│   │   ├── whatsapp.go                # Twilio API client
│   │   ├── push.go                    # FCM integration
│   │   └── nlp.go                     # NLP HTTP client
│   ├── repository/
│   │   └── couchdb.go                 # CouchDB data access (for admin operations)
│   ├── model/
│   │   ├── registro.go                # Domain types
│   │   ├── usuario.go
│   │   └── relatorio.go
│   ├── middleware/
│   │   ├── auth.go                    # JWT validation middleware for API routes
│   │   └── logging.go                 # Structured request logging
│   └── pdf/
│       └── generator.go              # chromedp HTML→PDF rendering
├── go.mod
├── go.sum
└── Dockerfile                         # Multi-stage: build → distroless or alpine
```

**`cmd/kanso-api/main.go` pattern:**
```go
func main() {
    cfg := config.Load()

    // Repository layer
    couchRepo := repository.NewCouchDB(cfg.CouchDBURL, cfg.CouchDBAdmin, cfg.CouchDBPassword)

    // Service layer
    authSvc := service.NewAuth(cfg.GoogleClientID, cfg.JWTSecret)
    reportSvc := service.NewReport(couchRepo, cfg.NLPServiceURL)
    whatsappSvc := service.NewWhatsApp(cfg.TwilioAccountSID, cfg.TwilioAuthToken)
    pushSvc := service.NewPush(cfg.FCMServerKey)
    nlpSvc := service.NewNLP(cfg.NLPServiceURL)

    // HTTP layer
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    // Public routes
    r.Post("/api/auth/google", handler.NewAuth(authSvc, couchRepo).HandleGoogleLogin)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.JWT(cfg.JWTSecret))
        r.Get("/api/usuario", handler.NewUsuario(couchRepo).Get)
        r.Put("/api/usuario", handler.NewUsuario(couchRepo).Update)
        r.Post("/api/relatorios", handler.NewRelatorio(reportSvc).Create)
        r.Get("/api/relatorios/{jobId}/status", handler.NewRelatorio(reportSvc).GetStatus)
        r.Get("/api/relatorios/{jobId}/download", handler.NewRelatorio(reportSvc).Download)
        r.Post("/api/whatsapp/enviar", handler.NewWhatsApp(whatsappSvc, reportSvc).Send)
        r.Post("/api/push/registrar", handler.NewPush(pushSvc).Register)
        r.Post("/api/analise/registro/{registroId}", handler.NewAnalise(nlpSvc, couchRepo).Analyze)
    })

    log.Fatal(http.ListenAndServe(":8080", r))
}
```

### Pattern 4: PDF Generation with chromedp

**What:** Use headless Chrome via chromedp to render HTML templates to PDF, preserving CSS styling.

**When:** Need pixel-perfect PDF from HTML with CSS support (tables, colors, fonts, SVG charts).

**Implementation:**
```go
// internal/pdf/generator.go
type Generator struct {
    allocCtx context.Context
    cancel   context.CancelFunc
}

func NewGenerator() (*Generator, error) {
    // Use chromedp/headless-shell in production
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
    )
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    return &Generator{allocCtx: allocCtx, cancel: cancel}, nil
}

func (g *Generator) GeneratePDF(htmlContent string) ([]byte, error) {
    ctx, cancel := chromedp.NewContext(g.allocCtx)
    defer cancel()

    var buf []byte
    err := chromedp.Run(ctx, chromedp.Tasks{
        chromedp.Navigate("about:blank"),
        chromedp.ActionFunc(func(ctx context.Context) error {
            // Set HTML content
            frameTree, err := page.GetFrameTree().Do(ctx)
            // ... inject HTML
            return nil
        }),
        chromedp.ActionFunc(func(ctx context.Context) error {
            var err error
            buf, _, err = page.PrintToPDF().
                WithPrintBackground(true).
                WithPaperWidth(210 / 25.4).  // A4
                WithPaperHeight(297 / 25.4).
                WithMarginTop(0.4).
                WithMarginBottom(0.4).
                WithMarginLeft(0.4).
                WithMarginRight(0.4).
                Do(ctx)
            return err
        }),
    })
    return buf, err
}
```

**Resource management:**
- Single `ExecAllocator` context reused across PDF generations
- New `Tab` context per PDF generation (isolated crashes)
- Timeout per PDF: 30 seconds
- For production: semaphore limiting concurrent Chrome tabs to 1-2 per container

### Pattern 5: NLP Service Isolation

**What:** Python FastAPI service on internal network only, called by Go backend.

**When:** A service running heavyweight ML models should not be publicly accessible.

**Service spec:**
```python
# nlp-service/main.py
from fastapi import FastAPI
from pydantic import BaseModel
from transformers import pipeline

app = FastAPI()

# Load model at startup (once, cached)
model = pipeline(
    "text-classification",
    model="pysentimiento/robertuito-emotion-analysis",
    return_all_scores=True,
    top_k=None
)

class AnalyzeRequest(BaseModel):
    texto: str
    idioma: str = "pt"

class Emotion(BaseModel):
    nome: str
    probabilidade: float

class AnalyzeResponse(BaseModel):
    emocoes: list[Emotion]

@app.post("/analyze", response_model=AnalyzeResponse)
async def analyze(req: AnalyzeRequest):
    results = model(req.texto)
    emocoes = [
        Emotion(nome=r["label"], probabilidade=r["score"])
        for r in results[0]
    ]
    return AnalyzeResponse(emocoes=emocoes)
```

**Docker integration:**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
# Pre-download model at build time (save startup time)
RUN python -c "from transformers import pipeline; pipeline('text-classification', model='pysentimiento/robertuito-emotion-analysis')"
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

**Go client:**
```go
// internal/service/nlp.go
type NLPClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

func (c *NLPClient) Analyze(ctx context.Context, texto string, idioma string) (*AnalyzeResponse, error) {
    body := map[string]string{"texto": texto, "idioma": idioma}
    jsonBody, _ := json.Marshal(body)
    
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/analyze", bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    // ... handle response
}
```

---

## CouchDB Document Modeling Strategy

### Database-per-Domain (not per-user)

Kanso uses **shared databases** with `userId` field for document ownership and CouchDB's security API for access control. This avoids database-per-user management overhead while maintaining isolation.

| Database | Purpose | Security |
|----------|---------|----------|
| `registros` | Emotion entries | `members.names: [user1, user2, ...]` |
| `sentimentos` | Custom emotion labels | Same |
| `usuarios` | User preferences | `members.names: [user1, ...]` |
| `relatorios` | Report metadata/jobs | Same |

**validate_doc_update function** (on each database):
```javascript
function(newDoc, oldDoc, userCtx, secObj) {
    // Only server admins bypass
    if (userCtx.roles.indexOf('_admin') !== -1) return;
    
    // Document must have userId matching authenticated user
    if (newDoc.userId !== userCtx.name) {
        throw { forbidden: 'Document userId must match your user' };
    }
    
    // Prevent changing userId on update
    if (oldDoc && newDoc.userId !== oldDoc.userId) {
        throw { forbidden: 'Cannot change userId' };
    }
}
```

### Mango Indexes

```json
// Design document for registros
{
  "_id": "_design/registros_indexes",
  "language": "query",
  "views": {
    "by_user_date": {
      "map": {
        "fields": {
          "userId": "asc",
          "dataHora": "desc"
        },
        "partial_filter_selector": {}
      },
      "reduce": "_count",
      "options": {
        "def": {
          "fields": ["userId", "dataHora"]
        }
      }
    }
  }
}
```

### Conflict Strategy

**For registros:** Append-only with UUID `_id` → conflicts are virtually impossible (no two users update the same doc). Each registry entry is created once and never modified (except by NLP pipeline which Go backend controls).

**For usuarios preferences:** Last-write-wins. Use upsert pattern in Go backend.

**For sentimentos:** Append-only (user creates new, never modifies existing). Same UUID approach.

---

## Traefik Configuration

### Docker Compose Labels Approach

```yaml
# docker-compose.yml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@kanso.app"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "letsencrypt:/letsencrypt"
    labels:
      # Global middleware: CORS
      - "traefik.http.middlewares.cors.headers.accesscontrolallowmethods=GET,POST,PUT,DELETE,OPTIONS"
      - "traefik.http.middlewares.cors.headers.accesscontrolallowheaders=Authorization,Content-Type"
      - "traefik.http.middlewares.cors.headers.accesscontrolalloworiginlist=*"
      - "traefik.http.middlewares.cors.headers.accesscontrolmaxage=600"
      # Rate limiting
      - "traefik.http.middlewares.ratelimit.ratelimit.average=100"
      - "traefik.http.middlewares.ratelimit.ratelimit.burst=50"

  couchdb:
    image: couchdb:3.4
    volumes:
      - "./infra/couchdb/local.ini:/opt/couchdb/etc/local.d/10-kanso.ini"
      - "couchdb-data:/opt/couchdb/data"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.couchdb.rule=Host(`api.kanso.app`) && PathPrefix(`/db/`)"
      - "traefik.http.routers.couchdb.entrypoints=websecure"
      - "traefik.http.routers.couchdb.tls.certresolver=letsencrypt"
      - "traefik.http.middlewares.couch-strip.stripprefix.prefixes=/db"
      - "traefik.http.routers.couchdb.middlewares=couch-strip,cors,ratelimit"
      - "traefik.http.services.couchdb.loadbalancer.server.port=5984"
    networks:
      - internal

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`api.kanso.app`) && PathPrefix(`/api/`)"
      - "traefik.http.routers.backend.entrypoints=websecure"
      - "traefik.http.routers.backend.tls.certresolver=letsencrypt"
      - "traefik.http.routers.backend.middlewares=cors,ratelimit"
      - "traefik.http.services.backend.loadbalancer.server.port=8080"
    environment:
      - COUCHDB_URL=http://couchdb:5984
      - COUCHDB_ADMIN=admin
      - COUCHDB_PASSWORD=${COUCHDB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
      - NLP_SERVICE_URL=http://nlp-service:8000
      - TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID}
      - TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN}
      - FCM_SERVER_KEY=${FCM_SERVER_KEY}
    networks:
      - internal

  nlp-service:
    build:
      context: ./nlp-service
      dockerfile: Dockerfile
    # NO Traefik labels — not publicly accessible
    deploy:
      resources:
        limits:
          memory: 2g    # PyTorch model requires ~1-2GB RAM
          cpus: "2"
    networks:
      - internal

networks:
  internal:
    driver: bridge

volumes:
  letsencrypt:
  couchdb-data:
```

### CouchDB Configuration for JWT Auth

```ini
# infra/couchdb/local.ini
[chttpd]
port = 5984
bind_address = 0.0.0.0
authentication_handlers = {chttpd_auth, jwt_authentication_handler}, {chttpd_auth, default_authentication_handler}

[jwt_auth]
required_claims = exp,iat

[jwt_keys]
; HMAC-SHA256 - the _default key is used when JWT has no "kid" claim
; Value is base64-encoded shared secret
hmac:_default = ${JWT_SECRET_BASE64}

[couchdb]
uuid = kanso-dev
```

---

## Monorepo Structure

```
kanso/
├── frontend/                   # React PWA (Vite + TypeScript + Tailwind)
│   ├── public/
│   │   ├── manifest.json      # PWA manifest
│   │   ├── sw.js              # Service Worker
│   │   └── icons/             # PWA icons
│   ├── src/
│   │   ├── components/        # Reusable UI components
│   │   ├── hooks/             # React hooks (usePouchDB, useAuth, etc.)
│   │   ├── services/
│   │   │   ├── pouchdb.ts     # PouchDB init, sync, CRUD
│   │   │   ├── auth.ts        # Google OAuth + JWT management
│   │   │   └── api.ts         # Go backend HTTP client
│   │   ├── pages/
│   │   │   ├── Register.tsx
│   │   │   ├── History.tsx
│   │   │   └── Profile.tsx
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   └── vite.config.ts
├── backend/                    # Go API (chi)
│   ├── cmd/
│   │   └── kanso-api/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   ├── middleware/
│   │   └── pdf/
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── nlp-service/               # Python FastAPI NLP
│   ├── main.py
│   ├── requirements.txt
│   └── Dockerfile
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   ├── traefik/
│   │   └── traefik.yml
│   └── couchdb/
│       └── local.ini
├── .github/
│   └── workflows/
│       └── deploy.yml
├── .env.example
└── README.md
```

---

## Anti-Patterns to Avoid

### 1. Go Backend as CRUD Proxy for CouchDB

**What:** Creating Go endpoints that mirror CouchDB CRUD operations (create registro, get registro, update registro).

**Why bad:** Doubles the code path, adds latency, and defeats the purpose of PouchDB↔CouchDB direct sync. The offline-first architecture is designed specifically so the browser syncs directly with CouchDB.

**Instead:** PouchDB syncs directly to CouchDB. Go backend only handles operations that need server-side processing (auth, PDF, WhatsApp, NLP). The `repository/couchdb.go` in Go is for **admin operations** (NLP result writing, report data fetching), not for exposing CRUD to the client.

### 2. Database-Per-User in CouchDB

**What:** Creating a separate CouchDB database for each user (`db_user_123`, `db_user_456`, etc.)

**Why bad:** Management overhead (creating/permissioning databases on user registration), CouchDB is optimized for a moderate number of databases, harder to query across users for analytics.

**Instead:** Shared databases with `userId` field + `_security` members list + `validate_doc_update` function. This scales better and maintains cross-user queryability for future admin features.

### 3. Blocking PDF Generation in Request Handler

**What:** Generating the PDF synchronously in the HTTP handler (user waits for chromedp to finish).

**Why bad:** chromedp takes 3-10 seconds per PDF. HTTP request timeout will fire for anything over 30s. User sees no feedback during generation.

**Instead:** Async job pattern: create job (immediate) → poll status → download when ready. Return `202 Accepted` with job ID immediately.

### 4. Exposing NLP Service Publicly

**What:** Creating a Traefik route for `/nlp/*` → NLP service.

**Why bad:** NLP service has no auth middleware (designed for internal use). A public endpoint means anyone can run inference on your model (cost, abuse).

**Instead:** Keep NLP service on internal network only. Go backend calls it via Docker DNS (`http://nlp-service:8000`). No Traefik route.

---

## Scalability Considerations

| Concern | At 1-10 users (MVP) | At 100 users | At 1000+ users |
|---------|---------------------|--------------|----------------|
| CouchDB storage | Single node | Single node | Cluster mode (3+ nodes) |
| Go backend | 1 replica | 2-3 replicas | Horizontal scaling behind Traefik |
| NLP service | 1 replica (CPU) | 1 replica with GPU | Multiple GPU workers |
| chromedp PDF | 1 concurrent (mutex) | 1-2 concurrent | Dedicated PDF worker service |
| CouchDB sync | Direct sync | Direct sync | Connection limits, filter replication per user |
| Sessions | In-memory | In-memory | Redis session store |

**For MVP (1 user):** Everything runs on a single VPS with 2-4GB RAM. The NLP model is the heaviest component (~1-2GB for PyTorch). CouchDB runs in relaxed mode (no clustering needed).

---

## Build Order and Dependencies

### Phase 1: Foundation — CouchDB + Traefik + Go Skeleton
**Dependencies:** None
**Delivers:** Running infrastructure with health-check endpoints
- `docker-compose.yml` with CouchDB, Traefik, backend skeleton
- CouchDB config (JWT auth, per-database security)
- Traefik routing config
- Go `cmd/kanso-api/main.go` with chi router, health-check
- Empty Go packages (handler, service, repository stubs)

### Phase 2: Authentication
**Dependencies:** Phase 1
**Delivers:** Login flow, JWT issuance, PouchDB sync with auth
- Google OAuth integration (Go: `handler/auth.go`, `service/auth.go`)
- JWT signing (Go)
- PouchDB sync setup (PWA: `services/pouchdb.ts`)
- PWA login page
- CouchDB JWT auth verification

### Phase 3: Core CRUD + Sync
**Dependencies:** Phase 2
**Delivers:** Register emotions, History view, offline-first working
- PWA: Register tab (form)
- PWA: History tab (list with PouchDB queries)
- PWA: PouchDB CRUD operations
- CouchDB design docs (Mango indexes)
- Emotion analysis placeholder (untriggered yet)

### Phase 4: NLP Service
**Dependencies:** Phase 3
**Delivers:** Emotion analysis pipeline
- Python FastAPI service with model
- Go NLP client (`service/nlp.go`)
- Trigger analysis on save or via API
- Display analysis results in PWA

### Phase 5: Reports (PDF + WhatsApp)
**Dependencies:** Phase 3
**Delivers:** PDF generation and WhatsApp delivery
- chromedp PDF generation (`internal/pdf/generator.go`)
- Async job system
- Report trigger in PWA (Profile tab)
- Twilio WhatsApp integration
- Therapist contact configuration

### Phase 6: Push Notifications
**Dependencies:** Phase 3
**Delivers:** Scheduled reminders
- FCM integration (Go + PWA Service Worker)
- Notification preferences UI
- Cron/scheduler for sending pushes at user-configured times

### Phase 7: Polish & Production
**Dependencies:** Phase 4, 5, 6
**Delivers:** Production-ready deployment
- CI/CD pipeline
- Monitoring (health checks, metrics)
- Backup strategy (CouchDB replication to backup)
- Error tracking
- Performance optimization

---

## Sources

- [CouchDB Proxy Authentication Documentation](https://docs.couchdb.org/en/stable/api/server/authn.html#proxy-authentication) — HIGH confidence
- [CouchDB JWT Authentication](https://docs.couchdb.org/en/stable/api/server/authn.html#jwt-authentication) — HIGH confidence
- [CouchDB Database Security (_security)](https://docs.couchdb.org/en/stable/api/database/security.html) — HIGH confidence
- [PouchDB Replication Guide](https://pouchdb.com/guides/replication.html) — HIGH confidence
- [PouchDB Conflicts Guide](https://pouchdb.com/guides/conflicts.html) — HIGH confidence
- [Go chi REST Example](https://github.com/go-chi/chi/blob/master/_examples/rest/main.go) — HIGH confidence (official example)
- [chromedp GitHub](https://github.com/chromedp/chromedp) — HIGH confidence (official repository, v0.15.1)
- [chromedp headless-shell Docker image](https://hub.docker.com/r/chromedp/headless-shell/) — HIGH confidence
- [Traefik ForwardAuth Middleware](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/) — MEDIUM confidence (Traefik v3 docs)
- [Traefik Headers Middleware](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/headers/) — MEDIUM confidence
- [CouchDB Security (validate_doc_update)](https://docs.couchdb.org/en/stable/ddocs/ddocs.html#validate-document-update-functions) — HIGH confidence
