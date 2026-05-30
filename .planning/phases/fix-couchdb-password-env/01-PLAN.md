---
phase: fix-couchdb-password-env
plan: 01
type: execute
wave: 1
depends_on:
  - fix-security-p2-01-credenciais
  - fix-docker-env-vars
files_modified:
  - infra/.env (new)
  - .env
autonomous: false
requirements:
  - "BUG-005: `make up` falha com `error while interpolating services.couchdb.environment.[]: required variable COUCHDB_PASSWORD is missing a value`"
must_haves:
  truths:
    - Docker Compose consegue interpolar ${COUCHDB_PASSWORD:?...} durante o parsing
    - COUCHDB_PASSWORD é uma senha forte (>= 16 chars base64) — sem default fraco
    - .env (raiz) permanece no .gitignore — credenciais não versionadas
    - infra/.env permanece no .gitignore (padrão .env já cobre)
    - Nenhuma senha é duplicada entre arquivos — cada .env contém o mesmo valor
    - Makefile permanece inalterado — cd infra && docker compose up -d funciona sem flags
  artifacts:
    - path: infra/.env
      provides: "COUCHDB_PASSWORD para interpolação do Docker Compose"
      contains: "COUCHDB_PASSWORD="
    - path: .env
      provides: "COUCHDB_PASSWORD para env_file dos serviços api/scheduler"
      contains: "COUCHDB_PASSWORD="
---

<objective>
Corrigir o erro `required variable COUCHDB_PASSWORD is missing a value` ao rodar `make up` criando `infra/.env` com a variável necessária.

**Causa raiz:** O Docker Compose procura o arquivo `.env` no diretório do projeto (`infra/`), não na raiz. A Task 2 do `fix-docker-env-vars` removeu as vars do bloco `environment`, e o `fix-security-p2-01-credenciais` trocou `:-admin123` por `:?` — mas nenhum dos dois criou `infra/.env` com o valor.

**Consistência com correções anteriores:**
- ME-04 (fix-security-p2-01): Removeu `:-admin123`, tornou a var obrigatória via `:?`
- fix-docker-env-vars: Estabeleceu `env_file: ../.env` para api/scheduler (container runtime)
- CR-01 (fix-security-p0): `.env` gitignorado — credenciais seguras
</objective>

<execution_context>
@infra/docker-compose.yml
@.env
@.gitignore
</execution_context>

<context>
**Arquitetura atual de env vars:**

```
infra/.env  (NOVO — criado por este plano)
  └── COUCHDB_PASSWORD  ──►  compose interpola ${COUCHDB_PASSWORD:?...}
                                │
                                ├─► couchdb (environment block)
                                ├─► api (environment block + env_file: ../.env)
                                └─► scheduler (environment block + env_file: ../.env)

.env (raiz)  (atualizado — descomentar COUCHDB_PASSWORD)
  └── COUCHDB_PASSWORD  ──►  env_file: ../.env → api/scheduler (container runtime)
```

**Por que `infra/.env`?**
- Docker Compose v2 lê `.env` do **project directory** (diretório do compose file = `infra/`)
- O `environment` block usa `${COUCHDB_PASSWORD:?...}` — isso é **interpolação do compose**, resolvida durante o parsing, NÃO dentro do container
- O `env_file: ../.env` nos serviços api/scheduler resolve variáveis para o **container runtime**, não para o parse do compose
- Sem `infra/.env`, o compose nunca encontra a variável e falha com o erro reportado
</context>

<tasks>

<task type="manual" tdd="false">
  <name>Task 1 — Gerar senha forte para CouchDB</name>
  <files></files>
  <action>
    Executar no terminal:
    ```bash
    openssl rand -base64 12
    ```
    O resultado será algo como `3hK9xP2mQ7rL8vN5` — usar nos passos seguintes.
  </action>
  <verify>
    <automated>Senha com >= 16 caracteres base64</automated>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 2 — Criar infra/.env com COUCHDB_PASSWORD</name>
  <files>infra/.env (new)</files>
  <action>
    Criar `infra/.env` com:
    ```
    COUCHDB_PASSWORD=<senha-gerada>
    ```

    Este arquivo é coberto pelo `.gitignore` (padrão `.env` em qualquer diretório).
  </action>
  <verify>
    <automated>test -f infra/.env</automated>
    <automated>grep -q "COUCHDB_PASSWORD=" infra/.env</automated>
    <automated>git check-ignore infra/.env</automated>
    <expected>true (gitignored)</expected>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 3 — Descomentar COUCHDB_PASSWORD no .env raiz</name>
  <files>.env</files>
  <action>
    Em `.env` (raiz), alterar:
    ```diff
    - # COUCHDB_PASSWORD=admin123
    + COUCHDB_PASSWORD=<senha-gerada>
    ```

    **Nota:** O valor `admin123` é fraco e não deve ser usado. A senha gerada em Task 1 é a mesma usada em `infra/.env`.
  </action>
  <verify>
    <automated>grep -q "^COUCHDB_PASSWORD=" .env</automated>
    <automated>test "$(grep "^COUCHDB_PASSWORD=" .env | cut -d= -f2)" != "admin123"</automated>
  </verify>
</task>

</tasks>

<verification>

### Build / Config
- `cd infra && docker compose config` — deve mostrar COUCHDB_PASSWORD definido, sem erros de interpolação
- `cd infra && docker compose config | grep -c "admin123"` — 0 (senha fraca não aparece)

### Runtime
- `make up` — sobe sem erro
- `docker exec kanso-couchdb curl -u admin:<senha> http://localhost:5984/_up` — 200 OK
- `docker exec kanso-api sh -c 'echo $COUCHDB_PASSWORD'` — mostra a senha configurada

### Security
- `git check-ignore infra/.env` — `.env` (ignorado)
- `git check-ignore .env` — `.env` (ignorado)
- Nenhuma senha em `git diff --cached`

</verification>

<success_criteria>

1. `make up` funciona sem erro
2. `COUCHDB_PASSWORD` é senha forte (> 16 chars base64)
3. `infra/.env` não é versionado no git
4. Mesma senha é usada por todos os 3 serviços (couchdb, api, scheduler)
5. Nenhuma alteração no Makefile ou docker-compose.yml
6. Consistente com ME-04 (var obrigatória, sem default)

</success_criteria>
