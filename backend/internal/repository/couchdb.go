package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CouchDB struct {
	baseURL    string
	adminUser  string
	adminPass  string
	httpClient *http.Client
}

type UserDoc struct {
	ID        string `json:"_id,omitempty"`
	Rev       string `json:"_rev,omitempty"`
	Type      string `json:"type"`
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func NewCouchDB(baseURL, adminUser, adminPass string) *CouchDB {
	return &CouchDB{
		baseURL:    baseURL,
		adminUser:  adminUser,
		adminPass:  adminPass,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CouchDB) CreateOrUpdateUser(doc *UserDoc) error {
	doc.Type = "usuario"
	now := time.Now().UTC().Format(time.RFC3339)

	existing, err := c.GetUser(doc.ID)
	if err != nil {
		return fmt.Errorf("couchdb get before put: %w", err)
	}

	if existing != nil {
		doc.Rev = existing.Rev
		doc.CreatedAt = existing.CreatedAt
	} else {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now

	url := fmt.Sprintf("%s/usuarios/%s", c.baseURL, doc.ID)
	body, _ := json.Marshal(doc)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couchdb put: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("couchdb put status: %d", resp.StatusCode)
	}
	return nil
}

func (c *CouchDB) GetUser(id string) (*UserDoc, error) {
	url := fmt.Sprintf("%s/usuarios/%s", c.baseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.adminUser, c.adminPass)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couchdb get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("couchdb get status: %d", resp.StatusCode)
	}

	var doc UserDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("couchdb decode: %w", err)
	}
	return &doc, nil
}
