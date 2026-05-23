# Phase 07-03: Integração — Research

**Researched:** 2026-05-23
**Domain:** Go backend _changes feed watcher + CouchDB enrichment + frontend emotion display
**Confidence:** HIGH

## Summary

This phase connects the NLP inference service (07-02) with the existing Go backend and React frontend. The core mechanics are well-understood and follow established patterns in the codebase:

1. **Backend watcher** (`service/watcher.go`) — a goroutine that long-polls CouchDB's `_changes` feed, calls the gRPC NLP client for each new `type:registro` doc, and stores results as `analise:{registroId}` docs in the same CouchDB database. Follows the `report.go` pattern (`sync.Mutex` + goroutine + `Start()` from `main.go`).
2. **Frontend enrichment** — PouchDB auto-syncs `analise_nlp` docs to the frontend. The `getRegistros()` service does an in-memory merge of registro + analysis data. `RegistroCard` renders emotion chips when analysis is available.
3. **PDF reports** — template updated to show emotion summary section + per-registro emotions.

**Primary recommendation:** Implement the _changes watcher as a new `service/watcher.go` following the `report.go` pattern (struct with mutex, `Start()` goroutine from `main.go`). Add `_changes` feed parsing and checkpoint CRUD to `repository/couchdb.go`. Frontend merge in `registros.ts`, display in `RegistroCard.tsx`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| NLP-01 | Backend analyzes registration text (sensações + contexto + pensamentos) for emotion patterns | Watcher goroutine calls `nlp.Client.Analyze()` for each new registro via _changes feed |
| NLP-02 | Detected emotions are stored alongside the registration | Stored as `analise:{registroId}` doc in same `registros` DB (D-44); frontend merges via PouchDB |
| NLP-03 | Analysis runs asynchronously — does not block registration | _changes feed is async; goroutine pattern (D-36); no user-facing errors (D-54); exponential backoff retry (D-53) |
</phase_requirements>

## User Constraints (from CONTEXT.md)

### Locked Decisions (D-36 through D-55)
- **D-36:** Goroutine inside `kanso-api` — not a separate microservice. Follows `report.go` pattern.
- **D-37:** New `service/watcher.go` file in `backend/internal/service/`. Started from `main.go` via constructor + `Start()`.
- **D-38:** Long-poll mode (`GET /registros/_changes?since={seq}&timeout=25000&feed=longpoll`).
- **D-39:** Filter by `type: "registro"` on the client side — skip other doc types.
- **D-40:** Checkpoint persistence in CouchDB doc (`checkpoint:nlp_watcher` in `registros` DB, type `checkpoint`).
- **D-41:** `since=0` on first run — single code path for backfill and live.
- **D-42:** Rate limit ~50ms between gRPC calls via `time.Sleep` or ticker.
- **D-43:** Same `registros` CouchDB database for analise docs.
- **D-44:** Doc schema: `_id: "analise:{registroId}"`, `type: "analise_nlp"`, fields from gRPC response.
- **D-45:** PouchDB local merge — no new API endpoint.
- **D-46:** No new Go API endpoint needed for enriched registros.
- **D-47:** Colored emotion chips in RegistroCard, below sentimentoNome, always visible.
- **D-48:** No chips when no analysis exists (graceful degradation).
- **D-49:** Chip styling: small rounded pills with emotion-appropriate colors.
- **D-50:** Summary section at top of PDF report with aggregate emotion frequency.
- **D-51:** Per-registro emotions in PDF report.
- **D-52:** Report service fetches analysis docs from CouchDB, passes to template.
- **D-53:** Exponential backoff retry (1s, 4s, 16s, 3 attempts) for failed NLP gRPC calls.
- **D-54:** No user-facing error for NLP failures. Checkpoint advances regardless of failures.
- **D-55:** Resume from saved `last_seq` checkpoint on restart.

### the agent's Discretion
- Exact rate limit value (50ms or adjust based on testing)
- Checkpoint doc schema and exact field names
- Emotion chip colors (palette mapping per emotion)
- Report summary aggregation logic (top N emotions, minimum threshold)
- Test file organization and test patterns
- `RegistroDoc` type update in frontend types (add optional `analise` field)

