package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edson/kanso-api/internal/repository"
)

func TestPushService_Subscribe_CreatesNewPrefs(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "id": "push_prefs:user123", "rev": "1-abc"})
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-key")

	err := svc.Subscribe("user123", "fcm-token-123", "America/Sao_Paulo")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPushService_Subscribe_UpdatesExisting(t *testing.T) {
	callCount := 0
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if r.Method == http.MethodGet && callCount == 1 {
			json.NewEncoder(w).Encode(repository.PushPrefsDoc{
				ID:      "push_prefs:user123",
				Rev:     "1-abc",
				Type:    "push_prefs",
				UserSub: "user123",
				Enabled: true,
				Times:   []string{"10:00", "20:00"},
			})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "id": "push_prefs:user123", "rev": "2-def"})
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-key")

	err := svc.Subscribe("user123", "new-token", "America/New_York")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPushService_GetPreferences_ReturnsDefaultsWhenNone(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-key")

	prefs, err := svc.GetPreferences("user123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !prefs.Enabled {
		t.Error("expected enabled to be true by default")
	}
	if len(prefs.Times) != 3 {
		t.Errorf("expected 3 default times, got %d", len(prefs.Times))
	}
}

func TestPushService_UpdatePreferences_SavesChanges(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "id": "push_prefs:user123", "rev": "1-abc"})
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-key")

	err := svc.UpdatePreferences("user123", false, []string{"14:00"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPushService_Send_ReturnsErrorWhenNoToken(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-key")

	err := svc.Send("user123")
	if err == nil {
		t.Fatal("expected error when no FCM token, got nil")
	}
}

func TestPushService_Send_MakesFCMRequest(t *testing.T) {
	var fcmBody map[string]interface{}
	var authHeader string

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository.PushPrefsDoc{
			ID:       "push_prefs:user123",
			Rev:      "1-abc",
			Type:     "push_prefs",
			UserSub:  "user123",
			Enabled:  true,
			Times:    []string{"12:00", "18:00", "23:00"},
			FCMToken: "fcm-token-123",
		})
	}))
	defer mockCouch.Close()

	fcmMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&fcmBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": 1})
	}))
	defer fcmMock.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewPushService(couchRepo, "test-server-key")
	svc.HTTPClient = fcmMock.Client()
	svc.FCMURL = fcmMock.URL

	err := svc.Send("user123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if authHeader == "" {
		t.Error("expected Authorization header")
	}
	if fcmBody == nil {
		t.Error("expected FCM request body")
	}
}
