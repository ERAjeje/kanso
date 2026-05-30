---
phase: 09-deploy-vps
type: execute
depends_on: [08-v3-integracao-qualidade]
files_modified:
  - infra/docker-compose.yml
  - infra/traefik/traefik.yml
  - infra/traefik/dynamic.yml
  - .env
files_created:
  - infra/.env.production
  - infra/scripts/setup-vps.sh
  - infra/scripts/deploy.sh
  - infra/scripts/backup-couchdb.sh
autonomous: true
must_haves:
  truths:
    - The app is accessible at https://kanso.edsonajeje.cloud
    - TLS is configured via Let's Encrypt (Traefik)
    - All services start with a single docker compose up on the VPS
    - Google OAuth works with the production domain
    - CouchDB data is backed up weekly
    - Deploy is done via SSH + git pull + docker compose up --build
  artifacts:
    - path: infra/docker-compose.yml
      provides: "Production-ready docker-compose with memory limits, restart policies, healthchecks"
    - path: infra/traefik/dynamic.yml
      provides: "Updated for production domain + Let's Encrypt"
    - path: infra/scripts/setup-vps.sh
      provides: "One-shot script to prepare the VPS"
    - path: infra/scripts/deploy.sh
      provides: "Repeatable deploy script"
---

<objective>
Colocar o Kanso em produção na VPS Hostinger (kanso.edsonajeje.cloud) com Docker, Traefik + Let's Encrypt, backup do CouchDB, e processo de deploy manual via SSH.

**Infra:** Hostinger KVM 1 — Ubuntu 24.04 LTS, 1 vCPU, 4GB RAM, 50GB SSD
**Domínio:** kanso.edsonajeje.cloud → apontar DNS para IP da VPS
**Deploy:** Manual — SSH + git pull + docker compose up --build
</objective>

<execution_context>
@.planning/PROJECT.md
@infra/docker-compose.yml
@infra/traefik/traefik.yml
@infra/traefik/dynamic.yml
@.env
@backend/Dockerfile
@nlp-service/Dockerfile
@scheduler/Dockerfile
</execution_context>

<decisions>
## Decisões de Deploy

### D-01: Build na VPS
O deploy é feito via SSH: `git pull && docker compose up --build -d`. Sem registry intermediário. O repositório é clonado na VPS e as imagens são buildadas localmente.

### D-02: Let's Encrypt via Traefik
Traefik faz TLS termination com certificados Let's Encrypt automáticos. DNS `kanso.edsonajeje.cloud` aponta para o IP da VPS.

### D-03: Backup via volumes Docker + snapshot
O servidor já tem backup semanal. Para o CouchDB, adicionalmente: script que exporta o volume Docker para um arquivo .tar.gz. Snapshots do provedor como redundância.

### D-04: Sem CI/CD (manual)
Processo manual documentado: SSH → git pull → docker compose build → docker compose up -d.

### D-05: NLP Service com 4GB de RAM
O NLP Service precisa de memória para o modelo BERTimbau. Na VPS com 4GB RAM total, o NLP é o serviço mais pesado. Monitorar uso.
</decisions>

<tasks>

<task>
  <name>Task 1: Setup VPS — script de provisionamento</name>
  <files>
    - infra/scripts/setup-vps.sh (novo)
  </files>
  <action>
    Criar script `infra/scripts/setup-vps.sh` que automatiza a preparação inicial:

    ```bash
    #!/bin/bash
    set -euo pipefail

    # 1. Atualizar sistema
    apt update && apt upgrade -y

    # 2. Instalar Docker + Docker Compose plugin
    apt install -y docker.io docker-compose-v2 git curl ufw

    # 3. Firewall (UFW)
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow ssh
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw --force enable

    # 4. Criar diretório do projeto
    mkdir -p /opt/kanso
    ```

    **Instruções de uso no PLAN.md:**
    1. SSH para a VPS
    2. `sudo ./infra/scripts/setup-vps.sh`
    3. Clonar repositório: `git clone <repo-url> /opt/kanso`
  </action>
  <verify>
    - Script é executável (`chmod +x`)
    - UFW status: active, portas 22, 80, 443 abertas
    - `docker --version` e `docker compose version` funcionam
  </verify>
