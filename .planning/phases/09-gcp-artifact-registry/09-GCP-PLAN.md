---
plan_id: 09-gcp-artifact-registry
type: execute
depends_on: [09-deploy-vps]
files_modified:
  - infra/docker-compose.yml
  - infra/scripts/deploy.sh
  - Makefile
files_created:
  - infra/docker-compose.prod.yml
  - backend/.dockerignore
  - scheduler/.dockerignore
autonomous: false
must_haves:
  truths:
    - Custom images (api, scheduler, nlp, chromedp) are published to africa-south1-docker.pkg.dev/kanso-496617/kanso-repo
    - VPS pulls images from GCP Artifact Registry instead of building locally
    - Local build and push is a single Makefile command
    - docker-compose.prod.yml overrides build: with image: for production
    - VPS authenticates via service account JSON key
  artifacts:
    - path: Makefile
      provides: "docker-publish target for build+tag+push all images"
    - path: infra/docker-compose.prod.yml
      provides: "Production compose override pointing image: to GCP"
    - path: infra/scripts/deploy.sh
      provides: "Updated to docker compose pull instead of build"
---

<objective>
Migrar o deploy das imagens custom do Kanso de "build na VPS" para "build local + push para GCP Artifact Registry + pull na VPS".

**Registry:** `africa-south1-docker.pkg.dev/kanso-496617/kanso-repo`
**Imagens publicadas:** `api`, `scheduler`, `nlp`, `chromedp` (frontend mantém bind mount com nginx:alpine)
**Tag:** `latest`
**Build/Push:** Manual via Makefile na máquina local
**Deploy VPS:** `docker compose pull` em vez de `docker compose build`
</objective>

<execution_context>
@infra/docker-compose.yml
@infra/scripts/deploy.sh
@infra/scripts/setup-vps.sh
@Makefile
@backend/Dockerfile
@scheduler/Dockerfile
@nlp-service/Dockerfile
@infra/chromedp/Dockerfile
</execution_context>

<decisions>

### D-01: Substitui decisão D-01 do 09-PLAN.md
Antes: build na VPS. Agora: build local → push GCP → pull na VPS.

### D-02: docker-compose.prod.yml separado
Criar arquivo de override em vez de modificar o docker-compose.yml original. O `docker-compose.prod.yml` sobrescreve apenas os serviços com `image:` apontando para o GCP. O compose original mantém `build:` para dev.

### D-03: Tag latest
Simplicidade. Builds manuais → tag `latest` sempre. Se no futuro precisar de versionamento, adicionamos SHA ou semver.

### D-04: Frontend fora do registry
Frontend continua com `nginx:alpine` + bind mount do `dist/`. O build do frontend é feito localmente e o diretório `dist/` vai versionado ou copiado manualmente.

### D-05: Service account `kanso-puller` no GCP
Criar service account dedicada com permissão mínima de `artifactregistry.reader`. JSON key copiada para a VPS.

</decisions>

<tasks>

<task>
  <name>Task 1: Adicionar .dockerignore para backend e scheduler</name>
  <files>
    - backend/.dockerignore (novo)
    - scheduler/.dockerignore (novo)
  </files>
  <action>
    Criar `.dockerignore` para evitar enviar contexto desnecessário para o Docker daemon.

    **backend/.dockerignore:**
    ```
    __pycache__
    *.pyc
    .pytest_cache
    tests
    .git
    .gitignore
    *.md
    .env
    ```

    **scheduler/.dockerignore:**
    ```
    .git
    .gitignore
    *.md
    .env
    tests
    ```
  </action>
  <verify>
    - Arquivos criados e ignorando diretórios corretos
    - `docker build` no backend/scheduler com contexto reduzido
  </verify>
</task>

<task>
  <name>Task 2: Autenticação local com GCP</name>
  <files>
    - (nenhum — comando único)
  </files>
  <action>
    Configurar autenticação Docker para o Artifact Registry:

    ```bash
    gcloud auth login
    gcloud config set project kanso-496617
    gcloud auth configure-docker africa-south1-docker.pkg.dev
    ```

    Verificar funcionamento:
    ```bash
    docker pull africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/hello-world  # se existir
    # ou apenas verificar que o auth está configurado:
    cat ~/.docker/config.json | grep africa-south1-docker.pkg.dev
    ```
  </action>
  <verify>
    - `docker login` configurado para o registry
    - `cat ~/.docker/config.json` contém entrada para `africa-south1-docker.pkg.dev`
  </verify>
