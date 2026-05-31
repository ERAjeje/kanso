---
phase: 08-v3-integracao-qualidade
plan: sec-hardening-01
type: execute
wave: 1
depends_on:
  - fix-security-p0
  - fix-security-p1-01
  - fix-security-p2-01
  - fix-security-p2-02
  - fix-security-p2-03
  - fix-security-p2-04
  - fix-security-p2-05
files_modified:
  - backend/internal/pdf/generator.go
  - infra/traefik/dynamic.yml
  - infra/traefik/traefik.yml
  - infra/docker-compose.yml
  - backend/internal/nlp/client.go
  - nlp-service/src/server.py
  - infra/docker-compose.yml
  - frontend/vite.config.ts
  - .env
  - frontend/.env
autonomous: false
requirements:
  - SEC-CR05: Chromedp sem flags inseguras
  - SEC-HI01: Traefik sem Docker socket (file provider)
  - SEC-HI02: gRPC com TLS auto-assinado
  - SEC-ME01: Vite proxy /db removido, PouchDB via Traefik HTTPS
user_setup:
  - Confiar certificado auto-assinado no browser (dev)
must_haves:
  truths:
    - generator.go não usa disable-web-security nem allow-file-access-from-files
    - Traefik descobre serviços via dynamic.yml (não via Docker labels)
    - docker.sock não está montado no Traefik
    - gRPC entre Go API e NLP service usa TLS com certificado auto-assinado
    - Vite não tem proxy /db — PouchDB sync via https://kanso.local/db
    - CORS configurado no Traefik para localhost:5173 (dev) e kanso.local (prod)
  artifacts:
    - path: backend/internal/pdf/generator.go
      provides: "Chromedp sem flags inseguras"
      not_contains: "disable-web-security"
      not_contains: "allow-file-access-from-files"
    - path: infra/traefik/dynamic.yml
      provides: "Rotas declarativas + middlewares CORS"
      contains: "routers.api"
      contains: "routers.couchdb"
      contains: "cors"
    - path: infra/traefik/traefik.yml
      provides: "Apenas file provider"
      not_contains: "docker"
    - path: backend/internal/nlp/client.go
      provides: "gRPC client com TLS"
      contains: "credentials.NewClientTLSFromCert"
    - path: nlp-service/src/server.py
      provides: "gRPC server com TLS"
      contains: "server.add_secure_port"
    - path: frontend/vite.config.ts
      provides: "Sem proxy /db"
      not_contains: "/db"
    - path: .env
      provides: "VITE_COUCHDB_URL apontando para Traefik"
      contains: "https://kanso.local/db"
---

<objective>
Corrigir os 4 itens remanescentes do SECURITY-AUDIT.md: CR-05 (chromedp flags), HI-01 (Docker socket → file provider), HI-02 (gRPC TLS auto-assinado), ME-01 (Vite proxy /db → Traefik).
</objective>

<execution_context>
@SECURITY-AUDIT.md
@.planning/phases/08-v3-integracao-qualidade/08-CONTEXT.md
</execution_context>

<context>
<interfaces>

From `backend/internal/pdf/generator.go:106-107`:
- Duas flags inseguras: `disable-web-security` + `allow-file-access-from-files`

From `infra/traefik/traefik.yml:7-11`:
- Docker provider + File provider ativos simultaneamente
- Routes via Docker labels nos containers couchdb e api

From `backend/internal/nlp/client.go:47-49`:
- `grpc.WithTransportCredentials(insecure.NewCredentials())`

From `nlp-service/src/server.py:85`:
- `server.add_insecure_port(f"[::]:{port}")`

From `frontend/vite.config.ts:36-40`:
- `/db` proxy para `http://localhost:5984` (rota quebrada — CouchDB não expõe porta ao host)

From `.env:51`:
- `VITE_COUCHDB_URL=/db` (precisa mudar para `https://kanso.local/db`)

From `infra/docker-compose.yml:92-98`:
- Labels Traefik no api e couchdb services

From `infra/traefik/dynamic.yml:6-39`:
- Já tem middlewares secHeaders, redirect-https, auth-ratelimit, db-ratelimit
- Precisa de routers + services + CORS middleware
</interfaces>
</context>

<tasks>