</task>

<task>
  <name>Task 2: Configurar DNS + Traefik para produção</name>
  <files>
    - infra/traefik/traefik.yml
    - infra/traefik/dynamic.yml
  </files>
  <action>
    **dynamic.yml — Atualizar regras:**
    - Substituir `Host("kanso.local")` por `Host("kanso.edsonajeje.cloud")`
    - Adicionar middleware `redirect-https` no router HTTP (porta 80 → 443)
    - Manter CORS para o domínio de produção: `https://kanso.edsonajeje.cloud`
    - Remover CORS para `http://localhost:5173` (apenas produção)

    **traefik.yml — Adicionar Let's Encrypt:**
    ```yaml
    certificatesResolvers:
      letsencrypt:
        acme:
          email: seu-email@dominio.com  # configurável
          storage: /etc/traefik/acme.json
          httpChallenge:
            entryPoint: web
    ```

    **dynamic.yml — Adicionar TLS com Let's Encrypt:**
    ```yaml
    tls:
      certificates:
        - certFile: /etc/traefik/certs/kanso.local.pem
          keyFile: /etc/traefik/certs/kanso.local-key.pem
      # Remover certificado auto-assinado, adicionar:
      # (Traefik resolve automaticamente via letsencrypt)
    ```

    Atualizar router `api` e `couchdb`:
    ```yaml
    tls:
      certResolver: letsencrypt
    ```

    **Ajuste no docker-compose.yml para Traefik:**
    - Adicionar volume para `acme.json`: `./traefik/acme.json:/etc/traefik/acme.json`
    - Adicionar porta 443 mapeada (já existe)
    - Config para Let's Encrypt com HTTP challenge na porta 80
  </action>
  <verify>
    - `docker compose config` válido
    - DNS `kanso.edsonajeje.cloud` resolvendo para o IP da VPS
    - `curl -I https://kanso.edsonajeje.cloud` retorna 200
    - TLS válido (não auto-assinado)
  </verify>
</task>

<task>
  <name>Task 3: Ajustar docker-compose.yml para produção</name>
  <files>
    - infra/docker-compose.yml
  </files>
  <action>
    Modificações no `docker-compose.yml` para produção:

    1. **Adicionar `network_mode: host` no Traefik** (ou manter bridge com port bind) — manter bridge com `ports: ["80:80", "443:443"]` já está correto.

    2. **Remover volumes de desenvolvimento**:
      - `grpc-certs` volume: em produção, usar TLS já configurado ou manter auto-assinado via script de setup

    3. **Ajustar restart policies** (já estão `unless-stopped` — OK)

    4. **Adicionar `.env.production`** como env_file principal:
      ```yaml
      env_file: ../.env.production
      ```
      (vs `env_file: ../.env` que é usado em dev)

    5. **NLP Service — mem_limit**: A VPS tem 4GB RAM total. NLP com 4g pode matar o sistema. Ajustar:
      ```yaml
      mem_limit: 2g
      ```
      (BERTimbau base ~500MB, com overhead de 1-2GB)

    6. **API — healthcheck** (adicionar):
      ```yaml
      healthcheck:
        test: ["CMD", "wget", "-qO-", "http://localhost:8080/api/health"]
        interval: 30s
        timeout: 5s
        retries: 3
      ```

    7. **Adicionar logging** com rotação:
      ```yaml
      logging:
        driver: "json-file"
        options:
          max-size: "10m"
          max-file: "3"
      ```
      (Aplicar a todos os serviços)

    8. **CouchDB — volume de backup**:
      Adicionar label para identificar volume de backup:
      ```yaml
      volumes:
        couchdb-data:
          labels:
            - "kanso.backup=true"
      ```
  </action>
  <verify>
    - `docker compose config` output válido
    - Nenhuma quebra de compatibilidade com serviços existentes
  </verify>
</task>

