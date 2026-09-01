# 09 — Di trú VPS (VPS mới, GIỮ nhà cung cấp)

> Chốt: **VPS MỚI cho sàn gạo, GIỮ nguyên nhà cung cấp cũ.** Ngày dự kiến **15/09/2026**.
> KHÔNG resize tại chỗ. Mục tiêu: **0 mất dữ liệu, downtime ~vài phút**. App mobile
> KHÔNG cần build lại (dùng domain). Cutover nhẹ nhờ Cloudflare (đổi origin IP).

> ⭐ **VÌ CÙNG NCC → hỏi họ 2 điều có thể làm mọi thứ DỄ HƠN NHIỀU:**
> 1. **Snapshot/clone VPS cũ sang VPS mới** — nhiều NCC làm giúp → khỏi backup/restore tay (bỏ qua bước B,C bên dưới).
> 2. **Gán (reassign) IP cũ sang VPS mới** — nếu được thì KHỎI đổi gì ở Cloudflare, IP không đổi.
> Nếu cả 2 đều được: di trú gần như bằng "clone + đổi IP", cực nhẹ. Nếu không, theo runbook đầy đủ.

## Bối cảnh + số liệu (đo 29/08/2026)
- **Từ:** `14.225.213.73` (NCC cũ, 4 CPU / 8 GB). **Sang:** VPS mới **2 CPU / 4 GB / 40 GB SSD**.
- **Dữ liệu nhỏ (~700 MB):** Postgres **12 MB** (vol 67M) · MongoDB **503 MB** · MinIO ảnh **96 MB** (bucket `rice-images`) · Redis 25M (cache, BỎ). certbot data BỎ (dùng Cloudflare Origin Cert).
- Repo `/opt/sangiagao`; DB `rice_marketplace` user `rice_user`; env ở `infras/.env.backend` + `.env.chat` + `.env.production` (secret, KHÔNG trong git).
- Deploy: `infras/scripts/quick-deploy.sh` (build `--no-cache` TRÊN VPS → cần 4GB **+ swap**).

## Chuẩn bị VPS mới (làm TRƯỚC ngày cutover, không vội)
1. VPS Ubuntu 22.04, **2 CPU / 4 GB / 40 GB SSD**, thêm SSH key.
2. **Thêm 4 GB swap (BẮT BUỘC — build Next.js dễ OOM):**
   ```bash
   fallocate -l 4G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
   echo '/swapfile none swap sw 0 0' >> /etc/fstab
   ```
3. Cài Docker + compose plugin.
4. `ufw allow 22` (SSH). 80/443 khóa CF ở bước F.

## Ngày cutover (chọn giờ ít người)
### A. Code + secret lên máy mới
```bash
git clone <repo> /opt/sangiagao          # hoặc rsync từ máy cũ
# copy secret (KHÔNG có trong git) từ máy cũ:
scp root@14.225.213.73:/opt/sangiagao/infras/.env.backend  /opt/sangiagao/infras/
scp root@14.225.213.73:/opt/sangiagao/infras/.env.chat     /opt/sangiagao/infras/
scp root@14.225.213.73:/opt/sangiagao/.env.production      /opt/sangiagao/
```
### B. Backup dữ liệu trên máy CŨ
```bash
docker exec rice_postgres pg_dump -U rice_user rice_marketplace | gzip > /root/pg.sql.gz
docker exec rice_mongodb mongodump --archive --gzip > /root/mongo.gz   # thêm -u/-p nếu Mongo có auth
tar -C /var/lib/docker/volumes/rice_prod_minio_data/_data -czf /root/minio.tgz .
```
### C. Chuyển + restore trên máy MỚI
```bash
scp root@14.225.213.73:/root/{pg.sql.gz,mongo.gz,minio.tgz} /root/
# dựng hạ tầng trước (postgres/mongo/minio/redis) — qua quick-deploy hoặc compose
# Postgres:
zcat /root/pg.sql.gz | docker exec -i rice_postgres psql -U rice_user -d rice_marketplace
# Mongo:
docker exec -i rice_mongodb mongorestore --archive --gzip < /root/mongo.gz
# MinIO ảnh:
docker run --rm -v rice_prod_minio_data:/d -v /root:/b alpine sh -c 'tar -C /d -xzf /b/minio.tgz'
```
> DB dump đã có full schema → KHÔNG cần chạy lại migrations. (Nếu dựng DB rỗng thì áp `infras/migrations/*` 015→043 thủ công bằng psql.)

### D. Dựng app (build cần swap đã bật)
```bash
cd /opt/sangiagao
bash infras/scripts/quick-deploy.sh backend
bash infras/scripts/quick-deploy.sh web
bash infras/scripts/quick-deploy.sh admin
bash infras/scripts/quick-deploy.sh chat
# verify nội bộ: curl -s localhost/health ; kiểm log DB connect
```
### E. SSL origin — Cloudflare Origin Certificate (dễ nhất sau CF)
- CF dashboard → **SSL/TLS → Origin Server → Create Certificate** → cài cert+key vào nginx máy mới (15 năm, không cần Let's Encrypt/challenge). Đảm bảo SSL mode ở CF là **Full (strict)**.

### F. Firewall origin (tái tạo — firewall là cục bộ từng máy)
```bash
# tạo lại /root/cf-origin-firewall.sh (nội dung ở [[sangiagao-vps-cloudflare-downsize]])
IFACE=<iface-máy-mới> /root/cf-origin-firewall.sh apply     # kiểm tên iface bằng: ip route get 8.8.8.8
# + systemd cf-origin-firewall.service (After=docker) để bền qua reboot
# + LẦN NÀY chặn luôn cổng 8080 (đừng publish backend ra ngoài — bind nội bộ)
```
### G. Cutover DNS (Cloudflare — tức thì)
- CF dashboard → **DNS** → sửa bản ghi A origin (`sangiagao.vn` + subdomain) → **IP máy mới** → Save.
- Cloudflare cập nhật vài giây, KHÔNG chờ DNS toàn cầu.
- **Verify qua CF:** đăng nhập, chat (wss), ảnh hiện, đăng tin, thanh toán.

### H. Sau cutover
- **Giữ VPS cũ chạy 1–2 ngày.** Ổn → tắt/hủy máy cũ. Lỗi → đổi origin IP về máy cũ trong CF = **rollback tức thì**.

## Secret KHÔNG được quên (mất là hỏng)
- **PHONE_ENCRYPT_KEY** (mất = không giải mã SĐT đã lưu).
- Zalo: APP_ID / SECRET / TEMPLATE_ID / REFRESH_TOKEN (OTP).
- MinIO ACCESS/SECRET + `MINIO_PUBLIC_URL=https://sangiagao.vn/images`.
- DB password, JWT secret, IAP (nếu có).
- (Keystore Android / Apple key KHÔNG ở VPS → không liên quan.)

## Vì sao app mobile an toàn tuyệt đối
- API `https://sangiagao.vn/api/v1`, chat `wss://sangiagao.vn/socket/websocket`, ảnh `sangiagao.vn/images` — **toàn domain, 0 IP cứng** → đổi VPS trong suốt với app. Không build lại, không nộp store.
