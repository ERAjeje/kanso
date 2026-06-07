# 09-gcp-artifact-registry — Summary

**Executed:** 2026-06-07
**Approved:** 2026-06-07

## What was created / modified

### Created
- `backend/.dockerignore` — Exclui node_modules, pycache, tests, .git do build context
- `scheduler/.dockerignore` — Exclui .git, .env, tests do build context
- `infra/docker-compose.prod.yml` — Override production com `image:` apontando para GCP (api, scheduler, nlp, chromedp)
- `.planning/phases/09-gcp-artifact-registry/09-GCP-PLAN.md`

### Modified
- `Makefile` — Adicionados targets: `docker-build`, `docker-push`, `docker-publish`, `docker-verify`
- `infra/scripts/deploy.sh` — Substituído `build` por `pull` via GCP, adicionado `-f docker-compose.prod.yml`
- `infra/scripts/setup-vps.sh` — Next steps atualizados com GCP auth
- `.gitignore` — Adicionado `infra/secrets/`

## Pending (manual)
- [ ] **Task 6**: Criar service account `kanso-puller` no GCP Console, baixar JSON key, copiar para VPS em `infra/secrets/gcp-key.json`
- [ ] **Task 6**: Executar `docker login` na VPS com a JSON key
- [ ] **Task 8**: Executar `make docker-publish` na máquina local para fazer o primeiro push das imagens
- [ ] Verificar no VPS: `docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml pull`