<task>
  <name>Task 4: Criar .env.production com secrets reais</name>
  <files>
    - infra/.env.production (novo)
  </files>
  <action>
    Criar `infra/.env.production` com os valores corretos para produção:

    ```bash
    # CouchDB — usar senha forte diferente do dev
    COUCHDB_PASSWORD=<senha-forte-para-producao>

    # JWT — trocar para produção!
    JWT_SECRET=<novo-secret-256-bit>

    # Google OAuth — mesmo Client ID (já configurado para domínios)
    GOOGLE_CLIENT_ID=721313208894-8tk11j8838outa181pbsj9d5fppqku0m.apps.googleusercontent.com

    # Scheduler API Key
    SCHEDULER_API_KEY=<nova-api-key-producao>

    # Backend
    PORT=8080
    COUCHDB_URL=http://couchdb:5984
    COUCHDB_USER=admin

    # Frontend (prefixo VITE_)
    VITE_GOOGLE_CLIENT_ID=721313208894-8tk11j8838outa181pbsj9d5fppqku0m.apps.googleusercontent.com
    VITE_API_URL=/api
    VITE_COUCHDB_URL=/db
    VITE_VAPID_PUBLIC_KEY=<se-houver>
    ```

    **⚠️ Segurança:**
    - `JWT_SECRET` deve ser 256-bit codificado em base64
    - `COUCHDB_PASSWORD` deve ser forte (20+ caracteres)
    - `SCHEDULER_API_KEY` deve ser token aleatório

    **Nunca commit este arquivo.** Adicionar ao `.gitignore`.

    **Google OAuth:** Verificar no Google Cloud Console se o URI de redirecionamento inclui `https://kanso.edsonajeje.cloud` e a origin `https://kanso.edsonajeje.cloud`.
  </action>
  <verify>
    - Arquivo não está versionado (git status mostra untracked)
    - `grep -c "JWT_SECRET" infra/.env.production` > 0
    - `grep -c "COUCHDB_PASSWORD" infra/.env.production` > 0
  </verify>
</task>

<task>
  <name>Task 5: Criar script de deploy</name>
  <files>
    - infra/scripts/deploy.sh (novo)
  </files>
  <action>
    Criar `infra/scripts/deploy.sh`:

    ```bash
    #!/bin/bash
    set -euo pipefail

    # Kanso Deploy Script
    # Uso: ./infra/scripts/deploy.sh
    # Executar na VPS em /opt/kanso

    echo "=== Kanso Deploy ==="

    # 1. Pull latest code
    echo "[1/5] Pulling latest code..."
    git pull origin master

    # 2. Build images
    echo "[2/5] Building Docker images..."
    docker compose -f infra/docker-compose.yml build

    # 3. Stop old containers
    echo "[3/5] Stopping old containers..."
    docker compose -f infra/docker-compose.yml down

    # 4. Start new containers
    echo "[4/5] Starting new containers..."
    docker compose -f infra/docker-compose.yml up -d

    # 5. Health check
    echo "[5/5] Waiting for health check..."
    sleep 10
    curl -f http://localhost:8080/api/health && echo " OK" || echo " FAILED"

    echo "=== Deploy complete ==="
    ```
  </action>
  <verify>
    - Script executável
    - `bash -n infra/scripts/deploy.sh` válido
  </verify>
</task>

<task>
  <name>Task 6: Script de backup do CouchDB</name>
  <files>
    - infra/scripts/backup-couchdb.sh (novo)
  </files>
  <action>
    Criar `infra/scripts/backup-couchdb.sh`:

    ```bash
    #!/bin/bash
    set -euo pipefail

    # Kanso CouchDB Backup
    # Backup do volume Docker couchdb-data para .tar.gz
    # Uso: ./infra/scripts/backup-couchdb.sh
    #
    # Recomendado: cron semanal
    #   sudo crontab -e
    #   0 3 * * 0 /opt/kanso/infra/scripts/backup-couchdb.sh

    BACKUP_DIR="/opt/kanso/backups"
    TIMESTAMP=$(date +%Y-%m-%d_%H%M%S)
    BACKUP_FILE="${BACKUP_DIR}/couchdb-${TIMESTAMP}.tar.gz"
    RETENTION_DAYS=90

    mkdir -p "$BACKUP_DIR"

    echo "[Backup] Stopping CouchDB..."
    docker compose -f /opt/kanso/infra/docker-compose.yml stop couchdb

    echo "[Backup] Creating archive..."
    tar czf "$BACKUP_FILE" -C /var/lib/docker/volumes couchdb-data

    echo "[Backup] Starting CouchDB..."
    docker compose -f /opt/kanso/infra/docker-compose.yml start couchdb

    echo "[Backup] Cleaning old backups (>${RETENTION_DAYS} days)..."
    find "$BACKUP_DIR" -name "couchdb-*.tar.gz" -mtime +${RETENTION_DAYS} -delete

    echo "[Backup] Done: ${BACKUP_FILE}"
    ls -lh "$BACKUP_FILE"
    ```
  </action>
  <verify>
    - Script executável
    - Backup funcional (testar com `bash -n`)
  </verify>
