# Fix: Relatório — 400 invalid request body (Contract Mismatch)

## Root Cause

`POST /api/reports` no backend espera body `{"periodStart":"...","periodEnd":"..."}`, mas o frontend envia POST sem body. O `json.Decode` falha com `io.EOF` → retorna 400.

## Fix

Remover `periodStart`/`periodEnd` do contrato. O backend calcula as datas internamente.

### Task 1: Repository — Add `GetLastCompletedReport(sub)`

- **File**: `backend/internal/repository/couchdb.go`
- Add method `GetLastCompletedReport(sub string) (*ReportJobDoc, error)`
- Mango query: `type=relatorio`, `userSub=sub`, `status=done`
- Sort by `createdAt` desc, limit 1
- Returns nil if no completed report found

### Task 2: Service — Compute dates internally

- **File**: `backend/internal/service/report.go`
- Change `RequestReport` signature: remove `periodStart, periodEnd` params
- Inside `RequestReport`:
  - Call `s.couchRepo.GetLastCompletedReport(sub)` to get previous report
  - `periodStart`: if previous report exists → use its `periodEnd` (or `createdAt`); else use `""`
  - `periodEnd`: `time.Now().UTC().Format(time.RFC3339)`
- Pass computed dates to `generatePDF`
- Update `generatePDF` signature accordingly

### Task 3: Handler — Remove body parsing

- **File**: `backend/internal/handler/report.go`
- Remove the `req` struct with `PeriodStart`/`PeriodEnd`
- Remove the `json.NewDecoder(r.Body).Decode(&req)` block
- Call `h.svc.RequestReport(r.Context(), sub)` without dates
- Return error from `RequestReport` directly (no need for EOF/null body guard)

### Task 4: Frontend — No changes needed

- `reports.ts` already sends POST without body — this is now the correct contract

### Task 5: Tests

- **File**: `backend/internal/handler/report_test.go`
- Update `TestReportHandler_RequestReport_Returns202`: send POST with empty body, expect 202
- Remove `TestReportHandler_RequestReport_BadRequest`: no longer applicable
- Update `mockCouchDBServer` to handle `_find` query for user's completed reports (for GetLastCompletedReport)
- Add a test case for empty body explicitly (no body at all)
- Keep all other existing tests unchanged

## Files Affected

| File | Action |
|------|--------|
| `backend/internal/handler/report.go` | Remove body parsing |
| `backend/internal/handler/report_test.go` | Update tests |
| `backend/internal/service/report.go` | Compute dates internally |
| `backend/internal/repository/couchdb.go` | Add `GetLastCompletedReport` |

## Dependencies

- Phase 3 (Reports) — existing infrastructure stays

## Verification

- `cd backend && go test ./...` — all tests pass
- `cd frontend && npm run test` — all tests pass
