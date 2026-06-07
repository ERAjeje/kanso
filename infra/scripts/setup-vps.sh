#!/bin/bash
# ============================================================
# Kanso — VPS Provisioning Script
# Uso: ./infra/scripts/setup-vps.sh
# Executar UMA VEZ no VPS Hostinger (Ubuntu 24.04 LTS)
# ============================================================
set -euo pipefail

# Garante execução com bash (pipefail é bash-specific)
if [ -z "$BASH_VERSION" ]; then
  echo "ERROR: This script must be run with bash, not sh."
  echo "Usage: bash $0"
  exit 1
fi

echo "=== Kanso VPS Setup ==="

# 1. System update
echo "[1/7] Updating system packages..."
apt update && apt upgrade -y

# 2. Install dependencies
echo "[2/7] Installing Docker, Compose, git, curl, ufw..."
apt install -y docker.io docker-compose-v2 git curl ufw

# 3. Enable Docker on boot
echo "[3/7] Enabling Docker service..."
systemctl enable --now docker

# 4. Firewall
echo "[4/7] Configuring UFW firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# 5. Clone repository
echo "[5/7] Cloning repository..."
if [ -d /opt/kanso ]; then
  if [ -d /opt/kanso/.git ]; then
    echo "Repository already cloned. Pulling latest..."
    cd /opt/kanso && git pull origin master
  else
    echo "WARNING: /opt/kanso exists but is not a git repo. Skipping clone."
    echo "         Remove or move it first, then re-run this script."
  fi
else
  git clone https://github.com/edsonaraujo/kanso.git /opt/kanso
fi

# 6. Create acme.json for Let's Encrypt (Traefik)
echo "[6/7] Creating acme.json for Let's Encrypt..."
mkdir -p /opt/kanso/infra/traefik
touch /opt/kanso/infra/traefik/acme.json
chmod 600 /opt/kanso/infra/traefik/acme.json

# 7. Verify installation
echo "[7/7] Verifying..."
docker --version
docker compose version
ufw status verbose
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. cd /opt/kanso"
echo "  2. Create infra/.env.production with secrets"
echo "  3. Copy GCP service account JSON key to infra/secrets/gcp-key.json"
echo "  4. Authenticate Docker with GCP Artifact Registry:"
echo "     cat infra/secrets/gcp-key.json | docker login -u _json_key --password-stdin"
echo "       https://africa-south1-docker.pkg.dev"
echo "  5. Verify DNS: kanso.edsonajeje.cloud -> $(curl -s ifconfig.me || echo '<VPS-IP>')"
echo "  6. Deploy: ./infra/scripts/deploy.sh"
