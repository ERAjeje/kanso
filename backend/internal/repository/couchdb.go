package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DBRegistros    = "registros"
	DBSentimentos  = "sentimentos"
	DBPreferencias = "preferencias"
	DBRelatorios   = "relatorios"
	DBUsuarios     = "usuarios"
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

type PushPrefsDoc struct {
	ID        string   `json:"_id,omitempty"`
	Rev       string   `json:"_rev,omitempty"`
	Type      string   `json:"type"`
	UserSub   string   `json:"userSub"`
	Enabled   bool     `json:"enabled"`
	Times     []string `json:"times"`
	Timezone  string   `json:"timezone"`
	FCMToken  string   `json:"fcmToken"`
	CreatedAt string   `json:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
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

	url := fmt.Sprintf("%s/"+DBUsuarios+"/%s", c.baseURL, doc.ID)
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

// --- Push Preferences Repository Methods ---

func (c *CouchDB) GetPushPrefs(sub string) (*PushPrefsDoc, error) {
	url := fmt.Sprintf("%s/"+DBPreferencias+"/push_prefs:%s", c.baseURL, sub)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get status: %d", resp.StatusCode)
	}

	var doc PushPrefsDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &doc, nil
}

func (c *CouchDB) SavePushPrefs(doc *PushPrefsDoc) error {
	doc.Type = "push_prefs"
	if doc.ID == "" {
		doc.ID = "push_prefs:" + doc.UserSub
	}
	if doc.CreatedAt == "" {
		doc.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return c.putDoc(DBPreferencias, doc.ID, doc)
}

func (c *CouchDB) GetAllPushPrefs() ([]PushPrefsDoc, error) {
	selector := map[string]interface{}{
		"type":    "push_prefs",
		"enabled": true,
	}
	query := mangoQuery{
		Selector: selector,
		Limit:    1000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/"+DBPreferencias+"/_find", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("find status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	prefs := make([]PushPrefsDoc, 0, len(mResp.Docs))
	for _, raw := range mResp.Docs {
		var p PushPrefsDoc
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

type ReportJobStatus string

const (
	StatusPending   ReportJobStatus = "pending"
	StatusProcessing ReportJobStatus = "processing"
	StatusDone      ReportJobStatus = "done"
	StatusFailed    ReportJobStatus = "failed"
)

type ReportJobDoc struct {
	ID            string          `json:"_id,omitempty"`
	Rev           string          `json:"_rev,omitempty"`
	Type          string          `json:"type"`
	UserSub       string          `json:"userSub"`
	Status        ReportJobStatus `json:"status"`
	PeriodStart   string          `json:"periodStart,omitempty"`
	PeriodEnd     string          `json:"periodEnd,omitempty"`
	CreatedAt     string          `json:"createdAt,omitempty"`
	CompletedAt   string          `json:"completedAt,omitempty"`
	FileName      string          `json:"fileName,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// --- Report Job Repository Methods ---

type couchDBPutResponse struct {
	ID  string `json:"id"`
	Rev string `json:"rev"`
	OK  bool   `json:"ok"`
}

type mangoQuery struct {
	Selector   map[string]interface{} `json:"selector"`
	Sort       []map[string]string    `json:"sort,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Execution  bool                   `json:"execution_stats,omitempty"`
}

type mangoResponse struct {
	Docs []json.RawMessage `json:"docs"`
}

func (c *CouchDB) putDoc(db, id string, doc interface{}) error {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, db, id)
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("put: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("put status: %d", resp.StatusCode)
	}
	return nil
}

func (c *CouchDB) CreateReportJob(job *ReportJobDoc) error {
	job.Type = "relatorio"
	if job.Status == "" {
		job.Status = StatusPending
	}
	if job.CreatedAt == "" {
		job.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return c.putDoc(DBRelatorios, job.ID, job)
}

func (c *CouchDB) GetReportJob(id string) (*ReportJobDoc, error) {
	url := fmt.Sprintf("%s/"+DBRelatorios+"/%s", c.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get status: %d", resp.StatusCode)
	}

	var doc ReportJobDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &doc, nil
}

func (c *CouchDB) UpdateReportJobStatus(id, rev string, status ReportJobStatus, fileName, errorMsg string) error {
	url := fmt.Sprintf("%s/"+DBRelatorios+"/%s", c.baseURL, id)

	// First get current doc to update
	getReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("new get request: %w", err)
	}
	getReq.SetBasicAuth(c.adminUser, c.adminPass)
	getResp, err := c.httpClient.Do(getReq)
	if err != nil {
		return fmt.Errorf("get before update: %w", err)
	}
	defer getResp.Body.Close()

	var current ReportJobDoc
	if err := json.NewDecoder(getResp.Body).Decode(&current); err != nil {
		return fmt.Errorf("decode current: %w", err)
	}

	current.Status = status
	current.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	current.FileName = fileName
	current.ErrorMessage = errorMsg

	body, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("put: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("put status: %d", resp.StatusCode)
	}
	return nil
}

func (c *CouchDB) ListReportJobsByUser(sub string) ([]ReportJobDoc, error) {
	selector := map[string]interface{}{
		"type":    "relatorio",
		"userSub": sub,
	}
	query := mangoQuery{
		Selector: selector,
		Sort:     []map[string]string{{"createdAt": "desc"}},
		Limit:    50,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/"+DBRelatorios+"/_find", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("find status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	jobs := make([]ReportJobDoc, 0, len(mResp.Docs))
	for _, raw := range mResp.Docs {
		var job ReportJobDoc
		if err := json.Unmarshal(raw, &job); err != nil {
			return nil, fmt.Errorf("unmarshal job: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (c *CouchDB) GetLastCompletedReport(sub string) (*ReportJobDoc, error) {
	selector := map[string]interface{}{
		"type":    "relatorio",
		"userSub": sub,
		"status":  StatusDone,
	}
	query := mangoQuery{
		Selector: selector,
		Sort:     []map[string]string{{"createdAt": "desc"}},
		Limit:    1,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/"+DBRelatorios+"/_find", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("find status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(mResp.Docs) == 0 {
		return nil, nil
	}

	var job ReportJobDoc
	if err := json.Unmarshal(mResp.Docs[0], &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}
	return &job, nil
}

func (c *CouchDB) ReportJobExists(id string) (bool, error) {
	job, err := c.GetReportJob(id)
	if err != nil {
		return false, err
	}
	return job != nil, nil
}

func (c *CouchDB) GetUser(id string) (*UserDoc, error) {
	url := fmt.Sprintf("%s/"+DBUsuarios+"/%s", c.baseURL, id)
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

// --- NLP Watcher Types ---

type ChangesResult struct {
	Seq     string              `json:"seq"`
	ID      string              `json:"id"`
	Changes []map[string]string `json:"changes"`
	Doc     json.RawMessage     `json:"doc,omitempty"`
	Deleted bool                `json:"_deleted,omitempty"`
}

type ChangesResponse struct {
	Results  []ChangesResult `json:"results"`
	LastSeq  string          `json:"last_seq"`
}

type CheckpointDoc struct {
	ID        string `json:"_id,omitempty"`
	Rev       string `json:"_rev,omitempty"`
	Type      string `json:"type"`
	Watcher   string `json:"watcher"`
	LastSeq   string `json:"last_seq"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type EmotionScore struct {
	Emotion string  `json:"emotion"`
	Score   float32 `json:"score"`
}

type AnaliseDoc struct {
	ID               string             `json:"_id,omitempty"`
	Rev              string             `json:"_rev,omitempty"`
	Type             string             `json:"type"`
	UserSub          string             `json:"userSub"`
	RegistroID       string             `json:"registroId"`
	EmotionPrincipal string             `json:"emotionPrincipal"`
	Emotions         []EmotionScore     `json:"emotions"`
	Scores           map[string]float32 `json:"scores"`
	Intensidade      float32            `json:"intensidade"`
	ModeloVersao     string             `json:"modeloVersao"`
	AnalisadoEm      string             `json:"analisadoEm,omitempty"`
}

type PeriodRegistroDoc struct {
	ID          string `json:"_id,omitempty"`
	Rev         string `json:"_rev,omitempty"`
	Type        string `json:"type"`
	UserSub     string `json:"userId"`
	DataHora    string `json:"dataHora"`
	Sensacoes   string `json:"sensacoes"`
	Sentimento  string `json:"sentimentoNome"`
	Contexto    string `json:"contexto"`
	Pensamentos string `json:"pensamentos"`
}

// --- NLP Watcher Repository Methods ---

// GetChanges fetches the _changes feed for a database using long-poll.
// `since` is the last sequence ID (use "0" for initial/backfill).
// Uses a 30s HTTP timeout (25s long-poll + 5s buffer).
// Always uses include_docs=true so we can filter by doc.type.
func (c *CouchDB) GetChanges(db, since string) (*ChangesResponse, error) {
	url := fmt.Sprintf("%s/%s/_changes?since=%s&timeout=25000&feed=longpoll&include_docs=true", c.baseURL, db, since)

	// Use longer timeout for long-poll (25s poll + buffer)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get changes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get changes status: %d", resp.StatusCode)
	}

	var changes ChangesResponse
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, fmt.Errorf("decode changes: %w", err)
	}
	return &changes, nil
}

