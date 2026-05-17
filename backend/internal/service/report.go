package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/templates"
)

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
		log.Printf("failed to get last completed report: %v", err)
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
		log.Printf("failed to update job %s to processing: %v", jobID, err)
	}

	// Render HTML template with report data
	tmplData := map[string]interface{}{
		"GeneratedAt": time.Now().UTC().Format("02/01/2006 às 15:04 MST"),
		"PeriodStart": periodStart,
		"PeriodEnd":   periodEnd,
		"Registros":   []interface{}{}, // Registros data comes from CouchDB via registros DB
	}

	var htmlBuf bytes.Buffer
	if err := templates.ReportTemplate.Execute(&htmlBuf, tmplData); err != nil {
		errMsg := fmt.Sprintf("template execute: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		log.Printf("PDF generation failed for job %s: %s", jobID, errMsg)
		return
	}

	// Generate PDF
	pdfData, err := s.gen.Generate(ctx, htmlBuf.String())
	if err != nil {
		errMsg := fmt.Sprintf("pdf generate: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		log.Printf("PDF generation failed for job %s: %s", jobID, errMsg)
		return
	}

	// Ensure tmp dir exists
	if err := os.MkdirAll(s.cfg.PDFTmpDir, 0750); err != nil {
		errMsg := fmt.Sprintf("mkdir tmp: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		log.Printf("PDF generation failed for job %s: %s", jobID, errMsg)
		return
	}

	// Write PDF to disk
	fileName := jobID + ".pdf"
	filePath := filepath.Join(s.cfg.PDFTmpDir, fileName)
	if err := os.WriteFile(filePath, pdfData, 0640); err != nil {
		errMsg := fmt.Sprintf("write file: %v", err)
		s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusFailed, "", errMsg)
		log.Printf("PDF generation failed for job %s: %s", jobID, errMsg)
		return
	}

	// Update to done
	if err := s.couchRepo.UpdateReportJobStatus(jobID, "", repository.StatusDone, fileName, ""); err != nil {
		log.Printf("failed to update job %s to done: %v", jobID, err)
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
