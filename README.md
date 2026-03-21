# BMAD Method - Tạo PRD từ Tài Liệu Thô

Bộ công cụ sử dụng **BMAD METHOD** để chuyển đổi tài liệu thô từ khách hàng thành **Product Requirements Document (PRD)** chuyên nghiệp, sẵn sàng cho đội phát triển.

---

## Mục Lục

1. [Giới thiệu](#giới-thiệu)
2. [Cách hoạt động](#cách-hoạt-động)
3. [Cài đặt](#cài-đặt)
4. [Hướng dẫn sử dụng](#hướng-dẫn-sử-dụng)
5. [Ví dụ mẫu](#ví-dụ-mẫu)
6. [Cấu trúc thư mục](#cấu-trúc-thư-mục)
7. [Các lệnh BMAD](#các-lệnh-bmad)
8. [Câu hỏi thường gặp](#câu-hỏi-thường-gặp)

---

## Giới thiệu

**BMAD METHOD** (Build More Architect Dreams) là framework phát triển phần mềm hướng AI với các agent chuyên biệt đóng vai trò như các chuyên gia trong team. Skill `/bmad` trong dự án này kích hoạt **John** - agent Product Manager - để:

- Đọc và phân tích tài liệu thô từ khách hàng
- Đặt câu hỏi làm rõ các điểm mơ hồ
- Tạo ra PRD đầy đủ, có cấu trúc chuẩn
- Lưu kết quả vào thư mục `_bmad-output/`

---

## Cách Hoạt Động

```
raw_docs/               ──►  Skill /bmad  ──►  _bmad-output/
(Tài liệu khách hàng)       (John - PM Agent)    PRD.md
                                  │
                            Phân tích → Hỏi làm rõ → Viết PRD → Review
```

**Quy trình 5 bước:**

1. **Khám phá**: Đọc tất cả file trong `raw_docs/`
2. **Phân tích**: Xác định yêu cầu, điểm mơ hồ, thiếu thông tin
3. **Làm rõ**: Hỏi bạn các câu hỏi cụ thể để điền vào chỗ trống
4. **Tạo PRD**: Viết tài liệu đầy đủ theo chuẩn BMAD
5. **Review & Lưu**: Bạn xem lại, chỉnh sửa nếu cần, rồi lưu file

---

## Cài Đặt

### Yêu cầu

- [Claude Code](https://claude.ai/code) (CLI của Anthropic)
- Mở dự án này trong Claude Code

### Kiểm tra skill đã sẵn sàng

```bash
# Kiểm tra skill đã tồn tại
ls .claude/skills/
# Kết quả mong đợi: bmad.md
```

Skill `/bmad` đã được cài đặt sẵn trong `.claude/skills/bmad.md`. Không cần cài đặt thêm gì.

---

## Hướng Dẫn Sử Dụng

### Bước 1: Chuẩn bị tài liệu thô

Đặt tất cả tài liệu từ khách hàng vào thư mục `raw_docs/`:

```
raw_docs/
├── bien-ban-hop-khach-hang.md      # Biên bản cuộc họp
├── yeu-cau-nghiep-vu.txt           # Tài liệu yêu cầu
├── email-trao-doi.md               # Email/chat với khách
└── he-thong-hien-tai.md            # Mô tả hệ thống cũ (nếu có)
```

> **Mẹo**: Tài liệu càng chi tiết, PRD càng chính xác. Không cần định dạng đẹp - cứ copy paste nội dung thô vào là được.

### Bước 2: Mở Claude Code và chạy skill

Trong terminal của Claude Code, gõ:

```
/bmad
```

### Bước 3: Trả lời câu hỏi làm rõ

John (PM Agent) sẽ:
1. Thông báo đã đọc xong các tài liệu
2. Tóm tắt những gì hiểu được
3. Hỏi 3-7 câu hỏi cụ thể về các điểm còn mơ hồ

Trả lời các câu hỏi này bằng tiếng Việt hoặc tiếng Anh.

### Bước 4: Xem xét PRD được tạo ra

Sau khi bạn trả lời, John sẽ tạo PRD hoàn chỉnh. Bạn có thể:

| Lệnh | Tác dụng |
|------|----------|
| `A` | Phân tích sâu hơn (Advanced Elicitation) |
| `V` | Kiểm tra tính đầy đủ của PRD |
| `E` | Chỉnh sửa một phần cụ thể |
| `S` | Lưu và hoàn thành |

### Bước 5: Lưu kết quả

PRD sẽ được lưu tại: `_bmad-output/PRD.md`

---

## Ví Dụ Mẫu

### Tình huống: Xây dựng app quản lý đơn hàng cho tiệm bánh

**Bước 1**: Chuẩn bị file `raw_docs/yeu-cau-tiem-banh.md` với nội dung:

```markdown
# Yêu cầu hệ thống quản lý đơn hàng - Tiệm Bánh Hoa

## Vấn đề hiện tại
- Nhận đơn qua Zalo, điện thoại, đến tận nơi
- Hay nhầm đơn, quên đơn, không biết tình trạng
- Mỗi tháng làm khoảng 200-300 đơn bánh kem, bánh mì
- Có 3 thợ bánh, 1 người nhận đơn, 1 chủ tiệm

## Mong muốn
- App để nhận và theo dõi đơn hàng
- Khách hàng tự đặt được
- Thợ biết mình cần làm gì hôm nay
- Chủ tiệm thấy tổng quan doanh thu

## Đặc thù sản phẩm
- Bánh cần đặt trước 2 ngày
- Có bánh theo yêu cầu (customize) và bánh có sẵn
- Mùa cao điểm: sinh nhật, tết, valentine

## Budget & Timeline
- Ngân sách: khoảng 50 triệu
- Cần xong trong 3 tháng
- Dùng trên điện thoại là chính
```

**Bước 2**: Gõ `/bmad` trong Claude Code

**Bước 3**: John phân tích và đặt câu hỏi:

```
Xin chào! Tôi là John, PM Agent từ BMAD METHOD.

Tôi đã đọc xong tài liệu "yeu-cau-tiem-banh.md". Đây là những gì tôi hiểu được:
✓ Tiệm bánh cần hệ thống quản lý đơn thay thế Zalo/điện thoại
✓ ~200-300 đơn/tháng, 3 thợ bánh + 1 người nhận đơn + chủ tiệm
✓ Cần app mobile, budget 50tr, timeline 3 tháng

Để tạo PRD chính xác, tôi cần làm rõ một số điểm:

**Về người dùng:**
1. Khách hàng đặt bánh có cần tạo tài khoản không, hay có thể đặt như khách vãng lai?
2. Thợ bánh có cần thấy đơn của nhau không hay chỉ thấy đơn được giao cho mình?

**Về nghiệp vụ:**
3. Quy trình thanh toán thế nào? Tiền mặt khi nhận hàng hay cần hỗ trợ chuyển khoản?
4. Khách customize bánh thì mô tả bằng text hay upload ảnh tham khảo?
5. Ai được phép hủy đơn - chỉ chủ tiệm hay cả khách hàng?

**Về kỹ thuật:**
6. Tiệm hiện có website chưa? Hay chỉ cần app mobile?
7. Cần tích hợp thêm gì không - Zalo Business, giao hàng, kế toán?
```

**Bước 4**: Bạn trả lời, ví dụ:

```
1. Khách vãng lai được, không cần tài khoản nhưng cần nhập SĐT
2. Thợ chỉ thấy đơn của mình, chủ tiệm thấy tất cả
3. Tiền mặt và chuyển khoản, cần xác nhận đã thanh toán
4. Upload ảnh tham khảo được
5. Khách hủy được trước 24 tiếng, sau đó chỉ chủ tiệm mới hủy được
6. Chỉ cần app mobile trước, web admin cho chủ tiệm thì tốt
7. Chưa cần tích hợp gì thêm
```

**Bước 5**: John tạo PRD và xuất ra file `_bmad-output/PRD.md`:

```markdown
# Product Requirements Document: Hệ Thống Quản Lý Đơn Hàng Tiệm Bánh Hoa

## Document Information
- Version: 1.0
- Date: 03/03/2026
- Author: John (PM Agent) via BMAD METHOD
- Status: Draft

## 1. Executive Summary
Hệ thống quản lý đơn hàng cho Tiệm Bánh Hoa giúp số hóa toàn bộ quy trình
từ đặt hàng đến giao bánh, giảm thiểu sai sót và tăng hiệu quả vận hành.

## 2. Problem Statement
Tiệm Bánh Hoa đang xử lý 200-300 đơn/tháng qua Zalo và điện thoại,
dẫn đến nhầm lẫn, quên đơn, và thiếu khả năng theo dõi tình trạng.

## 4. Goals & Success Metrics
| Goal | Metric | Target |
|------|--------|--------|
| Giảm nhầm đơn | % đơn bị nhầm | < 1% |
| Tăng đơn online | % đơn qua app | > 70% trong 3 tháng |
| Cải thiện trải nghiệm | NPS khách hàng | > 8/10 |

## 7. Functional Requirements

### FR-001: Đặt Hàng Online (Khách Hàng)
- **Priority**: Critical
- **User Story**: Là khách hàng, tôi muốn đặt bánh qua app mà không cần gọi điện
- **Acceptance Criteria**:
  - Given: Khách truy cập app, chưa đăng nhập
  - When: Chọn sản phẩm, nhập SĐT và thông tin đặt hàng
  - Then: Đơn được tạo, khách nhận xác nhận qua SMS

### FR-002: Quản Lý Đơn Hàng (Chủ Tiệm)
- **Priority**: Critical
- **User Story**: Là chủ tiệm, tôi muốn xem tất cả đơn hàng và trạng thái
- **Acceptance Criteria**:
  - Given: Chủ tiệm đăng nhập vào web admin
  - When: Xem danh sách đơn
  - Then: Thấy tất cả đơn với trạng thái, có thể lọc theo ngày/trạng thái

... (tiếp tục 15+ requirements khác)
```

---

## Cấu Trúc Thư Mục

```
sangiagao/
├── .claude/
│   └── skills/
│       └── bmad.md              # Skill định nghĩa BMAD PM Agent
├── raw_docs/                    # ← Đặt tài liệu khách hàng vào đây
│   ├── README.md                # Hướng dẫn thư mục
│   └── (tài liệu của bạn...)
├── _bmad-output/                # Kết quả đầu ra (tự động tạo)
│   ├── PRD.md                   # PRD được tạo ra
│   ├── product-brief.md         # Tóm tắt điều hành (nếu có)
│   └── epics/                   # User stories chi tiết (nếu có)
└── README.md                    # File này
```

---

## Các Lệnh BMAD

### Lệnh chính

| Lệnh | Tác dụng |
|------|----------|
| `/bmad` | Khởi động PM Agent, đọc raw_docs và tạo PRD |

### Lệnh trong phiên làm việc với agent

Sau khi PRD được tạo, bạn có thể gõ các lệnh sau trong chat:

| Lệnh | Tác dụng |
|------|----------|
| `VP` | Validate PRD - Kiểm tra tính đầy đủ, phát hiện thiếu sót |
| `EP` | Edit PRD - Chỉnh sửa một phần cụ thể của PRD |
| `CE` | Create Epics - Tạo user stories chi tiết từ PRD |
| `A`  | Advanced Elicitation - Phân tích sâu hơn bằng reasoning frameworks |
| `S`  | Save - Lưu PRD vào `_bmad-output/PRD.md` |

### Ví dụ câu lệnh tự do trong chat

```
# Yêu cầu kiểm tra lại PRD
"Kiểm tra lại PRD, xem có thiếu non-functional requirements không?"

# Yêu cầu chỉnh sửa một phần
"Phần Risk Analysis chưa đầy đủ, thêm risk về bảo mật dữ liệu khách hàng"

# Tạo user stories chi tiết
"Tạo user stories cho epic FR-001 theo chuẩn Given/When/Then"

# Xuất tóm tắt cho stakeholder
"Viết product-brief.md - bản tóm tắt 1 trang cho ban lãnh đạo"

# Yêu cầu viết bằng ngôn ngữ khác
"Dịch PRD sang tiếng Anh và lưu vào _bmad-output/PRD-en.md"
```

---

## Câu Hỏi Thường Gặp

**Q: Tài liệu thô phải viết theo format gì?**

Không cần format đặc biệt. Copy paste email, ghi chú tay, biên bản họp, bất cứ thứ gì vào file `.md` hoặc `.txt` là được. Agent sẽ tự phân tích.

**Q: Có thể đặt nhiều file trong raw_docs không?**

Có, nên đặt tất cả tài liệu liên quan. Ví dụ: 1 file biên bản họp + 1 file yêu cầu kỹ thuật + 1 file mô tả quy trình hiện tại. Agent sẽ đọc và tổng hợp tất cả.

**Q: PRD tạo ra có thể chỉnh sửa trực tiếp không?**

Có thể. Sau khi lưu vào `_bmad-output/PRD.md`, bạn có thể mở và chỉnh sửa tự do. Hoặc yêu cầu agent chỉnh sửa một phần trong cùng phiên chat.

**Q: Nếu tài liệu khách hàng bằng tiếng Anh thì sao?**

Agent sẽ phát hiện ngôn ngữ và phản hồi tương ứng. Bạn cũng có thể chỉ định rõ ngôn ngữ output mong muốn.

**Q: Có thể chạy lại /bmad với tài liệu mới không?**

Có. Mỗi lần chạy `/bmad` là một phiên mới. Xóa hoặc cập nhật file trong `raw_docs/` rồi chạy lại lệnh là được.

**Q: Skill này khác gì so với BMAD gốc từ npx?**

Skill này được tùy chỉnh riêng cho workflow: đọc raw_docs → tạo PRD. BMAD gốc có thêm nhiều agent khác (Architect Winston, Developer Amelia, QA Quinn...) cho các giai đoạn sau. Tham khảo thêm tại [BMAD METHOD](https://github.com/bmad-code-org/BMAD-METHOD).

---

## Về BMAD METHOD

Skill này được xây dựng dựa trên [BMAD METHOD](https://github.com/bmad-code-org/BMAD-METHOD) - framework phát triển phần mềm hướng AI mã nguồn mở với 12+ agent chuyên biệt và 34+ workflow có cấu trúc, bao gồm toàn bộ vòng đời dự án từ ideation đến deployment.
