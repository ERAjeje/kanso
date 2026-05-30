---
phase: fix-chromedp-separation
plan: 01
type: execute
wave: 1
depends_on:
  - fix-security-p1
files_modified:
  - infra/chromedp/Dockerfile
  - infra/docker-compose.yml
  - backend/internal/pdf/generator.go
  - backend/Dockerfile
  - backend/.env.example
files_created:
  - infra/chromedp/Dockerfile
autonomous: false
requirements:
  - CHROME-01: chromedp/headless-shell roda em container separado com remote debugging
  - CHROME-02: API container NÃO contém chromedp — roda como appuser, sem seccomp=unconfined
  - CHROME-03: PDF generator conecta ao chromedp remoto via CDP WebSocket
  - CHROME-04: Build e testes passam
must_haves:
  truths:
    - chromedp container expõe porta 9222 (CDP WebSocket)
    - API container descobre WebSocket URL via /json/version no startup
    - Nenhum chromedp binary no API container
    - API container roda como appuser
    - docker-compose build + up funciona sem erros
  artifacts:
    - path: infra/chromedp/Dockerfile
      provides: "Headless-shell container com remote debugging"
    - path: infra/docker-compose.yml
      provides: "kanso-chromedp service + API sem chromedp"
    - path: backend/internal/pdf/generator.go
      provides: "Remote allocator via NewRemoteAllocator"
---

<objective>
Separar o chromedp/headless-shell em um container dedicado (kanso-chromedp),
permitindo que o container da API rode como usuário não-root sem seccomp=unconfined.
</objective>

<execution_context>
@backend/internal/pdf/generator.go
@backend/Dockerfile
@infra/docker-compose.yml
@backend/internal/service/report.go
</execution_context>

---

## Tasks

### Task 1: Criar infra/chromedp/Dockerfile

**Arquivos:** `infra/chromedp/Dockerfile` (novo)

**Ação:**

```dockerfile
FROM chromedp/headless-shell:latest

EXPOSE 9222

CMD ["/headless-shell/headless-shell", "--remote-debugging-port=9222", "--headless", "--disable-gpu", "--no-sandbox"]
```

Container minimalista — só headless-shell com remote debugging.

---

### Task 2: Atualizar docker-compose.yml

**Arquivos:** `infra/docker-compose.yml`

**Ação:**

2a. Adicionar serviço `chromedp`:
```yaml
  chromedp:
    build:
      context: ./chromedp
      dockerfile: Dockerfile
    container_name: kanso-chromedp
    expose:
      - "9222"
    mem_limit: 512m
    restart: unless-stopped
    security_opt:
      - seccomp=unconfined
```

2b. Atualizar service `api`:
- Remover `CHROMEDP_PATH` do environment
- Remover `PDF_TMP_DIR` do environment (default em config.go já é `/tmp/kanso-pdf`)
- Remover `seccomp=unconfined`
- Remover volumes `pdf-tmp`
- Adicionar `CHROMEDP_WS_URL` env: `ws://chromedp:9222/devtools/browser/`

2c. Adicionar `depends_on: chromedp` condicional no `api`.

**Verificação:**
- `docker compose config` mostra 4 services: traefik, couchdb, api, chromedp
- API não tem `security_opt: seccomp=unconfined`
- chromedp tem `seccomp=unconfined`

---

### Task 3: Adicionar suporte a remote allocator no PDF generator

**Arquivos:** `backend/internal/pdf/generator.go`

**Ação:**

Adicionar `NewRemoteGenerator` que conecta via `chromedp.NewRemoteAllocator`:

```go
const defaultRemoteURL = "ws://chromedp:9222/devtools/browser/"

// NewRemoteGenerator creates a Generator that connects to a remote headless-shell.
// remoteURL is the WebSocket URL (e.g., "ws://chromedp:9222/devtools/browser/").
func NewRemoteGenerator(remoteURL string, timeout time.Duration) *Generator {
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    return &Generator{
        execPath:  "",
        remoteURL: remoteURL,
        timeout:   timeout,
    }
}
```

Modificar `Generate` para detectar se `g.remoteURL` está preenchido e usar `NewRemoteAllocator` em vez de `NewExecAllocator`:

```go
if g.remoteURL != "" {
    allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, g.remoteURL)
} else {
    // existing local allocator logic
}
```

Adicionar campo `remoteURL string` na struct `Generator`.

**Discovery do WebSocket URL:**

