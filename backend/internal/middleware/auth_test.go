package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTRequired_ValidToken(t *testing.T) {
	secret := []byte("test-secret-key-1234567890")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "user123",
		"email": "user@test.com",
		"name":  "Test User",
	})
	tokenStr, _ := token.SignedString(secret)

	handler := JWTRequired(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(UserContextKey).(jwt.MapClaims)
		if claims["sub"] != "user123" {
			t.Error("expected sub=user123")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestJWTRequired_NoAuthHeader(t *testing.T) {
	handler := JWTRequired([]byte("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTRequired_InvalidSignature(t *testing.T) {
	secret := []byte("test-secret-key-1234567890")
	wrongSecret := []byte("different-secret-key!!!!!!")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
	})
	tokenStr, _ := token.SignedString(wrongSecret)

	handler := JWTRequired(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTRequired_WrongAlgorithm(t *testing.T) {
	secret := []byte("test-secret-key-1234567890")

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "user123",
	})
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	handler := JWTRequired(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for alg=none token, got %d", rr.Code)
	}
}

func TestJWTRequired_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-1234567890")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
		"exp": 1000000000,
	})
	tokenStr, _ := token.SignedString(secret)

	handler := JWTRequired(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", rr.Code)
	}
}

func TestJWTRequired_WrongAlgoIntegration(t *testing.T) {
	secret := []byte("test-secret-key-1234567890")

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"sub": "user123",
	})
	tokenStr, _ := token.SignedString(secret)

	handler := JWTRequired(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for HS384 token when only HS256 is accepted, got %d", rr.Code)
	}
}

func TestJWTRequired_MissingBearerPrefix(t *testing.T) {
	handler := JWTRequired([]byte("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "token-without-bearer")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTRequired_SecurityLogging(t *testing.T) {
	var logBuf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn})))

	handler := JWTRequired([]byte("secret"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}

	logOutput := logBuf.String()
	if logOutput == "" {
		t.Error("expected security log output, got empty")
	}
	if !bytes.Contains([]byte(logOutput), []byte("auth")) && !bytes.Contains([]byte(logOutput), []byte("token")) {
		t.Log("warning: log output may not contain expected keywords:", logOutput)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})))
}

func TestUserContextKey_Consistent(t *testing.T) {
	if UserContextKey != "user" {
		t.Errorf("expected UserContextKey to be 'user', got '%s'", UserContextKey)
	}
}

func contextWithClaims(sub string) context.Context {
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": sub + "@test.com",
		"name":  "Test User",
	}
	return context.WithValue(context.Background(), UserContextKey, claims)
}

func TestContextWithClaims(t *testing.T) {
	ctx := contextWithClaims("user123")
	claims := ctx.Value(UserContextKey).(jwt.MapClaims)
	if claims["sub"] != "user123" {
		t.Errorf("expected sub=user123, got %v", claims["sub"])
	}
}
