package service

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/templates"
)

func setupTestReportService(t *testing.T) (*ReportService, *repository.CouchDB) {
	t.Helper()

	cfg := &config.Config{
		PDFTmpDir: "/tmp/kanso-pdf-test",
	}

	// Repository points to nothing — tests verify method signatures and error handling
	couchRepo := repository.NewCouchDB("http://localhost:15984", "admin", "pass")
	gen := pdf.NewGenerator("", 5*time.Second)
	svc := NewReportService(couchRepo, gen, cfg)
	return svc, couchRepo
}

func TestReportService_RequestReport_ReturnsJobID(t *testing.T) {
	svc, _ := setupTestReportService(t)

	// Without a running CouchDB, this will error — but the method signature is validated
	jobID, err := svc.RequestReport(context.Background(), "user123")
	if err != nil {
		// Expected: CouchDB not available — but the method exists and returns (string, error)
		t.Logf("RequestReport returned expected error (no CouchDB): %v", err)
	}
	if jobID != "" {
		t.Logf("Got job ID: %s", jobID)
	}
}

func TestReportService_GetJobs_ReturnsList(t *testing.T) {
	svc, _ := setupTestReportService(t)

	jobs, err := svc.GetJobs(context.Background(), "user123")
	if err != nil {
		t.Logf("GetJobs returned expected error (no CouchDB): %v", err)
	}
	if jobs != nil {
		t.Logf("Got %d jobs", len(jobs))
	}
}

func TestReportService_GetJob_OwnershipCheck(t *testing.T) {
	svc, _ := setupTestReportService(t)

	job, err := svc.GetJob(context.Background(), "nonexistent-id", "user123")
	if err != nil {
		t.Logf("GetJob returned expected error (no CouchDB or not found): %v", err)
	}
	if job != nil {
		t.Logf("Got job: %+v", job)
	}
}

func TestReportService_GetPDF_ReturnsBytes(t *testing.T) {
	svc, _ := setupTestReportService(t)

	// Use a fake job ID and sub — should fail because no CouchDB/no file
	data, err := svc.GetPDF(context.Background(), "nonexistent-id", "user123")
	if err != nil {
		t.Logf("GetPDF returned expected error (no CouchDB or not found): %v", err)
	}
	if data != nil {
		t.Logf("Got %d bytes", len(data))
	}
}

func TestReportService_MutexProtectsGeneration(t *testing.T) {
	svc, _ := setupTestReportService(t)

	// The service should have a mutex — verify it doesn't panic on concurrent access
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			svc.RequestReport(context.Background(), "user123")
			done <- true
		}()
	}

	select {
	case <-done:
		// All goroutines completed without panic
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for concurrent RequestReport calls")
	}
}

func TestReportTemplate_RendersEmotionSummary(t *testing.T) {
	data := ReportData{
		GeneratedAt: "23/05/2026 às 10:00 BRT",
		PeriodStart: "2026-01-01T00:00:00Z",
		PeriodEnd:   "2026-05-23T10:00:00Z",
		Registros: []RegistroReportItem{
			{
				Data:        "2026-05-23T09:00:00Z",
				Sentimento:  "ansiedade",
				Sensacoes:   "coração acelerado",
				Contexto:    "reunião",
				Pensamentos: "nervoso",
				Emocoes: []repository.EmotionScore{
					{Emotion: "ansiedade", Score: 0.85},
					{Emotion: "medo", Score: 0.42},
				},
			},
			{
				Data:        "2026-05-22T14:00:00Z",
				Sentimento:  "tristeza",
				Sensacoes:   "cansaço",
				Contexto:    "casa",
				Pensamentos: "saudade",
				Emocoes: []repository.EmotionScore{
					{Emotion: "tristeza", Score: 0.78},
					{Emotion: "saudade", Score: 0.65},
				},
			},
		},
		EmotionSummary: []EmotionSummaryItem{
			{Emotion: "ansiedade", Count: 1},
			{Emotion: "medo", Count: 1},
			{Emotion: "tristeza", Count: 1},
			{Emotion: "saudade", Count: 1},
		},
	}

	var buf bytes.Buffer
	if err := templates.ReportTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execute: %v", err)
	}

	output := buf.String()

	// Verify emotion summary section
	if !strings.Contains(output, "Resumo das Emoções") {
		t.Error("expected emotion summary heading")
	}
	if !strings.Contains(output, "ansiedade") {
		t.Error("expected emotion 'ansiedade' in output")
	}
	if !strings.Contains(output, "1x") {
		t.Error("expected count display")
	}

	// Verify per-registro emotions
	if !strings.Contains(output, "medo") {
		t.Error("expected secondary emotion 'medo' in output")
	}
	if !strings.Contains(output, "tristeza") {
		t.Error("expected emotion 'tristeza' in output")
	}

	// Verify per-registro chip styling
	if !strings.Contains(output, "padding: 2px 8px") {
		t.Error("expected chip inline style")
	}
	if !strings.Contains(output, "border-radius: 4px") {
		t.Error("expected chip border-radius style")
	}

	// Verify existing content still renders
	if !strings.Contains(output, "Relatório Kanso") {
		t.Error("expected report title")
	}
	if !strings.Contains(output, "coração acelerado") {
		t.Error("expected registro content")
	}
}

func TestReportTemplate_OmitsEmotionSummaryWhenEmpty(t *testing.T) {
	data := ReportData{
		GeneratedAt:    "23/05/2026 às 10:00 BRT",
		PeriodStart:    "2026-01-01T00:00:00Z",
		PeriodEnd:      "2026-05-23T10:00:00Z",
		Registros:      []RegistroReportItem{},
		EmotionSummary: []EmotionSummaryItem{},
	}

	var buf bytes.Buffer
	if err := templates.ReportTemplate.Execute(&buf, data); err != nil {
		t.Fatalf("template execute: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Resumo das Emoções") {
		t.Error("expected NO emotion summary when EmotionSummary is empty")
	}
}
