# Chương trình Đối tác / Affiliate — Tài liệu kỹ thuật

> Cập nhật: 2026-06-27. Hệ thống hoa hồng giới thiệu (referral) của Sàn Giá Gạo.
> Code chính: `internal/service/commission_engine.go`, `internal/service/referral_service.go`,
> `internal/repository/affiliate_repo.go`, `internal/handler/{referral_handler,admin_referral_handler}.go`.

---

## 1. Vai trò & đăng ký đối tác

Mọi tài khoản đăng nhập đều **tham gia chương trình đối tác được** (có mã + hưởng hoa hồng).
Hàm `ReferralService.BecomeAffiliate`:

| Vai trò khi đăng ký | Kết quả | Ghi chú |
|---|---|---|
| `member` / `seller` | đổi role → **`aff`** | Lên aff để có dashboard |
| `owner` / `admin` / `editor` (staff) | **GIỮ role gốc** + tạo mã | Đổi sẽ mất quyền hệ thống |
| `aff` | idempotent | Đã tham gia |

- "Là aff" (nhãn role) = chỉ member/seller sau khi đăng ký.
- "Tham gia đối tác" (có mã + hoa hồng) = **mọi vai trò**.
- Không còn `ErrRoleNotEligible` (trước đây 'seller' bị chặn 403 — đã sửa).

**Hiển thị UI** (app `profile_screen`/`referral_screen`, web `tai-khoan`/`gioi-thieu-ban`):
- Dashboard "Hoa hồng giới thiệu" (mã + link + thống kê): hiện cho `aff + owner + admin + editor`.
- Nút "Đăng ký làm Đối tác": chỉ cho `member + seller`.

Mã giới thiệu (`referral_codes`): 6 ký tự, tạo lười qua `GetOrCreateCode` (loại bỏ 0/O/1/I/L).
Link chia sẻ: `https://sangiagao.vn/r/{code}`.

---

## 2. Công thức tính hoa hồng (mỗi lần thanh toán)

`commission_engine.go` → `Calculate()`. Khi referee thanh toán 1 gói:

```
gross        = số tiền referee trả
phí nền tảng = gross × platformFeePct      (Apple 0.30, Google ~0.30, SePay/web 0)
net          = gross − phí nền tảng
cơ sở (base) = net  (base_type='net', mặc định)  hoặc  gross
tỷ lệ (rate) = theo LẦN thanh toán (paymentSequence), KHÔNG theo thời gian:
                 Lần 1   → stage1_pct
                 Lần 2   → stage2_pct
                 Lần 3+  → stage3_pct  (giới hạn bởi stage3_cap_months — xem §3)
hoa hồng     = round(base × rate)
```

`paymentSequence` = (số `commission_records` chưa cancelled của cặp referrer–referee) + 1.
Đổi từ time-based sang **payment-count-based** từ 2026-05-19 (cột `stage1_days/stage2_days`
còn trong DB nhưng **vô dụng** — engine bỏ qua, UI ẩn).

**Cấu hình mặc định hiện hành** (admin → Cài đặt quy tắc): 45% / 30% / 15%, cap 24 tháng,
base = ròng, ngưỡng rút 500.000đ. (Tất cả cấu hình được, versioned theo `active_from/active_to`.)

### Ví dụ — gói 100.000đ, lần đầu
| Nguồn | Phí | Net | Rate | Hoa hồng |
|---|---|---|---|---|
| SePay/web (0%) | 0 | 100.000 | 45% | **45.000đ** |
| Apple (30%) | 30.000 | 70.000 | 45% | **31.500đ** |

---

## 3. Giới hạn thời gian hoa hồng lần 3 (cap)

Cột `commission_rules.stage3_cap_months` (migration 041, mặc định **24**, **0 = vĩnh viễn**).

Chính sách (từ 2026-06-27): hoa hồng lần 3+ **không còn vĩnh viễn** — chỉ trả cho các lần
thanh toán trong vòng N tháng kể từ **ngày referee được giới thiệu/đăng ký** (`referred_at`,
fallback `created_at`).

Trong `RecordForPayment`: nếu `cap > 0 && occurred_at > referred_at + cap tháng` → **bỏ qua
hoàn toàn** (return nil, không tạo bản ghi — "chặn toàn bộ sau mốc", mọi stage). Hoa hồng đã
ghi nhận trước đó **giữ nguyên, không truy thu**.

