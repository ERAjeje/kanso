# Approval State

Controlled by `gsd-orchestrator`. Tracks the approval lifecycle for all plans.

---

## Pending Approval

*None — all plans approved and executed.*

---

## Approved Plans

| Plan ID | Description | Approved At | Executed | 
| 07-01 | Infra NLP — Python FastAPI/gRPC scaffold, Docker, model download | 2026-05-23 | 2026-05-23 |
| 07-02-01 | Data pipeline — mappings, curated phrases, config | 2026-05-23 | — |
| 07-02-02 | Training — back-translation + BERTimbau fine-tuning | 2026-05-23 | — |
| 07-02-03 | Inference — classifier, server, validation, Docker | 2026-05-23 | — |
| 04-01 | Resolve All Tech Debt (P0 + P1 + P2) | 2026-05-17 | 2026-05-17 |
| fix-01 | Fix Google Sign-In button not rendering (GIS script missing in index.html) | 2026-05-17 | 2026-05-17 |
| fix-02 | Fix Docker env vars: GOOGLE_CLIENT_ID não chegava ao container | 2026-05-17 | 2026-05-17 |
| fix-03 | Fix TabBar routing — layout aninhado pós-login | 2026-05-17 | 2026-05-17 |
| 05-01 | Histórico de Registros — lista cronológica com cards expansíveis | 2026-05-17 | 2026-05-17 |
| fix-relatorio-contract | Remover periodStart/periodEnd do contrato — backend calcula datas internamente | 2026-05-17 | 2026-05-17 |
| 06-01 | Push Notifications — backend + frontend (SW + settings) + scheduler + infra | 2026-05-17 | 2026-05-17 |
| fix-push-ux | Push notification UX: toasts para feedback visual | 2026-05-17 | 2026-05-17 |
| 06-02 | Push preferences via PouchDB sync (offline-first) | 2026-05-17 | 2026-05-17 |

---

## Rejected / Cancelled

| Plan ID | Description | Status | At |
|---------|-------------|--------|----|
| — | — | — | — |

---

## Log

| Date | Plan | Action | Detail |
| 2026-05-23 | 07-01 | approved | Infra NLP — Python FastAPI/gRPC scaffold, Docker, model download |
| 2026-05-23 | 07-01 | executed | Infra NLP — 6 tasks: proto, FastAPI+gRPC, Docker, docker-compose, Go client, docs |
| 2026-05-23 | 07-02-01 | approved | Data pipeline — mappings, curated phrases, config |
| 2026-05-23 | 07-02-02 | approved | Training — back-translation + BERTimbau fine-tuning |
| 2026-05-23 | 07-02-03 | approved | Inference — classifier, server, validation, Docker |
| 2026-05-17 | 04-01 | approved | Tech Debt resolution |
| 2026-05-17 | 04-01 | executed | Tech Debt resolution |
| 2026-05-17 | fix-01 | approved | Login Google Sign-In fix |
| 2026-05-17 | fix-01 | executed | Login Google Sign-In fix (script GIS + fallback button + testes) |
| 2026-05-17 | fix-02 | approved | Docker: GOOGLE_CLIENT_ID não passava ao container por precedência de env vars |
| 2026-05-17 | fix-02 | executed | Docker: removido env vars do environment bloco → env_file, corrigido entrypoint do Dockerfile |
| 2026-05-17 | fix-03 | approved | TabBar routing — layout aninhado |
| 2026-05-17 | fix-03 | executed | TabBar routing — App.tsx reestruturado com layout aninhado, rotas /register, /history, /profile dentro de TabBar |
| 2026-05-17 | 05-01 | approved | Histórico de Registros — lista cronológica com cards expansíveis |
| 2026-05-17 | 05-01 | executed | Histórico de Registros — implementado + testado |
| 2026-05-17 | fix-relatorio-contract | approved | Remover periodStart/periodEnd do contrato — backend calcula datas internamente |
| 2026-05-17 | fix-relatorio-contract | executed | Handler sem body parse + service computa datas + repository GetLastCompletedReport |
| 2026-05-17 | 06-01 | approved | Push Notifications — backend + frontend + scheduler + infra |
| 2026-05-17 | 06-01 | executed | Push Notifications — implementado + testado |
| 2026-05-17 | fix-push-ux | approved | Push UX: toasts para feedback visual em erros silenciosos |
| 2026-05-17 | fix-push-ux | executed | useAuth.tsx + Profile.tsx: toasts adicionados em falhas silenciosas |
| 2026-05-17 | 06-02 | approved | Push preferences via PouchDB sync |
| 2026-05-17 | 06-02 | executed | PouchDB preferenciasDB + push.ts rewrite + backend atualizado |

---

## Session Reset Log

*Auto-generated when user requests a new session. Helps the orchestrator distinguish deliberate resets from normal workflow.*

| Timestamp | Reason | State File |
|-----------|--------|------------|
| 2026-05-23 | session_reset | Auto-saved state before /new. Session handoff. NLP context gathered. |
| 2026-05-23 | session_reset | Auto-saved state before /new. 07-03 context gathered, ready for planning. |