// GetCheckpoint reads the NLP watcher checkpoint document.
// Returns ("", nil) if no checkpoint exists yet.
func (c *CouchDB) GetCheckpoint() (*CheckpointDoc, error) {
	url := fmt.Sprintf("%s/"+DBRegistros+"/checkpoint:nlp_watcher", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get checkpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get checkpoint status: %d", resp.StatusCode)
	}

	var doc CheckpointDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode checkpoint: %w", err)
	}
	return &doc, nil
}

// SaveCheckpoint creates or updates the NLP watcher checkpoint document.
// It does a GET-then-PUT to handle _rev for updates.
func (c *CouchDB) SaveCheckpoint(seq string) error {
	// Try to get existing checkpoint for _rev
	existing, err := c.GetCheckpoint()
	if err != nil {
		return fmt.Errorf("get existing checkpoint before save: %w", err)
	}

	doc := &CheckpointDoc{
		Type:      "checkpoint",
		Watcher:   "nlp",
		LastSeq:   seq,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if existing != nil {
		doc.ID = existing.ID
		doc.Rev = existing.Rev
	} else {
		doc.ID = "checkpoint:nlp_watcher"
	}

	return c.putDoc(DBRegistros, doc.ID, doc)
}

// SaveAnalise writes an analysis document to the sentimentos database.
// ID format: "analise:{registroId}"
func (c *CouchDB) SaveAnalise(doc *AnaliseDoc) error {
	doc.Type = "analise_nlp"
	if doc.ID == "" {
		doc.ID = "analise:" + doc.RegistroID
	}
	if doc.AnalisadoEm == "" {
		doc.AnalisadoEm = time.Now().UTC().Format(time.RFC3339)
	}
	return c.putDoc(DBSentimentos, doc.ID, doc)
}

// FindRegistrosByPeriod queries registros for a user between two timestamps.
// Uses _find with a selector on type, userSub, and dataHora range.
// LIMITATION: No CouchDB index on dataHora — acceptable for user-scale datasets (<5000 docs).
func (c *CouchDB) FindRegistrosByPeriod(userSub, periodStart, periodEnd string) ([]PeriodRegistroDoc, error) {
	selector := map[string]interface{}{
		"type":    "registro",
		"userId": userSub,
		"dataHora": map[string]interface{}{
			"$gte": periodStart,
			"$lte": periodEnd,
		},
	}
	query := mangoQuery{
		Selector: selector,
		Sort:     []map[string]string{{"dataHora": "desc"}},
		Limit:    1000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/"+DBRegistros+"/_find", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("find registros: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("find registros status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	docs := make([]PeriodRegistroDoc, 0, len(mResp.Docs))
	for _, raw := range mResp.Docs {
		var doc PeriodRegistroDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("unmarshal registro: %w", err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// FindAnaliseByRegistroIds queries analysis documents by their associated registro IDs.
// Accepts a slice of registro IDs and returns matching analise_nlp docs.
func (c *CouchDB) FindAnaliseByRegistroIds(ids []string) ([]AnaliseDoc, error) {
	if len(ids) == 0 {
		return []AnaliseDoc{}, nil
	}

	selector := map[string]interface{}{
		"type": "analise_nlp",
		"registroId": map[string]interface{}{
			"$in": ids,
		},
	}
	query := mangoQuery{
		Selector: selector,
		Limit:    1000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/"+DBSentimentos+"/_find", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.adminUser, c.adminPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("find analise: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("find analise status: %d", resp.StatusCode)
	}

	var mResp mangoResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	docs := make([]AnaliseDoc, 0, len(mResp.Docs))
	for _, raw := range mResp.Docs {
		var doc AnaliseDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("unmarshal analise: %w", err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
