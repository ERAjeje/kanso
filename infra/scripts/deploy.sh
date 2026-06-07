#!/bin/bash
# ============================================================
# Kanso — Deploy Script
# Uso: ./infra/scripts/deploy.sh
# Executar na VPS em /opt/kanso
# ============================================================
set -euo pipefail

COMPOSE_BASE="infra/docker-compose.yml"
COMPOSE_OVERRIDE="infra/docker-compose.prod.yml"
COMPOSE_FILES="-f $COMPOSE_BASE -f $COMPOSE_OVERRIDE"
ENV_FILE="infra/.env.production"
LOG_FILE="/tmp/kanso-deploy.log"

echo "=== Kanso Deploy ===" | tee "$LOG_FILE"

cd /opt/kanso

# 1. Pull latest code (frontend dist + configs)
echo "[1/6] Pulling latest code..." | tee -a "$LOG_FILE"
git pull origin master 2>&1 | tee -a "$LOG_FILE"

# 2. Pull images from GCP Artifact Registry
echo "[2/6] Pulling Docker images from GCP Artifact Registry..." | tee -a "$LOG_FILE"
docker compose $COMPOSE_FILES pull 2>&1 | tee -a "$LOG_FILE"

# 3. Stop old containers
echo "[3/6] Stopping old containers..." | tee -a "$LOG_FILE"
docker compose $COMPOSE_FILES down 2>&1 | tee -a "$LOG_FILE"

# 4. Start new containers
echo "[4/6] Starting new containers..." | tee -a "$LOG_FILE"
docker compose $COMPOSE_FILES --env-file "$ENV_FILE" up -d 2>&1 | tee -a "$LOG_FILE"

# 5. Wait and health check
echo "[5/6] Waiting for health checks..." | tee -a "$LOG_FILE"
sleep 15
HEALTH=$(curl -sf http://localhost:8080/api/health || echo "FAILED")
echo "Health: $HEALTH" | tee -a "$LOG_FILE"

# 6. Show container status
echo "[6/6] Container status:" | tee -a "$LOG_FILE"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | tee -a "$LOG_FILE"

echo ""
if [ "$HEALTH" = "FAILED" ]; then
  echo "⚠️  Health check FAILED — check logs: docker compose logs" | tee -a "$LOG_FILE"
  exit 1
fi
echo "=== Deploy complete ===" | tee -a "$LOG_FILE"
