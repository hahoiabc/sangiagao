#!/bin/bash
# ============================================
# Test kết nối tất cả services
# ============================================
set -e

echo "🧪 Test kết nối Rice Marketplace Infrastructure"
echo "================================================"

PASS=0
FAIL=0

# --- Test PostgreSQL ---
echo ""
echo "1️⃣  PostgreSQL (localhost:5432)"
if docker exec rice_postgres psql -U rice_user -d rice_marketplace -c "SELECT 1;" > /dev/null 2>&1; then
    echo "   ✅ Kết nối OK"

    # Kiểm tra tables
    TABLES=$(docker exec rice_postgres psql -U rice_user -d rice_marketplace -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';")
    TABLES=$(echo "$TABLES" | tr -d ' ')
    echo "   ✅ Số tables: $TABLES"

    # Kiểm tra admin user
    ADMIN=$(docker exec rice_postgres psql -U rice_user -d rice_marketplace -t -c "SELECT name FROM users WHERE role='admin' LIMIT 1;")
    ADMIN=$(echo "$ADMIN" | tr -d ' ')
    if [ -n "$ADMIN" ]; then
        echo "   ✅ Admin user: $ADMIN"
    else
        echo "   ⚠️  Chưa có admin user"
    fi

    # Kiểm tra extensions
    EXTS=$(docker exec rice_postgres psql -U rice_user -d rice_marketplace -t -c "SELECT extname FROM pg_extension WHERE extname IN ('uuid-ossp', 'pg_trgm');")
    echo "   ✅ Extensions:$(echo "$EXTS" | tr '\n' ',' | sed 's/,$//')"

    PASS=$((PASS + 1))
else
    echo "   ❌ Kết nối THẤT BẠI"
    FAIL=$((FAIL + 1))
fi

# --- Test Redis ---
echo ""
echo "2️⃣  Redis (localhost:6379)"
if docker exec rice_redis redis-cli ping | grep -q "PONG"; then
    echo "   ✅ Kết nối OK (PONG)"

    # Test set/get
    docker exec rice_redis redis-cli SET rice_test "hello" > /dev/null 2>&1
    VAL=$(docker exec rice_redis redis-cli GET rice_test 2>/dev/null)
    if [ "$VAL" = "hello" ]; then
        echo "   ✅ SET/GET hoạt động"
    fi
    docker exec rice_redis redis-cli DEL rice_test > /dev/null 2>&1

    PASS=$((PASS + 1))
else
    echo "   ❌ Kết nối THẤT BẠI"
    FAIL=$((FAIL + 1))
fi

# --- Test MongoDB ---
echo ""
echo "3️⃣  MongoDB (localhost:27017)"
if docker exec rice_mongodb mongosh --quiet -u rice_user -p rice_secret_dev --authenticationDatabase admin --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
    echo "   ✅ Kết nối OK"

    # Kiểm tra collections
    COLLS=$(docker exec rice_mongodb mongosh --quiet -u rice_user -p rice_secret_dev --authenticationDatabase admin rice_chat --eval "db.getCollectionNames().join(', ')" 2>/dev/null)
    echo "   ✅ Collections: $COLLS"

    # Kiểm tra indexes
    MSG_IDX=$(docker exec rice_mongodb mongosh --quiet -u rice_user -p rice_secret_dev --authenticationDatabase admin rice_chat --eval "db.messages.getIndexes().length" 2>/dev/null)
    echo "   ✅ Messages indexes: $MSG_IDX"

    PASS=$((PASS + 1))
else
    echo "   ❌ Kết nối THẤT BẠI"
    FAIL=$((FAIL + 1))
fi

# --- Summary ---
echo ""
echo "================================================"
TOTAL=$((PASS + FAIL))
if [ $FAIL -eq 0 ]; then
    echo "🎉 KẾT QUẢ: $PASS/$TOTAL services PASSED"
    echo "   Hạ tầng sẵn sàng cho development!"
else
    echo "⚠️  KẾT QUẢ: $PASS/$TOTAL passed, $FAIL/$TOTAL FAILED"
    echo "   Kiểm tra lại containers: docker compose -f infras/docker/docker-compose.yml logs"
fi
echo "================================================"

exit $FAIL