</task>

<task>
  <name>Task 7: Verificação Google OAuth para produção</name>
  <files>
    - (nenhum — configuração no Google Cloud Console)
  </files>
  <action>
    **No Google Cloud Console, verificar:**

    1. **Authorized JavaScript origins:**
       - `https://kanso.edsonajeje.cloud`

    2. **Authorized redirect URIs:**
       - `https://kanso.edsonajeje.cloud` (para o fluxo de redirect que o GIS usa)

    3. **Verificar se o Client ID usado (721313208894-...)**
       já permite essas origins ou se precisa criar um novo Client ID específico para produção.

    4. **IMPORTANTE:** Se for usar um Client ID diferente para produção, atualizar no `.env.production`.

    **Documentar** no PLAN.md: passo-a-passo do que verificar no console.
  </action>
  <verify>
    - Login Google funciona em https://kanso.edsonajeje.cloud
    - Sem erro `redirect_uri_mismatch` ou `origin_mismatch`
  </verify>
</task>

<task>
  <name>Task 8: Teste de produção — validação completa</name>
  <files>
    - (nenhum — validação manual)
  </files>
  <action>
    Validar o fluxo completo em produção:

    1. ✅ Acessar https://kanso.edsonajeje.cloud — carrega sem erro de TLS
    2. ✅ Fazer login com Google — redireciona corretamente
    3. ✅ Criar novo registro — salva no PouchDB local
    4. ✅ Verificar sync — registro aparece no CouchDB (via Fauxton)
    5. ✅ Histórico — registros aparecem na listagem
    6. ✅ NLP — análise automática aparece (chips de emoção)
    7. ✅ Gerar relatório PDF — download funciona
    8. ✅ Logout — funciona e limpa sessão
    9. ✅ Service Worker registrado — PWA instalável
    10. ✅ Offline — app carrega sem internet
    11. ✅ Docker healthchecks — todos verdes (`docker ps`)
    12. ✅ Logs sem erros críticos
  </action>
  <verify>
    - Checklist executado e documentado em SUMMARY.md
    - Nenhum erro no console do navegador
    - `docker ps` mostra todos containers "healthy" ou "Up"
  </verify>
</task>

</tasks>

<deploy_checklist>
## Checklist de Produção

### Antes do Deploy
- [ ] DNS `kanso.edsonajeje.cloud` aponta para o IP da VPS
- [ ] Portas 80 e 443 acessíveis (firewall)
- [ ] SSH funciona
- [ ] Docker e Docker Compose instalados
- [ ] Repositório clonado em `/opt/kanso`
- [ ] `.env.production` criado com secrets
- [ ] Google OAuth configurado com origins de produção
- [ ] `JWT_SECRET` diferente do dev
- [ ] `COUCHDB_PASSWORD` diferente do dev

### No Deploy
- [ ] `git pull origin master`
- [ ] `docker compose build`
- [ ] `docker compose up -d`
- [ ] Healthchecks verdes
- [ ] Teste de login

### Pós-Deploy
- [ ] Agendar backup semanal (cron)
- [ ] Monitorar uso de RAM (`docker stats`)
- [ ] Verificar logs semanalmente
</deploy_checklist>

<verification>
- `docker compose config` é válido
- `curl -I https://kanso.edsonajeje.cloud` retorna 200 com TLS
- Login Google funcional
- Criação e sync de registro funcional
- NLP analysis aparece no card
- Relatório PDF gera sem erro
- Backup script executável e testado
</verification>

<output>
After completion, create `.planning/phases/09-deploy-vps/09-SUMMARY.md`
</output>
