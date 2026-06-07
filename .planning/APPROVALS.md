# Approval State

Controlled by `gsd-orchestrator`. Tracks the approval lifecycle for all plans.

---

## Pending Approval

| Plan ID | Description | Created |
|---------|-------------|---------|
| fix-combobox-datepicker | Fix SentimentoCombobox vazio + custom DateTimePicker com date-fns | 2026-06-07 |
| fix-traefik-acme-vps | Corrigir ACME challenge no VPS — remover redirect entrypoint, adicionar catch-all HTTP router | 2026-05-31 |

---

## Approved Plans

| Plan ID | Description | Approved At | Executed |
| 09-gcp-artifact-registry | Migrar deploy para GCP Artifact Registry — Makefile, docker-compose.prod.yml, service account, deploy.sh pull | 2026-06-07 | 2026-06-07 |
| 10-sentiment-training | Sentiment Training — edição no History, coleta de dados, batch retraining, re-análise lazy | 2026-06-07 | 2026-06-07 |
| 09-deploy-vps | Deploy VPS Hostinger — DNS, Traefik/Let's Encrypt, docker-compose produção, backup, deploy script | 2026-05-30 | — |
| pwa-install-prompt | PWA install prompt — hook, banner, ícones, manifest | 2026-05-31 | 2026-05-31 | 
| fix-security-p2-01-credenciais | P2-01: CouchDB password hardening | 2026-05-30 | 2026-05-30 |
| fix-security-p2-02-correcoes-rapidas | P2-02: Correções rápidas | 2026-05-30 | 2026-05-30 |
| fix-security-p2-03-fcm-v1 | P2-03: FCM HTTP v1 migration | 2026-05-30 | 2026-05-30 |
| fix-security-p2-04-logging | P2-04: Log sanitization | 2026-05-30 | 2026-05-30 |
| fix-security-p2-05-arquitetura | P2-05: Decisões arquiteturais | 2026-05-30 | 2026-05-30 |
| 07-01 | Infra NLP — Python FastAPI/gRPC scaffold, Docker, model download | 2026-05-23 | 2026-05-23 |
| 07-02-01 | Data pipeline — mappings, curated phrases, config | 2026-05-23 | 2026-05-23 |
| 07-02-02 | Training — back-translation + BERTimbau fine-tuning | 2026-05-23 | 2026-05-23 |
| 07-02-03 | Inference — classifier, server, validation, Docker | 2026-05-23 | 2026-05-23 |
| 07-03-01 | Repository: CouchDB _changes, checkpoint, analise types + methods | 2026-05-23 | 2026-05-23 |
| 07-03-02 | Watcher: goroutine event loop, gRPC calls, exponential backoff, tests | 2026-05-23 | 2026-05-23 |
| 07-03-03 | Frontend: PouchDB merge, emotion chips in RegistroCard | 2026-05-23 | 2026-05-23 |
| 07-03-04 | PDF: emotion summary section + per-registro emotions | 2026-05-23 | 2026-05-23 |
| 04-01 | Resolve All Tech Debt (P0 + P1 + P2) | 2026-05-17 | 2026-05-17 |
| fix-01 | Fix Google Sign-In button not rendering (GIS script missing in index.html) | 2026-05-17 | 2026-05-17 |
| fix-02 | Fix Docker env vars: GOOGLE_CLIENT_ID não chegava ao container | 2026-05-17 | 2026-05-17 |
| fix-03 | Fix TabBar routing — layout aninhado pós-login | 2026-05-17 | 2026-05-17 |
| 05-01 | Histórico de Registros — lista cronológica com cards expansíveis | 2026-05-17 | 2026-05-17 |
| fix-relatorio-contract | Remover periodStart/periodEnd do contrato — backend calcula datas internamente | 2026-05-17 | 2026-05-17 |
| 06-01 | Push Notifications — backend + frontend (SW + settings) + scheduler + infra | 2026-05-17 | 2026-05-17 |
| fix-push-ux | Push notification UX: toasts para feedback visual | 2026-05-17 | 2026-05-17 |
| 06-02 | Push preferences via PouchDB sync (offline-first) | 2026-05-17 | 2026-05-17 |
| fix-couchdb-jwt-auth | Configure CouchDB JWT auth — local.ini, docker-compose mount, _security docs | 2026-05-23 | 2026-05-23 |
| fix-sentimentos-db | Move analise_nlp do registros DB para sentimentos DB | 2026-05-23 | — |
| fix-security-p1-01 | P1 Security Hardening — headers, rate limit, CouchDB isolation, JWT validation, dep update, logging, non-root containers | 2026-05-30 | 2026-05-30 |
| fix-chromedp-separation-01 | Separar chromedp em container dedicado — API roda como appuser, sem seccomp | 2026-05-30 | 2026-05-30 |
| sec-hardening-01 | Security Hardening — CR-05 + HI-01 + HI-02 + ME-01 | 2026-05-30 | 2026-05-30 |
| fix-deleted-client-oauth | Corrigir deleted_client — fonte única .env via envDir | 2026-05-30 | 2026-05-30 |
| fix-traefik-acme-vps | Corrigir ACME challenge no VPS — remover redirect entrypoint, adicionar catch-all HTTP router | 2026-05-31 | — |

