# Phase 07-03: Integração — Context

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Go backend _changes feed watcher that detects new registrations in CouchDB, sends text to the NLP Python gRPC service for emotion analysis, and stores results as `analise:{registroId}` documents. Frontend displays detected emotions as chips in RegistroCard and includes emotion data in PDF reports. Covers the integration layer that connects the NLP model (07-02) with the existing Go backend and frontend.

**Requirements addressed:** NLP-01 (async analysis), NLP-02 (emotion enrichment), NLP-03 (non-blocking)

</domain>

<decisions>
## Implementation Decisions

### Service Model
- **D-36:** **Goroutine inside `kanso-api`** — not a separate microservice. Follows the `report.go` pattern (`sync.Mutex` + goroutine for async processing). Watcher shares CouchDB connection and gRPC client with the main API. Simpler deploy, less surface area.
- **D-37:** New `service/watcher.go` file in `backend/internal/service/`. Started from `main.go` via a constructor + `Start()` goroutine.

### Changes Feed
- **D-38:** **Long-poll** mode (`GET /registros/_changes?since={seq}&timeout=25000&feed=longpoll`). Balances low latency (~1-2s) with simple connection management. No persistent connection needed.
- **D-39:** Filter by `type: "registro"` on the client side — skip `analise_nlp` and other doc types to avoid infinite analysis loops (analyzing an analysis doc).
- **D-40:** **Checkpoint persistence** — save `last_seq` from `_changes` response in a dedicated CouchDB doc (`checkpoint:nlp_watcher` in `registros` DB, type `checkpoint`). On restart, resume from saved sequence.

### Backfill Strategy
- **D-41:** **`since=0` on first run** — single code path for both backfill and live processing. `_changes?since=0` returns all existing docs sequentially. After backfill completes, long-poll takes over for new registrations.
- **D-42:** **Rate limit**: ~50ms minimum gap between NLP gRPC calls to avoid overwhelming the Python service. Implemented via `time.Sleep` or ticker between iterations.

### Analysis Storage
- **D-43:** **Same `registros` CouchDB database** — `analise:{registroId}` stored alongside registro docs, with `type: "analise_nlp"`. No new CouchDB database needed. Single `_changes` feed to watch.
- **D-44:** Doc schema: `_id: "analise:{registroId}"`, `type: "analise_nlp"`, fields from gRPC response (`emotionPrincipal`, `emotions[]`, `scores`, `intensidade`, `modeloVersao`) + `registroId`, `analisadoEm`.

### Frontend: Data Fetching
- **D-45:** **PouchDB local merge** — analysis docs auto-sync to the frontend via PouchDB (same `registros` DB). Frontend queries both `type: "registro"` and `type: "analise_nlp"` docs, merges in-memory by matching `analise:{registroId}` to `registro._id`. Works fully offline.
- **D-46:** No new Go API endpoint needed for enriched registros.

### Frontend: Emotion Display (RegistroCard)
- **D-47:** **Colored emotion chips below sentimentoNome** in the card header (always visible, even when collapsed). Shows `emotionPrincipal` as primary chip + secondary emotions as smaller chips.
- **D-48:** When no analysis exists yet: no chips shown (graceful degradation — same as "Buscando sentimento" pattern for missing sentimentos).
- **D-49:** Chip styling: small rounded pills with emotion-appropriate colors (e.g., `alegria` → green, `tristeza` → blue). Tailwind classes.

### PDF Report: Emotion Integration
- **D-50:** **Summary section at top** — aggregate emotion frequency across the report period (most common emotions with percentages).
- **D-51:** **Per-registro emotions** — each registro block in the report shows `emotionPrincipal` + secondary emotions list below the sentimento line.
- **D-52:** Report service (`service/report.go`) needs to fetch analysis docs from CouchDB when generating PDF and pass both registro + analysis data to the template.

### Error Handling
- **D-53:** **Exponential backoff retry** — failed NLP gRPC calls retried 3 times with 1s, 4s, 16s delays. After 3 failures: log the error, skip silently. The registro simply won't have emotion analysis.
- **D-54:** No user-facing error for NLP failures — analysis is async and non-blocking (NLP-03). The `_changes` checkpoint advances regardless of individual failures.
- **D-55:** On API restart: resume from saved `last_seq` checkpoint. Unprocessed registrations from the previous run are not retried (acceptable for MVP — next restart re-processes from checkpoint forward).