<task type="auto" tdd="false">
  <name>Task 1 — CR-05: Remove chromedp insecure flags</name>
  <files>backend/internal/pdf/generator.go</files>
  <action>
    In `backend/internal/pdf/generator.go`, remove lines 106-107:
    ```go
    chromedp.Flag("disable-web-security", true),
    chromedp.Flag("allow-file-access-from-files", true),
    ```
    Resultado esperado — allocOpts become:
    ```go
    allocOpts := []chromedp.ExecAllocatorOption{
        chromedp.NoFirstRun,
        chromedp.NoDefaultBrowserCheck,
        chromedp.Headless,
        chromedp.DisableGPU,
    }
    ```
  </action>
  <verify>
    <automated>go build ./internal/pdf/ && go vet ./internal/pdf/</automated>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 2 — HI-01: Migrate Traefik Docker provider → File provider</name>
  <files>infra/traefik/dynamic.yml, infra/traefik/traefik.yml, infra/docker-compose.yml</files>
  <action>
    2a. Em `infra/traefik/dynamic.yml`, adicionar routers, services e CORS middleware:

    ```yaml
    http:
      routers:
        api:
          rule: "Host(`kanso.local`) && PathPrefix(`/api`)"
          entrypoints:
            - websecure
          tls: {}
          service: api
          middlewares:
            - auth-ratelimit
            - secHeaders
            - cors-headers

        couchdb:
          rule: "Host(`kanso.local`) && PathPrefix(`/db`)"
          entrypoints:
            - websecure
          tls: {}
          service: couchdb
          middlewares:
            - couchdb-strip
            - secHeaders
            - db-ratelimit
            - cors-headers

      services:
        api:
          loadBalancer:
            servers:
              - url: "http://api:8080"

        couchdb:
          loadBalancer:
            servers:
              - url: "http://couchdb:5984"

      middlewares:
        couchdb-strip:
          stripPrefix:
            prefixes:
              - "/db"

        cors-headers:
          headers:
            accessControlAllowOriginList:
              - "http://localhost:5173"
              - "https://kanso.local"
            accessControlAllowMethods:
              - "GET"
              - "POST"
              - "PUT"
              - "DELETE"
              - "OPTIONS"
            accessControlAllowHeaders:
              - "Authorization"
              - "Content-Type"
            accessControlAllowCredentials: true
    ```

    2b. Em `infra/traefik/traefik.yml`, remover Docker provider:
    ```yaml
    providers:
      file:
        filename: /etc/traefik/dynamic.yml
    ```

    2c. Em `infra/docker-compose.yml`:
    - Remover volume `- /var/run/docker.sock:/var/run/docker.sock:ro` (linha 23)
    - Remover bloco `labels:` do service `couchdb` (linhas 46-53)
    - Remover bloco `labels:` do service `api` (linhas 92-98)
    - Adicionar `depends_on:` no service `traefik` não é necessário — Traefik descobre services via DNS
  </action>
  <verify>
    <automated>docker compose config --no-interpolate 2>&1 | head -5</automated>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 3 — HI-02: gRPC TLS auto-assinado</name>
  <files>
    infra/certs/gen-grpc-certs.sh (new),
    nlp-service/src/server.py,
    backend/internal/nlp/client.go,
    infra/docker-compose.yml
  </files>
  <action>
    3a. Criar script `infra/certs/gen-grpc-certs.sh`:
    ```bash
    #!/bin/sh
    set -e
    CERT_DIR="$(dirname "$0")"
    DAYS=3650

    # CA key + cert
    openssl req -x509 -newkey rsa:4096 -nodes \
      -keyout "$CERT_DIR/ca.key" -out "$CERT_DIR/ca.crt" \
      -subj "/CN=kanso-grpc-ca" -days $DAYS

    # Server key + CSR
    openssl req -newkey rsa:4096 -nodes \
      -keyout "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
      -subj "/CN=nlp"

    # Server cert signed by CA
    openssl x509 -req -in "$CERT_DIR/server.csr" \
      -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" \
      -CAcreateserial -out "$CERT_DIR/server.crt" -days $DAYS

    rm -f "$CERT_DIR/server.csr" "$CERT_DIR/ca.srl"
    chmod 600 "$CERT_DIR/ca.key" "$CERT_DIR/server.key"
    echo "gRPC certs generated in $CERT_DIR"
    ```

    3b. Em `infra/docker-compose.yml`, adicionar volume compartilhado:
    ```yaml
    volumes:
      grpc-certs:
        driver: local
        driver_opts:
          type: none
          device: ./certs
          o: bind
    
    services:
      nlp:
        volumes:
          - grpc-certs:/certs:ro
      
      api:
        volumes:
          - grpc-certs:/certs:ro
    ```

    3c. Em `nlp-service/src/server.py`, modificar `serve_grpc`:
    ```python
    def _load_grpc_credentials():
        cert_path = os.environ.get("GRPC_CERT_DIR", "/certs")
        with open(os.path.join(cert_path, "server.crt"), "rb") as f:
            server_cert = f.read()
        with open(os.path.join(cert_path, "server.key"), "rb") as f:
            server_key = f.read()
        return grpc.ssl_server_credentials([(server_key, server_cert)])

    async def serve_grpc(port: int = 50051):
        server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
        analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
        creds = _load_grpc_credentials()
        server.add_secure_port(f"[::]:{port}", creds)
        logger.info("gRPC secure server starting on port %d", port)
        await server.start()
        await server.wait_for_termination()
    ```
    Adicionar imports: `import os`, `import grpc` (já importado).

    3d. Em `backend/internal/nlp/client.go`, modificar `NewClient`:
    ```go
    func NewClient(addr string, caCertPath string) (*Client, error) {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        var opts []grpc.DialOption
        if caCertPath != "" {
            caCert, err := os.ReadFile(caCertPath)
            if err != nil {
                return nil, fmt.Errorf("read ca cert: %w", err)
            }
            certPool := x509.NewCertPool()
            if !certPool.AppendCertsFromPEM(caCert) {
                return nil, fmt.Errorf("failed to parse ca cert")
            }
            creds := credentials.NewClientTLSFromCert(certPool, "")
            opts = append(opts, grpc.WithTransportCredentials(creds))
        } else {
            opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
        }
        opts = append(opts, grpc.WithBlock())

        conn, err := grpc.DialContext(ctx, addr, opts...)
        if err != nil {
            return nil, err
        }
        return &Client{conn: conn, client: pb.NewAnalysisServiceClient(conn)}, nil
    }
    ```
    Adicionar imports: `"crypto/x509"`, `"os"`, `"fmt"`.

    3e. Atualizar `backend/cmd/kanso-api/main.go` — onde `nlp.NewClient` é chamado, passar `"/certs/ca.crt"` como segundo argumento.

    **Importante:** Usuário precisa executar `bash infra/certs/gen-grpc-certs.sh` uma vez antes de `docker compose up`.
  </action>
  <verify>
    <automated>
      go build ./internal/nlp/ && go vet ./internal/nlp/
      grep -c "add_secure_port" nlp-service/src/server.py
      grep -c "credentials.NewClientTLSFromCert" backend/internal/nlp/client.go
    </automated>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 4 — ME-01: Remove Vite proxy /db, usa Traefik HTTPS</name>
  <files>frontend/vite.config.ts, .env, frontend/.env</files>
  <action>
    4a. Em `frontend/vite.config.ts`, remover o bloco `/db` do proxy:
    ```typescript
    server: {
      host: true,
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
    ```

    4b. Em `.env` (raiz) e `frontend/.env`:
    ```
    VITE_COUCHDB_URL=https://kanso.local/db
    ```
  </action>
  <verify>
    <automated>grep -c "/db" frontend/vite.config.ts</automated>
    <expected>0</expected>
    <automated>grep -c "kanso.local/db" .env</automated>
    <expected>1</expected>
  </verify>
