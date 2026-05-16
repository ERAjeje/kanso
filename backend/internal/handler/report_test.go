package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
)

// mockCouchDBServer returns an httptest.Server that responds with CouchDB-compatible
// responses for the relatorios database operations used during testing.
func mockCouchDBServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":  true,
				"id":  "test-job-id",
				"rev": "1-abc123",
			})
		case http.MethodGet:
			json.NewEncoder(w).Encode(repository.ReportJobDoc{
				ID:          "test-job-id",
				Type:        "relatorio",
				UserSub:     "user123",
				Status:      repository.StatusDone,
				PeriodStart: "2026-01-01",
				PeriodEnd:   "2026-05-16",
				CreatedAt:   "2026-05-16T20:00:00Z",
				CompletedAt: "2026-05-16T20:01:00Z",
				FileName:    "test-job-id.pdf",
			})
		case http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"docs": []repository.ReportJobDoc{
					{
						ID:          "test-job-id",
						Type:        "relatorio",
						UserSub:     "user123",
						Status:      repository.StatusDone,
						PeriodStart: "2026-01-01",
						PeriodEnd:   "2026-05-16",
						CreatedAt:   "2026-05-16T20:00:00Z",
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func authenticatedRequest(r *http.Request, sub string) *http.Request {
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": sub + "@test.com",
		"name":  "Test User",
	}
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, claims)
	return r.WithContext(ctx)
}

func chiRouterWithHandler(h *ReportHandler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/reports", h.HandleRequestReport)
	r.Get("/api/reports", h.HandleListReports)
	r.Get("/api/reports/{id}", h.HandleGetReport)
	r.Get("/api/reports/{id}/download", h.HandleDownload)
	return r
}

func setupTestReportHandlerWithMock() (*ReportHandler, func()) {
	mockCouch := mockCouchDBServer()

	cfg := &config.Config{
		JWTSecret: "test-secret",
		PDFTmpDir: "/tmp/kanso-pdf-test",
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	gen := pdf.NewGenerator("", 5*time.Second)
	reportSvc := service.NewReportService(couchRepo, gen, cfg)
	handler := NewReportHandler(reportSvc, cfg)

	cleanup := func() {
		mockCouch.Close()
	}

	return handler, cleanup
}

func TestReportHandler_RequestReport_Returns202(t *testing.T) {
	h, cleanup := setupTestReportHandlerWithMock()
	defer cleanup()
	router := chiRouterWithHandler(h)

	body := `{"periodStart":"2026-01-01","periodEnd":"2026-05-16"}`
	req := httptest.NewRequest("POST", "/api/reports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, w.Body.String())
	}
	if resp["jobId"] == "" {
		t.Error("expected jobId in response")
	}
}

func TestReportHandler_ListReports_Returns200(t *testing.T) {
	h, cleanup := setupTestReportHandlerWithMock()
	defer cleanup()
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	var jobs []repository.ReportJobDoc
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, w.Body.String())
	}
}

func TestReportHandler_GetReport_Returns200(t *testing.T) {
	h, cleanup := setupTestReportHandlerWithMock()
	defer cleanup()
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports/test-job-id", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	var job repository.ReportJobDoc
	if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, w.Body.String())
	}
	if job.ID != "test-job-id" {
		t.Errorf("expected job ID test-job-id, got %s", job.ID)
	}
}

func TestReportHandler_GetReport_OwnershipDenied(t *testing.T) {
	h, cleanup := setupTestReportHandlerWithMock()
	defer cleanup()
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports/test-job-id", nil)
	req = authenticatedRequest(req, "other-user")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for ownership mismatch, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestReportHandler_Download_ReturnsPDF(t *testing.T) {
	tmpDir := t.TempDir()

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository.ReportJobDoc{
			ID:          "test-job-id",
			Type:        "relatorio",
			UserSub:     "user123",
			Status:      repository.StatusDone,
			PeriodStart: "2026-01-01",
			PeriodEnd:   "2026-05-16",
			CreatedAt:   "2026-05-16T20:00:00Z",
			CompletedAt: "2026-05-16T20:01:00Z",
			FileName:    "test-job-id.pdf",
		})
	}))
	defer mockCouch.Close()

	pdfContent := []byte("%PDF-1.4 fake pdf content for testing")
	filePath := fmt.Sprintf("%s/test-job-id.pdf", tmpDir)
	if err := os.WriteFile(filePath, pdfContent, 0644); err != nil {
		t.Fatalf("failed to create test PDF: %v", err)
	}

	cfg := &config.Config{
		JWTSecret: "test-secret",
		PDFTmpDir: tmpDir,
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	gen := pdf.NewGenerator("", 5*time.Second)
	reportSvc := service.NewReportService(couchRepo, gen, cfg)
	h := NewReportHandler(reportSvc, cfg)
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports/test-job-id/download", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %s", ct)
	}

	if w.Body.Len() == 0 {
		t.Error("expected non-empty PDF body")
	}
}

func TestReportHandler_RequestReport_BadRequest(t *testing.T) {
	h, cleanup := setupTestReportHandlerWithMock()
	defer cleanup()
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("POST", "/api/reports", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", w.Code)
	}
}
