---
phase: 07-03-integracao
plan: 03
subsystem: frontend
tags: [pouchdb, vitest, react, typescript, emotion-chips]

# Dependency graph
requires:
  - phase: 07-03-integracao
    provides: AnaliseNlpDoc type, getRegistros() merge
provides:
  - "PouchDB in-memory merge of analise_nlp docs with registro data"
  - "Colored emotion chips (principal + secondary) in RegistroCard"
  - "Emotion chip color palette for 13 emotions"
affects: [07-03-04-pdf, report.html template]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "PouchDB in-memory merge: allDocs both types, Map keyed by registroId"
    - "Emotion chip color mapping: Record<string, {bg, text}> with fallback"

key-files:
  created: []
  modified:
    - frontend/src/types/index.ts
    - frontend/src/services/registros.ts
    - frontend/src/services/registros.test.ts
    - frontend/src/components/RegistroCard.tsx
    - frontend/src/components/RegistroCard.test.tsx
    - frontend/src/pages/History.tsx

key-decisions:
  - "PouchDB in-memory merge via two allDocs calls + Map (per D-45, D-46)"
  - "Emotion colors hardcoded in EMOTION_CHIP_COLORS map (per UI-SPEC table)"
  - "Secondary chips: filter out principal, sort by score desc, max 3"
  - "Chips in card header (not expanded section) — always visible per D-47"

requirements-completed: [NLP-02]

# Metrics
duration: 5min
completed: 2026-05-23
---

# Phase 07-03 Plan 03: Emotion Display — Frontend Summary

**PouchDB in-memory merge of analise_nlp docs with registro data, plus colored emotion chips in RegistroCard**

## Performance

- **Duration:** 5 min
- **Started:** 2026-05-23T15:18:00Z
- **Completed:** 2026-05-23T15:22:50Z
- **Tasks:** 2 (TDD: 4 commits per TDD cycle)
- **Files modified:** 6

## Accomplishments
- Added `AnaliseNlpDoc`, `EmotionScore`, `RegistroWithAnalise` types to the frontend type system
- Updated `getRegistros()` to perform in-memory merge: fetches both `registro` and `analise_nlp` docs from PouchDB, builds an `analiseMap` keyed by registro ID, and returns `RegistroWithAnalise[]`
- Added `EMOTION_CHIP_COLORS` map with 13 emotion-specific Tailwind color pairs
- Rendered principal emotion chip + up to 3 secondary chips (sorted by score, filtered by ≠ principal) in RegistroCard header, always visible in collapsed state
- Graceful degradation: no chips rendered when `analise` is undefined

## Task Commits

Each task was committed atomically via TDD cycle:

1. **Task 1: AnaliseNlpDoc types + getRegistros merge** — TDD cycle:
   - `bcc4ee9b` (test) — Add failing test for getRegistros analise merge
   - `01118602` (feat) — Implement getRegistros analise merge
2. **Task 2: Emotion chips in RegistroCard** — TDD cycle:
   - `a92e8dcc` (test) — Add failing test for emotion chips in RegistroCard
   - `c03a9ac1` (feat) — Implement emotion chips in RegistroCard
3. **Type cleanup:** `c6644d9b` (fix) — Type cleanup after merge implementation

**Plan metadata:** *(committed after SUMMARY.md)*

_Note: TDD tasks have 2 commits each (test → feat)._

## Files Created/Modified
- `frontend/src/types/index.ts` — Added `EmotionScore`, `AnaliseNlpDoc`, `RegistroWithAnalise` interfaces
- `frontend/src/services/registros.ts` — Updated `getRegistros()`: imports new types, fetches analise_nlp docs, in-memory merge with Map keyed by registroId
- `frontend/src/services/registros.test.ts` — Added mock setup for `registrosDB.allDocs`, 3 merge test cases
- `frontend/src/components/RegistroCard.tsx` — Added `EMOTION_CHIP_COLORS` map, `getChipColors` helper, chip rendering JSX in card header
- `frontend/src/components/RegistroCard.test.tsx` — Added `makeEnrichedRegistro` helper, 5 emotion chip test cases
- `frontend/src/pages/History.tsx` — Updated state type to `RegistroWithAnalise[]`

## Decisions Made
- Used two separate `allDocs` calls (registros + analise_nlp) instead of one call with in-memory type filter — simpler, clearer, same performance at user-scale (<1000 docs)
- Emotion chip colors hardcoded as `Record<string, {bg, text}>` with gray fallback for unknown emotions — no runtime color computation needed
- Chips placed in card header (`flex-1 min-w-0` div) below date line, ensuring visibility in both collapsed and expanded states per D-47
- Secondary chips limited to 3, sorted by score descending, principal emotion excluded from secondaries

## Deviations from Plan

None - plan executed exactly as written.

### Test Fixes (documented during execution)

**1. "renders no chips" test — chip-only text selector**
- **Issue:** Test used `screen.queryByText('ansiedade')` which matched the `<h3>` heading (sentimentoNome), not just chips. Would always find the heading text even when no chips exist.
- **Fix:** Changed test to check for `screen.queryByText('medo')` — text that only appears as a secondary chip.
- **Verification:** Test passes: no chips rendered → `queryByText('medo')` returns null.

**2. "shows secondary emotion chips" test — duplicate text match**
- **Issue:** `screen.getByText('ansiedade')` found 2 elements (heading + principal chip), triggering "Found multiple elements" error.
- **Fix:** Changed to `screen.getAllByText('ansiedade')` with `.length >= 2` assertion.
- **Verification:** Test passes: principal chip + heading both present.

---

**Total deviations:** 0 auto-fixed. 2 test corrections documented during TDD refinement.
**Impact on plan:** None — test corrections were necessary for correct test logic. No scope change.

## Issues Encountered
- Two test logic issues fixed during TDD GREEN phase (documented above)
- Two TypeScript type errors fixed (implicit `any` in map callback, unused import) — compilation now clean

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Frontend emotion display complete — ready for Plan 04 (PDF: emotion summary section + per-registro emotions)
- Both `getRegistros()` and `RegistroCard` work when no analysis data exists (graceful degradation)
- NLP-02 (emotion enrichment) requirement now satisfied on frontend

---

## Self-Check: PASSED

- [x] All 6 modified files present on disk
- [x] 5 commits exist matching task descriptions
- [x] All 69 tests pass (11 test files)
- [x] TypeScript compiles with no errors (`npx tsc --noEmit`)
- [x] All acceptance criteria grep checks pass
- [x] NLP-02 requirement completed

---

*Phase: 07-03-integracao*
*Completed: 2026-05-23*
