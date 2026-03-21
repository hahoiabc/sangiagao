#!/bin/bash
set -euo pipefail

# ============================================
# Rice Marketplace - Production Deploy Script
# ============================================
# Tất cả services chạy trong Docker containers
# Sử dụng: bash infras/scripts/deploy.sh [command]

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.production.yml"
ENV_FILE="$PROJECT_ROOT/.env.production"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()   { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[X]${NC} $1"; exit 1; }

DC="docker compose -f $COMPOSE_FILE --env-file $ENV_FILE"

# --- Kiem tra dieu kien ---
check_env() {
    if [ ! -f "$ENV_FILE" ]; then
        error "Chua co .env.production! Copy tu .env.production.example:\n  cp $PROJECT_ROOT/.env.production.example $ENV_FILE"
    fi
    if grep -q "THAY_DOI" "$ENV_FILE"; then
        error "Van con gia tri mac dinh 'THAY_DOI_*' trong .env.production!"
    fi
    if grep -q "YOUR_DOMAIN" "$ENV_FILE"; then
        warn "Van con 'YOUR_DOMAIN' trong .env.production. Hay thay bang domain that."
    fi
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker chua cai dat"
    fi
    if ! docker info &> /dev/null; then
        error "Docker daemon chua chay"
    fi
}

# === SETUP (lan dau tren VPS) ===
cmd_setup() {
    echo "========================================="
    echo "  Rice Marketplace - Setup Production"
    echo "========================================="

    # Cai Docker
    if ! command -v docker &> /dev/null; then
        warn "Cai dat Docker..."
        curl -fsSL https://get.docker.com | sh
        sudo usermod -aG docker "$USER"
        log "Docker da cai. Logout va login lai."
    else
        log "Docker da co: $(docker --version)"
    fi

    # Firewall
    if command -v ufw &> /dev/null; then
        sudo ufw allow 80/tcp
        sudo ufw allow 443/tcp
        sudo ufw allow 22/tcp
        log "Firewall: chi mo port 80, 443, 22"
    fi

    # Tao .env.production neu chua co
    if [ ! -f "$ENV_FILE" ]; then
        cp "$PROJECT_ROOT/.env.production.example" "$ENV_FILE"
        warn "Da tao .env.production - HAY THAY DOI TAT CA MAT KHAU TRUOC KHI DEPLOY!"
    fi

    echo ""
    log "Setup xong! Buoc tiep theo:"
    echo "  1. Sua .env.production (thay het mat khau)"
    echo "  2. Sua domain trong infras/configs/nginx/nginx.prod.conf"
    echo "  3. bash infras/scripts/deploy.sh up"
    echo "  4. bash infras/scripts/deploy.sh ssl yourdomain.com"
}

# === KHOI DONG ===
cmd_up() {
    check_docker
    check_env

    echo "========================================="
    echo "  Rice Marketplace - Starting Production"
    echo "========================================="

    warn "Building va khoi dong tat ca services..."
    $DC up -d --build

    warn "Cho services khoi dong..."
    sleep 10

    echo ""
    log "=== Trang thai services ==="
    $DC ps
    echo ""
    log "Deploy thanh cong!"
    echo ""
    echo "  Xem logs:      bash $0 logs [service]"
    echo "  Dung:           bash $0 down"
    echo "  Backup ngay:    bash $0 backup"
    echo "  Trang thai:     bash $0 status"
}

# === DUNG ===
cmd_down() {
    check_docker
    warn "Dung tat ca services..."
    $DC down
    log "Da dung tat ca"
}

# === RESTART ===
cmd_restart() {
    SERVICE="${1:-}"
    check_docker
    if [ -n "$SERVICE" ]; then
        warn "Restart $SERVICE..."
        $DC restart "$SERVICE"
        log "$SERVICE da restart"
    else
        warn "Restart tat ca services..."
        $DC down
        $DC up -d --build
        sleep 10
        $DC ps
        log "Restart thanh cong!"
    fi
}

# === TRANG THAI ===
cmd_status() {
    check_docker
    echo ""
    echo "=== Docker Containers ==="
    $DC ps 2>/dev/null || echo "(chua chay)"

    echo ""
    echo "=== Disk usage ==="
    docker system df 2>/dev/null

    echo ""
    echo "=== Backup status ==="
    if docker exec rice_db_backup ls /backup/postgres/ 2>/dev/null; then
        PG_COUNT=$(docker exec rice_db_backup find /backup/postgres/ -name "*.sql.gz" 2>/dev/null | wc -l)
        MONGO_COUNT=$(docker exec rice_db_backup find /backup/mongo/ -maxdepth 1 -type d -name "rice_chat_*" 2>/dev/null | wc -l)
        TOTAL=$(docker exec rice_db_backup du -sh /backup/ 2>/dev/null | cut -f1)
        echo "  PostgreSQL: $PG_COUNT ban backup"
        echo "  MongoDB:    $MONGO_COUNT ban backup"
        echo "  Tong:       $TOTAL"
    else
        warn "Chua co backup nao"
    fi
}

# === BACKUP THU CONG ===
cmd_backup() {
    check_docker
    warn "Chay backup thu cong..."
    docker exec rice_db_backup /backup.sh
    log "Backup hoan tat!"
}

# === LOGS ===
cmd_logs() {
    check_docker
    SERVICE="${1:-}"
    if [ -n "$SERVICE" ]; then
        $DC logs -f --tail 100 "$SERVICE"
    else
        $DC logs -f --tail 50
    fi
}

# === SSL ===
cmd_ssl() {
    check_docker
    DOMAIN="${1:-}"
    if [ -z "$DOMAIN" ]; then
        error "Thieu domain. Vi du: bash $0 ssl yourdomain.com"
    fi

    warn "Thay nginx.prod.conf YOUR_DOMAIN bang $DOMAIN..."
    sed -i "s/YOUR_DOMAIN.com/$DOMAIN/g" "$PROJECT_ROOT/infras/configs/nginx/nginx.prod.conf"

    warn "Tao SSL certificate cho $DOMAIN..."
    $DC run --rm certbot \
        certbot certonly --webroot -w /var/www/certbot \
        -d "$DOMAIN" -d "admin.$DOMAIN" \
        --email "admin@$DOMAIN" --agree-tos --no-eff-email

    warn "Restart nginx..."
    $DC restart nginx
    log "SSL certificate da tao cho $DOMAIN!"
}

# === UPDATE (pull code moi va rebuild) ===
cmd_update() {
    check_docker
    check_env

    cd "$PROJECT_ROOT"
    if git rev-parse --git-dir > /dev/null 2>&1; then
        warn "Pull code moi nhat..."
        git pull origin main
        log "Code: $(git log --oneline -1)"
    fi

    warn "Rebuild va restart services..."
    $DC up -d --build
    sleep 10
    $DC ps
    log "Update thanh cong!"
}

# === MAIN ===
case "${1:-help}" in
    setup)     cmd_setup ;;
    up|start)  cmd_up ;;
    down|stop) cmd_down ;;
    restart)   cmd_restart "${2:-}" ;;
    status)    cmd_status ;;
    backup)    cmd_backup ;;
    logs)      cmd_logs "${2:-}" ;;
    ssl)       cmd_ssl "${2:-}" ;;
    update)    cmd_update ;;
    *)
        echo "Rice Marketplace - Production Deploy Tool"
        echo ""
        echo "Cach dung: bash infras/scripts/deploy.sh [command]"
        echo ""
        echo "Commands:"
        echo "  setup          Cai dat lan dau tren VPS (Docker, firewall)"
        echo "  up             Build va khoi dong tat ca services"
        echo "  down           Dung tat ca services"
        echo "  restart [svc]  Restart tat ca hoac 1 service"
        echo "  status         Kiem tra trang thai va backup"
        echo "  backup         Chay backup PostgreSQL + MongoDB thu cong"
        echo "  logs [svc]     Xem logs (backend|chat|admin|postgres|redis|mongodb|nginx)"
        echo "  ssl domain     Tao SSL certificate (Let's Encrypt)"
        echo "  update         Pull code moi, rebuild va deploy"
        ;;
esac
