#!/bin/bash
set -euo pipefail

# ============================================
# Quick Deploy — Sàn Giá Gạo
# ============================================
# Deploy từng service hoặc tất cả
# Sử dụng: bash infras/scripts/quick-deploy.sh [backend|web|admin|all]
#
# Script tự động:
# 1. git pull
# 2. Build image --no-cache
# 3. Stop → rm → run container mới (đúng pattern)
# 4. Restart nginx
# 5. Verify health + tất cả containers
# 6. Dọn build cache
# ============================================

PROJECT_ROOT="/opt/sangiagao"
NETWORK="rice_internal"
ENV_BACKEND="$PROJECT_ROOT/infras/.env.backend"
ENV_CHAT="$PROJECT_ROOT/infras/.env.chat"
FIREBASE_CRED="$PROJECT_ROOT/infras/firebase-credentials.json"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
fail() { echo -e "${RED}[X]${NC} $1"; }

ERRORS=0

# --- Helpers ---
deploy_service() {
    local name=$1 image=$2 build_dir=$3 port=$4
    shift 4
    local extra_args=("$@")

    warn "Building $name (--no-cache)..."
    if ! docker build --no-cache -t "$image" "$PROJECT_ROOT/$build_dir"; then
        fail "Build $name FAILED"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
    ok "Build $name OK"

    warn "Deploying $name..."
    docker stop "$name" 2>/dev/null || true
    docker rm "$name" 2>/dev/null || true
    docker run -d \
        --name "$name" \
        --hostname "$name" \
        --network "$NETWORK" \
        -p "$port" \
        "${extra_args[@]}" \
        "$image"
    ok "Container $name started"
}

verify() {
    echo ""
    warn "Verifying..."

    # Restart nginx (clear cached IPs)
    docker restart rice_nginx 2>/dev/null || warn "rice_nginx restart failed"
    sleep 2

    # Check all containers
    local required=("backend" "web" "admin" "chat" "rice_postgres" "rice_redis" "rice_mongodb" "minio" "rice_nginx")
    for c in "${required[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${c}$"; then
            ok "$c: running"
        else
            fail "$c: NOT running"
            ERRORS=$((ERRORS + 1))
        fi
    done

    # Health check (direct, bypass nginx)
    sleep 3
    local health
    health=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000")
    if [ "$health" = "200" ]; then
        ok "Backend health: OK"
    else
        fail "Backend health: HTTP $health"
        docker logs backend --tail 10 2>/dev/null
        ERRORS=$((ERRORS + 1))
    fi

    # Web check
    local web_code
    web_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3001 2>/dev/null || echo "000")
    if [ "$web_code" = "200" ] || [ "$web_code" = "307" ]; then
        ok "Web: HTTP $web_code"
    else
        fail "Web: HTTP $web_code"
        ERRORS=$((ERRORS + 1))
    fi

    # Admin check
    local admin_code
    admin_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 2>/dev/null || echo "000")
    if [ "$admin_code" = "200" ] || [ "$admin_code" = "307" ]; then
        ok "Admin: HTTP $admin_code"
    else
        fail "Admin: HTTP $admin_code"
        ERRORS=$((ERRORS + 1))
    fi

    # Cleanup
    warn "Cleaning up..."
    docker image prune -f > /dev/null 2>&1
    docker builder prune -f > /dev/null 2>&1
    ok "Cleanup done"

    # Disk check
    local disk_pct
    disk_pct=$(df / --output=pcent | tail -1 | tr -d ' %')
    if [ "$disk_pct" -gt 80 ]; then
        warn "Disk: ${disk_pct}% used — consider running: docker builder prune -af"
    else
        ok "Disk: ${disk_pct}% used"
    fi

    echo ""
    if [ "$ERRORS" -eq 0 ]; then
        ok "Deploy completed successfully!"
    else
        fail "Deploy completed with $ERRORS error(s)"
    fi
}

# --- Main ---
cd "$PROJECT_ROOT"

# Step 1: Pull
warn "Pulling latest code..."
git pull origin main
ok "Code: $(git log --oneline -1)"
echo ""

TARGET="${1:-help}"

case "$TARGET" in
    backend)
        deploy_service "backend" "sangiagao-backend" "backend" "8080:8080" \
            --env-file "$ENV_BACKEND" \
            -v "$FIREBASE_CRED:/app/firebase-credentials.json:ro"
        verify
        ;;
    web)
        deploy_service "web" "sangiagao-web" "web" "3001:3001" \
            -e "NODE_ENV=production"
        verify
        ;;
    admin)
        deploy_service "admin" "sangiagao-admin" "admin" "3000:3000" \
            -e "NODE_ENV=production"
        verify
        ;;
    chat)
        deploy_service "chat" "sangiagao-chat" "chat" "4000:4000" \
            --env-file "$ENV_CHAT"
        verify
        ;;
    migrate)
        warn "Running pending migrations..."
        source "$PROJECT_ROOT/.env.production"
        for f in "$PROJECT_ROOT"/infras/migrations/*.sql; do
            fname=$(basename "$f")
            warn "Applying $fname..."
            docker exec -i rice_postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$f" 2>&1 || true
        done
        ok "Migrations done"
        ;;
    all)
        deploy_service "backend" "sangiagao-backend" "backend" "8080:8080" \
            --env-file "$ENV_BACKEND" \
            -v "$FIREBASE_CRED:/app/firebase-credentials.json:ro"
        deploy_service "chat" "sangiagao-chat" "chat" "4000:4000" \
            --env-file "$ENV_CHAT"
        deploy_service "web" "sangiagao-web" "web" "3001:3001" \
            -e "NODE_ENV=production"
        deploy_service "admin" "sangiagao-admin" "admin" "3000:3000" \
            -e "NODE_ENV=production"
        verify
        ;;
    *)
        echo "Quick Deploy — Sàn Giá Gạo"
        echo ""
        echo "Sử dụng: bash infras/scripts/quick-deploy.sh [command]"
        echo ""
        echo "Commands:"
        echo "  backend    Deploy backend only"
        echo "  web        Deploy web only"
        echo "  admin      Deploy admin only"
        echo "  chat       Deploy chat (Elixir) only"
        echo "  migrate    Run all SQL migrations"
        echo "  all        Deploy backend + chat + web + admin"
        echo ""
        echo "Ví dụ:"
        echo "  bash infras/scripts/quick-deploy.sh all"
        echo "  bash infras/scripts/quick-deploy.sh backend"
        ;;
esac

exit $ERRORS