---

## Rejected / Cancelled

| Plan ID | Description | Status | At |
|---------|-------------|--------|----|
| 08-01 | Allowlist de emails (config:active_users) | cancelled — escopo alterado | 2026-05-30 |
| 08-02 | Hardening: headers, chromedp, CouchDB porta, secrets | cancelled — escopo alterado | 2026-05-30 |
| 08-03 | Push interno, proxy /db, JWT algorithm | cancelled — escopo alterado | 2026-05-30 |

---

## Log

| 2026-06-07 | 09-gcp-artifact-registry | approved | GCP Artifact Registry — Makefile, docker-compose.prod.yml, service account, deploy.sh |
| 2026-06-07 | 09-gcp-artifact-registry | executed | 8 tasks: .dockerignore, Makefile, compose.prod.yml, deploy.sh, setup-vps.sh, .gitignore |
| 2026-06-07 | 10-sentiment-training | approved | Sentiment Training — 4 waves, 13 tasks |
| 2026-06-07 | 10-sentiment-training | executed | 11 commits, 94/94 testes, 24 files (+1533/-13) |
| 2026-05-31 | fix-traefik-acme-vps | approved | Corrigir ACME challenge — remover redirect entrypoint, catch-all HTTP router |
| 2026-05-30 | fix-security-p2-01-credenciais | approved | P2-01: CouchDB password hardening — remover fallback admin123 |
| 2026-05-30 | fix-security-p2-01-credenciais | executed | docker-compose.yml: :-admin123 removido dos 3 serviços, validação obrigatória |
| 2026-05-30 | fix-security-p2-02-correcoes-rapidas | approved | P2-02: Correções rápidas — NLP subprocess, SW error, console.error, DB constants |
| 2026-05-30 | fix-security-p2-02-correcoes-rapidas | executed | LO-01: classifier.py sem subprocess. LO-03: SW error com warn. IN-01: console.error→warn. IN-03: DB constants |
| 2026-05-30 | fix-security-p2-03-fcm-v1 | approved | P2-03: Migrar FCM legacy API para HTTP v1 com OAuth2 |
| 2026-05-30 | fix-security-p2-03-fcm-v1 | executed | push.go: FCM v1 via OAuth2 + fallback legacy. Config: FCM_PROJECT_ID, FCM_SERVICE_ACCOUNT_B64 |
| 2026-05-30 | fix-security-p2-04-logging | approved | P2-04: Sanitizar logs do backend — slog estruturado, sem PII |
| 2026-05-30 | fix-security-p2-04-logging | executed | ~40 log.Printf substituídos por slog estruturado em 6 arquivos Go |
| 2026-05-30 | fix-security-p2-05-arquitetura | approved | P2-05: Decisões arquiteturais — Docker socket, gRPC TLS, Vite proxy, JWT cookie |
| 2026-05-30 | fix-security-p2-05-arquitetura | executed | Decisões: HI-01 aceito, HI-02 aceito, ME-01 db-ratelimit no Traefik, ME-05 adiado v4 |
| 2026-05-31 | fix-setup-vps-script | approved | setup-vps.sh — bash guard, subdirs, idempotência, step numbers |
| 2026-05-31 | fix-setup-vps-script | executed | infra/scripts/setup-vps.sh corrigido |

