package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/nlp"
	"github.com/edson/kanso-api/internal/repository"
)

// mockAnalyzer implements nlp.Analyzer for testing.
type mockAnalyzer struct {
	analyzeFn func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error)
	callCount int
}

func (m *mockAnalyzer) Analyze(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
	m.callCount++
	if m.analyzeFn != nil {
		return m.analyzeFn(ctx, req)
	}
	return &nlp.AnalyzeResponse{
		EmotionPrincipal: "alegria",
		Emotions:         []nlp.EmotionScore{{Emotion: "alegria", Score: 0.95}},
		Scores:           map[string]float32{"alegria": 0.95},
		Intensidade:      0.95,
		ModeloVersao:     "v1.0",
	}, nil
}

// testCouchDBHandler creates an httptest.Server that mocks CouchDB endpoints
// for watcher tests. The handler is customizable via the changesResponse,
// checkpointResponse, and onCheckpointSave/onAnaliseSave callbacks.
func testCouchDBHandler(
	t *testing.T,
	changesResponse string,
	checkpointStatus int,
	checkpointBody string,
	onCheckpointSave func(seq string),
	onAnaliseSave func(docJSON string),
) *httptest.Server {
	t.Helper()

	if changesResponse == "" {
		changesResponse = `{"results":[],"last_seq":"0"}`
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "_changes"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(changesResponse))

		case strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodGet:
			w.WriteHeader(checkpointStatus)
			if checkpointStatus == http.StatusOK {
				w.Write([]byte(checkpointBody))
			}

		case strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"id":"checkpoint:nlp_watcher","rev":"1-abc"}`))
			if onCheckpointSave != nil {
				var body struct {
					LastSeq string `json:"last_seq"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				onCheckpointSave(body.LastSeq)
			}

		case strings.Contains(r.URL.Path, "analise:") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"rev":"1-abc"}`))
			if onAnaliseSave != nil {
				body := make(map[string]interface{})
				json.NewDecoder(r.Body).Decode(&body)
				jsonStr, _ := json.Marshal(body)
				onAnaliseSave(string(jsonStr))
			}

		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true}`))
		}
	}))
}

func TestWatcher_NoCheckpoint_StartsWithSinceZero(t *testing.T) {
	var sinceSeen string

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "_changes") {
			sinceSeen = r.URL.Query().Get("since")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results":[],"last_seq":"0"}`))
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, &mockAnalyzer{}, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()
	time.Sleep(100 * time.Millisecond)
	svc.Stop()

	if sinceSeen != "0" {
		t.Errorf("expected since=0, got since=%s", sinceSeen)
	}
}

func TestWatcher_ResumesFromCheckpoint(t *testing.T) {
	var sinceSeen string

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "_changes") {
			sinceSeen = r.URL.Query().Get("since")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results":[],"last_seq":"3-abc"}`))
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"_id":"checkpoint:nlp_watcher","_rev":"1-abc","type":"checkpoint","watcher":"nlp","last_seq":"3-abc"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, &mockAnalyzer{}, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()
	time.Sleep(100 * time.Millisecond)
	svc.Stop()

	if sinceSeen != "3-abc" {
		t.Errorf("expected since=3-abc, got since=%s", sinceSeen)
	}
}

func TestWatcher_ProcessesRegistro(t *testing.T) {
	var analiseSaved bool
	analiseSavedCh := make(chan struct{}, 1)

	mockCouch := testCouchDBHandler(t,
		`{"results":[{"seq":"1-abc","id":"reg123","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"ansiedade no peito","contexto":"reuniao importante","pensamentos":"vai dar certo","dataHora":"2026-05-23T10:00:00Z"}}],"last_seq":"1-abc"}`,
		http.StatusNotFound, "",
		nil,
		func(docJSON string) {
			analiseSaved = true
			select {
			case analiseSavedCh <- struct{}{}:
			default:
			}
		},
	)
	defer mockCouch.Close()

	analyzerCalls := 0
	mockAnalyzer := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			analyzerCalls++
			if req.RegistroID != "reg123" {
				t.Errorf("expected RegistroID=reg123, got %s", req.RegistroID)
			}
			if req.Sensacoes != "ansiedade no peito" {
				t.Errorf("expected Sensacoes=ansiedade no peito, got %s", req.Sensacoes)
			}
			return &nlp.AnalyzeResponse{
				EmotionPrincipal: "ansiedade",
				Emotions:         []nlp.EmotionScore{{Emotion: "ansiedade", Score: 0.85}},
				Scores:           map[string]float32{"ansiedade": 0.85},
				Intensidade:      0.85,
				ModeloVersao:     "v1.0",
			}, nil
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, mockAnalyzer, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()

	select {
	case <-analiseSavedCh:
		// Success — analise was saved
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for analise to be saved")
	}

	svc.Stop()

	if !analiseSaved {
		t.Error("expected SaveAnalise to be called")
	}
	if analyzerCalls != 1 {
		t.Errorf("expected 1 Analyzer call, got %d", analyzerCalls)
	}
}