### Deferred Ideas (OUT OF SCOPE)
None.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Detect new registrations | Backend (Go) | — | CouchDB _changes feed, long-poll from kanso-api |
| Call NLP gRPC service | Backend (Go) | — | gRPC client in `nlp/` package, watcher goroutine |
| Store analysis results | Backend (Go) → CouchDB | — | PUT `analise:{registroId}` doc via couchRepo |
| Persist checkpoint | Backend (Go) → CouchDB | — | `checkpoint:nlp_watcher` doc in registros DB |
| Display emotion chips | Browser (React) | — | RegistroCard renders chips from merged analise data |
| Merge analysis with registro | Browser (PouchDB) | — | `getRegistros()` fetches both types, in-memory merge |
| Include emotions in PDF | Backend (Go) → chromedp | — | Report service fetches analise docs, passes to template |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| google.golang.org/grpc | v1.80.0 | gRPC client for NLP calls | Already in go.mod; `nlp.Client` uses it |
| Go stdlib `net/http` | Go 1.26 | CouchDB _changes feed HTTP long-poll | Already used across `couchdb.go`; no new deps needed |
| PouchDB-Browser | ^9.0.0 | Client-side merge of analise_nlp docs | Already in frontend project |
| Vitest | ^4.1.6 | Frontend testing | Already configured; used in all frontend tests |
| Go `testing` + `httptest` | stdlib | Backend testing | Pattern used in push_test.go, report_test.go |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Tailwind CSS | ^4.3.0 | Emotion chip styling | RegistroCard emotion chips (small rounded pills) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| _changes long-poll | WebSocket / continuous feed | Long-poll is simpler, no persistent connection; ~1-2s latency vs instant |
| Same DB for analise docs | Separate CouchDB DB | Same DB = single _changes feed, simpler; separate DB = cleaner isolation but 2 feeds |
| PouchDB merge | New Go API endpoint `/api/registros/enriched` | PouchDB merge works offline (D-45); API endpoint would break offline mode |

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CouchDB (registros DB)                       │
│                                                                     │
│  ┌──────────────┐     _changes feed     ┌──────────────────────┐   │
│  │  type:registro│ ◄─── long-poll ──────│  service/watcher.go  │   │
│  │  docs         │                       │  (goroutine)        │   │
│  └──────────────┘                       │                      │   │
│         ▲                               │  1. Read _changes    │   │
│         │ PouchDB sync                  │  2. Filter type=="   │   │
│         │                               │     registro"        │   │
│  ┌──────────────┐                       │  3. nlp.Client       │   │
│  │  type:analise│  ◄─── PUT result ─────│     .Analyze()       │   │
│  │ _nlp docs    │                       │  4. Save analise doc │   │
│  └──────────────┘                       │  5. Save checkpoint  │   │
│         ▲                               └──────┬───────────────┘   │
│         │                                      │                   │
└─────────┼──────────────────────────────────────┼───────────────────┘
          │                                      │ gRPC (port 50051)
          │ PouchDB sync                          ▼
          │                              ┌─────────────────┐
          │                              │ nlp-service      │
          │                              │ (Python/FastAPI) │
          │                              │  classifier.py   │
          │                              └─────────────────┘
          │
    ┌─────┴─────────────────────────────────────────────┐
    │              Frontend (PouchDB)                    │
    │                                                    │
    │  registros.ts: getRegistros()                      │
    │    1. allDocs where type="registro"                │
    │    2. allDocs where type="analise_nlp"             │
    │    3. merge in-memory by analise:{id} === _id      │
    │                                                    │
    │  History.tsx ──→ RegistroCard.tsx                  │
    │                    ├─ sentimentoNome (existing)    │
    │                    └─ emotion chips (NEW)          │
    │                                                    │
    │  ReportSection ──→ service/report.go ──→ report.html│
    │                    (fetches analise docs via CouchDB)│
    └────────────────────────────────────────────────────┘
```

### Recommended Project Structure

Changes/additions:

```
backend/
├── internal/
│   ├── repository/
│   │   └── couchdb.go          [ADD: GetChanges, SaveCheckpoint, GetCheckpoint,
│   │                             CreateAnalise methods]
│   ├── service/
│   │   └── watcher.go          [NEW: NlpWatcher struct + Start() goroutine]
│   ├── nlp/
│   │   └── client.go           [EXISTING: ready to use]
│   └── templates/
│       └── report.html         [MODIFY: add emotion summary + per-registro emotions]
└── cmd/
    └── kanso-api/
        └── main.go             [MODIFY: wire NlpWatcher, call Start()]