</task>

<task>
  <name>Task 3: Adicionar targets no Makefile para build + tag + push</name>
  <files>
    - Makefile
  </files>
  <action>
    Adicionar ao `Makefile`:

    ```makefile
    REGISTRY = africa-south1-docker.pkg.dev/kanso-496617/kanso-repo
    TAG = latest

    # Build todas as imagens
    docker-build:
        docker build -t $(REGISTRY)/api:$(TAG) -f backend/Dockerfile --target runtime backend
        docker build -t $(REGISTRY)/scheduler:$(TAG) -f scheduler/Dockerfile scheduler
        docker build -t $(REGISTRY)/nlp:$(TAG) -f nlp-service/Dockerfile nlp-service
        docker build -t $(REGISTRY)/chromedp:$(TAG) -f infra/chromedp/Dockerfile infra/chromedp

    # Push todas as imagens
    docker-push:
        docker push $(REGISTRY)/api:$(TAG)
        docker push $(REGISTRY)/scheduler:$(TAG)
        docker push $(REGISTRY)/nlp:$(TAG)
        docker push $(REGISTRY)/chromedp:$(TAG)

    # Build + Push em um comando
    docker-publish: docker-build docker-push
        @echo "Publicado: $(REGISTRY)/{api,scheduler,nlp,chromedp}:$(TAG)"

    # (Opcional) Verificar se as imagens estão no registry
    docker-verify:
        @echo "Images published:"
        @for img in api scheduler nlp chromedp; do \
            docker manifest inspect $(REGISTRY)/$$img:$(TAG) > /dev/null 2>&1 && \
            echo "  ✓ $$img:$(TAG)" || echo "  ✗ $$img:$(TAG)"; \
        done
    ```
  </action>
  <verify>
    - `make docker-build` executa sem erro
    - `make docker-publish` executa build + push
    - `make docker-verify` mostra todas as imagens como ✓
  </verify>
</task>

<task>
  <name>Task 4: Criar docker-compose.prod.yml</name>
  <files>
    - infra/docker-compose.prod.yml (novo)
  </files>
  <action>
    Criar `infra/docker-compose.prod.yml` que sobrescreve os serviços do `docker-compose.yml` para usar imagens do GCP em vez de build local:

    ```yaml
    # Kanso Production Override
    # Uso: docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
    #
    # Sobrescreve build: → image: para serviços custom
    # Imagens hospedadas em africa-south1-docker.pkg.dev/kanso-496617/kanso-repo

    services:
      chromedp:
        image: africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/chromedp:latest
        build: !reset null

      api:
        image: africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/api:latest
        build: !reset null

      scheduler:
        image: africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/scheduler:latest
        build: !reset null

      nlp:
        image: africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/nlp:latest
        build: !reset null
    ```

    **Nota:** `!reset null` limpa a chave `build:` herdada do compose principal. Se o Docker Compose da VPS não suportar YAML merge, usar:

    ```yaml
    services:
      chromedp:
        image: ...
      api:
        image: ...
      scheduler:
        image: ...
      nlp:
        image: ...
    ```

    (O build do compose original é ignorado quando image está presente no mesmo nível — mas o `!reset null` é mais explícito.)

    Frontend e CouchDB/Traefik não entram aqui — usam `nginx:alpine`, `couchdb:3.5` e `traefik:v3` diretamente do Docker Hub.
  </action>
  <verify>
    - `docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml config` não mostra chave `build` para nenhum serviço
    - Todos os serviços custom têm `image:` apontando para o GCP
  </verify>
</task>

