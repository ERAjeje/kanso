package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/templates"
)

type EmotionSummaryItem struct {
	Emotion string
	Count   int
}

type RegistroReportItem struct {
	Data        string
	Sentimento  string
	Sensacoes   string
	Contexto    string
	Pensamentos string
	Emocoes     []repository.EmotionScore
}

type ReportData struct {
	GeneratedAt    string
	PeriodStart    string
	PeriodEnd      string
	Registros      []RegistroReportItem
	EmotionSummary []EmotionSummaryItem
}

type ReportService struct {
	mu        sync.Mutex
	couchRepo *repository.CouchDB
	gen       *pdf.Generator
	cfg       *config.Config
}

func NewReportService(couchRepo *repository.CouchDB, gen *pdf.Generator, cfg *config.Config) *ReportService {
	return &ReportService{
		couchRepo: couchRepo,
		gen:       gen,
		cfg:       cfg,
	}
}

// RequestReport creates a new report job, generates the PDF synchronously,
// and updates the job status. The mutex ensures only one generation runs at a time.
// periodStart is computed from the last completed report (or empty if none).
// periodEnd is the current time.
func (s *ReportService) RequestReport(ctx context.Context, sub string) (string, error) {
	jobID := fmt.Sprintf("relatorio_%s_%d", sub, time.Now().UnixNano())
	now := time.Now().UTC().Format(time.RFC3339)

	// Compute periodStart from last completed report
	periodStart := ""
	last, err := s.couchRepo.GetLastCompletedReport(sub)
	if err != nil {
		slog.Warn("failed to get last completed report", "error", err)
	}
	if last != nil && last.PeriodEnd != "" {
		periodStart = last.PeriodEnd
	}

	periodEnd := now

	job := &repository.ReportJobDoc{
		ID:          jobID,
		Type:        "relatorio",
		UserSub:     sub,
		Status:      repository.StatusPending,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
	}

	if err := s.couchRepo.CreateReportJob(job); err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}

	// Acquire mutex — one PDF generation at a time
	s.mu.Lock()
	go func() {
		defer s.mu.Unlock()
		s.generatePDF(context.Background(), jobID, sub, periodStart, periodEnd)
	}()

	return jobID, nil
}

func (s *ReportService) generatePDF(ctx context.Context, jobID, sub, periodStart, periodEnd string) {
	// Update status to processing
	if err := s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusProcessing, "", ""); err != nil {
		slog.Warn("failed to update job to processing", "jobID", jobID, "error", err)
	}

	// Fetch registros for the period
	registros, err := s.couchRepo.FindRegistrosByPeriod(sub, periodStart, periodEnd)
	if err != nil {
		errMsg := fmt.Sprintf("fetch registros: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Collect registro IDs to fetch analysis docs
	registroIDs := make([]string, len(registros))
	for i, r := range registros {
		registroIDs[i] = r.ID
	}

	// Fetch analysis docs
	analiseDocs, err := s.couchRepo.FindAnaliseByRegistroIds(registroIDs)
	if err != nil {
		errMsg := fmt.Sprintf("fetch analise docs: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Build analise lookup map
	analiseMap := make(map[string]repository.AnaliseDoc)
	for _, a := range analiseDocs {
		analiseMap[a.RegistroID] = a
	}

	// Build report items with emotion data
	reportItems := make([]RegistroReportItem, 0, len(registros))
	for _, r := range registros {
		item := RegistroReportItem{
			Data:        r.DataHora,
			Sentimento:  r.Sentimento,
			Sensacoes:   r.Sensacoes,
			Contexto:    r.Contexto,
			Pensamentos: r.Pensamentos,
		}
		if a, ok := analiseMap[r.ID]; ok {
			item.Emocoes = a.Emotions
		}
		reportItems = append(reportItems, item)
	}

	// Compute emotion summary (aggregate frequency across all registros in period)
	emotionCounts := make(map[string]int)
	for _, item := range reportItems {
		for _, e := range item.Emocoes {
			emotionCounts[e.Emotion]++
		}
	}

	summary := make([]EmotionSummaryItem, 0, len(emotionCounts))
	for emotion, count := range emotionCounts {
		summary = append(summary, EmotionSummaryItem{Emotion: emotion, Count: count})
	}
	// Sort by count descending
	sort.Slice(summary, func(i, j int) bool {
		return summary[i].Count > summary[j].Count
	})

	// Build template data
	tmplData := ReportData{
		GeneratedAt:    time.Now().UTC().Format("02/01/2006 às 15:04 MST"),
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Registros:      reportItems,
		EmotionSummary: summary,
	}

	var htmlBuf bytes.Buffer
	if err := templates.ReportTemplate.Execute(&htmlBuf, tmplData); err != nil {
		errMsg := fmt.Sprintf("template execute: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Generate PDF
	pdfData, err := s.gen.Generate(ctx, htmlBuf.String())
	if err != nil {
		errMsg := fmt.Sprintf("pdf generate: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Ensure tmp dir exists
	if err := os.MkdirAll(s.cfg.PDFTmpDir, 0750); err != nil {
		errMsg := fmt.Sprintf("mkdir tmp: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Write PDF to disk
	fileName := jobID + ".pdf"
	filePath := filepath.Join(s.cfg.PDFTmpDir, fileName)
	if err := os.WriteFile(filePath, pdfData, 0640); err != nil {
		errMsg := fmt.Sprintf("write file: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		slog.Error("PDF generation failed", "jobID", jobID, "error", errMsg)
		return
	}

	// Update to done
	if err := s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusDone, fileName, ""); err != nil {
		slog.Warn("failed to update job to done", "jobID", jobID, "error", err)
	}
}

// GetJobs returns all report jobs for a given user.
func (s *ReportService) GetJobs(ctx context.Context, sub string) ([]repository.ReportJobDoc, error) {
	return s.couchRepo.ListReportJobsByUser(sub)
}

// GetJob returns a single job if it belongs to the given user.
// Returns nil, nil if the job is not found or belongs to another user.
func (s *ReportService) GetJob(ctx context.Context, id, sub string) (*repository.ReportJobDoc, error) {
	job, err := s.couchRepo.GetReportJob(id)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return nil, nil
	}
	if job.UserSub != sub {
		return nil, nil
	}
	return job, nil
}

// GetPDF returns the PDF file bytes for a completed job owned by the user.
// Path traversal is prevented via filepath.Join + filepath.Base.
func (s *ReportService) GetPDF(ctx context.Context, id, sub string) ([]byte, error) {
	job, err := s.GetJob(ctx, id, sub)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return nil, nil
	}
	if job.Status != repository.StatusDone {
		return nil, fmt.Errorf("job %s is not done (status: %s)", id, job.Status)
	}

	// Path traversal protection: filepath.Base strips any directory components
	safePath := filepath.Join(s.cfg.PDFTmpDir, filepath.Base(job.FileName))
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("read pdf: %w", err)
	}
	return data, nil
}
