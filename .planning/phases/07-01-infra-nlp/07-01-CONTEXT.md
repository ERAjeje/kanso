# Phase 07-01: Infra NLP — Context & Decisions

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Sub-phase Boundary

Scaffold the Python NLP service infrastructure: FastAPI + gRPC server, Dockerfile with model download at build time, docker-compose integration, and proto definitions. This is the foundation that sub-phases 07-02 (Modelo) and 07-03 (Integração) build upon.

**Requirements addressed:** NLP-01, NLP-02, NLP-03 (infra layer)

</domain>

<decisions>
## Sub-phase Decisions (extracted from milestone)

### Stack
- **D-09:** Serviço Python com **FastAPI + gRPC** (grpcio)
- **D-08:** Comunicação via **gRPC** (não HTTP/REST)

### Model Management
- **D-16:** Modelo baixado durante **Docker build** (não em runtime)
- **D-17:** Container Python com modelo incluído na imagem

### Proto Schema
- **D-10:** Request: registro completo (id, sensacoes, contexto, pensamentos, dataHora)
- **D-11:** Response: emocaoPrincipal, emocoesSecundarias[], scores, intensidade, analiseAdicional, metadadosModelo

### Service Topology
- NLP service is **internal only** (no public route via Traefik)
- Go backend proxies requests via gRPC on Docker internal network
- **Service isolation** — Each concern gets its own Docker service (api, scheduler, nlp)
- Pattern: follows `scheduler/` microservice structure

### the agent's Discretion
- Python dependency exact versions (grpcio, fastapi, torch, transformers)
- Directory structure within nlp-service/ (proto layout, module organization)
- Health check endpoint design for Docker compose
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Planning
- `.planning/REQUIREMENTS.md` §NLP — NLP-01, NLP-02, NLP-03
- `.planning/PROJECT.md` §Active — NLP feature description
- `.planning/ROADMAP.md` §Phase 7 — Phase structure

### Existing Patterns
- `infra/docker-compose.yml` — Service pattern (api, scheduler, couchdb, traefik)
- `scheduler/Dockerfile` — Multi-stage Go build pattern for microservice
- `scheduler/main.go` — Microservice entrypoint pattern
- `backend/internal/config/config.go` — Env-based config pattern
- `nlp-service/README.md` — Existing stub (update with current stack)

### Milestone Context
- `.planning/phases/v2-nlp-analysis/v2-nlp-CONTEXT.md` — Full milestone decisions
</canonical_refs>

---

*Phase: 07-01 — Infra NLP*
*Context gathered: 2026-05-23 via milestone discuss-phase*
