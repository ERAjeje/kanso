---
phase: 07-03
slug: integracao
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-23
---

# Phase 07-03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework (Backend)** | Go stdlib `testing` + `httptest` |
| **Framework (Frontend)** | Vitest v4 with jsdom |
| **Config file** | none — standard Go test convention / `vite.config.ts` test section |
| **Quick run (Go)** | `go test ./internal/service/ -run "TestWatcher|TestReport" -v -count=1` |
| **Quick run (Frontend)** | `npx vitest run src/services/registros.test.ts src/components/RegistroCard.test.tsx --reporter=verbose` |
| **Full suite command** | `go test ./... -v -count=1 && npm test` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run Go watcher tests + frontend merge/display tests
- **After every plan wave:** Run full suite (Go + frontend)
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 07-03-01 | 01 | 1 | NLP-01 | — | N/A — reads CouchDB _changes, filters by type | unit (Go) | `go test ./internal/service/ -run TestWatcher_ProcessesRegistro -v` | ❌ W0 | ⬜ pending |
| 07-03-01 | 01 | 1 | NLP-01 | — | N/A — skips non-registro doc types to avoid loops | unit (Go) | `go test ./internal/service/ -run TestWatcher_SkipsNonRegistro -v` | ❌ W0 | ⬜ pending |
| 07-03-01 | 01 | 1 | NLP-02 | — | N/A — stores analise:{registroId} doc | unit (Go) | `go test ./internal/service/ -run TestWatcher_SavesAnaliseDoc -v` | ❌ W0 | ⬜ pending |
| 07-03-01 | 01 | 1 | NLP-03 | — | N/A — retries failed NLP with backoff | unit (Go) | `go test ./internal/service/ -run TestWatcher_RetriesOnError -v` | ❌ W0 | ⬜ pending |
| 07-03-01 | 01 | 1 | NLP-03 | — | N/A — skips silently after 3 retries | unit (Go) | `go test ./internal/service/ -run TestWatcher_SkipsAfterRetries -v` | ❌ W0 | ⬜ pending |
| 07-03-01 | 01 | 1 | NLP-03 | — | N/A — checkpoint persists and resumes | unit (Go) | `go test ./internal/service/ -run TestWatcher_ResumesFromCheckpoint -v` | ❌ W0 | ⬜ pending |
| 07-03-02 | 02 | 1 | NLP-02 | — | N/A — frontend merges analise_nlp with registro | unit (Vitest) | `npx vitest run src/services/registros.test.ts -t "merges analise" --reporter=verbose` | ❌ W0 | ⬜ pending |
| 07-03-02 | 02 | 1 | NLP-02 | — | N/A — chips rendered when analise exists | unit (Vitest) | `npx vitest run src/components/RegistroCard.test.tsx -t "emotion chips" --reporter=verbose` | ❌ W0 | ⬜ pending |
| 07-03-03 | 03 | 2 | NLP-02 | — | N/A — PDF includes emotion summary + per-registro | unit (Go) | `go test ./internal/service/ -run TestReportService_WithAnalise -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/internal/service/watcher_test.go` — new file, all watcher tests
- [ ] `frontend/src/services/registros.test.ts` — add merge test case for analise_nlp docs
- [ ] `frontend/src/components/RegistroCard.test.tsx` — add emotion chip test cases

*Existing infrastructure covers all other phase requirements (Go test runner, Vitest, jsdom).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| CouchDB _changes long-poll connection lifecycle | NLP-01 | Requires running CouchDB instance | Start docker-compose, verify watcher connects and processes registrations |
| shadcn component integration (button.tsx) | — | Visual verification | Confirm button renders correctly in any screen |
| Emotion chip color rendering | NLP-02 | Visual verification | Confirm all 13 emotion colors render as expected |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