frontend/
├── src/
│   ├── types/
│   │   └── index.ts            [ADD: AnaliseNlpDoc, RegistroWithAnalise types]
│   ├── services/
│   │   └── registros.ts        [MODIFY: getRegistros() merge analise docs]
│   └── components/
│       └── RegistroCard.tsx     [MODIFY: render emotion chips with colors]
```

### Pattern 1: Async Goroutine Service (report.go pattern)

**What:** Service with `sync.Mutex`, constructor, and optional `Start()` goroutine for background processing. Mutex ensures only one batch of work runs at a time (or protects shared state).

**When to use:** Background tasks that share CouchDB connection and gRPC client with the main API.

**Example (report.go pattern):**
```go
type WatcherService struct {
    mu        sync.Mutex
    couchRepo *repository.CouchDB
    nlpClient *nlp.Client
    cfg       *config.Config
    stopChan  chan struct{}
}

func NewWatcherService(couchRepo *repository.CouchDB, nlpClient *nlp.Client, cfg *config.Config) *WatcherService {
    return &WatcherService{
        couchRepo: couchRepo,
        nlpClient: nlpClient,
        cfg:       cfg,
        stopChan:  make(chan struct{}),
    }
}

func (s *WatcherService) Start() {
    s.mu.Lock()
    go func() {
        defer s.mu.Unlock()
        s.run()
    }()
}
```

### Pattern 2: PouchDB In-Memory Merge

**What:** Fetch `analise_nlp` docs alongside `registro` docs from local PouchDB, merge by matching `analise:{registroId}` to `registro._id`.

**When to use:** Any offline-first scenario where related data lives in the same CouchDB database.

**Code sketch:**
```typescript
export async function getRegistros(): Promise<RegistroWithAnalise[]> {
  const allDocs = await registrosDB.allDocs<RegistroDoc>({ include_docs: true })
  const registros = allDocs.rows
    .map(r => r.doc!)
    .filter(d => d.type === 'registro')
    .sort((a, b) => new Date(b.dataHora).getTime() - new Date(a.dataHora).getTime())

  // Fetch analise_nlp docs
  const analiseResult = await registrosDB.allDocs<AnaliseNlpDoc>({ include_docs: true })
  const analiseMap = new Map<string, AnaliseNlpDoc>()
  for (const row of analiseResult.rows) {
    const doc = row.doc!
    if (doc.type === 'analise_nlp') {
      // doc._id is "analise:{registroId}" — extract registroId
      const registroId = doc._id.replace('analise:', '')
      analiseMap.set(registroId, doc)
    }
  }

  return registros.map(r => ({
    ...r,
    analise: analiseMap.get(r._id),
  }))
}
```

### Anti-Patterns to Avoid

- **Wrapping analise docs in the original registro.** The `_changes` response with `include_docs=true` returns the full doc. The watcher must NOT modify the original registro — it creates a separate `analise:` doc.
- **Blocking the main goroutine.** The watcher's event loop is in a goroutine; startup/teardown should not block `main()`.
- **Infinite analysis loop.** Without D-39 (client-side filter by `type:"registro"`), the watcher would see its own `analise_nlp` docs and analyze them, creating cascade.
- **Lost checkpoint on crash.** Save checkpoint *after* each individual analise doc is written, but before the next iteration. This way, on restart, the watcher resumes from the last fully-processed position. (At-most-once semantics per D-55.)

## CouchDB _changes Feed Mechanics

**Endpoint:** `GET {baseURL}/registros/_changes?since={seq}&timeout=25000&feed=longpoll`

**Response format:**
```json
{
  "results": [
    {
      "seq": "3-abc123",
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "changes": [{"rev": "3-def456"}]
    }
  ],
  "last_seq": "3-abc123"
}
```

With `include_docs=true`, each result also has a `"doc": { ... }` field with the full document body.

**Key behaviors:**
1. When `since` is behind the latest sequence, the response returns immediately with all pending changes (up to server batch limit).
2. When caught up, the connection stays open for up to `timeout` ms waiting for new changes.
3. If no changes arrive within the timeout, CouchDB returns `{"results": [], "last_seq": "{same_seq}"}`.
4. Long-poll mode is a regular HTTP request — use Go's standard `net/http` client.
5. The `since` value is an opaque string — store and send it back as-is.

**CouchDB doc-level access pattern:**
The `_changes` feed returns document IDs but not document content unless `include_docs=true`. Options:

| Approach | Pros | Cons |
|----------|------|------|
| `include_docs=true` | Single HTTP call per batch; no extra GETs | Larger response payload; CouchDB may skip docs on compaction |
| `include_docs=false` + per-doc GET | Reliable content even during compaction | N extra HTTP GETs per batch |

**Recommendation:** Use `include_docs=true`. The `doc` field is the document at the time the change was triggered. For the backfill scenario (which reads current docs, not history), this is reliable. The watcher filters by `doc.type === "registro"` so it only processes registration docs.

## Checkpoint Schema

The checkpoint doc in CouchDB:

```json
{
  "_id": "checkpoint:nlp_watcher",
  "type": "checkpoint",
  "watcher": "nlp",
  "last_seq": "3-abc123",
  "updatedAt": "2026-05-23T18:00:00Z"
}
```

Go struct:
```go
type CheckpointDoc struct {
    ID        string `json:"_id,omitempty"`
    Rev       string `json:"_rev,omitempty"`
    Type      string `json:"type"`
    Watcher   string `json:"watcher"`
    LastSeq   string `json:"last_seq"`
    UpdatedAt string `json:"updatedAt,omitempty"`
}
```

## Analise Doc Schema

```json
{
  "_id": "analise:550e8400-e29b-41d4-a716-446655440000",
  "type": "analise_nlp",
  "registroId": "550e8400-e29b-41d4-a716-446655440000",
  "emotionPrincipal": "ansiedade",
  "emotions": [
    {"emotion": "ansiedade", "score": 0.85},
    {"emotion": "medo", "score": 0.42}
  ],
  "scores": {
    "alegria": 0.02,
    "tristeza": 0.05,
    "raiva": 0.01,
    "medo": 0.42,
    "nojo": 0.0,
    "surpresa": 0.03,
    "ansiedade": 0.85,
    "vergonha": 0.01,
    "culpa": 0.0,
    "saudade": 0.0,
    "amor": 0.0,
    "gratidão": 0.0,
    "neutro": 0.08
  },
  "intensidade": 0.85,
  "modeloVersao": "v1.0",
  "analisadoEm": "2026-05-23T18:00:00Z"
}
```

Go struct:
```go
type EmotionScore struct {
    Emotion string  `json:"emotion"`
    Score   float32 `json:"score"`
}

