package pdf

import (
	"context"
	"time"
)

// Generator handles PDF generation from HTML using chromedp.
type Generator struct {
	execPath  string
	timeout   time.Duration
}

// NewGenerator creates a new PDF Generator.
// execPath is the path to Chromium/Chrome binary; empty means use system default.
// timeout controls the context timeout per generation.
func NewGenerator(execPath string, timeout time.Duration) *Generator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Generator{
		execPath: execPath,
		timeout:  timeout,
	}
}

// Generate produces PDF bytes from the given HTML content.
// The actual chromedp implementation is in Task 3.
func (g *Generator) Generate(ctx context.Context, htmlContent string) ([]byte, error) {
	return nil, nil
}
