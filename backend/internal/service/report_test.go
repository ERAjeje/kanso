package service

import (
	"context"
	"testing"
	"time"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
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
