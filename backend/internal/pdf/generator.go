package pdf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Generator handles PDF generation from HTML using chromedp.
type Generator struct {
	execPath  string
	remoteURL string
	timeout   time.Duration
}

// NewGenerator creates a new PDF Generator using a local chromedp instance.
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

// NewRemoteGenerator creates a Generator that connects to a remote headless-shell
// via the Chrome DevTools Protocol WebSocket URL (e.g., "ws://chromedp:9222/devtools/browser/").
// If remoteURL ends with "/browser/", the WebSocket URL is discovered automatically.
func NewRemoteGenerator(remoteURL string, timeout time.Duration) *Generator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Generator{
		remoteURL: remoteURL,
		timeout:   timeout,
	}
}

// discoverWebSocketURL fetches the WebSocket debugger URL from a headless-shell
// HTTP endpoint (e.g., http://chromedp:9222/json/version).
func discoverWebSocketURL(baseURL string) (string, error) {
	versionURL := baseURL + "/json/version"
	resp, err := http.Get(versionURL)
	if err != nil {
		return "", fmt.Errorf("discover ws url: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("discover ws url read body: %w", err)
	}

	var v struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", fmt.Errorf("discover ws url parse json: %w", err)
	}
	if v.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("discover ws url: empty webSocketDebuggerUrl")
	}
	return v.WebSocketDebuggerURL, nil
}

// Generate produces PDF bytes from the given HTML content using chromedp.
// It connects to a local or remote headless Chrome instance, renders the HTML, and prints to PDF.
func (g *Generator) Generate(ctx context.Context, htmlContent string) ([]byte, error) {
	if htmlContent == "" {
		return nil, fmt.Errorf("empty HTML content")
	}

	var allocCtx context.Context
	var allocCancel context.CancelFunc

	if g.remoteURL != "" {
		// Connect to remote headless-shell via CDP WebSocket
		wsURL := g.remoteURL
		// If the URL is an HTTP base, discover the WebSocket URL
		if len(wsURL) > 4 && wsURL[:4] == "http" {
			var err error
			wsURL, err = discoverWebSocketURL(g.remoteURL)
			if err != nil {
				return nil, fmt.Errorf("remote chromedp: %w", err)
			}
		}
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, wsURL)
	} else {
		// Local chromedp instance
		allocOpts := []chromedp.ExecAllocatorOption{
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Headless,
			chromedp.DisableGPU,
		}

		if g.execPath != "" {
			allocOpts = append(allocOpts, chromedp.ExecPath(g.execPath))
		}

		if g.execPath == "" {
			if envPath := os.Getenv("CHROMEDP_PATH"); envPath != "" {
				allocOpts = append(allocOpts, chromedp.ExecPath(envPath))
			}
		}

		allocCtx, allocCancel = chromedp.NewExecAllocator(ctx, allocOpts...)
	}
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
