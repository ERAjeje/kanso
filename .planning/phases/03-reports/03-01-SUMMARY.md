---
phase: 03
plan: 01
name: PDF Report Backend — Generator, API, and Infrastructure
subsystem: backend
tags: [pdf, reports, chromedp, docker, couchdb, go]
requires: []
provides: [report-endpoints, pdf-generator, report-templates, docker-infrastructure]
affects: [backend/cmd/kanso-api, backend/Dockerfile, infra/docker-compose.yml]
tech-stack:
  added: [chromedp, chromedp/headless-shell]
  patterns: [mutex-protected-generation, filepath-based-path-safety, html-template-auto-escaping]
key-files:
  created:
    - backend/internal/handler/report.go
    - backend/internal/handler/report_test.go
    - backend/internal/service/report.go
    - backend/internal/service/report_test.go
    - backend/internal/pdf/generator.go
    - backend/internal/pdf/generator_test.go
    - backend/internal/templates/report.html
    - backend/internal/templates/embed.go
    - infra/docker-compose.yml
    - backend/.gitignore
  modified:
    - backend/internal/config/config.go
    - backend/internal/repository/couchdb.go
    - backend/cmd/kanso-api/main.go
    - backend/Dockerfile
    - backend/go.mod
    - backend/go.sum
decisions:
  - chromedp chosen for headless PDF generation over wkhtmltopdf (better HTML5/CSS3 support)
  - chromedp/headless-shell:latest as Docker runtime (smaller than full Chromium)
  - //go:embed kept in templates/embed.go (not service/report.go) because Go embed requires file co-location
  - Handler tests use httptest.Server mock CouchDB instead of interface mocking
  - Mutex (sync.Mutex) protects concurrent PDF generation limit — only one at a time
metrics:
  duration: ~8 minutes
  commits: 3
  files_created: 10
  files_modified: 6
  total_lines: ~1328
  completed_date: 2026-05-16
---

# Phase 3 — Plan 01: PDF Report Backend Summary

## One-liner

Go backend for async PDF report generation — CouchDB job tracking, chromedp HTML-to-PDF, chi API with JWT ownership enforcement, and multi-stage Docker infrastructure with headless-shell.

## What Was Built

### 1. Repository Layer (couchdb.go — 5 new methods)
- **CreateReportJob** — PUT job document to `relatorios` DB with `type: "relatorio"`
- **GetReportJob** — GET job by document ID, returns `nil` on 404
- **UpdateReportJobStatus** — Fetches current doc, updates status/completedAt/error, PUTs back
- **ListReportJobsByUser** — Mango query on `relatorios/_find` with `type + userSub` selector
- **ReportJobExists** — Convenience wrapper around GetReportJob

### 2. Service Layer (report.go — 4 methods + mutex)
- **RequestReport** — Creates CouchDB job doc, acquires mutex, launches async goroutine for PDF generation
- **generatePDF** (private) — Template rendering → chromedp generation → file write → status update
- **GetJobs** — Delegates to `ListReportJobsByUser`
- **GetJob** — Delegates to `GetReportJob` with ownership filter (`job.UserSub != sub`)
- **GetPDF** — Ownership check → status check → path-safe file read using `filepath.Join(PDFTmpDir, filepath.Base(fileName))`
- **sync.Mutex** — Acquired in `RequestReport`, released in async goroutine after generation completes

### 3. Handler Layer (report.go — 4 endpoints)
- **POST /api/reports** — `HandleRequestReport`: Extracts JWT `sub`, decodes periodStart/periodEnd, returns 202 + `{jobId}`
- **GET /api/reports** — `HandleListReports`: Returns user's jobs as JSON array
- **GET /api/reports/{id}** — `HandleGetReport`: Returns single job with ownership check (404 on mismatch)
- **GET /api/reports/{id}/download** — `HandleDownload`: Returns PDF bytes with `application/pdf` Content-Type + Content-Disposition

### 4. PDF Generator (generator.go)
- chromedp-based HTML-to-PDF conversion with headless Chrome
- Data URI navigation with URL-encoded HTML content
- Configurable exec path via `CHROMEDP_PATH` env var or constructor parameter
- 30-second context timeout
- `page.PrintToPDF` with print background, 0.4in margins

### 5. HTML Template (report.html)
- Go `html/template` (not `template.HTML`) — automatic HTML escaping
- PT-BR locale: "Relatório Kanso", data format in Brazilian format
- CSS styling with indigo accent colors matching frontend design
- Print-friendly: `page-break-inside: avoid` on registro cards

### 6. Docker Infrastructure
- **Multi-stage Dockerfile**: `builder` (golang:1.26-alpine) → `runtime` (chromedp/headless-shell:latest) → `dev` (golang:1.26-alpine + air)
- **docker-compose.yml**: CouchDB 3.5 + API service with seccomp=unconfined, 2GB mem_limit

