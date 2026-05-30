---
phase: fix-security-p2-01-credenciais
plan: 01
type: execute
wave: 1
depends_on:
  - fix-security-p1-01
files_modified:
  - infra/docker-compose.yml
  - .env
autonomous: false
requirements:
  - ME-04: CouchDB admin password default removido — fallback `:-admin123` substituído por validação de env var obrigatória
must_haves:
  truths:
    - Nenhum serviço usa `admin123` como senha CouchDB — nem como fallback
    - docker-compose.yml falha no startup se COUCHDB_PASSWORD não estiver definida
    - .env.example documenta que a senha é obrigatória e sugere `openssl rand -base64 12`
  artifacts:
    - path: infra/docker-compose.yml
      provides: "COUCHDB_PASSWORD sem fallback inseguro"
      contains: "COUCHDB_PASSWORD removido :-admin123"
      contains: "validator obrigatório ou env check"
---

<objective>
Remover a senha padrão `admin123` do CouchDB e garantir que todos os serviços usem uma senha forte definida via env var.
</objective>

<execution_context>
@infra/docker-compose.yml
@.env
</execution_context>

---

## Tasks

### Task 1: Remover fallback `:-admin123` do docker-compose.yml

**Arquivo:** `infra/docker-compose.yml`

**Ação:**

Substituir todos os `:-admin123` por validação que falha se a var não estiver definida:

Linha 33 (`couchdb` service):
```yaml
- COUCHDB_PASSWORD=${COUCHDB_PASSWORD:?COUCHDB_PASSWORD is required}
```

Linha 81 (`api` service):
```yaml
- COUCHDB_PASSWORD=${COUCHDB_PASSWORD:?COUCHDB_PASSWORD is required}
```

Linha 109 (`scheduler` service):
```yaml
- COUCHDB_PASSWORD=${COUCHDB_PASSWORD:?COUCHDB_PASSWORD is required}
```

Isso faz o Docker Compose falhar imediatamente se a variável não estiver definida, em vez de usar um default inseguro.

### Task 2: Gerar senha forte e atualizar .env

**Arquivo:** `.env`

**Ação:**

1. Gerar nova senha:
   ```bash
   openssl rand -base64 12
   ```

2. Atualizar `.env`:
   ```
   COUCHDB_PASSWORD=<nova-senha>
   ```

3. Garantir que `.env.example` (se existir) documenta a necessidade:
   ```
   # COUCHDB_PASSWORD=<gerar com: openssl rand -base64 12>
   ```

**Verificação:**
- `docker compose config` não mostra mais o valor `admin123` em nenhum serviço
- `docker compose up` falha com erro claro se `COUCHDB_PASSWORD` não estiver definida
- Todos os 3 serviços (couchdb, api, scheduler) usam a mesma senha

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-ME-04-01 | Broken Authentication | CouchDB | mitigate | Senha forte aleatória, sem fallback default |
| T-ME-04-02 | Information Disclosure | docker-compose.yml | mitigate | `:-admin123` removido, var é obrigatória |

## Verification

- `docker compose config` não contém `admin123`
- `docker compose up` com `COUCHDB_PASSWORD` vazia → falha com erro
- CouchDB healthcheck passa com a nova senha

## Success Criteria

1. Nenhuma ocorrência de `:-admin123` no `docker-compose.yml`
2. `COUCHDB_PASSWORD` é obrigatória — startup falha sem ela
3. Senha forte (>12 chars base64) gerada e configurada no `.env`
4. Docker compose sobe sem erros com a nova senha
