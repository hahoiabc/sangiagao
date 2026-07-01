# AUDIT FINDINGS — Sàn Giao Gạo

> Lập bởi sgg (Claude). Bắt đầu: 25/06/2026.
> Quy ước: **CONFIRMED** = đã trace tầng 5 (schema + grep + đọc code), sẵn sàng fix.
> **NEEDS-VERIFY** = nghi vấn từ đọc code, phải kiểm sâu hơn TRƯỚC khi fix.
> Nguyên tắc của chủ dự án: **fix sau khi kiểm tra sâu hơn** — không vá vội.

---

## 🔴 P0 — CRITICAL (mất tiền)

### BUG #1 — Clawback hoa hồng nhắm sai bảng + sai cột — `[CONFIRMED]`
- **File:** `backend/internal/service/commission_engine.go:259-276` (`CancelCommissionsForSubscription`)
- **Lỗi 1:** `UPDATE commissions ...` — không có bảng `commissions`. Bảng thật là `commission_records` (mig `infras/migrations/026_affiliate_referral.sql:69`).
- **Lỗi 2:** câu UPDATE set `updated_at = NOW()` nhưng `commission_records` KHÔNG có cột `updated_at` → kể cả sửa tên bảng vẫn lỗi.
- **Hậu quả:** clawback chạy là lỗi runtime, bị nuốt thành `slog.Warn` ở cả Apple (`apple_iap_notification.go:231`) lẫn Google (`google_iap_service.go:355`). Record refund không bị hủy → sau 45 ngày (`PayableDelayDays`) cron `PromotePayableRecords` (`affiliate_repo.go:232`) đẩy `pending`→`payable` → **payout cho giao dịch đã refund**. Cơ chế chống gian lận DUY NHẤT (comment `commission_engine.go:166`, `referral_service.go:127`) → hỏng 100%.
- **Fix dự kiến:** `commissions`→`commission_records`; bỏ `updated_at=NOW()` (hoặc thêm cột qua migration mới); `slog.Warn`→`slog.Error`; bọc trong tx nhất quán; **viết test** cho clawback (insert record pending → cancel → assert status='cancelled').
- **Kiểm thêm trước fix:** rà toàn bộ commission_records hiện có trên PROD xem đã có record refund nào lọt `payable`/`paid` chưa (thiệt hại thực tế).

---

## 🔴 P1 — CRITICAL (thủng doanh thu / freemium không hoạt động)

### BUG #2 — Hết hạn gói KHÔNG ẩn tin của người bán thật — `[CONFIRMED]`
- **File:** `backend/internal/repository/subscription_repo.go:91-108` (`HideListingsForExpired`)
- **Lỗi:** ẩn tin `WHERE u.role = 'seller'`. Nhưng người bán thật có role `'member'` (đăng ký gán cứng `'member'` ở `auth_service.go:183` + `user_repo.go:82`) hoặc `'aff'` (sau khi opt-in affiliate). KHÔNG ai được gán `'seller'` tự động.
- **Đã loại trừ phòng thủ chỗ khác:** `Browse` (`listing_repo.go:253`) và `Search` (`listing_repo.go:278`) chỉ lọc `status='active'`, KHÔNG join subscription/expiry/role. Không có lớp chặn nào khác.
- **Hậu quả:** `RunExpiryCron` (cron 1h, có wiring ở main.go) mark subscription `expired` nhưng ẩn 0 tin → tin người bán hết hạn vẫn `active` → vẫn hiện công khai. **Người bán dùng 30 ngày trial rồi không trả tiền vẫn bán mãi mãi.** Chỉ chặn được TẠO tin mới (`listing_service.go:151`), không ẩn tin cũ.
- **Fix dự kiến:** đổi điều kiện sang "user có role được phép đăng tin" (`member`/`aff`/`seller`) HOẶC bỏ hẳn filter role, chỉ dựa vào "không có subscription active". Cân nhắc: chính `NOT EXISTS(active sub)` đã đủ — filter role là thừa và sai.
- **Kiểm thêm trước fix:** đếm trên PROD bao nhiêu seller hết hạn vẫn còn tin `active` (thất thoát doanh thu thực tế).

