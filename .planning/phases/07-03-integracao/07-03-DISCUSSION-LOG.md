# Phase 07-03: Integração — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Phase:** 07-03 — Integração
**Areas discussed:** Service model, Changes feed, Backfill strategy, Analysis storage, Frontend display, PDF report, Error handling

---

## Service Model

| Option | Description | Selected |
|--------|-------------|----------|
| Goroutine in api (like report.go) | Single Go binary, watcher goroutine in main.go | ✓ |
| New microservice (like scheduler/) | Separate binary, own Dockerfile and docker-compose entry | |

**User's choice:** Goroutine in api (like report.go)
**Notes:** Simpler deploy, proven pattern from report.go. Watcher shares CouchDB+gRPC connections with main API.

---

## Changes Feed

| Option | Description | Selected |
|--------|-------------|----------|
| Polling (like scheduler) | GET /{db}/_changes every N seconds | |
| Long-poll | GET /{db}/_changes?since=...&timeout=25000 | ✓ |

**User's choice:** Long-poll
**Notes:** Lower latency than polling without persistent connection complexity. Filter by type:registro to avoid loops.

---

## Backfill Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Sequential from _changes?since=0 | Single code path for backfill and live | ✓ |
| Dedicated batch job on startup | Query with Mango _find, batch processing | |

**User's choice:** Sequential from _changes?since=0
**Notes:** Simpler, single code path. Save sequence checkpoint in CouchDB for restart resilience.

---

## Analysis Storage

| Option | Description | Selected |
|--------|-------------|----------|
| Same registros DB | analise:{registroId} in registros database | ✓ |
| Separate analises DB | New CouchDB database with per-user isolation | |

**User's choice:** Same registros DB
**Notes:** Simpler deploy (no new DB), single _changes feed, auto-sync to PouchDB. _changes filter by type avoids infinite loop.

---

## Frontend Display

| Option | Description | Selected |
|--------|-------------|----------|
| PouchDB local merge, API optional | analise_nlp docs auto-sync, merge in-memory | ✓ |
| Go API endpoint enriched response | New backend endpoint for joined data | |

**User's choice:** PouchDB local merge, API optional
**Notes:** Fully offline, no new backend endpoint. Same DB sync means analysis docs arrive automatically.

**Emotion display:** Colored chips below sentimentoNome in card header (always visible).

---

## PDF Report

| Option | Description | Selected |
|--------|-------------|----------|
| Per-registro emotion tag | emotionPrincipal + secondary emotions per registro block | |
| Summary section + per-registro | Aggregate emotion frequency + per-registro emotions | ✓ |

**User's choice:** Summary section + per-registro
**Notes:** Both aggregate top emotions + per-registro detail for therapist overview.

---

## Error Handling

| Option | Description | Selected |
|--------|-------------|----------|
| Retry with exponential backoff, max 3 attempts | 1s, 4s, 16s, then skip silently | ✓ |
| Persistent retry queue in CouchDB | pending_analysis docs for failed items | |

**User's choice:** Retry with exponential backoff, max 3 attempts
**Notes:** Async analysis means silent failure is acceptable. Checkpoint advances regardless.

---

## the agent's Discretion

- Exact rate limit value (50ms or adjust based on testing)
- Checkpoint doc schema and exact field names
- Emotion chip colors (palette mapping per emotion)
- Report summary aggregation logic (top N emotions, minimum threshold)
- Test file organization and test patterns
- RegistroDoc type update in frontend types

## Deferred Ideas

None — discussion stayed within phase scope.