type AnaliseDoc struct {
    ID               string         `json:"_id,omitempty"`
    Rev              string         `json:"_rev,omitempty"`
    Type             string         `json:"type"`
    RegistroID       string         `json:"registroId"`
    EmotionPrincipal string         `json:"emotionPrincipal"`
    Emotions         []EmotionScore `json:"emotions"`
    Scores           map[string]float32 `json:"scores"`
    Intensidade      float32        `json:"intensidade"`
    ModeloVersao     string         `json:"modeloVersao"`
    AnalisadoEm      string         `json:"analisadoEm,omitempty"`
}
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CouchDB _changes feed parsing | Manual JSON parser | Go `encoding/json` with struct | Already used across codebase; standard library |
| gRPC client | Custom TCP/protobuf client | `nlp.Client` (already built) | `nlp/client.go` is ready — `NewClient(addr)` + `Analyze()` |
| gRPC connection management | Connection pooling | `grpc.ClientConn` built into `nlp.Client` | `grpc.DialContext` handles pooling |
| Exponential backoff timer | Manual sleep math | `time` + simple loop | Pattern is simple: 1s, 4s, 16s — 3 hardcoded delays |
| PouchDB fetching | Custom IndexedDB wrapper | Existing `registrosDB.allDocs()` | Already built and tested in `registros.ts` |

**Key insight:** The watcher does not need any new infrastructure. It reuses the existing `nlp.Client`, `repository.CouchDB`, and PouchDB sync. The only new component is the watcher loop logic itself, which is ~100 lines of Go.

## Watcher Lifecycle Details

### Startup Sequence

1. `main.go` constructs `nlp.NewClient(cfg.NLPGrpAddr)` — returns `*nlp.Client`
2. `main.go` constructs `service.NewWatcherService(couchRepo, nlpClient, cfg)` — returns `*WatcherService`
3. `main.go` calls `watcherSvc.Start()` — launches `run()` goroutine
4. `run()` reads checkpoint from CouchDB (`GET registros/checkpoint:nlp_watcher`)
   - If checkpoint exists: use `checkpoint.last_seq` as `since`
   - If no checkpoint: use `since=0` (triggers backfill)
5. Enter loop: `GET _changes?since={last_seq}&timeout=25000&feed=longpoll&include_docs=true`

### Event Loop Pseudocode

