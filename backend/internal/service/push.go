package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/edson/kanso-api/internal/repository"
)

type PushService struct {
	CouchRepo    *repository.CouchDB
	FCMServerKey string
	FCMURL       string
	HTTPClient   *http.Client
}

func NewPushService(couchRepo *repository.CouchDB, fcmServerKey string) *PushService {
	return &PushService{
		CouchRepo:    couchRepo,
		FCMServerKey: fcmServerKey,
		FCMURL:       "https://fcm.googleapis.com/fcm/send",
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *PushService) Subscribe(sub, fcmToken, timezone string) error {
	existing, err := 	s.CouchRepo.GetPushPrefs(sub)
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

	return 	s.CouchRepo.SavePushPrefs(doc)
}

func (s *PushService) GetPreferences(sub string) (*repository.PushPrefsDoc, error) {
	prefs, err := 	s.CouchRepo.GetPushPrefs(sub)
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
	existing, err := 	s.CouchRepo.GetPushPrefs(sub)
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

	return 	s.CouchRepo.SavePushPrefs(doc)
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
	prefs, err := 	s.CouchRepo.GetPushPrefs(sub)
	if err != nil {
		return fmt.Errorf("get prefs: %w", err)
	}
	if prefs == nil || prefs.FCMToken == "" {
		return fmt.Errorf("no FCM token for user %s", sub)
	}

	msg := fcmMessage{
		To: prefs.FCMToken,
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
