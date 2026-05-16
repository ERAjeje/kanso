package pdf

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Generator handles PDF generation from HTML using chromedp.
type Generator struct {
	execPath string
	timeout  time.Duration
}

// NewGenerator creates a new PDF Generator.
// execPath is the path to Chromium/Chrome binary; empty means use system default.
// timeout controls the context timeout per generation (default 30s).
func NewGenerator(execPath string, timeout time.Duration) *Generator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Generator{
		execPath: execPath,
		timeout:  timeout,
	}
}

// Generate produces PDF bytes from the given HTML content using chromedp.
// It creates a headless Chrome instance, renders the HTML, and prints to PDF.
func (g *Generator) Generate(ctx context.Context, htmlContent string) ([]byte, error) {
	if htmlContent == "" {
		return nil, fmt.Errorf("empty HTML content")
	}

	// Build chromedp context with allocator options
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-file-access-from-files", true),
	}

	// Use custom Chrome/Chromium path if provided
	if g.execPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(g.execPath))
	}

	// Use CHROMEDP_PATH env var if exec path not set
	if g.execPath == "" {
		if envPath := os.Getenv("CHROMEDP_PATH"); envPath != "" {
			allocOpts = append(allocOpts, chromedp.ExecPath(envPath))
		}
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	// Create context with timeout
	genCtx, genCancel := context.WithTimeout(allocCtx, g.timeout)
	defer genCancel()

	// Create Chrome tab context
	tabCtx, tabCancel := chromedp.NewContext(genCtx)
	defer tabCancel()

	// Navigate to a data URI with the encoded HTML content
	dataURI := "data:text/html," + url.PathEscape(htmlContent)

	var pdfBuf []byte
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(dataURI),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp pdf generation: %w", err)
	}

	if len(pdfBuf) == 0 {
		return nil, fmt.Errorf("chromedp produced empty PDF")
	}

	return pdfBuf, nil
}
