#!/bin/bash
##############################################
# Rice Marketplace - System Monitor
# Kiểm tra health tất cả services
# Dùng: bash infras/scripts/monitor.sh
# Cron: */5 * * * * /opt/sangiagao/infras/scripts/monitor.sh >> /var/log/rice_monitor.log 2>&1
##############################################

set -u

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
ERRORS=0

ok()   { echo -e "  ${GREEN}[OK]${NC} $1"; }
warn() { echo -e "  ${YELLOW}[!]${NC} $1"; }
fail() { echo -e "  ${RED}[X]${NC} $1"; ERRORS=$((ERRORS + 1)); }

echo "============================================"
echo "[$TIMESTAMP] System Monitor - Sàn Giá Gạo"
echo "============================================"

# ========================
# 1. Docker Containers
# ========================
echo ""
echo "--- Docker Containers ---"

REQUIRED_CONTAINERS="rice_postgres rice_redis rice_mongodb minio rice_nginx backend admin web rice_chat"

for name in $REQUIRED_CONTAINERS; do
    STATUS=$(docker inspect --format='{{.State.Status}}' "$name" 2>/dev/null)
    if [ "$STATUS" = "running" ]; then
        # Check if healthy (if healthcheck exists)
        HEALTH=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}no-check{{end}}' "$name" 2>/dev/null)
        if [ "$HEALTH" = "healthy" ] || [ "$HEALTH" = "no-check" ]; then
            ok "$name: running"
        else
            warn "$name: running but $HEALTH"
        fi
    elif [ -z "$STATUS" ]; then
        fail "$name: NOT FOUND"
    else
        fail "$name: $STATUS"
    fi
done

# ========================
# 2. Backend Health API
# ========================
echo ""
echo "--- Backend Health ---"

HEALTH_RESPONSE=$(curl -s --max-time 5 http://localhost:8080/health 2>/dev/null)
if [ $? -eq 0 ] && echo "$HEALTH_RESPONSE" | grep -q '"status"'; then
    STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    PG_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"postgres":"[^"]*"' | cut -d'"' -f4)
    REDIS_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"redis":"[^"]*"' | cut -d'"' -f4)

    if [ "$STATUS" = "ok" ]; then
        ok "API: $STATUS (postgres=$PG_STATUS, redis=$REDIS_STATUS)"
    else
        warn "API: $STATUS (postgres=$PG_STATUS, redis=$REDIS_STATUS)"
    fi
else
    fail "API: unreachable"
fi

# ========================
# 3. Website Check
# ========================
echo ""
echo "--- Websites ---"

for URL in "https://sangiagao.vn" "https://admin.sangiagao.vn"; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$URL" 2>/dev/null)
    if [ "$HTTP_CODE" = "200" ]; then
        ok "$URL: $HTTP_CODE"
    elif [ "$HTTP_CODE" = "000" ]; then
        fail "$URL: unreachable"
    else
        warn "$URL: HTTP $HTTP_CODE"
    fi
done

# ========================
# 4. Disk Usage
# ========================
echo ""
echo "--- Disk Usage ---"

DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | tr -d '%')
DISK_TOTAL=$(df -h / | awk 'NR==2 {print $2}')
DISK_AVAIL=$(df -h / | awk 'NR==2 {print $4}')

if [ "$DISK_USAGE" -lt 80 ]; then
    ok "Disk: ${DISK_USAGE}% used (${DISK_AVAIL} free / ${DISK_TOTAL} total)"
elif [ "$DISK_USAGE" -lt 90 ]; then
    warn "Disk: ${DISK_USAGE}% used (${DISK_AVAIL} free) - Consider cleanup"
else
    fail "Disk: ${DISK_USAGE}% used (${DISK_AVAIL} free) - CRITICAL!"
fi

# Docker disk
DOCKER_SIZE=$(docker system df --format '{{.Size}}' 2>/dev/null | head -1)
if [ -n "$DOCKER_SIZE" ]; then
    echo "  Docker images: $DOCKER_SIZE"
fi

BUILD_CACHE=$(docker system df --format '{{.Type}} {{.Size}} {{.Reclaimable}}' 2>/dev/null | grep "Build" | awk '{print $2, "(reclaimable:", $3 ")"}')
if [ -n "$BUILD_CACHE" ]; then
    echo "  Build cache: $BUILD_CACHE"
fi

# ========================
# 5. Backup Status
# ========================
echo ""
echo "--- Backup Status ---"

BACKUP_EXISTS=$(docker exec rice_db_backup ls /backup/postgres/ 2>/dev/null | wc -l)
if [ "$BACKUP_EXISTS" -gt 0 ]; then
    LATEST_PG=$(docker exec rice_db_backup ls -t /backup/postgres/ 2>/dev/null | head -1)
    PG_COUNT=$(docker exec rice_db_backup find /backup/postgres/ -name "*.sql.gz" 2>/dev/null | wc -l | tr -d ' ')
    TOTAL_SIZE=$(docker exec rice_db_backup du -sh /backup/ 2>/dev/null | cut -f1)
    ok "Backups: $PG_COUNT PostgreSQL files, total $TOTAL_SIZE"
    echo "  Latest: $LATEST_PG"
else
    warn "No backups found (db-backup container may not be running)"
fi

# ========================
# 6. Memory Usage
# ========================
echo ""
echo "--- Memory ---"

MEM_TOTAL=$(free -m 2>/dev/null | awk '/^Mem:/ {print $2}')
MEM_USED=$(free -m 2>/dev/null | awk '/^Mem:/ {print $3}')
if [ -n "$MEM_TOTAL" ] && [ -n "$MEM_USED" ]; then
    MEM_PCT=$((MEM_USED * 100 / MEM_TOTAL))
    if [ "$MEM_PCT" -lt 85 ]; then
        ok "Memory: ${MEM_USED}MB / ${MEM_TOTAL}MB (${MEM_PCT}%)"
    else
        warn "Memory: ${MEM_USED}MB / ${MEM_TOTAL}MB (${MEM_PCT}%) - HIGH"
    fi
fi

# ========================
# Summary
# ========================
echo ""
echo "============================================"
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}[$TIMESTAMP] All checks passed${NC}"
else
    echo -e "${RED}[$TIMESTAMP] $ERRORS issue(s) detected!${NC}"
fi
echo "============================================"

exit "$ERRORS"
