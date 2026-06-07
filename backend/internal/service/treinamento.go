package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/nlp"
	"github.com/edson/kanso-api/internal/repository"
)

// TreinamentoService manages training data change detection, model retraining,
// and lazy re-analysis of outdated registros.
type TreinamentoService struct {
	repo      *repository.CouchDB
	nlpClient nlp.Analyzer
	cfg       *config.Config
	httpClient *http.Client
}

// NewTreinamentoService creates a new training management service.
func NewTreinamentoService(repo *repository.CouchDB, nlpClient nlp.Analyzer, cfg *config.Config) *TreinamentoService {
	return &TreinamentoService{
		repo:      repo,
		nlpClient: nlpClient,
		cfg:       cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CheckAndTrain checks whether the training data has changed and triggers
// model retraining via the NLP service HTTP endpoint if needed.
func (s *TreinamentoService) CheckAndTrain(ctx context.Context) error {
	changed, hash, err := s.repo.HasTrainingChanged()
	if err != nil {
		return fmt.Errorf("check training changed: %w", err)
	}

	if !changed {
		slog.Info("treinamento: no changes detected, skipping training")
		return nil
	}

	slog.Info("treinamento: training data changed, triggering NLP training")

	// Call NLP service /train endpoint
	url := s.cfg.NLPHTTPAddr + "/train"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte{}))
	if err != nil {
		return fmt.Errorf("create train request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call nlp train: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read nlp train response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("nlp train returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get new model version
	var trainResp struct {
		Status       string `json:"status"`
		ModelVersion string `json:"model_version"`
		TrainedCount int    `json:"trained_count"`
	}
	if err := json.Unmarshal(body, &trainResp); err != nil {
		return fmt.Errorf("parse nlp train response: %w", err)
	}

	// Update checkpoint with new hash and version
	if err := s.repo.SaveTrainingCheckpoint(hash, trainResp.ModelVersion); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}

	slog.Info("treinamento: model retrained successfully",
		"model_version", trainResp.ModelVersion,
		"trained_count", trainResp.TrainedCount,
	)

	return nil
}

// GetCurrentModelVersion returns the current model version from the checkpoint.
// Returns an empty string if no checkpoint exists yet.
func (s *TreinamentoService) GetCurrentModelVersion() string {
	cp, err := s.repo.GetTrainingCheckpoint()
	if err != nil {
		slog.Warn("treinamento: failed to get checkpoint for version", "error", err)
		return ""
	}
	if cp == nil {
		return ""
	}
	return cp.ModelVersion
}

// ReanalyzeRegistros re-analyses all registros whose modeloVersao does not match
// the current model version. Rate-limited to 50ms between calls.
func (s *TreinamentoService) ReanalyzeRegistros(ctx context.Context) error {
	currentVersion := s.GetCurrentModelVersion()
	if currentVersion == "" {
		slog.Warn("treinamento: no model version set, skipping re-analysis")
		return nil
	}

	// Fetch all analysis docs from the sentimentos DB
	// But we don't have all registro IDs here. Let me think...

	// Approach: Get all analise_nlp docs from sentimentos DB via _find
	// We'll use the same selector pattern as ListReportJobsByUser but for sentimentos DB.

	// For now, implement the core logic: scan analise docs and re-analyze outdated ones.
	// We'll fetch registros that have analise_nlp with outdated modeloVersao.

	selector := map[string]interface{}{
		"type": "analise_nlp",
	}
	query := map[string]interface{}{
		"selector": selector,
		"limit":    10000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/%s/_find", s.cfg.CouchDBURL, "sentimentos")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(s.cfg.CouchDBUser, s.cfg.CouchDBPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("find analises: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("find analises status: %d", resp.StatusCode)
	}

	var mResp struct {
		Docs []json.RawMessage `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	reanalyzed := 0
	total := len(mResp.Docs)

	for _, raw := range mResp.Docs {
		var doc struct {
			ID             string `json:"_id"`
			Rev            string `json:"_rev"`
			RegistroID     string `json:"registroId"`
			UserSub        string `json:"userSub"`
			Sensacoes      string `json:"sensacoes"`
			Contexto       string `json:"contexto"`
			Pensamentos    string `json:"pensamentos"`
			ModeloVersao   string `json:"modeloVersao"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			slog.Warn("treinamento: failed to parse analise doc", "error", err)
			continue
		}

		if doc.ModeloVersao == currentVersion {
			continue
		}

		// Rate limit
		time.Sleep(50 * time.Millisecond)

		// Read the actual registro to get text content for analysis
		// For simplicity, we try to get the registro data
		regURL := fmt.Sprintf("%s/%s/%s", s.cfg.CouchDBURL, "registros", doc.RegistroID)
		regReq, err := http.NewRequestWithContext(ctx, "GET", regURL, nil)
		if err != nil {
			slog.Warn("treinamento: failed to create registro request", "id", doc.RegistroID, "error", err)
			continue
		}
		regReq.SetBasicAuth(s.cfg.CouchDBUser, s.cfg.CouchDBPass)

		regResp, err := s.httpClient.Do(regReq)
		if err != nil {
			slog.Warn("treinamento: failed to get registro", "id", doc.RegistroID, "error", err)
			continue
		}

		var registroData struct {
			Sensacoes   string `json:"sensacoes"`
			Contexto    string `json:"contexto"`
			Pensamentos string `json:"pensamentos"`
			DataHora    string `json:"dataHora"`
		}
		if err := json.NewDecoder(regResp.Body).Decode(&registroData); err != nil {
			regResp.Body.Close()
			slog.Warn("treinamento: failed to decode registro", "id", doc.RegistroID, "error", err)
			continue
		}
		regResp.Body.Close()

		// Send to gRPC NLP for analysis
		nlpReq := &nlp.AnalyzeRequest{
			RegistroID:  doc.RegistroID,
			Sensacoes:   registroData.Sensacoes,
			Contexto:    registroData.Contexto,
			Pensamentos: registroData.Pensamentos,
			DataHora:    registroData.DataHora,
		}

		result, err := s.nlpClient.Analyze(ctx, nlpReq)
		if err != nil {
			slog.Warn("treinamento: re-analysis failed", "id", doc.RegistroID, "error", err)
			continue
		}

		// Update the analysis doc with new version
		analiseDoc := &repository.AnaliseDoc{
			ID:               doc.ID,
			Rev:              doc.Rev,
			RegistroID:       doc.RegistroID,
			UserSub:          doc.UserSub,
			EmotionPrincipal: result.EmotionPrincipal,
			Emotions:         convertEmotionScores(result.Emotions),
			Scores:           result.Scores,
			Intensidade:      result.Intensidade,
			ModeloVersao:     result.ModeloVersao,
		}
		if err := s.repo.SaveAnalise(analiseDoc); err != nil {
			slog.Warn("treinamento: failed to save re-analysis", "id", doc.RegistroID, "error", err)
			continue
		}

		reanalyzed++
		slog.Info("treinamento: re-analyzed registro",
			"id", doc.RegistroID,
			"emotion", result.EmotionPrincipal,
			"version", result.ModeloVersao,
		)
	}

	slog.Info("treinamento: re-analysis complete",
		"reanalyzed", reanalyzed,
		"total", total,
	)

	return nil
}

// StartScheduler starts a periodic ticker that calls CheckAndTrain.
// interval is the time between checks (default 7 days via env TRAIN_INTERVAL).
func (s *TreinamentoService) StartScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		slog.Info("treinamento: scheduler started", "interval", interval.String())
		for {
			select {
			case <-ticker.C:
				slog.Info("treinamento: scheduler tick — checking for training data changes")
				if err := s.CheckAndTrain(ctx); err != nil {
					slog.Error("treinamento: scheduler training failed", "error", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("treinamento: scheduler stopped")
				return
			}
		}
	}()
}