```
for {
    resp, err := doChangesRequest(since)
    if err != nil {
        log error, sleep 5s, continue  // network error — retry
    }
    
    for _, result := range resp.Results {
        doc := result.Doc
        if doc.Type != "registro" {
            continue  // D-39: skip non-registro docs
        }
        
        // Rate limit: 50ms between calls (D-42)
        throttle()
        
        // Call NLP gRPC with retry
        var analysis *nlp.AnalyzeResponse
        for attempt := 0; attempt < 3; attempt++ {
            analysis, err = nlpClient.Analyze(ctx, buildRequest(doc))
            if err == nil {
                break
            }
            sleep(backoff[attempt])  // 1s, 4s, 16s
        }
        
        if err != nil {
            log.Printf("NLP analysis failed for %s after 3 retries: %v", doc.ID, err)
            continue  // D-54: skip silently, advance checkpoint anyway
        }
        
        // Save analysis doc
        analiseDoc := buildAnaliseDoc(doc.ID, analysis)
        couchRepo.SaveAnalise(analiseDoc)
    }
    
    // Save checkpoint (last_seq from response)
    if resp.LastSeq != "" {
        couchRepo.SaveCheckpoint(resp.LastSeq)
        since = resp.LastSeq
    }
}
```

### Shutdown

The watcher needs a clean shutdown mechanism. Add a `Stop()` method that closes the stop channel:

```go
func (s *WatcherService) Stop() {
    close(s.stopChan)
}
```

In the event loop, use a select on `stopChan` during the long-poll (wrapped in a cancellable context):

```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    <-s.stopChan
    cancel()
}()
// Use ctx for the HTTP request
```

## Emotion Chip Color Palette

From `model_config.py`: 13 emotion labels. Suggested palette (fine-tune during implementation):

| Emotion | Tailwind Class | Hex |
|---------|---------------|-----|
| alegria | `bg-emerald-100 text-emerald-700` | #d1fae5 / #047857 |
| tristeza | `bg-blue-100 text-blue-700` | #dbeafe / #1d4ed8 |
| raiva | `bg-red-100 text-red-700` | #fee2e2 / #b91c1c |
| medo | `bg-purple-100 text-purple-700` | #f3e8ff / #7e22ce |
| nojo | `bg-amber-100 text-amber-700` | #fef3c7 / #b45309 |
| surpresa | `bg-orange-100 text-orange-700` | #ffedd5 / #c2410c |
| ansiedade | `bg-yellow-100 text-yellow-700` | #fef9c3 / #a16207 |
| vergonha | `bg-pink-100 text-pink-700` | #fce7f3 / #be185d |
| culpa | `bg-rose-100 text-rose-700` | #ffe4e6 / #be123c |
| saudade | `bg-violet-100 text-violet-700` | #ede9fe / #6d28d9 |
| amor | `bg-pink-200 text-pink-800` | #fbcfe8 / #9d174d |
| gratidão | `bg-teal-100 text-teal-700` | #ccfbf1 / #0f766e |
| neutro | `bg-gray-100 text-gray-600` | #f3f4f6 / #4b5563 |

All chips use `text-xs font-medium px-2 py-0.5 rounded-full` for consistent sizing.

## Report Template Changes

### Summary Section (top of report, D-50)

Insert after the `.periodo` div:

```html
{{if .EmotionSummary}}
<div class="summary">
  <h2>Resumo das Emoções</h2>
  {{range .EmotionSummary}}
  <div class="emotion-bar">
    <span class="emotion-label">{{.Emotion}}</span>
    <span class="emotion-count">{{.Count}}x</span>
  </div>
  {{end}}
</div>
{{end}}
```

### Per-Registro Emotions (D-51)

Insert after `.sentimento` line in the registro block:

```html
<div class="sentimento">{{.Sentimento}}</div>
{{if .Emocoes}}
<div class="emocoes">
  {{range .Emocoes}}
  <span class="emocao-chip">{{.Emotion}}</span>
  {{end}}
</div>
{{end}}
```

### Go Template Data Changes

The report service needs to:
1. Query `registros` DB for `type:analise_nlp` docs whose `registroId` matches any registro in the period
2. Pass `EmotionSummary` (aggregated count per emotion) as top-level template data
3. Pass `Emocoes` (emotion list) per registro

```go
type EmotionSummaryItem struct {
    Emotion string
    Count   int
}

type RegistroReportItem struct {
    Data        string
    Sentimento  string
    Sensacoes   string
    Contexto    string
    Pensamentos string
    Emocoes     []nlp.EmotionScore  // NEW
}

type ReportData struct {
    GeneratedAt    string
    PeriodStart    string
    PeriodEnd      string
    Registros      []RegistroReportItem
    EmotionSummary []EmotionSummaryItem  // NEW
}
```

## Existing Code Integration Points

### `main.go` — Wire Watcher

After existing service initialization (around line 51):
```go
nlpClient, err := nlp.NewClient(cfg.NLPGrpAddr)
if err != nil {
    log.Printf("warning: nlp client not available: %v", err)
}
watcherSvc := service.NewWatcherService(couchRepo, nlpClient, cfg)
watcherSvc.Start()
```

