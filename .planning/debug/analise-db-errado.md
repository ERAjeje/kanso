---
status: resolved
trigger: "analise_nlp docs continuam sendo salvos no DB registros em vez de sentimentos"
created: 2026-05-23
updated: 2026-05-23
resolved_at: 2026-05-23
---

## Symptoms

- analise_nlp docs iam para registros DB
- sentimentos DB vazio
- Frontend sem emotion chips

## Root Cause (two-fold)

1. **Docker image stale:** `make build` compila para host, mas serviço roda em Docker container com imagem antiga
2. **Dados órfãos:** 7 analise:* docs existentes em registros DB nunca migrados

## Resolution

| # | Fix | Status |
|---|-----|--------|
| 1 | Rebuild Docker image: `docker compose build api && docker compose up -d api` | ✅ |
| 2 | Migrate 7 analise docs: registros → sentimentos | ✅ |

## Verification

- sentimentos DB: 7 analise docs ✅
- registros DB: 0 analise docs ✅
- Docker logs: "watcher: event loop started" ✅
