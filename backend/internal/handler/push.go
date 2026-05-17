package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/service"
)

type PushHandler struct {
	pushSvc *service.PushService
}

func NewPushHandler(pushSvc *service.PushService) *PushHandler {
	return &PushHandler{pushSvc: pushSvc}
}

type subscribeRequest struct {
	FCMToken string `json:"fcmToken"`
	Timezone string `json:"timezone"`
}

func (h *PushHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.pushSvc.Subscribe(sub, req.FCMToken, req.Timezone); err != nil {
		log.Printf("failed to subscribe: %v", err)
		http.Error(w, `{"error":"failed to subscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *PushHandler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	prefs, err := h.pushSvc.GetPreferences(sub)
	if err != nil {
		log.Printf("failed to get preferences: %v", err)
		http.Error(w, `{"error":"failed to get preferences"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

type updatePreferencesRequest struct {
	Enabled bool     `json:"enabled"`
	Times   []string `json:"times"`
}

func (h *PushHandler) HandleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.pushSvc.UpdatePreferences(sub, req.Enabled, req.Times); err != nil {
		log.Printf("failed to update preferences: %v", err)
		http.Error(w, `{"error":"failed to update preferences"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type sendRequest struct {
	UserID string `json:"userId"`
}

func (h *PushHandler) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.pushSvc.Send(req.UserID); err != nil {
		log.Printf("failed to send push: %v", err)
		http.Error(w, `{"error":"failed to send push"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