func TestWatcher_SkipsNonRegistro(t *testing.T) {
	analyzerCalls := 0

	mockCouch := testCouchDBHandler(t,
		`{"results":[
			{"seq":"1-abc","id":"analise:reg1","changes":[{"rev":"1-def"}],"doc":{"type":"analise_nlp","emotionPrincipal":"alegria"}},
			{"seq":"2-abc","id":"checkpoint:nlp_watcher","changes":[{"rev":"1-def"}],"doc":{"type":"checkpoint","watcher":"nlp"}}
		],"last_seq":"2-abc"}`,
		http.StatusNotFound, "",
		nil,
		nil,
	)
	defer mockCouch.Close()

	mockAnalyzer := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			analyzerCalls++
			return &nlp.AnalyzeResponse{}, nil
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, mockAnalyzer, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()
	time.Sleep(200 * time.Millisecond)
	svc.Stop()

	if analyzerCalls != 0 {
		t.Errorf("expected 0 Analyzer calls for non-registro docs, got %d", analyzerCalls)
	}
}

func TestWatcher_SkipsDeletedDoc(t *testing.T) {
	analyzerCalls := 0

	mockCouch := testCouchDBHandler(t,
		`{"results":[
			{"seq":"1-abc","id":"reg123","changes":[{"rev":"2-def"}],"doc":{"_deleted":true}}
		],"last_seq":"1-abc"}`,
		http.StatusNotFound, "",
		nil,
		nil,
	)
	defer mockCouch.Close()

	mockAnalyzer := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			analyzerCalls++
			return &nlp.AnalyzeResponse{}, nil
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, mockAnalyzer, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()
	time.Sleep(200 * time.Millisecond)
	svc.Stop()

	if analyzerCalls != 0 {
		t.Errorf("expected 0 Analyzer calls for deleted doc, got %d", analyzerCalls)
	}
}