Cấu hình theo từng thời kỳ + **theo từng đối tác** (per `referral_code_id`).

---

## 4. Quy tắc mặc định vs riêng từng đối tác

Bảng `commission_rules`. `GetActiveRule(codeID)`:
1. Tìm rule theo `referral_code_id = codeID AND active_to IS NULL` (override riêng).
2. Không có → fallback rule mặc định (`referral_code_id IS NULL AND active_to IS NULL`).

**Admin UI** `/admin/referrals/rules`:
- Quy tắc mặc định (áp mọi đối tác chưa có override).
- Quy tắc riêng từng đối tác: chọn đối tác từ leaderboard (cần `referral_code_id`), set %/cap/ngưỡng.
- `UpsertRule` versioned (đóng rule cũ `active_to`, chèn rule mới). `DeleteRule` = đóng override
  riêng → quay về mặc định (giữ lịch sử + `rule_id` cho các bản ghi cũ).
- Endpoint: `GET/POST /admin/referrals/rules`, `DELETE /admin/referrals/rules/:codeId`.
- Quyền: `referrals.manage_rules` (owner/admin).

---

## 5. Vòng đời 1 khoản hoa hồng

```
Ghi nhận → [pending]   payable_after = occurred_at + 45 ngày (PayableDelayDays, phòng refund)
   ↓ qua T+45 (cron PromotePayableRecords: pending → payable)
         → [payable]   (có thể chi)
   ↓ admin tạo payout (khi payable ≥ minimum_payout)
         → [paid]      (trừ phí chuyển khoản)
```

- **Refund**: referee được hoàn tiền trước `paid` → `CancelCommissionsForSubscription` đặt
  status `cancelled` (clawback pending/payable). **KHÔNG** clawback khoản đã `paid`.
- Idempotency: INSERT `ON CONFLICT (payment_source, payment_event_id) DO NOTHING`; tx có
  `SELECT ... FOR UPDATE` trên referee để serialize webhook song song (Apple retry + SePay).

---

## 6. Thống kê dashboard (`StatsForReferrer`)

Từ `commission_records WHERE referrer_user_id = $1`:
- **total_referrals** = COUNT(DISTINCT referee).
- **active_referees** = referee còn subscription active.
- **total_earned** = Σ tất cả.
- **pending** = Σ status `pending`.
- **payable** = Σ status `payable`.
- **paid** = Σ status `paid`.

Admin xem toàn bộ qua `Leaderboard` (`/admin/referrals/leaderboard`) — trả thêm
`referral_code_id` cho dropdown chọn đối tác.

---

## 7. Thỏa thuận đối tác (T&C)

- Endpoint `GET /me/aff-terms` (`referral_handler.GetTerms`) trả `current_version`, trạng thái
  `accepted`, và **rule sống** (stage %, `stage3_cap_months`, min_payout) → app/web hiển thị
  số động, KHÔNG hardcode.
- **Version hiện tại = "1.1"** (bump từ "1.0" khi đổi chính sách cap 2026-06-27).
- Đối tác đã ký 1.0 → `accepted=false` → **bắt đồng ý lại**: web `dieu-khoan-doi-tac` cho ký lại,
  banner nhắc ở web `gioi-thieu-ban` + app `referral_screen` (`_buildReacceptBanner`).
- Mục 7 thỏa thuận: hoa hồng đã ghi nhận tính theo điều khoản cũ (snapshot tại payment).
- **Đổi chính sách sau này → nhớ bump `current_version`** để bắt ký lại.

---

## 8. Phòng gian lận

- Self-referral chặn 2 lớp: `AttributeReferral` (code rỗng + self) + `RecordForPayment`.
- Phương án C (2026-05-18): mọi role có code đều earn — vector tự bơm tiền KHÔNG lời vì
  commission < số tiền trả; chỉ rủi ro refund Apple/Google → xử lý bằng clawback khi nhận
  webhook REFUND.

---

## 9. Lịch sử thay đổi chính

- 2026-05-18/19: payment-count model (bỏ time-based stage); mọi role earn (phương án C).
- 2026-06-27: mở đăng ký mọi role (bỏ chặn seller); cap stage 3 (mig 041, 24 tháng);
  UI quy tắc riêng từng đối tác + DeleteRule; thỏa thuận 1.1 + bắt ký lại; owner/staff thấy
  dashboard + link.
