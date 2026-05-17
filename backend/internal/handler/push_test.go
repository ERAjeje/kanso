package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
)

func setupTestPushHandler() (*PushHandler, *httptest.Server) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":  true,
				"id":  "push_prefs:user123",
				"rev": "1-abc",
			})
		}
	}))

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	pushSvc := service.NewPushService(couchRepo, "test-key")
	handler := NewPushHandler(pushSvc)

	return handler, mockCouch
}

func pushRouter(h *PushHandler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/push/subscribe", h.HandleSubscribe)
	r.Get("/api/push/preferences", h.HandleGetPreferences)
	r.Put("/api/push/preferences", h.HandleUpdatePreferences)
	r.Post("/api/push/send", h.HandleSend)
	return r
}

func authenticatedPushRequest(r *http.Request, sub string) *http.Request {
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": sub + "@test.com",
		"name":  "Test User",
	}
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, claims)
	return r.WithContext(ctx)
}

func TestPushHandler_Subscribe_Returns200(t *testing.T) {
	h, mockCouch := setupTestPushHandler()
	defer mockCouch.Close()
	router := pushRouter(h)

	body := `{"fcmToken":"token-123","timezone":"America/Sao_Paulo"}`
	req := httptest.NewRequest("POST", "/api/push/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedPushRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf(`expected status "ok", got %q`, resp["status"])
	}
}

func TestPushHandler_Subscribe_InvalidBody(t *testing.T) {
	h, mockCouch := setupTestPushHandler()
	defer mockCouch.Close()
	router := pushRouter(h)

	req := httptest.NewRequest("POST", "/api/push/subscribe", strings.NewReader(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedPushRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestPushHandler_GetPreferences_Returns200(t *testing.T) {
	h, mockCouch := setupTestPushHandler()
	defer mockCouch.Close()
	router := pushRouter(h)

	req := httptest.NewRequest("GET", "/api/push/preferences", nil)
	req = authenticatedPushRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var prefs repository.PushPrefsDoc
	json.NewDecoder(w.Body).Decode(&prefs)
	if prefs.UserSub != "user123" {
		t.Errorf("expected userSub user123, got %s", prefs.UserSub)
	}
}

func TestPushHandler_UpdatePreferences_Returns200(t *testing.T) {
	h, mockCouch := setupTestPushHandler()
	defer mockCouch.Close()
	router := pushRouter(h)

	body := `{"enabled":false,"times":["10:00","22:00"]}`
	req := httptest.NewRequest("PUT", "/api/push/preferences", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = authenticatedPushRequest(req, "user123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf(`expected status "ok", got %q`, resp["status"])
	}
}

func TestPushHandler_Send_Returns200(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository.PushPrefsDoc{
			ID:       "push_prefs:user123",
			Rev:      "1-abc",
			Type:     "push_prefs",
			UserSub:  "user123",
			Enabled:  true,
			Times:    []string{"12:00"},
			FCMToken: "token-123",
		})
	}))
	defer mockCouch.Close()

	fcmMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": 1})
	}))
	defer fcmMock.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	pushSvc := service.NewPushService(couchRepo, "test-key")
	pushSvc.HTTPClient = fcmMock.Client()
	pushSvc.FCMURL = fcmMock.URL

	h := NewPushHandler(pushSvc)
	router := pushRouter(h)

	body := `{"userId":"user123"}`
	req := httptest.NewRequest("POST", "/api/push/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
