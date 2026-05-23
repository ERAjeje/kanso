package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/nlp"
	"github.com/edson/kanso-api/internal/repository"
)

// WatcherService watches the registros _changes feed for new registrations,
// sends them to the NLP service for emotion analysis, and stores results
// as analise:{registroId} documents in the registros database.
type WatcherService struct {
	mu        sync.Mutex
	couchRepo *repository.CouchDB
	client    nlp.Analyzer
	cfg       *config.Config
	stopChan  chan struct{}
	started   bool

	// Test-friendly overrides for backoff and rate limiting
	backoff   []time.Duration
	rateLimit time.Duration
}

// NewWatcherService creates a new NLP watcher service.
// The client parameter must be non-nil; if the NLP service is unavailable,
// the watcher should not be created (see main.go wiring).
func NewWatcherService(couchRepo *repository.CouchDB, client nlp.Analyzer, cfg *config.Config) *WatcherService {
	return &WatcherService{
		couchRepo: couchRepo,
		client:    client,
		cfg:       cfg,
		stopChan:  make(chan struct{}),
		backoff:   []time.Duration{1 * time.Second, 4 * time.Second, 16 * time.Second},
		rateLimit: 50 * time.Millisecond,
	}
}

// Start launches the watcher goroutine. Safe to call multiple times.
// Returns immediately after starting the goroutine (non-blocking).
func (s *WatcherService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.stopChan = make(chan struct{})
	go s.run()
}

// Stop signals the watcher goroutine to exit. Safe to call multiple times.
// The goroutine may take up to one HTTP timeout period to fully exit.
func (s *WatcherService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		close(s.stopChan)
		s.started = false
	}
}

// run is the main watcher event loop. It runs in a goroutine launched by Start().
func (s *WatcherService) run() {
	// Determine starting sequence from checkpoint
	checkpoint, err := s.couchRepo.GetCheckpoint()
	since := "0"
	if err != nil {
		log.Printf("watcher: failed to get checkpoint: %v", err)
	} else if checkpoint != nil {
		since = checkpoint.LastSeq
	}

	for {
		// Check for stop signal before each iteration
		select {
		case <-s.stopChan:
			return
		default:
		}

		resp, err := s.couchRepo.GetChanges("registros", since)
		if err != nil {
			log.Printf("watcher: failed to get changes: %v", err)
			select {
			case <-s.stopChan:
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		for i, result := range resp.Results {
			// Check for stop signal between results
			select {
			case <-s.stopChan:
				return
			default:
			}

			// Skip deleted docs (D-39 safeguard)
			if result.Deleted {
				continue
			}

			// Skip if doc body is missing (CouchDB compaction)
			if len(result.Doc) == 0 {
				continue
			}

			// Parse type field — only process "registro" docs (D-39)
			var docType struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(result.Doc, &docType); err != nil {
				continue
			}
			if docType.Type != "registro" {
				continue
			}

			// Rate limit between consecutive NLP calls (D-42)
			if i > 0 {
				time.Sleep(s.rateLimit)
			}

			// Parse registro fields for the NLP request
			var registro struct {
				UserSub     string `json:"userSub"`
				Sensacoes   string `json:"sensacoes"`
				Contexto    string `json:"contexto"`
				Pensamentos string `json:"pensamentos"`
				DataHora    string `json:"dataHora"`
			}
			if err := json.Unmarshal(result.Doc, &registro); err != nil {
				log.Printf("watcher: failed to parse registro doc %s: %v", result.ID, err)
				continue
			}

			// Build NLP request
			req := &nlp.AnalyzeRequest{
				RegistroID:  result.ID,
				Sensacoes:   registro.Sensacoes,
				Contexto:    registro.Contexto,
				Pensamentos: registro.Pensamentos,
				DataHora:    registro.DataHora,
			}

			// Retry loop with exponential backoff (D-53)
			var analysis *nlp.AnalyzeResponse
			var analyzeErr error

			for attempt := 0; attempt < 3; attempt++ {
				ctx := context.Background()
				analysis, analyzeErr = s.client.Analyze(ctx, req)
				if analyzeErr == nil {
					break
				}
				log.Printf("watcher: analyze attempt %d/3 failed for %s: %v", attempt+1, result.ID, analyzeErr)
				if attempt < 2 {
					select {
					case <-s.stopChan:
						return
					case <-time.After(s.backoff[attempt]):
					}
				}
			}

			if analyzeErr != nil {
				// D-54: silent failure — checkpoint advances regardless
				log.Printf("watcher: NLP analysis failed for %s after 3 retries: %v", result.ID, analyzeErr)
				continue
			}

			// Save analysis document
			analiseDoc := &repository.AnaliseDoc{
				RegistroID:       result.ID,
				UserSub:          registro.UserSub,
				EmotionPrincipal: analysis.EmotionPrincipal,
				Emotions:         convertEmotionScores(analysis.Emotions),
				Scores:           analysis.Scores,
				Intensidade:      analysis.Intensidade,
				ModeloVersao:     analysis.ModeloVersao,
			}
			if err := s.couchRepo.SaveAnalise(analiseDoc); err != nil {
				log.Printf("watcher: failed to save analise for %s: %v", result.ID, err)
			}
		}

		// Save checkpoint after batch is fully processed (D-40)
		if resp.LastSeq != "" {
			if err := s.couchRepo.SaveCheckpoint(resp.LastSeq); err != nil {
				log.Printf("watcher: failed to save checkpoint: %v", err)
			}
			since = resp.LastSeq
		}
	}
}

// convertEmotionScores converts nlp.EmotionScore to repository.EmotionScore.
// Both types have identical fields but are defined in different packages.
func convertEmotionScores(src []nlp.EmotionScore) []repository.EmotionScore {
	dst := make([]repository.EmotionScore, len(src))
	for i, e := range src {
		dst[i] = repository.EmotionScore{Emotion: e.Emotion, Score: e.Score}
	}
	return dst
}