func TestWatcher_RetriesOnError(t *testing.T) {
	analyzerCalls := 0
	changesCallCount := 0
	saveCh := make(chan struct{}, 1)

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "_changes") {
			changesCallCount++
			if changesCallCount == 1 {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results":[{"seq":"1-abc","id":"reg123","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"test","contexto":"test","pensamentos":"test","dataHora":"2026-05-23T10:00:00Z"}}],"last_seq":"1-abc"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results":[],"last_seq":"1-abc"}`))
			}
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"id":"checkpoint:nlp_watcher","rev":"1-abc"}`))
			select { case saveCh <- struct{}{}: default: }
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer mockCouch.Close()

	mockAnalyzer := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			analyzerCalls++
			if analyzerCalls <= 2 {
				return nil, errors.New("temporary error")
			}
			return &nlp.AnalyzeResponse{
				EmotionPrincipal: "neutro",
				Emotions:         []nlp.EmotionScore{{Emotion: "neutro", Score: 1.0}},
				Scores:           map[string]float32{"neutro": 1.0},
				Intensidade:      0.5,
				ModeloVersao:     "v1.0",
			}, nil
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, mockAnalyzer, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()

	// Wait for checkpoint save (batch fully processed)
	select {
	case <-saveCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for checkpoint save")
	}

	svc.Stop()

	if analyzerCalls != 3 {
		t.Errorf("expected 3 Analyzer calls (2 fails + 1 success), got %d", analyzerCalls)
	}
}

func TestWatcher_SkipsAfterRetries(t *testing.T) {
	analyzerCalls := 0
	changesCallCount := 0
	saveCh := make(chan struct{}, 1)

	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "_changes") {
			changesCallCount++
			if changesCallCount == 1 {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results":[{"seq":"1-abc","id":"reg123","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"test","contexto":"test","pensamentos":"test","dataHora":"2026-05-23T10:00:00Z"}}],"last_seq":"1-abc"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results":[],"last_seq":"1-abc"}`))
			}
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"ok":true,"id":"checkpoint:nlp_watcher","rev":"1-abc"}`))
			select { case saveCh <- struct{}{}: default: }
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer mockCouch.Close()

	alwaysFail := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			analyzerCalls++
			return nil, errors.New("persistent error")
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, alwaysFail, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()

	// Wait for checkpoint to be saved (means batch was processed)
	select {
	case <-saveCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for checkpoint save")
	}

	svc.Stop()

	if analyzerCalls != 3 {
		t.Errorf("expected 3 Analyzer calls (all fail), got %d", analyzerCalls)
	}
}

func TestWatcher_SavesCheckpointAfterBatch(t *testing.T) {
	var checkpointSeq string
	checkpointSaved := make(chan struct{}, 1)

	mockCouch := testCouchDBHandler(t,
		`{"results":[
			{"seq":"1-abc","id":"reg1","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"a","contexto":"b","pensamentos":"c","dataHora":"2026-05-23T10:00:00Z"}},
			{"seq":"2-abc","id":"reg2","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"d","contexto":"e","pensamentos":"f","dataHora":"2026-05-23T11:00:00Z"}}
		],"last_seq":"2-abc"}`,
		http.StatusNotFound, "",
		func(seq string) {
			checkpointSeq = seq
			select {
			case checkpointSaved <- struct{}{}:
			default:
			}
		},
		nil,
	)
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, &mockAnalyzer{}, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	svc.Start()

	select {
	case <-checkpointSaved:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for checkpoint save")
	}

	svc.Stop()

	if checkpointSeq != "2-abc" {
		t.Errorf("expected checkpoint last_seq=2-abc, got %s", checkpointSeq)
	}
}

func TestWatcher_RateLimitsCalls(t *testing.T) {
	var callTimes []time.Time
	callTimesMu := make(chan struct{}, 10)

	mockCouch := testCouchDBHandler(t,
		`{"results":[
			{"seq":"1-abc","id":"reg1","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"a","contexto":"b","pensamentos":"c","dataHora":"2026-05-23T10:00:00Z"}},
			{"seq":"2-abc","id":"reg2","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"d","contexto":"e","pensamentos":"f","dataHora":"2026-05-23T11:00:00Z"}},
			{"seq":"3-abc","id":"reg3","changes":[{"rev":"1-def"}],"doc":{"type":"registro","sensacoes":"g","contexto":"h","pensamentos":"i","dataHora":"2026-05-23T12:00:00Z"}}
		],"last_seq":"3-abc"}`,
		http.StatusNotFound, "",
		nil,
		nil,
	)
	defer mockCouch.Close()

	mockAnalyzer := &mockAnalyzer{
		analyzeFn: func(ctx context.Context, req *nlp.AnalyzeRequest) (*nlp.AnalyzeResponse, error) {
			callTimes = append(callTimes, time.Now())
			callTimesMu <- struct{}{}
			return &nlp.AnalyzeResponse{
				EmotionPrincipal: "neutro",
				Emotions:         []nlp.EmotionScore{{Emotion: "neutro", Score: 1.0}},
				Scores:           map[string]float32{"neutro": 1.0},
				Intensidade:      0.5,
				ModeloVersao:     "v1.0",
			}, nil
		},
	}

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, mockAnalyzer, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	// Use default 50ms rate limit for this test

	svc.Start()

	// Wait for all 3 analyzer calls
	for i := 0; i < 3; i++ {
		select {
		case <-callTimesMu:
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for analyzer call %d", i+1)
		}
	}

	svc.Stop()

	if len(callTimes) < 3 {
		t.Fatalf("expected at least 3 analyzer calls, got %d", len(callTimes))
	}

	elapsed := callTimes[2].Sub(callTimes[0])
	// With 50ms rate limit and 3 docs (2 gaps), minimum is ~100ms
	minExpected := 100 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("expected elapsed >= %v between first and last call (2 gaps × 50ms), got %v", minExpected, elapsed)
	}
}

func TestWatcher_Stop_ExitsCleanly(t *testing.T) {
	mockCouch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "checkpoint:nlp_watcher") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return empty _changes immediately to avoid blocking
		if strings.Contains(r.URL.Path, "_changes") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results":[],"last_seq":"0"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer mockCouch.Close()

	couchRepo := repository.NewCouchDB(mockCouch.URL, "admin", "pass")
	svc := NewWatcherService(couchRepo, &mockAnalyzer{}, &config.Config{})
	svc.backoff = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond}
	svc.rateLimit = 1 * time.Millisecond

	// Start then immediately stop — should not hang
	svc.Start()
	svc.Stop()

	// Second stop should be safe
	svc.Stop()

	// Verify we can restart
	svc.Start()
	svc.Stop()

	t.Log("Start/Stop completed without hang")
}
