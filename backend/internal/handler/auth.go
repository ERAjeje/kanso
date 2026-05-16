package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authSvc  *service.AuthService
	jwtSecret string
}

type googleLoginRequest struct {
	IDToken string `json:"idToken"`
}

func NewAuth(authSvc *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, jwtSecret: jwtSecret}
}

func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req googleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	claims, err := h.authSvc.VerifyGoogleToken(r.Context(), req.IDToken)
	if err != nil {
		http.Error(w, `{"error":"invalid Google token"}`, http.StatusUnauthorized)
		return
	}

	result, err := h.authSvc.SignJWT(claims)
	if err != nil {
		http.Error(w, `{"error":"failed to sign JWT"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 3600,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jwt":       result.JWT,
		"expiresIn": result.ExpiresIn,
	})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sub":   claims["sub"],
		"email": claims["email"],
		"name":  claims["name"],
	})
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, `{"error":"no refresh token"}`, http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, `{"error":"invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	sub, _ := claims["sub"].(string)
	result, err := h.authSvc.SignJWT(&service.GoogleClaims{
		Sub: sub,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to sign JWT"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"jwt": result.JWT})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}
