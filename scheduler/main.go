package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	couchDBURL  string
	couchDBUser string
	couchDBPass string
	apiURL      string
	interval    time.Duration
}

type pushPrefsDoc struct {
	ID       string   `json:"_id"`
	Rev      string   `json:"_rev,omitempty"`
	Type     string   `json:"type"`
	UserSub  string   `json:"userSub"`
	Enabled  bool     `json:"enabled"`
	Times    []string `json:"times"`
	Timezone string   `json:"timezone"`
}

type mangoQuery struct {
	Selector map[string]interface{} `json:"selector"`
	Limit    int                    `json:"limit,omitempty"`
}

type mangoResponse struct {
	Docs []json.RawMessage `json:"docs"`
}

func loadConfig() config {
	getEnv := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	interval := 1 * time.Minute
	if v := os.Getenv("SCHEDULER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	return config{
		couchDBURL:  getEnv("COUCHDB_URL", "http://couchdb:5984"),
		couchDBUser: getEnv("COUCHDB_USER", "admin"),
		couchDBPass: getEnv("COUCHDB_PASSWORD", ""),
		apiURL:      getEnv("API_URL", "http://api:8080"),
		interval:    interval,
	}
}

func getPushPrefs(cfg config) ([]pushPrefsDoc, error) {
	query := mangoQuery{
		Selector: map[string]interface{}{
			"type":    "push_prefs",
			"enabled": true,
		},
		Limit: 1000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/preferencias/_find", cfg.couchDBURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(cfg.couchDBUser, cfg.couchDBPass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couchdb find: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("couchdb status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	prefs := make([]pushPrefsDoc, 0, len(mResp.Docs))
	for _, raw := range mResp.Docs {
		var p pushPrefsDoc
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func shouldNotify(pref pushPrefsDoc, now time.Time) bool {
	loc := time.UTC
	if pref.Timezone != "" {
		if l, err := time.LoadLocation(pref.Timezone); err == nil {
			loc = l
		}
	}

	localTime := now.In(loc)
	currentHourMin := fmt.Sprintf("%02d:%02d", localTime.Hour(), localTime.Minute())

	for _, t := range pref.Times {
		if t == currentHourMin {
			return true
		}
	}
	return false
}

func sendPush(cfg config, userSub string) error {
	body, _ := json.Marshal(map[string]string{"userId": userSub})

	url := fmt.Sprintf("%s/api/push/send", cfg.apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("api status: %d", resp.StatusCode)
	}
	return nil
}

func tick(cfg config) {
	prefs, err := getPushPrefs(cfg)
	if err != nil {
		log.Printf("failed to fetch push prefs: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, pref := range prefs {
		if !shouldNotify(pref, now) {
			continue
		}
		log.Printf("sending push to user %s (tz=%s, times=%v)", pref.UserSub, pref.Timezone, pref.Times)
		if err := sendPush(cfg, pref.UserSub); err != nil {
			log.Printf("failed to send push to %s: %v", pref.UserSub, err)
		}
	}
}

func main() {
	cfg := loadConfig()
	log.Printf("starting scheduler — interval=%s couchdb=%s api=%s", cfg.interval, cfg.couchDBURL, cfg.apiURL)

	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	tick(cfg)

	for range ticker.C {
		tick(cfg)
	}
}
