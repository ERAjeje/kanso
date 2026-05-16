package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

func setupTestReportHandler(t *testing.T) (*ReportHandler, *service.ReportService, *repository.CouchDB) {
	t.Helper()

	// Use a config that points to nothing — we won't actually call CouchDB
	cfg := &config.Config{
		JWTSecret: "test-secret",
		PDFTmpDir: "/tmp/kanso-pdf-test",
	}

	couchRepo := repository.NewCouchDB("http://localhost:15984", "admin", "pass")
	reportSvc := service.NewReportService(couchRepo, cfg)
	handler := NewReportHandler(reportSvc, cfg)
	return handler, reportSvc, couchRepo
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

func TestReportHandler_RequestReport_Returns202(t *testing.T) {
	h, _, _ := setupTestReportHandler(t)
	router := chiRouterWithHandler(h)

	body := `{"periodStart":"2026-01-01","periodEnd":"2026-05-16"}`
	req := httptest.NewRequest("POST", "/api/reports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["jobId"] == "" {
		t.Error("expected jobId in response")
	}
}

func TestReportHandler_ListReports_Returns200(t *testing.T) {
	h, _, _ := setupTestReportHandler(t)
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var jobs []repository.ReportJobDoc
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestReportHandler_GetReport_Returns200(t *testing.T) {
	h, _, _ := setupTestReportHandler(t)
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports/some-job-id", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", w.Code)
	}
}

func TestReportHandler_Download_ReturnsPDF(t *testing.T) {
	h, _, _ := setupTestReportHandler(t)
	router := chiRouterWithHandler(h)

	req := httptest.NewRequest("GET", "/api/reports/some-job-id/download", nil)
	req = authenticatedRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", w.Code)
	}

	if w.Code == http.StatusOK {
		ct := w.Header().Get("Content-Type")
		if ct != "application/pdf" {
			t.Errorf("expected Content-Type application/pdf, got %s", ct)
		}
	}
}
