---
phase: fix
plan_id: 02
wave: 1
depends_on: [fix-01]
files_modified:
  - infra/docker-compose.yml
  - backend/Dockerfile
  - .env
  - .gitignore
autonomous: false
requirements:
  - "BUG-003: GOOGLE_CLIENT_ID não chegava ao container por precedência de env vars no docker-compose"
  - "BUG-004: Go API nunca iniciava — entrypoint do Dockerfile passava CMD como argumento ao headless-shell"
---

# Plan 02: Fix Docker env vars and API entrypoint

## Objective

Fix two issues preventing the authentication flow from completing:

1. **Env var precedence**: docker-compose `environment` block usa `${GOOGLE_CLIENT_ID:-dev-client-id}` que faz substituição no parse do compose (lê do shell ou `infra/.env`), não do `env_file`. Resultado: `GOOGLE_CLIENT_ID=dev-client-id` dentro do container.

2. **Entrypoint do Dockerfile**: `chromedp/headless-shell:latest` tem entrypoint próprio (`/headless-shell/run.sh`). O `CMD ["kanso-api"]` virava argumento do `run.sh`, que o passa para o headless-shell — o Go API nunca executava.

3. **Credenciais no git**: Garantir que `.env` com credenciais esteja no `.gitignore`.

## Tasks

### Task 1: Fix Dockerfile entrypoint — run Go API alongside headless-shell

**type**: execute
**wave**: 1
**files_modified**:
  - backend/Dockerfile

<acceptance_criteria>
- Dockerfile runtime stage has a wrapper script that starts both `kanso-api` (background) and `headless-shell` (foreground)
- `docker logs kanso-api` shows "Starting server on :8080"
- `curl http://localhost:8080/api/health` returns `{"status":"ok"}`
</acceptance_criteria>

<action>
Replace the runtime stage CMD with a wrapper entrypoint:

```dockerfile
RUN printf '#!/bin/sh\n\
/usr/local/bin/kanso-api &\n\
exec /headless-shell/run.sh "$@"\n' > /entrypoint.sh && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```
</action>

---

### Task 2: Fix docker-compose env var precedence

**type**: execute
**wave**: 1
**files_modified**:
  - infra/docker-compose.yml
  - .env

<acceptance_criteria>
- `GOOGLE_CLIENT_ID` no container tem o valor real (não `dev-client-id`)
- `JWT_SECRET` tem valor real (não fallback)
- `docker exec kanso-api echo $GOOGLE_CLIENT_ID` retorna o client ID configurado
</acceptance_criteria>

<action>
1. Remover `GOOGLE_CLIENT_ID`, `JWT_SECRET`, `COUCHDB_PASSWORD` do bloco `environment` no docker-compose — manter apenas variáveis fixas (`COUCHDB_URL`, `COUCHDB_USER`, `PDF_TMP_DIR`, `CHROMEDP_PATH`)

2. Adicionar `env_file: ../.env` ao service `api`

3. No `.env` raiz:
   - Descomentar `JWT_SECRET=dev-secret-change-in-production`
   - Remover `GOOGLE_CLIENT_SECRET` (não usado pelo backend)
   - Remover `GOOGLE_CLIENT_ID` duplicado (seção Docker Compose)
</action>

---

### Task 3: Verificar .gitignore

**type**: execute
**wave**: 1
**files_modified**:
  - .gitignore

<acceptance_criteria>
- `.env` está listado no `.gitignore`
- `git check-ignore .env` retorna `.env`
- `git check-ignore frontend/.env` retorna `frontend/.env`
</acceptance_criteria>

<action>
Verificar que `.env` já existe no `.gitignore`. Já está coberto pela linha 3: `.env`
</action>

---

## Verification

1. `docker exec kanso-api sh -c 'echo $GOOGLE_CLIENT_ID'` mostra valor real
2. `curl http://localhost:8080/api/health` retorna 200 OK
3. `docker logs kanso-api` mostra "Starting server on :8080"
4. `git check-ignore .env` confirma ignorado
5. Login Google → /api/auth/google → 200 (não 401)
