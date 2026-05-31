# Fix: setup-vps.sh — bash shebang, subdiretórios, idempotência

## Problemas Identificados

1. **Shebang ignorado** — usuário executou com `sh` em vez de `bash` → `pipefail` é inválido no POSIX sh
2. **Subdiretórios ausentes** — `touch /opt/kanso/infra/traefik/acme.json` falha se `infra/traefik/` não existe
3. **Clone em dir existente** — `git clone` para `/opt/kanso` após `mkdir -p /opt/kanso` cria dir vazio, clone falha
4. **Step numbers inconsistentes** — steps 1-5 usam `[N/7]`, steps 6-8 usam `[N/8]`

## Correções

### Task 1: Guard bash check no topo
- **File**: `infra/scripts/setup-vps.sh`
- Adicionar validação no início: se não estiver rodando com bash, aborta com mensagem clara

### Task 2: mkdir -p para subdiretórios do acme.json
- **File**: `infra/scripts/setup-vps.sh`
- Adicionar `mkdir -p /opt/kanso/infra/traefik` antes do `touch acme.json`

### Task 3: Tornar clone idempotente
- **File**: `infra/scripts/setup-vps.sh`
- Se `/opt/kanso/.git` existe → `git pull`
- Se `/opt/kanso` não existe → `git clone`
- Se `/opt/kanso` existe mas sem `.git` → avisa e pula

### Task 4: Corrigir step numbers
- **File**: `infra/scripts/setup-vps.sh`
- Todos os steps passam a usar `[N/8]` consistentemente

### Task 5: Remover mkdir desnecessário (opcional)
- **File**: `infra/scripts/setup-vps.sh`
- Remover `mkdir -p /opt/kanso` da step 5, já que o clone (step 7) cria o diretório. Mas manter como fallback seguro para idempotência.

## Files Affected

| File | Action |
|------|--------|
| `infra/scripts/setup-vps.sh` | Corrigir 4 problemas |

## Dependencies

Nenhuma. Script standalone.

## Verification

```bash
bash infra/scripts/setup-vps.sh --dry-run  # ou revisão visual do diff
```
