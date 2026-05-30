# Phase 09: Deploy VPS — Summary

**Executed:** 2026-05-30
**Approved:** 2026-05-30

## What was created / modified

### Scripts (infra/scripts/)
- `setup-vps.sh` — Provisionamento one-shot da VPS (Docker, UFW, git, clone repo)
- `deploy.sh` — Deploy manual via git pull + build + up
- `backup-couchdb.sh` — Backup CouchDB via export de volume Docker

### Traefik configs
- `traefik.yml` — Adicionado Let's Encrypt (certificatesResolvers), redirect HTTP→HTTPS
- `dynamic.yml` — Rotas atualizadas para `kanso.edsonajeje.cloud`, CORS de produção, certResolver letsencrypt

### Docker Compose
- `docker-compose.yml` — Logging rotation em todos os serviços, NLP 4g→2g, API healthcheck, backup label no volume couchdb, volume acme.json

### Environment
- `.env.production` — Template com placeholders para secrets de produção
- `.gitignore` — Adicionado `infra/traefik/acme.json` e `infra/.env.production`

### Infra
- `infra/traefik/acme.json` — Arquivo vazio (chmod 600) para Traefik gerir certificados

## Pending (manual / VPS-side)
- [ ] Executar `setup-vps.sh` na VPS
- [ ] Configurar DNS: `kanso.edsonajeje.cloud` → IP da VPS
- [ ] Gerar secrets e preencher `.env.production`
- [ ] Verificar Google OAuth origins no Cloud Console
- [ ] `docker compose up -d` na VPS
- [ ] Validar fluxo completo (checklist no PLAN.md Task 8)
- [ ] Agendar cron para backup semanal
