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

func (h *PushHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	var req struct {
		FCMToken string `json:"fcmToken"`
		Timezone string `json:"timezone"`
	}
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

func (h *PushHandler) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
	}
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