### the agent's Discretion
- Exact rate limit value (50ms or adjust based on testing)
- Checkpoint doc schema and exact field names
- Emotion chip colors (palette mapping per emotion)
- Report summary aggregation logic (top N emotions, minimum threshold)
- Test file organization and test patterns
- `RegistroDoc` type update in frontend types (add optional `analise` field for merged display)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone & Requirements
- `.planning/REQUIREMENTS.md` §NLP — NLP-01, NLP-02, NLP-03 requirements
- `.planning/PROJECT.md` §Active — NLP feature description
- `.planning/phases/v2-nlp-analysis/v2-nlp-CONTEXT.md` — Full milestone decisions (D-01 through D-17)
- `.planning/ROADMAP.md` §Phase 7 — Phase structure and sub-phase boundaries

### Phase 07-01 (Infra — built)
- `.planning/phases/07-01-infra-nlp/07-01-CONTEXT.md` — Infra decisions (D-08 through D-17)
- `.planning/phases/07-01-infra-nlp/07-01-PLAN.md` — What was built (gRPC server, Dockerfile, proto)
- `nlp-service/proto/analysis.proto` — gRPC interface (AnalyzeRequest / AnalyzeResponse schema)
- `nlp-service/src/server.py` — Existing gRPC server with classifier integration
- `infra/docker-compose.yml` — NLP service already configured (`kanso-nlp`, port 50051)

### Phase 07-02 (Modelo — built)
- `.planning/phases/07-02-infra-nlp/07-02-CONTEXT.md` — Model decisions (D-18 through D-35)
- `nlp-service/src/classifier.py` — EmotionClassifier with predict() method
- `nlp-service/src/model_config.py` — 13 labels, threshold config

### Existing Go Backend Patterns
- `backend/cmd/kanso-api/main.go` — Chi router, service initialization, goroutine pattern for services
- `backend/internal/service/report.go` — Async goroutine pattern (`sync.Mutex` + goroutine)
- `backend/internal/service/push.go` — Service pattern with CouchDB repository
- `backend/internal/nlp/client.go` — Go gRPC client for NLP service (already built)
- `backend/internal/config/config.go` — Env-based config with `NLPGrpAddr`
- `backend/internal/repository/couchdb.go` — CouchDB operations, _find queries

### Frontend Patterns
- `frontend/src/components/RegistroCard.tsx` — Existing card component to receive emotion chips
- `frontend/src/services/registros.ts` — PouchDB operations for registros DB
- `frontend/src/types/index.ts` — RegistroDoc type

### Report Template
- `backend/internal/templates/report.html` — PDF report HTML template to receive emotion sections
- `backend/internal/pdf/generator.go` — chromedp PDF generation

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **nlp/client.go** — Full Go gRPC client (`nlp.Client`) with `Analyze()` method. Ready to use.
- **nlp/proto/analysis.pb.go** — Generated protobuf types. Ready to use.
- **service/report.go** — Async goroutine pattern (mutex + goroutine) — direct pattern for watcher.
- **repository/couchdb.go** — HTTP client with basic auth for CouchDB operations.
- **config/config.go** — `NLPGrpAddr` already defined ("nlp:50051").
- **docker-compose.yml** — `kanso-nlp` service already configured with `expose: ["50051", "8000"]`.

### Established Patterns
- **Goroutine service** — `report.go` shows pattern: constructor → struct with mutex → `Start()` called from `main.go`.
- **CouchDB query** — All CouchDB operations go through `repository.CouchDB` HTTP client wrapper.
- **PouchDB sync** — Frontend reads from PouchDB, auto-syncs to CouchDB. Analysis docs in same DB auto-sync.

### Integration Points
- **nlp/client.go Analyze()** — Watcher calls this for each new registro.
- **registros DB _changes feed** — Watcher listens here for new/updated registrations.
- **RegistroCard.tsx** — Add emotion chip display when analise data is available.
- **report.html** — Add summary section layout + per-registro emotion display.
- **registros.ts** — Add query for analise docs + merge logic with registro data.
- **main.go** — Wire watcher service alongside existing services.
- **CouchDB per-user isolation** — analise docs follow same `userId` constraint as registros (validate_doc_update ensures separation).

</code_context>

<specifics>
## Specific Ideas

- Emotion chip colors: alegria → green (emerald-500), tristeza → blue (blue-500), raiva → red (red-500), medo → purple (purple-500), nojo → amber, surpresa → orange, ansiedade → yellow, vergonha → pink, culpa → rose, saudade → violet, amor → pink-600, gratidão → teal, neutro → gray. Fine-tune during implementation.
- Checkpoint doc: `_id: "checkpoint:nlp_watcher"`, `type: "checkpoint"`, `last_seq: "..."`, `updatedAt: "..."`.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 07-03 — Integração*
*Context gathered: 2026-05-23*
