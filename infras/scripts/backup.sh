#!/bin/sh
##############################################
# Rice Marketplace - Database Backup Script
# Chạy tự động hàng ngày lúc 02:00 AM
# Lưu tại: /backup/postgres/ và /backup/mongo/
# Tự xóa file backup cũ hơn BACKUP_RETENTION_DAYS ngày
##############################################

set -e

# Trap errors and log them
trap 'echo "[ERROR] Backup thất bại tại dòng $LINENO. Exit code: $?" >&2' ERR

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DATE_LABEL=$(date +"%Y-%m-%d %H:%M:%S")
PG_BACKUP_DIR="/backup/postgres"
MONGO_BACKUP_DIR="/backup/mongo"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

mkdir -p "$PG_BACKUP_DIR" "$MONGO_BACKUP_DIR"

echo "============================================"
echo "[$DATE_LABEL] Bắt đầu backup database"
echo "============================================"

# ========================
# 1. Backup PostgreSQL
# ========================
echo "[PostgreSQL] Đang backup database ${POSTGRES_DB}..."

PG_FILE="$PG_BACKUP_DIR/${POSTGRES_DB}_${TIMESTAMP}.sql.gz"

PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
  -h "$POSTGRES_HOST" \
  -p "$POSTGRES_PORT" \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  --no-owner \
  --no-privileges \
  --format=custom \
  --compress=6 \
  -f "$PG_FILE"

if [ ! -s "$PG_FILE" ]; then
  echo "[PostgreSQL] ERROR: Backup file rỗng hoặc không tồn tại!" >&2
  exit 1
fi
PG_SIZE=$(du -h "$PG_FILE" | cut -f1)
echo "[PostgreSQL] Backup thành công: $PG_FILE ($PG_SIZE)"

# ========================
# 2. Backup MongoDB
# ========================
echo "[MongoDB] Đang backup database ${MONGO_DB}..."

MONGO_FILE="$MONGO_BACKUP_DIR/${MONGO_DB}_${TIMESTAMP}"

mongodump \
  --host="$MONGO_HOST" \
  --port="$MONGO_PORT" \
  --username="$MONGO_USER" \
  --password="$MONGO_PASSWORD" \
  --authenticationDatabase=admin \
  --db="$MONGO_DB" \
  --out="$MONGO_FILE" \
  --gzip

if [ ! -d "$MONGO_FILE" ]; then
  echo "[MongoDB] ERROR: Backup thư mục không tồn tại!" >&2
  exit 1
fi
MONGO_SIZE=$(du -sh "$MONGO_FILE" | cut -f1)
echo "[MongoDB] Backup thành công: $MONGO_FILE ($MONGO_SIZE)"

# ========================
# 3. Xóa backup cũ
# ========================
echo "[Cleanup] Xóa backup cũ hơn ${RETENTION_DAYS} ngày..."

PG_DELETED=$(find "$PG_BACKUP_DIR" -name "*.sql.gz" -mtime +"$RETENTION_DAYS" -delete -print | wc -l)
MONGO_DELETED=$(find "$MONGO_BACKUP_DIR" -maxdepth 1 -type d -name "${MONGO_DB}_*" -mtime +"$RETENTION_DAYS" -exec rm -rf {} + -print | wc -l)

echo "[Cleanup] Đã xóa: $PG_DELETED file PG, $MONGO_DELETED thư mục Mongo"

# ========================
# 4. Tổng kết
# ========================
PG_TOTAL=$(find "$PG_BACKUP_DIR" -name "*.sql.gz" | wc -l)
MONGO_TOTAL=$(find "$MONGO_BACKUP_DIR" -maxdepth 1 -type d -name "${MONGO_DB}_*" | wc -l)
TOTAL_SIZE=$(du -sh /backup | cut -f1)

echo "============================================"
echo "[$DATE_LABEL] Backup hoàn tất!"
echo "  PostgreSQL: $PG_TOTAL bản ($PG_BACKUP_DIR)"
echo "  MongoDB:    $MONGO_TOTAL bản ($MONGO_BACKUP_DIR)"
echo "  Tổng dung lượng: $TOTAL_SIZE"
echo "  Giữ lại: ${RETENTION_DAYS} ngày gần nhất"
echo "============================================"
