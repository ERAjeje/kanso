# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-17)

**Core value:** O usuário consegue nomear e registrar suas emoções no momento em que as sente, criando um histórico que torna o processo terapêutico mais concreto e orientado a dados.
**Current focus:** v2 features — próximos passos

## Current Position

Phase: 5 — Histórico de Registros (completo)
Branch: feat/05-historico-registros
Status: Phase 5 ✅ | Sentimento opcional ✅ | Relatório contract fix ✅
Last activity: 2026-05-17 — All fixes committed

Progress: [████████████] 05-01 — Serviço + RegistroCard + History page + testes

## Active Plans

| ID | Description | Files | Status |
|----|-------------|-------|--------|
| 05-01 | Histórico de Registros — lista cronológica + cards expansíveis | `registros.ts`, `History.tsx`, `RegistroCard.tsx` + tests | ✅ Executed |
| bug-sentimento-opcional | Sentimento deixou de ser obrigatório no formulário | `RegistrationForm.tsx`, `RegistrationForm.test.tsx` | ✅ Fixed |
| fix-relatorio-contract | Remover periodStart/periodEnd do contrato — backend calcula datas | `report.go` (handler + service), `couchdb.go` | ✅ Executed |

## Performance Metrics

**Velocity:**
- Total plans completed: 9
- Bug fixes: 5 (3 v1 + 2 v2: sentimento opcional + relatório contract)
- Last activity: 2026-05-17 — Fix relatório contract (handler/service/repository)

**Trend:**
- v1 MVP: ✅ All phases complete
- v2: Phase 5 — Histórico de Registros concluído ✅

## Decisions

- **Wrapper entrypoint** no Dockerfile para rodar Go API + headless-shell no mesmo container
- **env_file** no docker-compose em vez de `environment` com variáveis de configuração
- **env_file** define GOOGLE_CLIENT_ID, JWT_SECRET, COUCHDB_PASSWORD; environment mantém apenas vars fixas
- **Sentimento opcional** — campo sentimento no formulário não é obrigatório. Se vazio, salva como "" e no histórico exibe "Buscando sentimento"
- **Workflow reforçado** — HARD RULES adicionadas ao AGENTS.md: sem código sem `/approve`, sequência LOCKED, bugs via `/gsd-debug`, verificação S.T.O.P. pré-código
- **Relatório sem body** — periodStart/periodEnd removidos do contrato. Backend computa: periodStart = último relatório concluído; periodEnd = now

## Session Continuity

Last session: 2026-05-17 — Phase 5 + 2 bug fixes committed
Branch: feat/05-historico-registros
Resume with: Definir próxima feature v2 (push notifications, WhatsApp, NLP)
