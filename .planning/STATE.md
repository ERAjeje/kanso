# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-17)

**Core value:** O usuário consegue nomear e registrar suas emoções no momento em que as sente, criando um histórico que torna o processo terapêutico mais concreto e orientado a dados.
**Current focus:** v2 features — próximos passos

## Current Position

Phase: 7 — NLP Analysis (07-01 executado)
Milestone: v2 — NLP Analysis (07-01 complete)
Branch: master
Status: Phase 5 ✅ | Phase 6 ✅ | Phase 7: 07-01 ✅ (Infra NLP) | 07-02 ⏳ | 07-03 ⏳
Last activity: 2026-05-23 — Phase 07-01 executed: NLP infra scaffold (FastAPI + gRPC + Docker + Go client)

Progress: [████████████████████] 05-01 ✅ | 06-01 ✅ | 07-01 ✅

## Active Plans

| ID | Description | Files | Status |
|----|-------------|-------|--------|
| 05-01 | Histórico de Registros — lista cronológica + cards expansíveis | `registros.ts`, `History.tsx`, `RegistroCard.tsx` + tests | ✅ Executed |
| bug-sentimento-opcional | Sentimento deixou de ser obrigatório no formulário | `RegistrationForm.tsx`, `RegistrationForm.test.tsx` | ✅ Fixed |
| fix-relatorio-contract | Remover periodStart/periodEnd do contrato — backend calcula datas | `report.go` (handler + service), `couchdb.go` | ✅ Executed |
| **06-01** | **Push Notifications — infra completa (backend + SW + scheduler + UI)** | **9 tasks (ver PLAN.md)** | **✅ Executed** |
| **07-01** | **Infra NLP — FastAPI + gRPC + Docker + Go client + docker-compose** | **6 tasks** | **✅ Executed** |

## Performance Metrics

**Velocity:**
- Total plans completed: 9
- Bug fixes: 5 (3 v1 + 2 v2: sentimento opcional + relatório contract)
- Last activity: 2026-05-17 — Phase 6 planejada

**Trend:**
- v1 MVP: ✅ All phases complete
- v2: Phase 5 ✅ → Phase 6 ✅ (Push Notifications)

## Decisions

- **Wrapper entrypoint** no Dockerfile para rodar Go API + headless-shell no mesmo container
- **env_file** no docker-compose em vez de `environment` com variáveis de configuração
- **env_file** define GOOGLE_CLIENT_ID, JWT_SECRET, COUCHDB_PASSWORD; environment mantém apenas vars fixas
- **Sentimento opcional** — campo sentimento no formulário não é obrigatório. Se vazio, salva como "" e no histórico exibe "Buscando sentimento"
- **Workflow reforçado** — HARD RULES adicionadas ao AGENTS.md: sem código sem `/approve`, sequência LOCKED, bugs via `/gsd-debug`, verificação S.T.O.P. pré-código
- **Relatório sem body** — periodStart/periodEnd removidos do contrato. Backend computa: periodStart = último relatório concluído; periodEnd = now

## Session Continuity

Last session: 2026-05-23 — Phase 07-01 executada (Infra NLP)
Branch: master
Resume with: `/gsd-discuss-phase 07-02` para iniciar discussão da sub-fase 2 (Modelo NLP)
