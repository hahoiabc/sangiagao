#!/bin/bash
# Dừng toàn bộ hạ tầng Rice Marketplace
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRAS_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$INFRAS_DIR/docker"
ENV_FILE="$INFRAS_DIR/.env"

echo "🛑 Dừng Rice Marketplace Infrastructure..."

cd "$DOCKER_DIR"
docker compose --env-file "$ENV_FILE" down

echo "✅ Đã dừng tất cả services."
echo "   (Data vẫn được giữ lại trong Docker volumes)"
echo "   Muốn xóa sạch data: ./infras/scripts/reset.sh"
