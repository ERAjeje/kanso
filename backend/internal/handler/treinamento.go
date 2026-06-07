package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Trainer defines the interface for training operations used by TreinamentoHandler.
// This allows test mocking without depending on the service package.
type Trainer interface {
	CheckAndTrain(ctx context.Context) error
	GetCurrentModelVersion() string
	ReanalyzeRegistros(ctx context.Context) error
}

type TreinamentoHandler struct {
	trainer Trainer
}

func NewTreinamentoHandler(trainer Trainer) *TreinamentoHandler {
	return &TreinamentoHandler{trainer: trainer}
}

// HandleTrain checks for training data changes and triggers model retraining.
// POST /api/train
func (h *TreinamentoHandler) HandleTrain(w http.ResponseWriter, r *http.Request) {
	err := h.trainer.CheckAndTrain(r.Context())
	if err != nil {
		slog.Warn("treinamento: train failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"trained": false,
			"reason":  err.Error(),
		})
		return
	}

	version := h.trainer.GetCurrentModelVersion()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trained":      true,
		"modelVersion": version,
	})
}

// HandleTrainStatus returns the current training status.
// GET /api/train/status
func (h *TreinamentoHandler) HandleTrainStatus(w http.ResponseWriter, r *http.Request) {
	version := h.trainer.GetCurrentModelVersion()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"modelVersion": version,
		"trained":      version != "",
	})
}

// HandleReanalyze triggers lazy re-analysis of registros with outdated model versions.
// POST /api/reanalyze
func (h *TreinamentoHandler) HandleReanalyze(w http.ResponseWriter, r *http.Request) {
	err := h.trainer.ReanalyzeRegistros(r.Context())
	if err != nil {
		slog.Warn("treinamento: reanalyze failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reanalyzed": true,
	})
}
