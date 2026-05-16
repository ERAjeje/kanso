package service

import (
	"context"
	"log"
	"time"

	"github.com/edson/kanso-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
)

type AuthService struct {
	googleClientID string
	jwtSecret      []byte
	couchRepo      *repository.CouchDB
}

type GoogleClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type AuthResult struct {
	JWT          string `json:"jwt"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	Sub          string `json:"sub"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}

func NewAuth(googleClientID string, jwtSecret string, couchRepo *repository.CouchDB) *AuthService {
	return &AuthService{
		googleClientID: googleClientID,
		jwtSecret:      []byte(jwtSecret),
		couchRepo:      couchRepo,
	}
}

func (s *AuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, idToken, s.googleClientID)
	if err != nil {
		return nil, err
	}
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	return &GoogleClaims{
		Sub:   payload.Subject,
		Email: email,
		Name:  name,
	}, nil
}

func (s *AuthService) SignJWT(claims *GoogleClaims) (*AuthResult, error) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	jwtClaims := jwt.MapClaims{
		"sub":   claims.Sub,
		"email": claims.Email,
		"name":  claims.Name,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedJWT, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"sub": claims.Sub,
		"iat": now.Unix(),
		"exp": now.Add(30 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if claims.Sub != "" && s.couchRepo != nil {
		if err := s.couchRepo.CreateOrUpdateUser(&repository.UserDoc{
			ID:    claims.Sub,
			Sub:   claims.Sub,
			Email: claims.Email,
			Name:  claims.Name,
		}); err != nil {
			log.Printf("Warning: failed to save user: %v", err)
		}
	}

	return &AuthResult{
		JWT:          signedJWT,
		RefreshToken: signedRefresh,
		ExpiresIn:    3600,
		Sub:          claims.Sub,
		Email:        claims.Email,
		Name:         claims.Name,
	}, nil
}
