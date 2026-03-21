#!/bin/bash
# ============================================
# Khởi động toàn bộ hạ tầng Rice Marketplace
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRAS_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$INFRAS_DIR/docker"
ENV_FILE="$INFRAS_DIR/.env"

echo "🚀 Khởi động Rice Marketplace Infrastructure..."
echo "================================================"

# Kiểm tra Docker daemon
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker daemon chưa chạy!"
    echo "   → Mở Docker Desktop trước, rồi chạy lại script này."
    exit 1
fi

echo "✅ Docker daemon đang chạy"

# Kiểm tra .env file
if [ ! -f "$ENV_FILE" ]; then
    echo "❌ File .env không tồn tại!"
    echo "   → Chạy: cp infras/.env.example infras/.env"
    exit 1
fi

echo "✅ File .env tồn tại"

# Khởi động containers
echo ""
echo "📦 Pulling images và khởi động containers..."
cd "$DOCKER_DIR"
docker compose --env-file "$ENV_FILE" up -d

# Đợi healthcheck
echo ""
echo "⏳ Đợi services sẵn sàng..."

wait_for_service() {
    local name=$1
    local container=$2
    local max_wait=60
    local waited=0

    while [ $waited -lt $max_wait ]; do
        status=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "not_found")
        if [ "$status" = "healthy" ]; then
            echo "   ✅ $name: healthy"
            return 0
        fi
        sleep 2
        waited=$((waited + 2))
    done

    echo "   ❌ $name: timeout sau ${max_wait}s"
    return 1
}

wait_for_service "PostgreSQL" "rice_postgres"
wait_for_service "Redis" "rice_redis"
wait_for_service "MongoDB" "rice_mongodb"

echo ""
echo "================================================"
echo "🎉 Hạ tầng đã sẵn sàng!"
echo ""
echo "  PostgreSQL : localhost:5432  (user: rice_user, db: rice_marketplace)"
echo "  Redis      : localhost:6379"
echo "  MongoDB    : localhost:27017 (user: rice_user, db: rice_chat)"
echo ""
echo "  Dừng:       ./infras/scripts/stop.sh"
echo "  Trạng thái: ./infras/scripts/status.sh"
echo "  Test:       ./infras/scripts/test.sh"
echo "================================================"
