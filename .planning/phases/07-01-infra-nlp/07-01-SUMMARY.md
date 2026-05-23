---
phase: 07-01-infra-nlp
plan: 01
subsystem: infra
tags: [python, fastapi, grpc, docker, couchdb, bertimbau]
requires:
  - phase: 01-foundation-auth
    provides: docker infrastructure pattern (traefik, couchdb)
  - phase: 06-push-notifications
    provides: docker-compose microservice pattern (scheduler)
provides:
  - NLP service scaffold (FastAPI + gRPC)
  - Docker multi-stage build with BERTimbau at build time
  - docker-compose integration (nlp service)
  - Go gRPC client stub
affects: [07-02-modelo, 07-03-integracao]
tech-stack:
  added: [FastAPI, gRPC (grpcio), Uvicorn, PyTorch, Transformers, sentencepiece, google.golang.org/grpc]
  patterns: [multi-stage Docker with model download, gRPC internal service, Go gRPC client pattern]
key-files:
  created: [nlp-service/proto/analysis.proto, nlp-service/src/server.py, nlp-service/src/health.py, nlp-service/src/__main__.py, nlp-service/Dockerfile, nlp-service/download_model.py, backend/internal/nlp/client.go]
  modified: [infra/docker-compose.yml, backend/internal/config/config.go, backend/go.mod, nlp-service/README.md]
key-decisions:
  - "gRPC for service-to-service communication between Go backend and Python NLP"
  - "BERTimbau model downloaded at Docker build time, not runtime"
  - "NLP service is internal only (no Traefik public route)"
  - "NLP service follows scheduler/ microservice pattern"
patterns-established:
  - "Python microservice with FastAPI + gRPC dual-server entrypoint"
  - "Multi-stage Docker build with model caching in build stage"
  - "Go gRPC client with timeout context and insecure local connection"
requirements-completed: [NLP-01, NLP-02, NLP-03]
duration: 15min
completed: 2026-05-23
---

# Phase 07-01: Infra NLP Summary

**Python NLP service scaffold with FastAPI + gRPC, Docker multi-stage build with BERTimbau at build time, docker-compose integration, and Go gRPC client stub**

## Performance

- **Duration:** 15 min
- **Started:** 2026-05-23T12:44:00-03:00
- **Completed:** 2026-05-23T12:59:00-03:00
- **Tasks:** 6
- **Files modified:** 27 (created + modified)

## Accomplishments
- gRPC proto definition for AnalysisService with Analyze RPC
- FastAPI health endpoint (HTTP) + gRPC server dual entrypoint
- Multi-stage Dockerfile with BERTimbau download at build time
- docker-compose integration: nlp service with healthcheck
- Go gRPC client with NewClient/Analyze/Close methods
- All 8 Python tests passing (proto, server, health)

## Task Commits

1. **Task 1: gRPC proto + deps** - `4db73fed` (feat)
2. **Task 2: FastAPI + gRPC server** - `52c08211` (feat)
3. **Task 3: Dockerfile** - `00ddf357` (feat)
4. **Task 4: docker-compose + config.go** - `74ecb0ec` (feat)
5. **Task 5: Go gRPC client + proto stubs** - `331d75f7` (feat)
6. **Task 6: Project docs update** - `b90aa94e` (docs)

## Files Created/Modified
- `nlp-service/proto/analysis.proto` - gRPC proto with AnalysisService RPC
- `nlp-service/requirements.txt` - Python dependencies (torch, transformers, grpcio, fastapi)
- `nlp-service/gen_proto.sh` - Python gRPC stub generator
- `nlp-service/src/server.py` - gRPC AnalysisServicer with placeholder
- `nlp-service/src/health.py` - FastAPI /health endpoint
- `nlp-service/src/__main__.py` - Dual-server entrypoint (HTTP + gRPC)
- `nlp-service/Dockerfile` - Multi-stage build with model download
- `nlp-service/download_model.py` - BERTimbau download script
- `infra/docker-compose.yml` - Added nlp service (port 50051/8000, healthcheck)
- `backend/internal/config/config.go` - Added NLPGrpAddr field
- `backend/internal/nlp/client.go` - Go gRPC client
- `backend/internal/nlp/proto/analysis.pb.go` - Generated proto (Go)
- `backend/internal/nlp/proto/analysis_grpc.pb.go` - Generated gRPC stubs (Go)

## Decisions Made
- Followed plan as specified; no deviations

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Float precision in proto test (float32 vs float64) - fixed with math.isclose
- pytest-asyncio not installed for async tests - installed locally
- protoc not available - installed from GitHub releases

## Next Phase Readiness
- NLP service scaffold complete, ready for 07-02 (Modelo) - fine-tuning BERTimbau for emotion classification
- Go client ready for 07-03 (Integração) - _changes listener + CouchDB enrichment

---

*Phase: 07-01-infra-nlp*
*Completed: 2026-05-23*
