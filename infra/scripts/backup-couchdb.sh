#!/bin/bash
# ============================================================
# Kanso — CouchDB Backup Script
# Exporta o volume Docker couchdb-data para .tar.gz
#
# Agendar no cron (semanal):
#   sudo crontab -e
#   0 3 * * 0 /opt/kanso/infra/scripts/backup-couchdb.sh
# ============================================================
set -euo pipefail

BACKUP_DIR="/opt/kanso/backups"
TIMESTAMP=$(date +%Y-%m-%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/couchdb-${TIMESTAMP}.tar.gz"
RETENTION_DAYS=90
COMPOSE_DIR="/opt/kanso"
COMPOSE_FILE="infra/docker-compose.yml"

mkdir -p "$BACKUP_DIR"

echo "[Backup $(date)] Starting CouchDB backup..."

# Stop CouchDB for consistent snapshot
echo "  Stopping CouchDB..."
docker compose -f "${COMPOSE_DIR}/${COMPOSE_FILE}" stop couchdb

# Backup the Docker volume
echo "  Creating archive: ${BACKUP_FILE}"
docker run --rm \
  -v couchdb-data:/source:ro \
  -v "$BACKUP_DIR":/backup \
  alpine tar czf "/backup/couchdb-${TIMESTAMP}.tar.gz" -C /source .

# Start CouchDB
echo "  Starting CouchDB..."
docker compose -f "${COMPOSE_DIR}/${COMPOSE_FILE}" start couchdb

# Clean old backups
echo "  Cleaning backups older than ${RETENTION_DAYS} days..."
find "$BACKUP_DIR" -name "couchdb-*.tar.gz" -mtime +${RETENTION_DAYS} -delete

# Report
BACKUP_SIZE=$(ls -lh "$BACKUP_FILE" | awk '{print $5}')
echo "[Backup $(date)] Done: ${BACKUP_FILE} (${BACKUP_SIZE})"