| 2026-05-23 | fix-couchdb-jwt-auth | approved | PouchDB sync fix — CouchDB JWT auth config + _security docs |
| 2026-05-23 | fix-couchdb-jwt-auth | executed | infra/couchdb/local.ini + docker-compose mount + _security in main.go |
| 2026-05-23 | fix-sentimentos-db | approved | Move analise_nlp do registros DB para sentimentos DB |
| 2026-05-23 | fix-sentimentos-db | executed | couchdb.go + watcher.go + registros.ts — SaveAnalise→sentimentos, FindAnalise→sentimentos/_find, getRegistros→sentimentosDB |
| 2026-05-30 | fix-security-p0 | approved | P0 security: CR-01 OAuth secret, CR-02 JWT secret, CR-03 CouchDB validate_doc_update, CR-04 push auth |
| 2026-05-30 | fix-security-p0 | executed | P0 security implemented: JWT secret rotated, validate_doc_update deployed, gitignore for local.ini, push endpoint with JWT admin + API key |
| 2026-05-30 | fix-security-p1-01 | approved | P1 Security Hardening — headers, rate limit, CouchDB isolation, JWT validation, dep update, logging, non-root containers |
| 2026-05-30 | fix-chromedp-separation-01 | approved | Separar chromedp em container dedicado — API roda como appuser, sem seccomp |
| 2026-05-30 | fix-chromedp-separation-01 | executed | Chromedp separation: novo infra/chromedp/Dockerfile, remote allocator, API alpine + appuser, sem seccomp |
| 2026-05-23 | debug-analise-db-errado | fixed | Docker rebuild + restart + migrate 7 analise docs registros→sentimentos |
| 2026-05-23 | fix-mango-indexes | fixed | Add Mango indexes on relatorios (type+userSub+createdAt) and registros (type+userSub+dataHora) |
| 2026-05-23 | fix-frontend-contract-reports | fixed | Align ReportJob type with backend JSON; use downloadReport() with auth instead of <a href> |
| 2026-05-23 | fix-report-userId | fixed | PeriodRegistroDoc.userSub→userId; FindRegistrosByPeriod selector userId; index updated |
| Date | Plan | Action | Detail |
| 2026-05-23 | 07-01 | approved | Infra NLP — Python FastAPI/gRPC scaffold, Docker, model download |
| 2026-05-23 | 07-01 | executed | Infra NLP — 6 tasks: proto, FastAPI+gRPC, Docker, docker-compose, Go client, docs |
| 2026-05-23 | 07-03-01 | approved | Repository: CouchDB _changes, checkpoint, analise types + methods |
| 2026-05-23 | 07-03-02 | approved | Watcher: goroutine event loop, gRPC calls, exponential backoff, tests |
| 2026-05-23 | 07-03-03 | approved | Frontend: PouchDB merge, emotion chips in RegistroCard |
| 2026-05-23 | 07-03-04 | approved | PDF: emotion summary section + per-registro emotions |
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
| 2026-05-17 | fix-scheduler-go-sum | approved | Scheduler build quebra por go.sum ausente |
| 2026-05-17 | fix-scheduler-go-sum | executed | Dockerfile: COPY go.* ./ em vez de go.mod go.sum |

---

## Session Reset Log

*Auto-generated when user requests a new session. Helps the orchestrator distinguish deliberate resets from normal workflow.*

| Timestamp | Reason | State File |
|-----------|--------|------------|
| 2026-05-23 | session_reset | Auto-saved state before /new. Session handoff. NLP context gathered. |
| 2026-05-23 | session_reset | Auto-saved state before /new. 07-03 context gathered, ready for planning. |
| 2026-05-30 | 08-01 | cancelled | Phase 8 planos descartados — escopo redefinido pós-auditoria |
| 2026-05-30 | 08-02 | cancelled | Phase 8 planos descartados — escopo redefinido pós-auditoria |
| 2026-05-30 | 08-03 | cancelled | Phase 8 planos descartados — escopo redefinido pós-auditoria |
| 2026-05-30 | sec-hardening | pending | Plano consolidado: CR-05 + HI-01 + HI-02 + ME-01 |
| 2026-05-30 | sec-hardening-01 | approved | Security Hardening — 4 tasks |
| 2026-05-30 | sec-hardening-01 | executing | CR-05 chromedp flags → HI-01 Traefik → HI-02 gRPC TLS → ME-01 Vite proxy |
| 2026-05-30 | sec-hardening-01 | executed | All 4 tasks complete. Builds pass. generator.go limpo, Traefik file provider, gRPC TLS, Vite proxy removido |
| 2026-05-30 | fix-couchdb-password-env | approved | Criar infra/.env + descomentar COUCHDB_PASSWORD no .env raiz |
| 2026-05-30 | fix-couchdb-password-env | executing | Task 1: gerar senha → Task 2: infra/.env → Task 3: .env raiz |
| 2026-05-30 | fix-couchdb-password-env | executed | infra/.env criado, .env raiz atualizado, senha forte 6Hcw+OfIY83rA46R |
| 2026-05-30 | fix-deleted-client-oauth | approved | Corrigir deleted_client — fonte única .env via envDir |
| 2026-05-30 | fix-deleted-client-oauth | executed | frontend/vite.config.ts +envDir, frontend/.env removido. 69/69 testes OK |
| 2026-05-30 | fix-chromedp-healthcheck | approved | Corrigir healthcheck + Dockerfile do chromedp |
| 2026-05-30 | fix-chromedp-healthcheck | executed | Dockerfile limpo, healthcheck com socat, make up OK — 6/6 containers healthy |
| 2026-05-31 | pwa-install-prompt | approved | PWA install prompt — hook, banner, ícones, manifest |
| 2026-05-31 | pwa-install-prompt | executed | hook useInstallPrompt + InstallBanner + ícones Pillow + manifest unificado + 9+4 testes |
| 2026-06-07 | nlp-global-syntaxerror | fixed | SyntaxError: `global _current_model_version` após uso na linha 27. Movido para linha 26. 4/4 health tests ✅ |