If `nlpClient` is nil (NLP not available), the watcher should log a warning and become a no-op until the next restart.

### `couchdb.go` — New Repository Methods

Need to add:
```go
// GetChanges fetchs _changes feed
func (c *CouchDB) GetChanges(db, since string) (*ChangesResponse, error)

// SaveCheckpoint persists the last_seq
func (c *CouchDB) SaveCheckpoint(seq string) error

// GetCheckpoint reads saved last_seq; returns "", nil if none
func (c *CouchDB) GetCheckpoint() (string, error)

// SaveAnalise writes the analysis result doc
func (c *CouchDB) SaveAnalise(doc *AnaliseDoc) error
```

### `RegistroCard.tsx` — Emotion Chips

The card receives `registro: RegistroWithAnalise`. In the header, below `sentimentoNome`, add:

```tsx
{registro.analise && (
  <div className="flex flex-wrap gap-1.5 mt-1.5">
    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700">
      {registro.analise.emotionPrincipal}
    </span>
    {registro.analise.emotions
      .filter(e => e.emotion !== registro.analise!.emotionPrincipal)
      .slice(0, 3)
      .map(e => (
        <span key={e.emotion} className="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">
          {e.emotion}
        </span>
      ))
    }
  </div>
)}
```

The actual chip color is computed from a mapping function (emotion → Tailwind classes) at the agent's discretion.

## Common Pitfalls

### Pitfall 1: _changes long-poll timeout handling
**What goes wrong:** HTTP request hangs for 25s on each poll, blocking goroutine.
**Why it happens:** CouchDB long-poll holds the connection open for `timeout` ms. If no changes arrive, it returns an empty response after 25s.
**How to avoid:** The timeout is *desired behavior* — it means no new docs. After receiving the empty response, immediately loop back with the same `since`. Add a per-request timeout of 30s (slightly above the 25s poll timeout) to prevent connection leaks.
**Warning signs:** Watcher goroutine not advancing; checkpoint not updating.

### Pitfall 2: Checkpoint saved before analise docs committed
**What goes wrong:** Crash after saving checkpoint but before analise doc reaches CouchDB. Result: analise doc is lost forever (D-55 semantics).
**Why it happens:** Race between checkpoint write and analise doc write on crash.
**How to avoid:** Save checkpoint AFTER all analise docs in the batch are committed. Accept at-most-once semantics (D-55 explicitly says lost docs are acceptable for MVP). For stronger guarantees, persist each individual analise doc+checkpoint atomically or use a batch approach where checkpoint only advances after batch confirmation.

### Pitfall 3: include_docs=true + document compaction
**What goes wrong:** CouchDB compaction removes old document bodies from `_changes` history. A watcher starting from an old `since` may get results without `doc` fields.
**Why it happens:** When `since` points to a sequence before the compaction point, CouchDB returns `{"results": []}` for those entries — the docs are lost.
**How to avoid:** Not a problem here — the watcher runs continuously and saves checkpoints. On restart (D-55), it resumes from saved `last_seq`, which is always current. Only relevant if `since` is days/weeks old, which won't happen with continuous operation.
**Mitigation:** If `include_docs=true` result has no `doc`, fall back to a direct GET `/{db}/{id}`.

### Pitfall 4: Goroutine never stops gracefully
**What goes wrong:** On API server shutdown (`SIGINT`/`SIGTERM`), the watcher goroutine continues running or is killed mid-checkpoint-write.
**Why it happens:** Go doesn't kill goroutines on process exit — the process exits when `main()` returns. But if `main()` doesn't wait for goroutines, in-flight HTTP requests may be cut.
**How to avoid:** Use a `stopChan` pattern with `http.Request` cancellation. On shutdown, close the channel to cancel in-flight requests, then the goroutine exits cleanly.

## Testing Strategy

### Backend Tests (Go)

**Pattern to follow:** `push_test.go` and `report_test.go` — use `httptest.NewServer` to mock CouchDB HTTP responses.