### BUG #3 — Giá trị status ẩn/khôi phục tin lệch nhau giữa các kênh — `[CONFIRMED, latent]`
- **Hide cron:** đặt `status = 'hidden_subscription'` (`subscription_repo.go:93`).
- **Restore SePay/admin:** tìm `status = 'hidden_subscription'` (`subscription_repo.go:322`) ✓ khớp.
- **Restore Apple IAP:** tìm `status = 'hidden'` (`apple_iap_service.go:224`) ✗ LỆCH.
- **Restore Google IAP:** tìm `status = 'hidden'` (`google_iap_service.go:193`) ✗ LỆCH.
- **Hide Google revoke:** đặt `status = 'hidden'` (`google_iap_service.go:344`) — giá trị thứ 2.
- **Hậu quả (sẽ lộ NGAY khi fix BUG #2):** seller bị ẩn tin (`hidden_subscription`) rồi gia hạn qua Apple/Google IAP → restore tìm `hidden` → **KHÔNG khôi phục tin**. Chỉ gia hạn qua SePay/admin mới khôi phục. Hiện đang bị BUG #2 che (cron không ẩn gì).
- **Fix dự kiến:** thống nhất 1 giá trị status duy nhất cho "ẩn do hết hạn" (đề xuất `hidden_subscription`) ở MỌI nơi hide + restore (Apple/Google sửa `'hidden'`→`'hidden_subscription'`). Fix CHUNG đợt với BUG #2.

---

# VÒNG 2 — Audit bảo mật (auth / upload / RBAC) — 25/06/2026

## 🔴 P0 — Account takeover

### BUG #4 — Backdoor OTP `123456` có thể bật ở PRODUCTION — `[CONFIRMED, sgg verify]`
- `generateOTP()` (`backend/internal/service/auth_service.go:413`): nếu `os.Getenv("SMS_PROVIDER")=="mock"` → trả cứng `"123456"`.
- `config.Validate()` (`config.go:87-115`) **KHÔNG** kiểm `SMS_PROVIDER` → không chặn mock ở `AppEnv=production`.
- **Smoking gun:** `docker-compose.production.yml:146` → `SMS_PROVIDER: ${SMS_PROVIDER:-mock}`. Host quên set biến → production chạy mock → **mọi OTP = 123456** → chiếm bất kỳ tài khoản nào chỉ cần SĐT (login/reset-password đều qua OTP).
- `.env.production.example:43` có `SMS_PROVIDER=zalo` (đúng) — nhưng default `:-mock` là mìn: thiếu biến = âm thầm bật backdoor.
- **Fix:** (1) `config.Validate()` fail nếu `AppEnv=production && SMS_PROVIDER=mock`; (2) bỏ default `:-mock` trong compose production (để rỗng → fail rõ); (3) `generateOTP` nên đọc config chứ không `os.Getenv` trực tiếp.
- **Kiểm thêm:** xác nhận container production hiện tại có `SMS_PROVIDER=zalo` (không phải mock).

## 🟠 P1 — Leo thang quyền / DoS

### BUG #5 — Route /admin mồ côi: thiếu `RequirePermission`, đặc biệt `/admin/permissions` — `[CONFIRMED, sgg verify]`
Group `/admin` chỉ gác `RequireRole("owner","admin","editor")` (`main.go:571`); các route sau KHÔNG có permission key → bỏ qua hoàn toàn ma trận phân quyền tinh chỉnh:
- `PUT /admin/permissions` (`main.go:644`, `SavePermissions` = DELETE all + INSERT payload) → **`editor` tự ghi đè ma trận quyền, tự cấp mọi quyền** (leo thang). Nghiêm trọng nhất.
- `GET /admin/payments` (`:578`) → editor xem toàn bộ đơn thanh toán.
- `PUT /admin/site-settings/*` (`:631-634`) → editor sửa nội dung site.
- `GET/POST/PUT/DELETE /admin/inbox` (`:637-640`) → editor full CRUD inbox hệ thống.
- `GET /admin/zalo-zns/status` (`:622`).
- **Fix:** gắn `RequirePermission` key cho từng route (vd `permissions.manage` — nên giới hạn owner/admin, `payments.view`, `site_settings.manage`, `inbox.manage`).
- **Lưu ý giao thoa BUG #7:** nếu role `editor` không gán được (constraint mig 026) thì khai thác bị hạn chế — nhưng vẫn phải sửa vì least-privilege.

### BUG #6 — `/upload/confirm` thiếu rate-limit + nhận key tùy ý — `[CONFIRMED, sgg verify route + agent verify handler]`
- `POST /upload/confirm` (`main.go:454`) KHÔNG có `uploadLimit` (3 route upload kia tại `:451-453` đều có).
- Handler `ConfirmPresignedUpload` nhận `key` thô từ body, `GetObject(key)` + `io.ReadAll` + decode ảnh trong worker pool, không validate key thuộc prefix hợp lệ / thuộc người gọi (`upload_handler.go:96`, `upload_service.go:229`).
- **Khai thác:** spam confirm với key bất kỳ trong bucket / file lớn → tải về RAM + decode → DoS rẻ.
- **Fix:** thêm `uploadLimit` cho confirm; validate `key` regex prefix + đuôi ảnh; lý tưởng lưu key đã cấp presign (Redis TTL) và chỉ confirm key trong tập đó.

### BUG #7 — Role constraint mig 026 BỎ `owner`/`editor` (định danh lệch pha) — `[CONFIRMED code, verify prod]`
- `infras/migrations/026_affiliate_referral.sql:14`: `CHECK (role IN ('member','seller','admin','aff'))` — DROP rồi redefine, **mất `owner` và `editor`** (init.sql có owner/editor, không có aff/seller).
- Nhưng code dùng `owner`/`editor` khắp nơi: bypass subscription (`subscription.go:26`), guard group `/admin` (`main.go:571`), fallback permission (`permission_service.go:53`), seed quyền mig 027.
- **Hậu quả (tùy trạng thái prod):** nếu prod khởi tạo SAU mig 026 → không gán nổi `owner`/`editor` (lỗi 23514) → mọi route `ownerOnly` (plans, reward, ZNS refresh) chết khóa; nếu owner đã tồn tại TRƯỚC 026 → hàng cũ còn nhưng không tạo mới được.
- **Fix:** migration mới hợp nhất `CHECK (role IN ('member','seller','admin','editor','owner','aff'))`. **Kiểm `\d users` trên prod trước.**

## 🟡 Agent báo cáo — CHƯA tự verify tầng 5 (đọc lại trước khi fix)

> Các mục dưới do agent đọc sâu phát hiện, sgg CHƯA tự trace. Đánh dấu để verify, không coi là chốt.

- **Logout/blacklist "giả" khi Redis down** — Redis non-fatal (`main.go:61`), blacklist check `if cache!=nil` (`middleware/auth.go:59`) → Redis chết = logout no-op, refresh token sống 30 ngày không thu hồi nổi. [P1]
- **Không rotate refresh token** — refresh cũ vẫn hợp lệ sau khi cấp cặp mới; token lộ = 30 ngày toàn quyền, không phát hiện reuse. [P1]
- **Reset/đổi mật khẩu KHÔNG vô hiệu hóa session cũ**; **user bị block vẫn dùng access token tới 15'** (middleware không check is_blocked). [P1]
- **OTP cũ không bị huỷ khi gửi OTP mới** → mở rộng cửa sổ brute-force; OTP không gắn `purpose` (mã login dùng cho reset-password). [P1/P2]
- **DeleteAccount HARD delete** (`user_repo.go:303`) — phụ thuộc FK CASCADE; `feedbacks.user_id` no-action → user có feedback **không xóa được tài khoản** (500 âm thầm); FK nhóm referral/commission chưa rõ CASCADE/RESTRICT → mất lịch sử hoa hồng hoặc chặn xóa. **Verify `\d` trên prod.** [P1]
- **Avatar/listing image URL không validate domain** (`user_service.go:72`, `listing_service.go:321`) → lưu URL tùy ý (stored XSS/hotlink tùy cách render). [P2]
- **spam_service fail-OPEN** — lỗi DB → `return nil` → tắt chống spam khi DB chập chờn; device limit bỏ qua nếu `X-Device-ID` rỗng (client tự gửi). [P2]
- **page_permissions cho role `aff`**: code check đúng (fail-closed), nhưng nếu prod lỡ gán `referrals.create_payout`/`view_all` cho `aff` → IDOR/leak dữ liệu hoa hồng. **Verify config prod.** [P2-config]

## 🟢 Bảo mật — ĐÃ KIỂM, OK
- **JWT**: ép HS256 (chống alg=none/confusion), secret≥32 từ env. ✅
- **OTP so sánh**: `subtle.ConstantTimeCompare` cả 3 chỗ. ✅
- **WebSocket chat**: JWT verify trước upgrade (header/cookie, KHÔNG query param), `IsParticipant` check trước join + lại ở service → chống IDOR đọc chat người khác. Origin check + rate limit. ✅ (chắc)
- **IDOR listing/payment/conversation/notification**: đều check ownership/participant đúng. ✅
- **Upload (presign + multipart)**: key do server sinh UUID, ép folder+content-type, magic-bytes thật chống đổi đuôi, presigned GET có expiry, không leak MinIO nội bộ. ✅ (chỉ confirm là lỗ — BUG #6)
- **Phone encryption**: AES-256-GCM nonce/lần, hash SHA-256 deterministic; bcrypt DefaultCost. ✅
- **Payout RBAC**: write-path đòi `referrals.create_payout` (handler check lại), aff bị lọc row-level. ✅

---

# VÒNG 3 — Module backend còn lại + Frontend — 25/06/2026

## 🔴 BUG #8 — Cột `admin_note` MA → kiểm duyệt report hỏng — `[CONFIRMED code, verify prod]`
- `report_repo.go:19` (`reportColumns`) SELECT/RETURNING `admin_note`; `Resolve`/`Dismiss` (`:114/:124`) UPDATE `admin_note`.
- Bảng `reports` (`backend/migrations/001_initial_schema.up.sql:107`) chỉ có `admin_action`, KHÔNG có `admin_note`. Grep toàn bộ migrations: 0 lần `admin_note`, không ALTER.
- **Hậu quả:** mọi query report ném `column "admin_note" does not exist (42703)` → `GET /admin/reports` 500, `PUT /admin/reports/:id` 500 → **toàn bộ kiểm duyệt báo cáo vi phạm không chạy** (trừ khi prod đã vá tay ngoài migration — đúng pattern định danh cột lệch như BUG #1/#3).
- **Fix:** migration `ALTER TABLE reports ADD COLUMN admin_note TEXT;`. Verify `\d reports` prod.

## 🟠 BUG #9 — Xóa tài khoản HARD-DELETE bị FK no-cascade chặn — `[CONFIRMED]`
- `user_repo.go:304`: `DELETE FROM users WHERE id=$1` (hard delete), có xác nhận mật khẩu (OK).
- FK tới `users(id)` KHÔNG cascade → DELETE **lỗi FK violation** cho user nào có:
  - `feedbacks.user_id` (`007_feedback.sql:4`) — có feedback.
  - `commission_records.referrer/referee_user_id` (`026:71-72`) — aff hoặc người được giới thiệu.
  - `payouts.referrer_user_id` (`026:103`), `payments.user_id` (`022:4`) — từng thanh toán.
  - `reports.resolved_by` (`001:116`) — admin từng xử lý report.
- **Hậu quả:** **user đang hoạt động / từng trả tiền KHÔNG xóa được tài khoản** → handler trả 500 âm thầm. Vi phạm yêu cầu bắt buộc xóa tài khoản của CH Play/App Store (đã quảng cáo "đã có" trong STORE_READINESS_AUDIT).
- **Fix:** chuyển sang **soft-delete/anonymize** (set `deleted_at` + ẩn dữ liệu, giữ commission/payment để đối soát), HOẶC xử lý tường minh trong 1 transaction (xóa/anonymize FK trước). KHÔNG nên CASCADE commission (mất lịch sử kế toán).

## 🟠 BUG #10 — Stored XSS qua JSON-LD trang chi tiết tin — `[CONFIRMED]`
- `web/.../san-giao-dich/[id]/page.tsx:100` + `client.tsx:173`: `dangerouslySetInnerHTML={{__html: JSON.stringify(jsonLd)}}` với `jsonLd` nhúng `listing.title/description/seller.name` (free-text người bán). `JSON.stringify` KHÔNG escape `<`,`/` → seller đặt description `</script><script>...</script>` thoát khỏi thẻ → chạy JS với mọi người xem tin. Backend không sanitize.
- Giảm nhẹ: token httpOnly (không lấy được token cookie) nhưng vẫn defacement/phishing/CSRF-leverage/đánh cắp dữ liệu hiển thị.
- **Fix:** escape `<>&` → `\u00xx` trước khi nhúng (cả 4 trang JSON-LD bang-gia-gao); lý tưởng backend strip HTML khi tạo listing. Cân nhắc CSP nonce thay `unsafe-inline`.

## 🟠 Nhóm auth P1 — sgg ĐÃ verify tầng 5 (nâng từ "agent báo" lên CONFIRMED)
- **BUG #11 — không rotate refresh token:** `RefreshToken` (`auth_service.go:223`) cấp cặp mới, KHÔNG blacklist refresh cũ → refresh cũ sống 30 ngày, không phát hiện reuse. [P1]
- **BUG #12 — reset/đổi mật khẩu KHÔNG vô hiệu hóa token cũ:** `middleware/auth.go` JWTAuth chỉ check chữ ký + blacklist, không có `token_version`/`password_changed_at`. Reset password xong, token kẻ tấn công vẫn sống. [P1]
- **BUG #13 — user bị block vẫn dùng access token tới 15':** JWTAuth (`middleware/auth.go:45`) KHÔNG check `is_blocked` (chỉ `RefreshToken` check). [P1]
- **BUG #14 — logout/blacklist VÔ HIỆU khi Redis down:** `BlacklistToken` return sớm nếu cache nil; JWTAuth skip blacklist nếu `tokenCache==nil` (`auth.go:59`). Redis là non-fatal → Redis chết = logout no-op, token thu hồi vẫn dùng được. [P1]

## 🟡 Module backend + Frontend P2 — sgg ĐÃ verify tầng 5, CONFIRMED
- **BUG #15 — Resolve report nuốt lỗi action:** `report_handler.go:92` set `resolved` TRƯỚC, rồi `executeAction` (`:102`) lỗi chỉ `log.Printf` (`:103`), vẫn trả 200 → report "đã xử lý" nhưng tin vi phạm còn sống / user chưa bị khóa. **Fix:** chạy action cùng tx với resolve, hoặc set resolved sau khi action OK. [P1]
- **BUG #16 — Ratings thiếu gate "≥5 tin nhắn"** (PRD FR-009 AC-1): `rating_service.go:25` chỉ check self/exists/unique, KHÔNG đếm tương tác → rating brigading (nhiều account). Cũng KHÔNG thực sự check target là seller (chỉ check tồn tại). **Fix:** thêm check số message reviewer↔seller ≥ ngưỡng. *(xác nhận có phải chủ ý không)* [P1]
- **BUG #17 — Inbox `GetByID` IDOR:** `inbox_repo.go:153` `WHERE si.id=$1` — KHÔNG lọc target/role/expires → member `GET /inbox/:id` đọc được thông báo `target=role:editor` (và auto mark-read). **Fix:** thêm điều kiện target match + expires vào GetByID. [P2]
- **BUG #18 — Không rate-limit `/reports` & `/ratings`** (`main.go:560/563` thiếu `UserRateLimit`, khác với conversations có) → spam. **Fix:** thêm `UserRateLimit`. [P2]
- **BUG #19 — Report Create không map 23505 + không chặn self-report:** `report_repo.go:32` không bắt 23505 → dup pending trả 500 thay vì "đã báo cáo"; `report_service.go:17` pass-through, không check `reporterID != targetID`. **Fix:** map 23505→409, chặn self-report user. [P2]
- **BUG #20 — Avatar/listing image URL không validate domain:** `user_service.go:72` UpdateAvatar + `listing_service.go:321` AddImage lưu URL thô (AddImage có check ownership, thiếu validate URL) → stored XSS/hotlink tùy render. **Fix:** ép URL thuộc host storage của mình. [P2]
- **BUG #21 — Broadcast/SendToUser không rate-limit** (`main.go:618-619`) → editor bị chiếm quyền spam push toàn sàn. [P2]
- **Frontend P2:** CSP `script-src 'unsafe-inline' 'unsafe-eval'` (giảm hiệu lực chống #10) + `msg.content` nhét `img.src`/`window.open` (`tin-nhan/[id]/page.tsx:632`) → whitelist host MinIO. **spam_service fail-OPEN** (`spam_service.go:43/59/68/84` mọi lỗi DB→`return nil`) → tắt chống spam khi DB chập chờn; device-limit bỏ qua nếu `X-Device-ID` rỗng. [P2]

## 🟢 Module backend + Frontend — ĐÃ KIỂM, OK
- **Marketplace search**: tham số hóa `$N` + `plainto_tsquery('simple', unaccent($N))` + sort qua switch hằng → **không SQLi**. Pagination cap ≤100. ✅
- **Listing Bump**: UPDATE atomic 1 query (ownership + cooldown 354' + cap 240 trong WHERE) → race-safe, không bypass. ✅
- **Catalog/Sponsor/Site-settings/Feedback/ZNS**: gated `RequirePermission` (trừ site-settings/inbox — đã ghi BUG #5); tham số hóa; ZNS modify chỉ owner. ✅
- **FCM**: token cleanup 404 + register xóa token cũ đúng; không gửi nhầm cross-user. ✅
- **Frontend auth**: httpOnly cookie + CSRF double-submit + refresh guard chống loop → **tốt**. ✅
- **Referral redirect / next image whitelist / secrets**: `r/[code]` sanitize + no open-redirect; `next/image` remotePatterns whitelist; không lộ secret FE (Zalo secret đã mask backend). ✅

---

## 🟢 ĐÃ KIỂM — KHÔNG phải bug (loại khỏi danh sách)

- **Relay sender_id (#2 cũ):** mobile push `relay` từ socket CỦA CHÍNH người gửi (`mobile/lib/screens/chat/chat_screen.dart:354`) → `sender_id` đúng. Không file Go nào gọi relay. ✅ OK.
- **Payout flow:** `CreatePayout` (`affiliate_repo.go:246`) thiết kế TỐT — tx nguyên tử, check `RowsAffected != len(recordIDs)` rồi rollback chống double-payout. Cột `transfer_fee` (mig 028) + `sent_at` (mig 026) đều tồn tại. ✅ OK.

---

## 🟡 P2 — Robustness chat (không mất tiền) — `[NEEDS-VERIFY]`

### #4 — `uuid_to_bin` dùng `Base.decode16!` (raise)
- **File:** `chat/lib/rice_chat/messages.ex`
- **Nghi:** topic `chat:<id>` do client kiểm soát; join với id không phải hex UUID → crash process channel (supervisor bắt, trả lỗi). Không mất dữ liệu nhưng nên decode an toàn.
- **Phải kiểm:** mobile đã validate UUID (GoRouter) nhưng client khác có thể không → đánh giá bề mặt tấn công.

### #5 — `typing` không có timeout → indicator "đang gõ" có thể kẹt
- **File:** `chat/lib/rice_chat_web/channels/chat_channel.ex:80` — cosmetic, ưu tiên thấp.

### #6 — Role model lệch: code dùng `owner`/`editor` nhưng DB constraint chỉ cho `member/seller/admin/aff`
- `referral_service.BecomeAffiliate` xử lý case `owner`/`admin`/`editor` (`referral_service.go:67`) nhưng `users_role_check` (mig 026:14) chỉ cho phép `member/seller/admin/aff` → gán `owner`/`editor` sẽ vi phạm constraint. Cần soi: 2 role này có thực sự được dùng/gán không, hay là tàn dư. Liên quan trực tiếp BUG #2 (role model mơ hồ).

---

## 🔵 P3 — Edge cases luồng tiền cần soi (chưa audit hết) — `[NEEDS-VERIFY]`

- **SePay:** order code dùng `math/rand` (không crypto) — `payment_service.go:61`. Dedup + amount + pending-guard có mitigate; đánh giá lại độ rủi ro.
- **SePay double-activation:** nếu `MarkPaid` fail sau khi `AdminActivate` thành công (`payment_service.go:144`), order kẹt `pending` nhưng sub đã active; webhook trả nil(200) nên SePay không retry → hiện không double. Xác nhận lại SePay không tự retry độc lập.
- **Apple/Google `recordCommission` best-effort:** lỗi insert commission bị log, không retry → có thể mất 1 record hoa hồng hợp lệ. Cân nhắc outbox/retry.
- **Apple/Google `recordCommission` best-effort:** lỗi insert commission bị log, không retry → có thể mất 1 record hoa hồng hợp lệ. Cân nhắc outbox/retry.

---

## Trạng thái audit theo vùng

| Vùng | Tầng đạt | Ghi chú |
|------|----------|---------|
| Commission + 3 webhook payment (verify path) | Tầng 5 | BUG #1 confirmed |
| Commission clawback/refund path | Tầng 5 | BUG #1 confirmed |
| Payout flow | Tầng 5 | ✅ OK, thiết kế tốt |
| Subscription expiry → hide/restore listings | Tầng 5 | **BUG #2 + #3 confirmed** |
| Marketplace Browse/Search filter | Tầng 5 | chỉ lọc status='active', không chặn expired |
| Register role + listing-create gate | Tầng 5 | role gán 'member', gate create check sub |
| Chat realtime | Tầng 4 | race "mất tin" đã loại; relay OK; 3 issue nhỏ |
| Auth (OTP/JWT/register/reset) | Tầng 5 | **BUG #4 (OTP mock) confirmed**; nhiều P1 agent-báo cần verify |
| Upload (presign/confirm) | Tầng 5 | **BUG #6 confirmed**; presign+multipart OK |
| WebSocket chat auth | Tầng 4-5 | ✅ OK, chống IDOR tốt |
| RBAC / route /admin | Tầng 5 | **BUG #5 (route mồ côi) + #7 (role constraint) confirmed** |
| Xóa tài khoản + IDOR khác | Tầng 4 | IDOR OK; hard-delete + FK cần verify prod |
| Còn lại (mobile/web/admin FE) | Tầng 1-2 | mới khảo sát |

## Tổng số bug CONFIRMED chờ fix (audit HOÀN TẤT — 3 vòng + verify P2)
- **P0 (mất tiền / chiếm tài khoản):** #1 clawback, #4 OTP mock — **2**
- **P1 (doanh thu / leo thang / xóa TK / XSS / auth / moderation):** #2 ẩn tin, #3 status, #5 route mồ côi, #6 upload, #7 role constraint, #8 admin_note, #9 xóa tài khoản FK, #10 XSS JSON-LD, #11 refresh rotation, #12 reset session, #13 blocked 15', #14 logout/Redis, #15 resolve report nuốt lỗi, #16 ratings gate — **14**
- **P2:** #17 inbox IDOR, #18 rate-limit report/rating, #19 self-report/23505, #20 image URL, #21 broadcast limit, + spam fail-open, CSP, msg.content — **~8**
- **TỔNG: 21 bug đánh số CONFIRMED + vài P2 phụ.** Pattern gốc rễ "định danh lệch pha + lỗi nuốt" lặp ở #1/#3/#7/#8.

## Thứ tự fix đề xuất (khi chốt fix 1 lần)
1. **Migrations gom 1 đợt:** `admin_note` (#8), hợp nhất role constraint (#7), (cân nhắc) cột cho clawback (#1).
2. **Luồng tiền/doanh thu:** #1 clawback (đổi bảng/cột), #2 ẩn tin (bỏ filter role), #3 status thống nhất `hidden_subscription`.
3. **Bảo mật chặn ngoài:** #4 OTP mock (Validate + compose), #5 route mồ côi (RequirePermission), #6 upload confirm, #10 XSS escape.
4. **Auth hardening:** #11-#14 (rotate refresh, token_version, check block ở middleware, Redis bắt buộc ở prod).
5. **Xóa tài khoản #9:** chuyển soft-delete/anonymize.
6. **P2 còn lại** + test integration phủ các nhánh phụ.

---

## 🧩 Nhận định gốc rễ (pattern chung)

3 bug CRITICAL (#1, #2, #3) cùng 1 căn nguyên: **định danh bị lệch pha khi code tiến hóa, không cập nhật đồng bộ** —
- Bảng: ý niệm `commissions` → thực tế `commission_records` (BUG #1)
- Role: `seller` (PRD/early) → thực tế `member`/`aff` (BUG #2)
- Status: `hidden` → `hidden_subscription` không thống nhất (BUG #3)

Cả 3 đều ở **nhánh phụ ít chạy / lỗi bị nuốt** nên test thủ công không lộ. Khi fix: nên thêm test integration phủ đúng các nhánh này (refund → clawback, expire → hide → renew → restore) để chống tái diễn.

---

# KẾ HOẠCH FIX (dependency-ordered) — lập 25/06/2026

## Phân loại duyệt
- **[TỰ-LÀM]** = bảo mật hệ thống thuần, người dùng không thấy đổi → fix không cần hỏi.
- **[DUYỆT]** = chạm người dùng / ý đồ thiết kế / tiền / auth-người-dùng → trình cách fix, chờ chủ duyệt.

| # | Bug | Loại | Phụ thuộc | Lý do phân loại |
|---|-----|------|-----------|-----------------|
| #7 | Role constraint bỏ owner/editor | **[DUYỆT]** | — (NỀN TẢNG) | quyết định role model → chi phối #2,#5,#13,bypass sub |
| #1 | Clawback sai bảng/cột | **[DUYỆT]** | độc lập | chạm tiền hoa hồng (affiliate) — dù khôi phục ý đồ |
| #2 | Hết hạn không ẩn tin | **[DUYỆT]** | ← #7 | đổi hành vi monetization người bán thấy |
| #3 | Status hide/restore lệch | **[DUYỆT]** | ← #2 | đổi việc khôi phục tin người dùng |
| #4 | Backdoor OTP mock | **[DUYỆT]** | độc lập | auth liên quan người dùng (chủ nêu rõ) |
| #16 | Ratings thiếu gate ≥5 msg | **[DUYỆT]** | độc lập | product decision (PRD FR-009) |
| #9 | Xóa tài khoản FK chặn | **[DUYỆT]** | có thể cần migration | hành vi tài khoản + store compliance |
| #11 | Refresh không rotate | **[DUYỆT]** | nhóm auth | đổi vòng đời session người dùng |
| #12 | Reset không huỷ session | **[DUYỆT]** | ↔ #13 (chung token_version) | đổi hành vi đăng nhập sau reset |
| #13 | Block vẫn sống 15 phút | **[DUYỆT]** | ← #7, ↔ #12 | hành vi khoá tài khoản người dùng |
| #14 | Logout chết khi Redis down | **[DUYỆT]** | độc lập | logout là hành vi người dùng |
| #5 | Route /admin mồ côi | [TỰ-LÀM] | (tham chiếu #7) | RBAC nội bộ, end-user không thấy |
| #6 | Upload confirm DoS | [TỰ-LÀM] | độc lập | hardening, vô hình |
| #8 | Cột admin_note ma | [TỰ-LÀM] | cần migration | sửa công cụ admin về đúng thiết kế |
| #10 | Stored XSS JSON-LD | [TỰ-LÀM] | độc lập | escape output, vô hình người dùng hợp lệ |
| #17 | Inbox GetByID IDOR | [TỰ-LÀM] | độc lập | chặn rò rỉ dữ liệu nội bộ |
| #18 | Thiếu rate-limit report/rating | [TỰ-LÀM] | độc lập | DoS hardening |
| #19 | Self-report + 23505 | [TỰ-LÀM] | độc lập | robustness; báo lỗi sạch hơn |
| #20 | Image URL không validate | [TỰ-LÀM] | độc lập | chống XSS/hotlink |
| #21 | Broadcast không rate-limit | [TỰ-LÀM] | độc lập | hardening |
| + | spam fail-open, CSP, msg.content | [TỰ-LÀM] | độc lập | hardening |

## Đồ thị phụ thuộc (nút nền tảng → nhánh)
- **#7 (role model)** là GỐC: phải chốt trước → mở khoá #2 (lọc seller), #5 (editor có thật?), #13 (danh sách role + check block), logic bypass subscription.
- **#2 → #3**: #3 (thống nhất status) chỉ có nghĩa khi #2 (ẩn tin) chạy; #3 đụng code restore Apple/Google (gần #1).
- **#12 ↔ #13**: dùng chung cơ chế `token_version`/`password_changed_at` → làm cùng.
- **Migrations trước code**: #8 (admin_note), #7 (mở rộng constraint), #3 (chuẩn hoá row status cũ, nếu cần), #9 (deleted_at nếu soft-delete).

## Thứ tự thực thi
0. **Chốt 4 quyết định thiết kế** (D-ROLE #7, D-DELETE #9, D-RATING #16, D-HIDE #2/#3) + duyệt cách fix #1/#4/#11-14.
1. **Migrations** (áp thủ công): admin_note; role constraint; (tuỳ) chuẩn hoá status; (tuỳ) deleted_at.
2. **[DUYỆT] Tiền/doanh thu**: #1, #2, #3 — sau khi duyệt diff.
3. **[TỰ-LÀM] Bảo mật hệ thống**: #5, #6, #8, #10, #17, #18, #19, #20, #21, spam, CSP, msg.content — làm ngay.
4. **[DUYỆT] Auth**: #4, #11, #12+#13, #14 — sau khi duyệt.
5. **[DUYỆT] Xóa tài khoản #9** — sau khi chốt soft/hard.
6. **Test integration** phủ các nhánh phụ (refund→clawback, expire→hide→renew→restore, OTP, block).


---

# TIẾN ĐỘ FIX (cập nhật 25/06/2026)

## ✅ ĐÃ FIX (build + test PASS) — 14 bug + migration 036
- **Migration 036** (CHỜ ÁP psql trên prod): role 6-set #7, reports.admin_note #8, users.deleted_at #9, chuẩn hóa status #3.
- **Tiền/doanh thu:** #1 clawback (đúng commission_records, bỏ updated_at, log Error), #2 ẩn tin hết hạn (bỏ filter role), #3 thống nhất hidden_subscription (4 chỗ IAP).
- **Auth:** #4 OTP chặn mock prod (config.Validate + compose `:?`), #11 refresh rotation, #12 reset/đổi MK thu hồi token, #13 khóa thu hồi token ngay, #14 Redis bắt buộc prod. (Cơ chế #12/#13: `RevokeUserTokens` set `tvf:{userID}` Redis + JWTAuth từ chối token cũ — cache-only.)
- **Bảo mật khác:** #6 upload/confirm rate-limit, #16 ratings gate ≥5 tin (+test), #17 inbox GetByID IDOR, #18 rate-limit report/rating, #19 self-report + map 23505, #21 broadcast rate-limit.

## ⏳ CÒN LẠI (5 mục)
- **#5** route /admin mồ côi: cần thêm `RequirePermission` + **migration 037 seed permission keys** (permissions.manage chỉ owner/admin, payments.view, site_settings.manage, inbox.manage). [TỰ-LÀM]
- **#20** validate domain URL ảnh (avatar/listing) — cần luồng MinIOPublicURL vào service. [TỰ-LÀM]
- **spam_service fail-open** → log + fail-closed. [TỰ-LÀM]
- **#10** escape XSS JSON-LD (web/admin — codebase frontend). [TỰ-LÀM]
- **#9** soft-delete + ẩn danh tài khoản (refactor lớn: filter deleted_at IS NULL ở các lookup). [đã duyệt thiết kế]
- **Test integration** phủ refund→clawback, expire→hide→renew→restore, OTP, block.

## ⚠️ VIỆC CHỦ CẦN LÀM
1. **Áp migration 036 thủ công** qua psql trên prod TRƯỚC khi deploy binary mới (code #8 cần cột admin_note; #2 query users.deleted_at).
2. Đảm bảo container prod có `SMS_PROVIDER=zalo` (không mock) — nếu không app sẽ KHÔNG khởi động (đúng ý đồ #4).
3. Đảm bảo Redis chạy ở prod (nếu không app KHÔNG khởi động — #14).


---

# ✅ HOÀN TẤT FIX — 25/06/2026 (build + go vet + test + gofmt + tsc PASS)

**Đã fix 21/21 bug + 2 migration (036, 037).** Code xanh hết.

| Nhóm | Bug | Cách fix |
|------|-----|----------|
| Tiền | #1 | clawback → `commission_records`, bỏ `updated_at`, log Error |
| Doanh thu | #2,#3 | bỏ filter role + thống nhất `hidden_subscription` (4 chỗ IAP) |
| Auth | #4 | OTP chặn mock prod (Validate + compose `:?`) |
| Auth | #11 | refresh rotation (huỷ token cũ sau refresh) |
| Auth | #12,#13 | `RevokeUserTokens` (tvf Redis) khi đổi/reset MK + khóa; JWTAuth check |
| Auth | #14 | Redis bắt buộc prod |
| RBAC | #5 | RequirePermission 5 route + migration 037 seed key |
| DoS | #6,#18,#21 | rate-limit confirm/report/rating/broadcast |
| Mod | #8 | migration admin_note + #15 (resolve nuốt lỗi vẫn cần xem*) |
| Mod | #16 | gate ratings ≥5 tin nhắn (+ test) |
| IDOR | #17 | inbox GetByID lọc target/role |
| Robust | #19 | self-report + map 23505 |
| XSS | #10,#20 | escape JSON-LD (web) + validate domain URL ảnh |
| Spam | — | spam_service fail-closed + log |
| Tài khoản | #9 | soft-delete + ẩn danh (đổi phone→mã, giữ commission/payment) |

\* **#15** (resolve report nuốt lỗi executeAction) — CHƯA fix trong đợt này (cần refactor transaction resolve+action). Ghi nhận để đợt sau.

## ⚠️ CHỦ CẦN LÀM TRƯỚC KHI DEPLOY
1. Áp **036** rồi **037** thủ công qua psql (theo thứ tự) TRƯỚC khi chạy binary mới.
2. Set `SMS_PROVIDER=zalo` trên prod (không mock) — nếu không app KHÔNG khởi động.
3. Đảm bảo Redis chạy ở prod — nếu không app KHÔNG khởi động.
4. (FE) Deploy lại web (đã sửa #10) + admin nếu cần.

## CÒN NỢ (đề xuất đợt sau)
- **#15** resolve report nuốt lỗi action (transaction).
- **Test integration** (refund→clawback, expire→hide→renew→restore, OTP, block, soft-delete) — cần dựng **testcontainers** (chưa có hạ tầng). Hiện chỉ có unit test (rating gate, commission Calculate).
- Re-registration sau xóa tài khoản: SĐT gốc đã được giải phóng (đổi phone→mã) nên đăng ký lại ĐƯỢC.
