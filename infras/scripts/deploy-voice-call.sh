#!/bin/bash
set -euo pipefail

# Deploy voice call feature — chạy 1 lần trên VPS
# Sử dụng: bash /opt/sangiagao/infras/scripts/deploy-voice-call.sh

cd /opt/sangiagao

echo "=== 1. Git pull ==="
git pull origin main

echo ""
echo "=== 2. Migration 016 (call_logs table) ==="
source .env.production
docker exec -i rice_postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < infras/migrations/016_call_logs.sql 2>&1 || echo "Migration already applied or error (OK if table exists)"

echo ""
echo "=== 3. Deploy backend ==="
bash infras/scripts/quick-deploy.sh backend

echo ""
echo "=== 4. Deploy chat ==="
bash infras/scripts/quick-deploy.sh chat

echo ""
echo "=== 5. Deploy coturn TURN server ==="
bash infras/scripts/quick-deploy.sh coturn

echo ""
echo "=== 6. Firewall — open port 3478 ==="
ufw allow 3478/udp 2>/dev/null || true
ufw allow 3478/tcp 2>/dev/null || true
echo "Port 3478 opened"

echo ""
echo "=== DONE ==="
docker ps --format "table {{.Names}}\t{{.Status}}" | head -15
