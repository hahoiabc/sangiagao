#!/bin/bash
# Kiểm tra trạng thái hạ tầng Rice Marketplace
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRAS_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$INFRAS_DIR/docker"
ENV_FILE="$INFRAS_DIR/.env"

echo "📊 Trạng thái Rice Marketplace Infrastructure"
echo "================================================"

cd "$DOCKER_DIR"
docker compose --env-file "$ENV_FILE" ps

echo ""
echo "--- Health Status ---"
for container in rice_postgres rice_redis rice_mongodb; do
    status=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "not running")
    printf "  %-15s : %s\n" "$container" "$status"
done

echo ""
echo "--- Resource Usage ---"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" rice_postgres rice_redis rice_mongodb 2>/dev/null || echo "  (containers not running)"