### 7. Configuration
- `PDFTmpDir` config field with `/tmp/kanso-pdf` default
- Routes wired behind JWTRequired middleware

## Verification Results

| Criteria | Result |
|----------|--------|
| `go build ./cmd/kanso-api` | ✅ PASS |
| `go test ./internal/handler/ -run TestReport -v` | ✅ 6/6 PASS |
| `go test ./internal/service/ -run TestReport -v` | ✅ 5/5 PASS |
| `go vet ./internal/...` | ✅ PASS |
| Multi-stage Dockerfile (builder/runtime/dev) | ✅ Verified |
| All 4 API endpoints wired | ✅ Verified |
| Mutex in report.go | ✅ Verified (line 20) |
| JWT ownership checks in handler | ✅ Verified (4 endpoints) |
| Path traversal protection | ✅ Verified (filepath.Base in GetPDF) |
| chromedp dependency | ✅ Verified (go.mod) |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 — Missing critical functionality] Handler tests needed mock CouchDB**
- **Found during:** Task 2
- **Issue:** Handler tests failed because no CouchDB running — tests tried to connect to localhost:15984
- **Fix:** Rewrote tests to use `httptest.NewServer` as mock CouchDB, returning realistic responses for PUT/GET/POST operations
- **Files modified:** `backend/internal/handler/report_test.go`
- **Commit:** `8ded834b`

**2. [Rule 1 — Bug] `//go:embed` not possible in service/report.go**
- **Found during:** Task 2
- **Issue:** Plan specified `//go:embed` in service/report.go, but Go's embed directive requires the file to be in the same directory tree as the embedded file
- **Fix:** Kept `//go:embed` in `templates/embed.go` (where `report.html` lives) — templates package is imported by service
- **Files modified:** `backend/internal/templates/embed.go`
- **Commit:** `8ded834b`

**3. [Rule 2 — Missing critical functionality] `.gitignore` pattern too broad**
- **Found during:** Task 2 staging
- **Issue:** `kanso-api` in `.gitignore` matched both `backend/kanso-api` (binary) and `backend/cmd/kanso-api/` (package directory)
- **Fix:** Changed to `/kanso-api` (root-relative only)
- **Files modified:** `backend/.gitignore`
- **Commit:** `8ded834b`

**4. [Rule 1 — Bug] chromedp API changed in v0.15.1**
- **Found during:** Task 3 compilation
- **Issue:** `chromedp.PrintToPDF` was removed; new API uses `page.PrintToPDF()` from `github.com/chromedp/cdproto/page`
- **Fix:** Updated import and method call to use `page.PrintToPDF()`
- **Files modified:** `backend/internal/pdf/generator.go`
- **Commit:** `3b98b662`

### Design Decisions Verified

1. **Mock server for handler tests** — Using `httptest.NewServer` instead of interface mocks keeps the test focused on HTTP behavior without refactoring the repository into an interface. The mock returns CouchDB-compatible JSON for PUT/GET/POST operations.

2. **Template embed location** — Embedded in `templates/embed.go` (not `service/report.go`) because Go's `//go:embed` requires the source file to be in the same directory as the embedded file.

## Known Stubs

| Stub | File | Line | Reason |
|------|------|------|--------|
| Empty Registros array in template data | `service/report.go` | 83 | Emotion data fetching from CouchDB `registros` DB is deferred — the template renders "Nenhum registro encontrado" |

## Threat Flags

| Flag | File | Description |
|------|------|------------|
| `threat_flag: new_auth_endpoints` | `cmd/kanso-api/main.go` | 4 new API endpoints added behind JWTRequired middleware — all enforce ownership via `sub` claim |
| `threat_flag: file_access_pattern` | `service/report.go` | PDF file access via `os.ReadFile` within PDFTmpDir with filepath.Base sanitization |

## Commit History

| Commit | Description |
|--------|-------------|
| `131b8f51` | `test(3-01):` Add types, config, test contracts, and stubs for PDF report backend |
| `8ded834b` | `feat(3-01):` Implement report handler, service, repository, templates, and wire routes |
| `3b98b662` | `feat(3-01):` Add chromedp PDF generator, multi-stage Docker, docker-compose |

## Self-Check: PASSED

All files verified:
- `backend/internal/handler/report.go` ✅ exists
- `backend/internal/handler/report_test.go` ✅ exists
- `backend/internal/service/report.go` ✅ exists
- `backend/internal/service/report_test.go` ✅ exists
- `backend/internal/pdf/generator.go` ✅ exists
- `backend/internal/pdf/generator_test.go` ✅ exists (with //go:build integration)
- `backend/internal/templates/report.html` ✅ exists
- `backend/internal/templates/embed.go` ✅ exists
- `infra/docker-compose.yml` ✅ exists
- `backend/.gitignore` ✅ exists

All commits verified:
- `131b8f51` ✅
- `8ded834b` ✅
- `3b98b662` ✅
