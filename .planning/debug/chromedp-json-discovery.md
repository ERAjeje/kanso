---
status: resolved
trigger: "PDF generation fails: 'invalid character H looking for beginning of value'"
created: 2026-05-30
updated: 2026-05-30
resolved: 2026-05-30
---

# Debug Session: chromedp-json-discovery

## Symptoms

- **Expected behavior**: `POST /api/reports` generates and downloads a PDF report
- **Actual behavior**: HTTP error response with message `pdf generate: remote chromedp: discover ws url parse json: invalid character 'H' looking for beginning of value`
- **Error chain**:
  1. `service/report.go` wraps: `"pdf generate: %v"`
  2. `pdf/generator.go:95` wraps: `"remote chromedp: %w"`
  3. `pdf/generator.go:69` wraps: `"discover ws url parse json: %w"`
  4. `pdf/generator.go:68` — `json.Unmarshal` fails — response is NOT JSON
- **Timeline**: After security fixes (commit `d2c06801`) that separated chromedp into its own container
- **Reproduction**: Any PDF generation request with the stack running via docker-compose

## Current Focus

- **hypothesis**: The container `kanso-chromedp` passes its TCP healthcheck (socat port check) but headless-shell on port 9223 is NOT properly serving the HTTP DevTools API (`/json/version`). The response body starts with `'H'` instead of `{`, causing `json.Unmarshal` to fail.
- **test**: Verify what `http://chromedp:9222/json/version` actually returns — is it HTML, plain text, or garbage?
- **expecting**: The response is likely a plain-text error message from headless-shell or a misconfigured socat forward
- **next_action**: Inspect the actual response from the endpoint and fix based on findings

## Evidence

### Known architecture
- `infra/chromedp/Dockerfile`: Only `FROM chromedp/headless-shell:latest` + `EXPOSE 9222` — no CMD
- Base image entrypoint (`run.sh`):
  - Starts headless-shell on **port 9223** with `--remote-debugging-port=9223`
  - Starts `socat` TCP forward **9222 → 9223**
- `infra/docker-compose.yml:61`: Healthcheck uses `socat /dev/null TCP4:localhost:9222` — **TCP only, not HTTP**
- `infra/docker-compose.yml:76`: `CHROMEDP_WS_URL=http://chromedp:9222` — HTTP URL triggers discovery path
- `backend/internal/pdf/generator.go:52-75`: `discoverWebSocketURL()` does `http.Get(baseURL + "/json/version")` and fails on JSON parse
- `backend/cmd/kanso-api/main.go:187-191`: Reads `CHROMEDP_WS_URL`, creates remote generator

### Why healthcheck passes but HTTP fails
The healthcheck uses `socat /dev/null TCP4:localhost:9222` — this only verifies:
- socat is listening on port 9222
- TCP handshake succeeds

It does NOT verify:
- headless-shell is running on port 9223
- The HTTP API at `/json/version` responds with valid JSON
- socat forwarding is functional for bidirectional data (not just connection setup)

### What 'H' could mean
The JSON parser sees first byte `0x48` (ASCII 'H'). Possible response bodies:
- "Headless..." — headless-shell might output startup text to the socket
- "Host..." — a proxy/server response line
- Some other plain-text message

## Eliminated

- *(none yet — need to verify actual response body)*

## Resolution

- **root_cause**: headless-shell rejeita requests com `Host` header não-IP/não-localhost (proteção contra DNS rebinding). `http.Get("http://chromedp:9222/json/version")` envia `Host: chromedp:9222`, que é rejeitado com HTTP 500 e body de texto "Host header is specified and is not an IP address or localhost.", causando falha no `json.Unmarshal`.
- **fix**: Em `discoverWebSocketURL`, criar `http.NewRequest` com `req.Host = "localhost:9222"` para satisfazer a validação, e depois substituir `localhost:9222` na URL WebSocket retornada pelo hostname real (ex: `chromedp:9222`).
- **verification**: `go build ./...` exit 0, `go vet ./...` exit 0, `go test -short ./...` pass. Teste manual com Go program confirmou que a descoberta funciona via IP e retorna URL WebSocket com host correto.
- **files_changed**: `backend/internal/pdf/generator.go` — `discoverWebSocketURL` agora usa `http.NewRequest` com `Host` override + URL replacement pós-parse.