O headless-shell com `--remote-debugging-port=9222` expõe:
- `http://chromedp:9222/json/version` → retorna JSON com `webSocketDebuggerUrl`
- `http://chromedp:9222/json` → lista de alvos disponíveis

A URL do browser pode mudar entre reboots (UUID diferente). Para evitar esse problema, usar `chromedp.NewRemoteAllocator` que aceita uma URL base e descobre o browser automaticamente quando termina com `/`.

**Alternativa**: Em vez de um `remoteURL` fixo, descobrir dinamicamente no startup:
```go
func discoverWebSocketURL(baseURL string) (string, error) {
    resp, err := http.Get(baseURL + "/json/version")
    if err != nil { return "", err }
    defer resp.Body.Close()
    var v struct {
        WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
        return "", err
    }
    return v.WebSocketDebuggerURL, nil
}
```

Abordagem híbrida: Tenta `CHROMEDP_WS_URL` env var; se vazio, descobre via HTTP.

**Verificação:**
- `go build ./...` exit 0
- `go vet ./...` exit 0
- Testes de unidade passam

---

### Task 4: Atualizar backend/Dockerfile

**Arquivos:** `backend/Dockerfile`

**Ação:**

Mudar runtime stage de `chromedp/headless-shell:latest` para `golang:1.26-alpine`:

```dockerfile
FROM golang:1.26-alpine AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /kanso-api /usr/local/bin/kanso-api
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 8080
CMD ["/usr/local/bin/kanso-api"]
```

Remove:
- chromedp/headless-shell base image
- Wrapper entrypoint script
- apt-get installs
- `ENV CHROMEDP_PATH`

**Verificação:**
- `docker compose build api` termina sem chromedp
- Container não tem `/headless-shell/`
- `docker compose run api whoami` retorna `appuser`

---

### Task 5: Atualizar main.go e config

**Arquivos:** `backend/cmd/kanso-api/main.go`

**Ação:**

Mudar de `pdf.NewGenerator(cfg.CHROMEDPPath, timeout)` para:
```go
remoteURL := os.Getenv("CHROMEDP_WS_URL")
if remoteURL == "" {
    remoteURL = "http://chromedp:9222"
}
pdfGen := pdf.NewRemoteGenerator(remoteURL, 30*time.Second)
```

Ou manter `pdf.NewGenerator` e passar via env var. Ajustar conforme API da Task 3.

**Verificação:**
- `go build ./...` exit 0
- Nenhuma referência a `CHROMEDP_PATH` no código

---

### Task 6: Atualizar integration tests

**Arquivos:** `backend/internal/pdf/generator_test.go`

**Ação:**

Atualizar teste para usar remote generator quando `CHROMEDP_WS_URL` estiver setada:
```go
func TestGenerator_RemoteConnection(t *testing.T) {
    remoteURL := os.Getenv("CHROMEDP_WS_URL")
    if remoteURL == "" {
        t.Skip("CHROMEDP_WS_URL not set")
    }
    gen := NewRemoteGenerator(remoteURL, 30*time.Second)
    // ...
}
```

**Verificação:**
- `go test -tags=integration ./internal/pdf/` com CHROMEDP_WS_URL setada passa
- `go test -short ./internal/pdf/` passa sem conexão

---

## Files Changed Summary

| File | Change |
|------|--------|
| `infra/chromedp/Dockerfile` | **Novo** — headless-shell com remote debugging |
| `infra/docker-compose.yml` | +chromedp service; api sem chromedp, USER appuser |
| `backend/internal/pdf/generator.go` | +remoteURL field + NewRemoteGenerator + remote allocator |
| `backend/Dockerfile` | Base trocada para golang:alpine + USER appuser |
| `backend/cmd/kanso-api/main.go` | usa remote generator |
| `backend/internal/pdf/generator_test.go` | test p/ remote connection |

## Verification

- `go build ./...` exit 0
- `go vet ./...` exit 0
- `go test ./...` passa
- `docker compose build` sem erros
- Container API roda como `appuser`
- Container chromedp roda headless-shell com remote debugging

## Success Criteria

1. chromedp/headless-shell isolado em container próprio
2. API container não contém chromedp — imagem menor, mais segura
3. API container roda como `appuser` (não-root)
4. API container não precisa `seccomp=unconfined`
5. Geração de PDF funciona via conexão remota CDP
6. Build e vet passam, testes existentes continuam funcionando