<task>
  <name>Task 5: Atualizar deploy.sh para usar pull do GCP</name>
  <files>
    - infra/scripts/deploy.sh
  </files>
  <action>
    Modificar o script de deploy para usar `pull` em vez de `build`:

    ```bash
    #!/bin/bash
    set -euo pipefail

    # Kanso Deploy Script
    # Uso: ./infra/scripts/deploy.sh
    # Executar na VPS em /opt/kanso

    echo "=== Kanso Deploy ==="

    # 1. Pull latest code (para frontend dist e configs)
    echo "[1/5] Pulling latest code..."
    git pull origin master

    # 2. Pull images from GCP Artifact Registry
    echo "[2/5] Pulling Docker images from GCP..."
    docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml pull

    # 3. Stop old containers
    echo "[3/5] Stopping old containers..."
    docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml down

    # 4. Start new containers
    echo "[4/5] Starting new containers..."
    docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml up -d

    # 5. Health check
    echo "[5/5] Waiting for health check..."
    sleep 10
    curl -f http://localhost:8080/api/health && echo " OK" || echo " FAILED"

    echo "=== Deploy complete ==="
    ```
  </action>
  <verify>
    - Script executável (`chmod +x`)
    - `bash -n infra/scripts/deploy.sh` válido
  </verify>
</task>

<task>
  <name>Task 6: Service account + autenticação na VPS</name>
  <files>
    - (nenhum — instruções para execução manual)
  </files>
  <action>
    **No Google Cloud Console:**

    1. Criar service account `kanso-puller` no projeto `kanso-496617`
    2. Conceder papel `roles/artifactregistry.reader` no repositório `kanso-repo`
    3. Gerar e baixar JSON key (`kanso-puller-key.json`)

    **Na VPS:**

    1. Copiar JSON key para a VPS:
       ```bash
       scp kanso-puller-key.json root@<vps-ip>:/opt/kanso/infra/secrets/gcp-key.json
       ```

    2. Configurar Docker login com a service account:
       ```bash
       cat /opt/kanso/infra/secrets/gcp-key.json | docker login -u _json_key --password-stdin \
         https://africa-south1-docker.pkg.dev
       ```

    3. Verificar:
       ```bash
       docker pull africa-south1-docker.pkg.dev/kanso-496617/kanso-repo/api:latest
       ```

    4. (Opcional) Adicionar ao crontab para renovar o login periódico se necessário.

    **Segurança:**
    - A JSON key tem permissão **somente leitura** no Artifact Registry
    - Não comitar a key no git (adicionar `infra/secrets/` ao `.gitignore`)
    - A key fica apenas na VPS
  </action>
  <verify>
    - VPS consegue `docker pull` do GCP sem erro
    - `docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml pull` funcional
  </verify>
</task>

<task>
  <name>Task 7: Atualizar setup-vps.sh com instalação do Docker Credential Helper</name>
  <files>
    - infra/scripts/setup-vps.sh
  </files>
  <action>
    Adicionar ao `setup-vps.sh` a instalação do `docker-credential-gcr` (ou `docker-credential-gcloud`) para que a VPS consiga autenticar no Artifact Registry sem depender de `docker login` manual:

    ```bash
    # 5. Instalar docker-credential-gcr para GCP Artifact Registry
    # (opcional — pode usar docker login com JSON key)
    ```

    **Decisão:** Como a VPS já tem Docker instalado e o `docker login` com JSON key é mais simples (não precisa instalar SDK), **manter `docker login` com JSON key** como método principal. Apenas documentar no PLAN.md.
  </action>
  <verify>
    - (N/A — apenas documentação)
  </verify>
</task>

<task>
  <name>Task 8: Primeiro publish — build + push manual</name>
  <files>
    - (nenhum — execução manual)
  </files>
  <action>
    Executar na máquina local:

    ```bash
    # 1. Verificar autenticação
    gcloud auth configure-docker africa-south1-docker.pkg.dev

    # 2. Build + Push
    make docker-publish

    # 3. Verificar
    make docker-verify

    # 4. (Opcional) Listar imagens no GCP
    gcloud artifacts docker images list \
      africa-south1-docker.pkg.dev/kanso-496617/kanso-repo
    ```
  </action>
  <verify>
    - `make docker-publish` executa sem erro
    - Imagens aparecem no `gcloud artifacts docker images list`
    - `docker manifest inspect` retorna success para cada imagem
  </verify>
</task>

</tasks>

<verification>
- `make docker-publish` → 4 imagens no GCP Artifact Registry
- `docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml config` sem erros
- `infra/scripts/deploy.sh` executa `pull` em vez de `build`
- VPS consegue `docker pull` do GCP autenticado com service account
- `docker compose up -d` na VPS usando imagens do GCP funciona
- Healthcheck passa após deploy
</verification>

<output>
After execution, create `.planning/phases/09-gcp-artifact-registry/09-GCP-SUMMARY.md`
</output>
