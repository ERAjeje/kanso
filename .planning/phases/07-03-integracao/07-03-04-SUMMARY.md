---
phase: 07-03-integracao
plan: 04
subsystem: api
tags: [pdf, report, emotion, couchdb, analise_nlp, go]

# Dependency graph
requires:
  - phase: 07-03-integracao-01
    provides: CouchDB types (PeriodRegistroDoc, AnaliseDoc) and methods (FindRegistrosByPeriod, FindAnaliseByRegistroIds)
provides:
  - Report service fetches registros and analise_nlp docs from CouchDB instead of using empty slice
  - Emotion summary section in PDF report with aggregate emotion frequency
  - Per-registro emotion chips in PDF report
affects: []

# Tech tracking
tech-stack:
  added: [sort (stdlib)]
  patterns: [Structured ReportData type replacing map[string]interface{} for template data]

key-files:
  created: []
  modified:
    - backend/internal/service/report.go
    - backend/internal/templates/report.html
    - backend/internal/service/report_test.go

key-decisions:
  - "Template data type changed from map[string]interface{} to typed ReportData struct"
  - "Emotion summary computed server-side by aggregating frequency across all analise_nlp docs in period"
  - "Inline styles used for emotion sections to ensure PDF-safe rendering via chromedp"

requirements-completed: [NLP-02]

# Metrics
duration: 1min
completed: 2026-05-23
---

# Phase 07-03 Plan 04: Emotion Data in PDF Reports

**Report service fetches registros + analise_nlp docs from CouchDB and renders aggregate emotion summary + per-registro emotion chips in the PDF template**

## Performance

- **Duration:** 1 min
- **Started:** 2026-05-23T18:24:12Z
- **Completed:** 2026-05-23T18:26:04Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments

- Report `generatePDF()` now fetches real registros data via `FindRegistrosByPeriod()` instead of passing an empty slice (fixes a pre-existing gap where registros were never fetched from CouchDB)
- Emotion summary section (`"Resumo das Emoções"`) with aggregate frequency count, sorted by count descending, shown only when emotions exist in period
- Per-registro emotion chips appear below the sentimento line in each registro block, gated by `{{if .Emocoes}}`
- Template rendering tests verify both present and empty emotion summary scenarios

## Task Commits

Each task was committed atomically:

1. **Task 1: Update report.go** — `14674b28` (feat)
   - Add EmotionSummaryItem, RegistroReportItem, ReportData types
   - Rewrite generatePDF to fetch registros + analise docs, compute emotion summary
   - Use structured ReportData instead of map[string]interface{}
2. **Task 2: Update report.html** — `a2b6d277` (feat)
   - Add "Resumo das Emoções" summary section with frequency list
   - Add per-registro emotion chips with inline styles for PDF rendering
3. **Task 3: Add tests** — `1116af5a` (test)
   - TestReportTemplate_RendersEmotionSummary — validates emotion sections render
   - TestReportTemplate_OmitsEmotionSummaryWhenEmpty — validates empty guard works

**Plan metadata:** (not yet committed separately — orchestrator will commit STATE/ROADMAP updates)

## Files Created/Modified

- `backend/internal/service/report.go` — Added 3 types, 87 lines of new data-fetching logic; replaced map with ReportData struct
- `backend/internal/templates/report.html` — Added 19 lines: emotion summary section + per-registro emotion chips
- `backend/internal/service/report_test.go` — Added 103 lines: 2 template rendering test functions

## Decisions Made

- **Structured ReportData type**: Replaced `map[string]interface{}` with a typed `ReportData` struct for compile-time safety. Go templates work identically with struct fields accessed via `.FieldName`.
- **Emotion aggregation**: Summary computed by counting frequency of each emotion across all analise_nlp docs in the period, sorted descending by count. This gives a quick overview of which emotions were most prevalent.
- **Inline styles for PDF**: Used inline `style` attributes for emotion chips and summary rows rather than CSS classes, ensuring chromedp renders them reliably in the PDF.

## Deviations from Plan

None — plan executed exactly as written.

## Threat Surface Scan

No security-relevant surface added beyond what was already in the plan's threat model:
- Emotion names come from the NLP model (13 hardcoded pt-BR strings from `model_config.py`), not user input — no template injection risk beyond existing registro fields
- `go test ./...` passes (full suite green)

## Self-Check: PASSED

- [x] All 3 modified files exist
- [x] All 3 commits found in git log
- [x] All acceptance criteria pass (structs, imports, template elements, test functions)
- [x] Full test suite passes (`go test ./... -count=1`)
- [x] SUMMARY.md created

---

*Phase: 07-03-integracao*
*Completed: 2026-05-23*
