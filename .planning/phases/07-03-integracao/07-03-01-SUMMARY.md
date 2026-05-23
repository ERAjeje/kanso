---
phase: 07-03-integracao
plan: 01
subsystem: repository
tags: couchdb, changes-feed, checkpoint, nlp-watcher, analise
requires:
  - phase: 07-01-infra-nlp
    provides: gRPC server, Docker config, protobuf types
  - phase: 07-02-modelo-nlp
    provides: sentiment model, EmotionScore type, Analyzer interface
provides:
  - CouchDB _changes feed long-poll method (GetChanges)
  - Checkpoint persistence (GetCheckpoint/SaveCheckpoint) for watcher resume
  - Analysis document storage (SaveAnalise) for NLP results
  - Registro period queries (FindRegistrosByPeriod) for PDF report enrichment
  - Bulk analysis lookup (FindAnaliseByRegistroIds) for PDF emotion data
affects: 07-03-integracao plan 02 (watcher), plan 04 (report)

tech-stack:
  added: []
  patterns:
    - CouchDB _changes feed long-poll with 30s HTTP timeout
    - GET-then-PUT checkpoint pattern with _rev handling
    - MangoQuery _find for time-range and $in queries
    - Separate HTTP client for long-poll vs short-lived requests

key-files:
  created: []
  modified:
    - backend/internal/repository/couchdb.go

key-decisions:
  - "GetChanges uses dedicated 30s http.Client (not c.httpClient's 10s) to accommodate 25s long-poll timeout"
  - "Checkpoint uses GET-then-PUT pattern for safe _rev handling on updates"
  - "FindRegistrosByPeriod uses string comparison for ISO 8601 date ranges (valid for lexicographic ordering)"
  - "PeriodRegistroDoc.Sentimento uses json: sentimentoNome to match CouchDB field name"
  - "FindAnaliseByRegistroIds returns empty slice (not nil) for empty input to simplify callers"

requirements-completed:
  - NLP-01
  - NLP-02
  - NLP-03

duration: 3 min
completed: 2026-05-23
---

# Phase 07-03 Plan 01: CouchDB repository types and methods for NLP integration

**CouchDB _changes feed long-poll, checkpoint persistence, analysis document storage, and registro/analise query methods for the NLP watcher service**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-23T18:10:52Z
- **Completed:** 2026-05-23T18:13:00Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments

- Added 6 Go types for NLP integration: `ChangesResult`, `ChangesResponse`, `CheckpointDoc`, `EmotionScore`, `AnaliseDoc`, `PeriodRegistroDoc`
- Implemented `GetChanges` — CouchDB _changes long-poll feed with 30s timeout (25s poll + 5s buffer)
- Implemented `GetCheckpoint` / `SaveCheckpoint` — checkpoint persistence with GET-then-PUT pattern for safe _rev handling
- Implemented `SaveAnalise` — analysis document storage with auto-ID generation (`analise:{registroId}`)
- Implemented `FindRegistrosByPeriod` — _find query with type, userSub, and ISO 8601 date range selection
- Implemented `FindAnaliseByRegistroIds` — _find query using $in operator for bulk analysis lookup
- All methods follow existing codebase patterns (basic auth, error wrapping, mangoQuery responses)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add CouchDB repository types and methods for NLP watcher** - `055ff280` (feat)

**Plan metadata:** `(pending)`

_Note: Single-task plan — all changes in one commit._

## Files Created/Modified

- `backend/internal/repository/couchdb.go` - Added 272 lines: 6 new types and 6 new methods for NLP watcher integration and report enrichment

## Decisions Made

- **Separate HTTP client for long-poll:** `GetChanges` uses a dedicated `&http.Client{Timeout: 30 * time.Second}` instead of the CouchDB struct's 10s client. This is necessary because the _changes long-poll holds connections open for up to 25s.
- **GET-then-PUT checkpoint pattern:** `SaveCheckpoint` first calls `GetCheckpoint` to obtain the current `_rev`, then uses `putDoc` for the update. On first call (no existing checkpoint), it creates the doc without `_rev`.
- **ISO 8601 string comparison for date ranges:** `FindRegistrosByPeriod` uses `$gte`/`$lte` on `dataHora` string fields. ISO 8601 lexicographic ordering makes this correct.
- **sentimentoNome JSON tag:** `PeriodRegistroDoc.Sentimento` uses `json:"sentimentoNome"` to match the actual CouchDB field name in registro documents.
- **Empty input handling:** `FindAnaliseByRegistroIds` returns an empty slice (not nil) when given an empty `ids` slice, simplifying caller code.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Repository layer complete and verified (build + vet pass). Ready for Plan 02 (watcher service) which depends on these types and methods. The watcher will import `ChangesResponse`, call `GetChanges` for the feed loop, use `GetCheckpoint`/`SaveCheckpoint` for resume, and `SaveAnalise` for storing results.

---

*Phase: 07-03-integracao*
*Completed: 2026-05-23*