| Test | Approach | File |
|------|----------|------|
| Watcher processes changes | Mock CouchDB `_changes` endpoint to return 1 registro; mock gRPC (or skip NLP call); assert analise doc PUT | `service/watcher_test.go` |
| Watcher skips non-registro | Mock `_changes` returns analise_nlp + checkpoint docs; assert no NLP call made | `service/watcher_test.go` |
| Checkpoint saved after batch | Mock verifies PUT to `checkpoint:nlp_watcher` | `service/watcher_test.go` |
| Since=0 on first run | Mock shows no checkpoint doc; verifies `since=0` in URL | `service/watcher_test.go` |
| Resume from checkpoint | Mock returns checkpoint doc; verifies `since={last_seq}` | `service/watcher_test.go` |
| Exponential backoff retry | Mock gRPC to fail 2 times, succeed on 3rd; verify 3 calls made with delays | `service/watcher_test.go` |
| NLP failures skip silently | Mock gRPC to always fail; assert analise doc NOT written, checkpoint advances | `service/watcher_test.go` |
| Report template with emotions | Test data includes analise docs; verify template renders emotion chips | `service/report_test.go` |

**gRPC client mocking:** The watcher's `nlpClient` field is an interface or struct. For testing, extract an interface or use a test double:

```go
type Analyzer interface {
    Analyze(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error)
}
```

The `nlp.Client` satisfies this interface. Tests provide a mock implementation.

### Frontend Tests (Vitest)

**Pattern to follow:** `registros.test.ts` and `RegistroCard.test.tsx`.

| Test | Approach | File |
|------|----------|------|
| getRegistros merges analise | Mock `allDocs` to return registros + analise_nlp docs; assert merged output | `services/registros.test.ts` |
| RegistroCard shows emotion chips | Pass registro with analise data; assert chips rendered | `components/RegistroCard.test.tsx` |
| RegistroCard no chips when no analise | Pass registro without analise; assert no chips | `components/RegistroCard.test.tsx` |
| RegistroCard principal chip style | Assert principal chip has correct color class | `components/RegistroCard.test.tsx` |
| History page renders emotion chips | Mock merge to return enriched registros; assert chips visible | `pages/History.test.tsx` |

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| NLP gRPC service slow on first request (model loading) | High (lazy init in server.py) | First analise doc delayed 5-10s | First call timeout is 10s; after that model is cached. Acceptable for MVP. |
| _changes long-poll connection timeout | Low (25s poll + 30s request timeout) | Watcher reconnects immediately | Request timeout 30s > poll timeout 25s; next iteration restarts immediately. |
| Checkpoint lost on crash between docs | Medium (unexpected restart) | Some analise docs never written | D-55 explicitly accepts this. MVP semantics: at-most-once. No user-facing data loss (registros intact, just missing analysis). |
| Large backfill (since=0) processes too fast | Low | NLP service overwhelmed | Rate limit 50ms between calls (D-42). ~20 registros/second max. |
| PouchDB merge with many analise docs | Low | UI pause during merge | `allDocs` returns all docs; filter by type O(n). Acceptable for user-scale data (<1000 docs). |
| _changes doc without `_deleted` filter | Low | Watcher tries to analyze deleted docs | Check `doc._deleted === true` and skip; check `doc.type` before NLP call. |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | CouchDB `_changes?include_docs=true` returns current document content for non-compacted sequences | CouchDB patterns | Minor — fallback to per-doc GET |
| A2 | The `nlp.Client` connection is long-lived and can be reused across many `Analyze()` calls | gRPC client | High — if connection breaks, need reconnect logic |
| A3 | `analise_nlp` docs auto-sync to the frontend via existing PouchDB live sync | PouchDB merge | Medium — must confirm `createSyncedDB('registros')` includes doc type filtering or not |
| A4 | The existing PouchDB `allDocs` query returns all doc types including `analise_nlp` | Frontend | Medium — if PouchDB filters by type, would need separate query |

## Open Questions

1. **gRPC connection resilience**
   - What we know: `nlp.NewClient` dials once with `grpc.WithBlock()` and 5s timeout. If NLP service is down at startup, `NewClient` returns error.
   - What's unclear: Should watcher handle gRPC reconnect if connection drops mid-operation? `grpc.ClientConn` has built-in reconnection by default.
   - Recommendation: Accept built-in gRPC reconnection. If `NewClient` fails, log warning and the watcher becomes no-op until next restart.

2. **include_docs behavior for deleted docs**
   - What we know: `_changes` includes `_deleted: true` docs. The `doc` field may be minimal.
   - What's unclear: Whether we need explicit `_deleted` check or `doc.type` check is sufficient.
   - Recommendation: Check `doc._deleted` explicitly AND `doc.type === "registro"` before processing. Deleted docs won't have `type` field, but explicit check is safer.

3. **PouchDB merge performance with allDocs**
   - What we know: `allDocs({ include_docs: true })` returns all docs in the database.
   - What's unclear: Whether fetching both types in one call vs. separate calls is better.
   - Recommendation: Single `allDocs()` call with in-memory filtering. At user-scale (<1000 total docs), performance difference is negligible.

