# Milestone v2: NLP Analysis — Context & Decisions

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Milestone Boundary

NLP emotion analysis service that processes diary entries (sensações + contexto + pensamentos) using a Portuguese-language emotion classification model (BERTimbau fine-tuned), and enriches registrations with detected emotion tags. Runs asynchronously triggered by new registrations syncing to CouchDB.

**V2 Requirements addressed:** NLP-01, NLP-02, NLP-03

**3 sub-phases:**
1. **Infra** — Python/FastAPI scaffold, gRPC service, Docker, model download at build
2. **Modelo** — BERTimbau fine-tuning, emotion classification pipeline (10-15 Portuguese emotions)
3. **Integração** — Go backend _changes feed listener, CouchDB enrichment, UI display (histórico + relatório)

</domain>

<decisions>
## Implementation Decisions

### Trigger Model
- **D-01:** Análise disparada **a cada novo registro** — assim que o doc sync para o CouchDB
- **D-02:** Go backend escuta **CouchDB _changes feed** para detectar novos registros (não cria endpoint Go para registros)
- **D-03:** **Backfill** de todos os registros existentes na primeira execução do NLP
- **D-04:** Análise **assíncrona** — não bloqueia o registro do usuário (NLP-03)

### Modelo de Emoções
- **D-05:** **BERTimbau** (neuralmind.ai) — BERT pré-treinado em português
- **D-06:** Fine-tuning para classificação de **10-15 emoções do português** (alegria, tristeza, raiva, medo, nojo, surpresa, ansiedade, vergonha, culpa, saudade, amor, gratidão + neutro)
- **D-07:** Dataset inicial: **GoEmotions-PT** (traduzido) + expansão com dados reais do app conforme acumula

### Arquitetura Go → Python
- **D-08:** Comunicação via **gRPC** (não HTTP/REST)
- **D-09:** Serviço Python com **FastAPI + gRPC** (grpcio)
- **D-10:** Request schema: registro completo (id, sensacoes, contexto, pensamentos, dataHora)
- **D-11:** Response schema: emocaoPrincipal, emocoesSecundarias[], scores, intensidade, analiseAdicional, metadadosModelo

### Schema de Armazenamento (CouchDB)
- **D-12:** Resultado da análise armazenado em **documento separado** (`analise:{registroId}`)
  - Tipo: `analise_nlp`
  - Campos: emocaoPrincipal, emocoes[], scores, modelo, analisadoEm, registroId
- **D-13:** Não modifica o documento original do registro

### Enriquecimento
- **D-14:** Emoções detectadas aparecem **no relatório PDF E no histórico** (RegistroCard)
- **D-15:** Sub-fase 3 (Integração) inclui mudanças no frontend para exibir emoções no histórico

### Gestão do Modelo
- **D-16:** Modelo baixado durante **Docker build** (não em runtime)
- **D-17:** Container Python com modelo incluído na imagem

### Sub-fases do Milestone
- **Sub-fase 1 — Infra:** Scaffold Python/FastAPI/gRPC + Dockerfile + download do modelo no build
- **Sub-fase 2 — Modelo:** Fine-tuning BERTimbau + pipeline de classificação + testes do modelo
- **Sub-fase 3 — Integração:** Go _changes listener + CouchDB enrichment + frontend display

### the agent's Discretion
- Tamanho exato do modelo e estratégia de cache de inferência no Python
- Estratégia de polling vs long-poll no _changes feed
- Detalhes de UI para exibição das emoções no RegistroCard
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Planning
- `.planning/REQUIREMENTS.md` §NLP — NLP-01, NLP-02, NLP-03 requirements
- `.planning/PROJECT.md` §Active — NLP feature description
- `nlp-service/README.md` — Existing stub README with planned stack

### Existing Patterns (consume, do not duplicate)
- `infra/docker-compose.yml` — Service pattern (api, scheduler, couchdb, traefik). NLP service follows same pattern.
- `backend/cmd/kanso-api/main.go` — Chi router, middleware, handler/service/repository layers
- `backend/internal/service/report.go` — Async job pattern for reference
- `backend/internal/config/config.go` — Env-based config pattern
- `backend/internal/repository/couchdb.go` — CouchDB operations pattern
- `frontend/src/services/registros.ts` — Registro data model and PouchDB operations
- `.planning/phases/03-reports/03-01-PLAN.md` — Async job infrastructure (PDF gen pattern)

### Prior Phase Context
- `.planning/phases/06-push-notifications/06-CONTEXT.md` — Scheduler microservice pattern, docker-compose service addition
- `.planning/contexts/4-CONTEXT.md` — Tech debt context, nlp-service README creation
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **📁 scheduler/** — Pattern for separate microservice (Dockerfile + docker-compose entry)
- **backend/internal/repository/couchdb.go** — CouchDB _changes feed available via existing HTTP client
- **backend/internal/config/config.go** — Env config pattern (add NLP_SERVICE_URL, etc.)

### Established Patterns
- **Service isolation** — Each concern gets its own Docker service (api, scheduler, nlp)
- **CouchDB as datastore** — New doc type `analise_nlp` follows `relatorio`, `push_prefs` pattern
- **gRPC** — New pattern for the project (currently all internal comm is HTTP or CouchDB direct)
- **Async processing** — Report service uses sync.Mutex + goroutine; NLP can use same pattern or dedicated worker

### Integration Points
- **CouchDB registros DB** — _changes feed watched by Go backend
- **Go backend** — New service that watches _changes and calls Python gRPC
- **Report generation** — NLP-enriched data available when generating PDF
- **Frontend RegistroCard** — Display detected emotions in history view
- **Report template** — Include emotion analysis in PDF output
</code_context>

<specifics>
## Specific Ideas

- BERTimbau fine-tuned para 10-15 emoções portuguesas (incluindo saudade, vergonha, culpa — culturalmente relevantes)
- Dataset GoEmotions-PT como partida, evolução com dados reais
- gRPC streaming futuramente se necessário para análise em tempo real
</specifics>

<deferred>
## Deferred Ideas

- **WhatsApp integration** — Próxima feature v2 após NLP. Enviar relatório PDF via Twilio
- **Análise em tempo real (streaming)** — Possível evolução futura do gRPC, não para v1 do NLP
- **Modelo fine-tuned com dados do usuário específico** — Personalização por perfil, futura melhoria
</deferred>

---

*Milestone: v2 — NLP Analysis*
*Context gathered: 2026-05-23*
