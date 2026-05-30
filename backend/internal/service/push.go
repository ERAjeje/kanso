package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/edson/kanso-api/internal/repository"
)

type PushService struct {
	CouchRepo            *repository.CouchDB
	FCMServerKey         string
	FCMURL               string
	FCMProjectID         string
	FCMServiceAccountB64 string
	HTTPClient           *http.Client
}

func NewPushService(couchRepo *repository.CouchDB, fcmServerKey, fcmProjectID, fcmSvcAcctB64 string) *PushService {
	return &PushService{
		CouchRepo:            couchRepo,
		FCMServerKey:         fcmServerKey,
		FCMURL:               "https://fcm.googleapis.com/fcm/send",
		FCMProjectID:         fcmProjectID,
		FCMServiceAccountB64: fcmSvcAcctB64,
		HTTPClient:           &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *PushService) Subscribe(sub, fcmToken, timezone string) error {
	existing, err := s.CouchRepo.GetPushPrefs(sub)
	if err != nil {
		return fmt.Errorf("get existing: %w", err)
	}

	doc := &repository.PushPrefsDoc{
		UserSub:  sub,
		FCMToken: fcmToken,
		Timezone: timezone,
	}

	if existing != nil {
		doc.Rev = existing.Rev
		doc.CreatedAt = existing.CreatedAt
		doc.Enabled = existing.Enabled
		doc.Times = existing.Times
	} else {
		doc.Times = []string{"12:00", "18:00", "23:00"}
		doc.Enabled = true
	}

	return s.CouchRepo.SavePushPrefs(doc)
}

func (s *PushService) GetPreferences(sub string) (*repository.PushPrefsDoc, error) {
	prefs, err := s.CouchRepo.GetPushPrefs(sub)
	if err != nil {
		return nil, fmt.Errorf("get prefs: %w", err)
	}
	if prefs == nil {
		return &repository.PushPrefsDoc{
			UserSub: sub,
			Enabled: true,
			Times:   []string{"12:00", "18:00", "23:00"},
		}, nil
	}
	return prefs, nil
}

func (s *PushService) UpdatePreferences(sub string, enabled bool, times []string) error {
	existing, err := s.CouchRepo.GetPushPrefs(sub)
	if err != nil {
		return fmt.Errorf("get existing: %w", err)
	}

	doc := &repository.PushPrefsDoc{
		UserSub: sub,
		Enabled: enabled,
		Times:   times,
	}

	if existing != nil {
		doc.Rev = existing.Rev
		doc.CreatedAt = existing.CreatedAt
		doc.FCMToken = existing.FCMToken
		doc.Timezone = existing.Timezone
	} else {
		doc.Times = []string{"12:00", "18:00", "23:00"}
		doc.Enabled = true
	}

	return s.CouchRepo.SavePushPrefs(doc)
}

type fcmMessage struct {
	To           string          `json:"to"`
	Notification fcmNotification `json:"notification"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *PushService) Send(sub string) error {
	prefs, err := s.CouchRepo.GetPushPrefs(sub)
	if err != nil {
		return fmt.Errorf("get prefs: %w", err)
	}
	if prefs == nil || prefs.FCMToken == "" {
		return fmt.Errorf("no FCM token for user %s", sub)
	}

	if s.FCMServiceAccountB64 != "" && s.FCMProjectID != "" {
		return s.sendV1(prefs.FCMToken)
	}

	return s.sendLegacy(prefs.FCMToken)
}

func (s *PushService) sendV1(token string) error {
	ctx := context.Background()

	credsJSON, err := base64.StdEncoding.DecodeString(s.FCMServiceAccountB64)
	if err != nil {
		return fmt.Errorf("decode service account: %w", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, credsJSON, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("credentials from json: %w", err)
	}

	client := oauth2.NewClient(ctx, creds.TokenSource)

	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"token": token,
			"notification": map[string]string{
				"title": "Kanso",
				"body":  "Como você está se sentindo agora?",
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.FCMProjectID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fcm v1 request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("fcm v1 status: %d", resp.StatusCode)
	}

	return nil
}

func (s *PushService) sendLegacy(token string) error {
	msg := fcmMessage{
		To: token,
		Notification: fcmNotification{
			Title: "Kanso",
			Body:  "Como você está se sentindo agora?",
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", s.FCMURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "key="+s.FCMServerKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("fcm status: %d", resp.StatusCode)
	}

	return nil
}