## Validation Architecture

> nyquist_validation: true (from .planning/config.json)

### Test Framework

| Property | Value |
|----------|-------|
| Go Framework | Go stdlib `testing` + `httptest` |
| Go Config file | none — standard Go test convention |
| Go quick run | `go test ./internal/service/ -run TestWatcher -v -count=1` |
| Go full suite | `go test ./... -v -count=1` |
| Frontend Framework | Vitest v4 with jsdom |
| Frontend Config | vite.config.ts (test section) |
| Frontend quick run | `npx vitest run --reporter=verbose src/services/registros.test.ts src/components/RegistroCard.test.tsx` |
| Frontend full suite | `npm test` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| NLP-01 | Watcher detects new registros via _changes, calls NLP gRPC | Unit (Go) | `go test ./internal/service/ -run TestWatcher_ProcessesRegistro -v` |
| NLP-01 | Watcher skips non-registro doc types | Unit (Go) | `go test ./internal/service/ -run TestWatcher_SkipsNonRegistro -v` |
| NLP-02 | Analysis result stored as `analise:{registroId}` doc | Unit (Go) | `go test ./internal/service/ -run TestWatcher_SavesAnaliseDoc -v` |
| NLP-02 | Frontend merges analise_nlp with registro data | Unit (Vitest) | `npx vitest run src/services/registros.test.ts -t "merges analise" --reporter=verbose` |
| NLP-02 | RegistroCard shows emotion chips when analise exists | Unit (Vitest) | `npx vitest run src/components/RegistroCard.test.tsx -t "emotion chips" --reporter=verbose` |
| NLP-02 | PDF report includes emotion summary + per-registro emotions | Unit (Go) | `go test ./internal/service/ -run TestReportService_WithAnalise -v` |
| NLP-03 | Watcher retries failed NLP calls with backoff | Unit (Go) | `go test ./internal/service/ -run TestWatcher_RetriesOnError -v` |
| NLP-03 | Failed NLP skips silently, checkpoint advances | Unit (Go) | `go test ./internal/service/ -run TestWatcher_SkipsAfterRetries -v` |
| NLP-03 | Checkpoint persists and resumes on restart | Unit (Go) | `go test ./internal/service/ -run TestWatcher_ResumesFromCheckpoint -v` |

### Sampling Rate

- **Per task commit:** `go test ./internal/service/ -run "TestWatcher|TestReport" -v -count=1` (Go) + `npx vitest run src/services/registros.test.ts src/components/RegistroCard.test.tsx --reporter=verbose` (Frontend)
- **Per wave merge:** `go test ./... -v -count=1` + `npm test`

### Wave 0 Gaps

- [ ] `backend/internal/service/watcher_test.go` — new file, all watcher tests
- [ ] `frontend/src/services/registros.test.ts` — add merge test case
- [ ] `frontend/src/components/RegistroCard.test.tsx` — add emotion chip test cases
- [ ] `frontend/src/services/registros.test.ts` — needs mock setup for analise_nlp allDocs response

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `backend/internal/service/report.go`, `push.go` — goroutine pattern
- Codebase analysis: `backend/internal/nlp/client.go` — gRPC client interface
- Codebase analysis: `backend/internal/repository/couchdb.go` — HTTP client pattern
- Codebase analysis: `nlp-service/proto/analysis.proto` — gRPC contract
- Codebase analysis: `nlp-service/src/classifier.py`, `model_config.py` — 13 emotion labels
- Codebase analysis: `frontend/src/services/registros.ts` — PouchDB pattern
- Codebase analysis: `frontend/src/components/RegistroCard.tsx` — existing card component
- Codebase analysis: `backend/internal/templates/report.html` — PDF template
- Codebase analysis: `backend/internal/service/push_test.go` — httptest mock pattern
- Codebase analysis: `frontend/src/components/RegistroCard.test.tsx` — frontend test pattern

### Secondary (MEDIUM confidence)
- CouchDB `_changes` API: long-poll mode — verified against CouchDB 3.x documentation [CITED: docs.couchdb.org/en/stable/api/database/changes.html]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in go.mod/package.json, no new deps
- Architecture: HIGH — patterns well-established in codebase (report.go, push.go, couchdb.go)
- Pitfalls: HIGH — based on codebase patterns and common CouchDB _changes issues
- Frontend: HIGH — existing test patterns (vitest, jsdom, vi.mock) well-documented

**Research date:** 2026-05-23
**Valid until:** 2026-06-23 (stable; CouchDB _changes API hasn't changed in years)