</task>

</tasks>

<threat_model>

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-SEC-01 | Elevation of Privilege | Chromedp | mitigate | disable-web-security removido — chromedp respeita Same-Origin Policy |
| T-SEC-02 | Information Disclosure | Chromedp | mitigate | allow-file-access-from-files removido — sem acesso a arquivos locais |
| T-SEC-03 | Privilege Escalation | Traefik | mitigate | Docker socket removido — Traefik sem acesso ao daemon Docker |
| T-SEC-04 | Eavesdropping | gRPC | mitigate | TLS auto-assinado — tráfego criptografado na rede interna |
| T-SEC-05 | Spoofing | gRPC | mitigate | Certificado server-side — cliente valida identidade do servidor |
| T-SEC-06 | Unauthorized Access | CouchDB proxy | mitigate | Proxy /db removido do Vite. PouchDB sync via Traefik com JWT auth + HTTPS |
| T-SEC-07 | Unauthorized Access | CouchDB CORS | mitigate | CORS restrito a localhost:5173 e kanso.local |

</threat_model>

<verification>

### Build
- `go build ./...` compila sem erros
- `go vet ./...` passa
- `grep -c "disable-web-security" backend/internal/pdf/generator.go` → 0
- `grep -c "insecure.NewCredentials" backend/internal/nlp/client.go` → 0 (ou só no fallback)
- `grep -c "add_insecure_port" nlp-service/src/server.py` → 0
- `grep -c "/db" frontend/vite.config.ts` → 0
- `grep "kanso.local/db" .env` → encontrado

### Runtime
- `docker compose config` sem docker.sock mount
- `curl -sk https://kanso.local/api/health` → resposta 200 com security headers
- `curl -sk https://kanso.local/db/_up` → resposta 200 (via Traefik)
- NLP container sobe com gRPC TLS
- API container conecta ao NLP via TLS
- PouchDB sync funciona via `https://kanso.local/db`

</verification>

<success_criteria>

1. Chromedp sem flags inseguras — PDF generation funciona com sandbox padrão
2. Traefik sem Docker socket — rotas definidas em dynamic.yml
3. gRPC criptografado com TLS auto-assinado entre Go API e NLP service
4. Vite sem proxy /db — PouchDB conecta via Traefik HTTPS com JWT
5. CORS configurado no Traefik para dev (localhost:5173) e prod (kanso.local)
6. All builds and vet pass

</success_criteria>

<output>
After completion, create `.planning/phases/08-v3-integracao-qualidade/sec-hardening-01-SUMMARY.md`
</output>
