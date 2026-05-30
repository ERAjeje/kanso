#!/bin/bash
# ============================================================
# Kanso — VPS Provisioning Script
# Uso: sudo ./infra/scripts/setup-vps.sh
# Executar UMA VEZ no VPS Hostinger (Ubuntu 24.04 LTS)
# ============================================================
set -euo pipefail

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

# 5. Create project directory
echo "[5/7] Creating /opt/kanso..."
mkdir -p /opt/kanso

# 6. Create acme.json for Let's Encrypt (Traefik)
echo "[6/8] Creating acme.json for Let's Encrypt..."
touch /opt/kanso/infra/traefik/acme.json
chmod 600 /opt/kanso/infra/traefik/acme.json

# 7. Clone repository
echo "[7/8] Cloning repository..."
if [ ! -d /opt/kanso/.git ]; then
  git clone https://github.com/edsonaraujo/kanso.git /opt/kanso
else
  echo "Repository already cloned. Pulling latest..."
  cd /opt/kanso && git pull origin master
fi

# 8. Verify installation
echo "[8/8] Verifying..."
docker --version
docker compose version
ufw status verbose
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. cd /opt/kanso"
echo "  2. Create infra/.env.production with secrets"
echo "  3. Verify DNS: kanso.edsonajeje.cloud -> $(curl -s ifconfig.me || echo '<VPS-IP>')"
echo "  4. docker compose -f infra/docker-compose.yml up -d"
