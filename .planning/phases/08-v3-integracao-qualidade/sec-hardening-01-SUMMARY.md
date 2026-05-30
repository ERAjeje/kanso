# sec-hardening-01 — Summary

**Executed:** 2026-05-30
**Plan:** CR-05 + HI-01 + HI-02 + ME-01

## Results

### Task 1 — CR-05: Chromedp flags ✅
- `backend/internal/pdf/generator.go`: removed `disable-web-security` + `allow-file-access-from-files`
- `go build ./internal/pdf/` ✅

### Task 2 — HI-01: Traefik File Provider ✅
- `infra/traefik/traefik.yml`: removed `providers.docker`
- `infra/traefik/dynamic.yml`: added routers (api, couchdb), services, couchdb-strip middleware, cors-headers middleware (CORS for localhost:5173 + kanso.local)
- `infra/docker-compose.yml`: removed `docker.sock` mount + all `traefik.*` labels from couchdb and api services

### Task 3 — HI-02: gRPC TLS ✅
- `infra/certs/gen-grpc-certs.sh`: new script to generate self-signed CA + server cert
- `infra/docker-compose.yml`: added `grpc-certs` volume (bind mount to `./certs`), mounted in nlp + api
- `nlp-service/src/server.py`: added `_load_grpc_credentials()` + `add_secure_port()` with fallback to insecure
- `backend/internal/nlp/client.go`: `NewClient` now accepts `caCertPath`, uses TLS if provided with fallback
- `backend/cmd/kanso-api/main.go`: passes `GRPC_CA_CERT` env var to `NewClient`

### Task 4 — ME-01: Vite proxy /db → Traefik ✅
- `frontend/vite.config.ts`: removed `/db` proxy
- `.env`: `VITE_COUCHDB_URL` changed from `/db` to `https://kanso.local/db`
- `frontend/.env`: same change

## Build Verification
- `go build ./...` ✅
- `go vet ./...` ✅

## Manual Step Required
```bash
bash infra/certs/gen-grpc-certs.sh
```
Run once before `docker compose up` to generate gRPC TLS certificates.

## Security Audit Coverage
- 19/24 audit items fixed (4 new in this plan)
- 2 deferred (ME-05 JWT localStorage v4, CR-01 OAuth secret — não necessário para SPA)
- 0 outstanding code fixes
