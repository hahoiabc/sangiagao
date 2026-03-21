#!/bin/bash
# ============================================
# XÓA SẠCH dữ liệu và tạo lại hạ tầng từ đầu
# ⚠️ CẢNH BÁO: Sẽ mất toàn bộ data trong databases!
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRAS_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$INFRAS_DIR/docker"
ENV_FILE="$INFRAS_DIR/.env"

echo "⚠️  CẢNH BÁO: Sẽ xóa toàn bộ data (PostgreSQL, Redis, MongoDB)!"
read -p "   Bạn chắc chắn? (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "   Đã hủy."
    exit 0
fi

echo ""
echo "🗑️  Xóa containers và volumes..."
cd "$DOCKER_DIR"
docker compose --env-file "$ENV_FILE" down -v

echo ""
echo "🚀 Khởi động lại..."
"$SCRIPT_DIR/start.sh"
