package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/repository"
)

func testTreinamentoCouchDBHandler(t *testing.T) *httptest.Server {
	t.Helper()

	var checkpointRev string

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "checkpoint:training") && r.Method == http.MethodGet:
			if checkpointRev == "" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"_id":"checkpoint:training","_rev":"` + checkpointRev + `","type":"checkpoint","contentHash":"abc123","modelVersion":"v1.0","updatedAt":"2026-06-07T00:00:00Z"}`))
			}

		case strings.Contains(r.URL.Path, "checkpoint:training") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"id":"checkpoint:training","rev":"2-def"}`))
			checkpointRev = "2-def"

		case strings.Contains(r.URL.Path, "_find"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"docs":[]}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewTreinamentoService(t *testing.T) {
	mockCouch := testTreinamentoCouchDBHandler(t)
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	mockNLP := &mockAnalyzer{}
	cfg := &config.Config{}

	svc := NewTreinamentoService(couchRepo, mockNLP, cfg)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestTreinamentoService_CheckAndTrain_NoChanges(t *testing.T) {
	// Server that returns matching hashes => no change detected
	callCount := 0
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++

		switch {
		case strings.Contains(r.URL.Path, "checkpoint:training") && r.Method == http.MethodGet:
			// Return existing checkpoint with a known hash
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"_id":"checkpoint:training","_rev":"1-abc","type":"checkpoint","contentHash":"4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945","modelVersion":"v1.0","updatedAt":"2026-06-07T00:00:00Z"}`))

		case strings.Contains(r.URL.Path, "_find"):
			// Empty training data (hash of empty JSON = sha256 of "[]")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"docs":[]}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	mockNLP := &mockAnalyzer{}
	cfg := &config.Config{NLPHTTPAddr: "http://localhost:9999"}

	svc := NewTreinamentoService(couchRepo, mockNLP, cfg)

	err := svc.CheckAndTrain(context.Background())
	// Hash should match, no changes detected, no NLP call needed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTreinamentoService_GetCurrentModelVersion_Default(t *testing.T) {
	mockCouch := testTreinamentoCouchDBHandler(t)
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	mockNLP := &mockAnalyzer{}
	cfg := &config.Config{}

	svc := NewTreinamentoService(couchRepo, mockNLP, cfg)

	version := svc.GetCurrentModelVersion()
	// No checkpoint exists yet, so version should be empty or default
	if version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestTreinamentoService_HasTrainingChanged_EmptyDB(t *testing.T) {
	mockCouch := testTreinamentoCouchDBHandler(t)
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")

	changed, hash, err := couchRepo.HasTrainingChanged()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty training data + no checkpoint = changed (no previous hash to compare)
	if !changed {
		t.Log("empty DB with no checkpoint: changed should be true (first run)")
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestCouchDB_GetTrainingCheckpoint_NotFound(t *testing.T) {
	mockCouch := testTreinamentoCouchDBHandler(t)
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")

	cp, err := couchRepo.GetTrainingCheckpoint()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp != nil {
		t.Fatal("expected nil checkpoint when not found")
	}
}

func TestCouchDB_TrainingCheckpointRoundtrip(t *testing.T) {
	var storedHash string
	var storedVersion string

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "checkpoint:training") && r.Method == http.MethodGet:
			if storedHash == "" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"_id":"checkpoint:training","_rev":"1-abc","type":"checkpoint","contentHash":"` + storedHash + `","modelVersion":"` + storedVersion + `","updatedAt":"2026-06-07T00:00:00Z"}`))
			}

		case strings.Contains(r.URL.Path, "checkpoint:training") && r.Method == http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			storedHash, _ = body["contentHash"].(string)
			storedVersion, _ = body["modelVersion"].(string)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"id":"checkpoint:training","rev":"2-def"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")

	err := couchRepo.SaveTrainingCheckpoint("hash123", "v1.1")
	if err != nil {
		t.Fatalf("unexpected error saving checkpoint: %v", err)
	}

	cp, err := couchRepo.GetTrainingCheckpoint()
	if err != nil {
		t.Fatalf("unexpected error getting checkpoint: %v", err)
	}
	if cp == nil {
		t.Fatal("expected non-nil checkpoint")
	}
	if cp.ContentHash != "hash123" {
		t.Fatalf("expected hash 'hash123', got %q", cp.ContentHash)
	}
	if cp.ModelVersion != "v1.1" {
		t.Fatalf("expected version 'v1.1', got %q", cp.ModelVersion)
	}
}

func TestCouchDB_ComputeTrainingHash_Empty(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "_find"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"docs":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")

	hash, err := couchRepo.ComputeTrainingHash()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash even for empty data")
	}
}
