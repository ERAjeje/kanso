//go:build integration

package pdf

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestGenerator_GeneratesNonEmptyPDF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gen := NewGenerator("", 30*time.Second)

	html := "<html><body><h1>Test Report</h1><p>This is a test</p></body></html>"
	pdfBytes, err := gen.Generate(context.Background(), html)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("Generate() returned empty PDF")
	}

	// PDFs start with %PDF
	if string(pdfBytes[:4]) != "%PDF" {
		t.Fatalf("Generate() output does not start with %%PDF, got: %s", string(pdfBytes[:20]))
	}
}

func TestGenerator_ErrorsOnInvalidHTML(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gen := NewGenerator("", 5*time.Second)

	// Extremely broken HTML might cause chromedp to error
	_, err := gen.Generate(context.Background(), "")
	if err == nil {
		t.Fatal("Generate() expected error for empty HTML, got nil")
	}
}

func TestRemoteGenerator_GeneratesNonEmptyPDF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	remoteURL := os.Getenv("CHROMEDP_WS_URL")
	if remoteURL == "" {
		t.Skip("CHROMEDP_WS_URL not set — skipping remote generator test")
	}

	gen := NewRemoteGenerator(remoteURL, 30*time.Second)

	html := "<html><body><h1>Remote Test Report</h1><p>This is a remote test</p></body></html>"
	pdfBytes, err := gen.Generate(context.Background(), html)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("Generate() returned empty PDF")
	}

	if string(pdfBytes[:4]) != "%PDF" {
		t.Fatalf("Generate() output does not start with %%PDF, got: %s", string(pdfBytes[:20]))
	}
}
