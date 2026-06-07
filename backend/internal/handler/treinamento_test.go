package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/middleware"
)

type mockTreinamentoService struct {
	checkAndTrainErr     error
	currentModelVersion string
	reanalyzeErr         error
}

func (m *mockTreinamentoService) CheckAndTrain(ctx context.Context) error {
	return m.checkAndTrainErr
}

func (m *mockTreinamentoService) GetCurrentModelVersion() string {
	return m.currentModelVersion
}

func (m *mockTreinamentoService) ReanalyzeRegistros(ctx context.Context) error {
	return m.reanalyzeErr
}

func treinamentoRouter(h *TreinamentoHandler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/train", h.HandleTrain)
	r.Get("/api/train/status", h.HandleTrainStatus)
	r.Post("/api/reanalyze", h.HandleReanalyze)
	return r
}

func authenticatedTrainRequest(r *http.Request, sub string) *http.Request {
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": sub + "@test.com",
		"name":  "Test User",
	}
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, claims)
	return r.WithContext(ctx)
}

func TestHandleTrain_Returns200(t *testing.T) {
	mockSvc := &mockTreinamentoService{}
	h := NewTreinamentoHandler(mockSvc)
	router := treinamentoRouter(h)

	req := httptest.NewRequest("POST", "/api/train", nil)
	req = authenticatedTrainRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["trained"] == nil {
		t.Fatal("expected 'trained' field in response")
	}
}

func TestHandleTrainStatus_Returns200(t *testing.T) {
	mockSvc := &mockTreinamentoService{
		currentModelVersion: "v1.1",
	}
	h := NewTreinamentoHandler(mockSvc)
	router := treinamentoRouter(h)

	req := httptest.NewRequest("GET", "/api/train/status", nil)
	req = authenticatedTrainRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["modelVersion"] != "v1.1" {
		t.Fatalf("expected modelVersion 'v1.1', got %v", resp["modelVersion"])
	}
}

func TestHandleReanalyze_Returns200(t *testing.T) {
	mockSvc := &mockTreinamentoService{}
	h := NewTreinamentoHandler(mockSvc)
	router := treinamentoRouter(h)

	req := httptest.NewRequest("POST", "/api/reanalyze", nil)
	req = authenticatedTrainRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["reanalyzed"] == nil {
		t.Fatal("expected 'reanalyzed' field in response")
	}
}
