# Product Requirements Document: Sàn Giao Dịch Gạo Trực Tuyến (Rice Marketplace)

## Document Information
- **Version**: 1.1 (Sau Validation)
- **Date**: 03/03/2026
- **Author**: John (PM Agent) via BMAD METHOD
- **Status**: Draft — Chờ xác nhận Open Questions

---

## Changelog
| Version | Thay đổi |
|---------|----------|
| 1.0 | Bản draft ban đầu |
| 1.1 | Fix C-1, C-2; Thêm FR-013, FR-014; Bổ sung AC thiếu vào FR-001, FR-003, FR-006, FR-007, FR-009, FR-011; Bổ sung NFR về OS, browser, offline, listing limit; Cập nhật Success Metrics |

---

## 1. Executive Summary

Rice Marketplace là nền tảng sàn giao dịch gạo trực tuyến kết nối trực tiếp người sản xuất (nông dân, hợp tác xã), thương nhân và người mua trên toàn lãnh thổ Việt Nam. Nền tảng cung cấp môi trường minh bạch để đăng tin sản phẩm, tìm kiếm nguồn hàng và trao đổi trực tiếp qua chat realtime — không xử lý giao dịch tài chính. Mô hình kinh doanh là subscription freemium: dùng thử 30 ngày miễn phí, sau đó 30.000 VND/tháng cho Seller.

---

## 2. Problem Statement

Chuỗi cung ứng gạo tại Việt Nam hiện vẫn phụ thuộc nhiều vào thương lái trung gian, dẫn đến: người nông dân bán giá thấp, người mua mua giá cao, thông tin thị trường thiếu minh bạch, và chi phí giao dịch cao. Không có nền tảng số tập trung để các bên trong chuỗi cung ứng kết nối trực tiếp, đàm phán và so sánh giá một cách hiệu quả.

---

## 3. Product Vision

Trở thành sàn giao dịch gạo số 1 Việt Nam — nơi bất kỳ ai trong chuỗi cung ứng gạo (từ nông dân đến doanh nghiệp xuất khẩu) có thể tìm kiếm, kết nối và thỏa thuận trực tiếp mà không cần qua trung gian, trong vòng vài phút thay vì vài ngày.

---

## 4. Goals & Success Metrics

| Goal | Metric | Target (Cuối tháng 3) |
|------|--------|----------------------|
| Tăng trưởng người dùng | Số tài khoản đăng ký | 1.000 users |
| Kích hoạt người bán | Số Seller có ít nhất 1 tin đăng active | 300 sellers |
| Chuyển đổi subscription | % Seller chuyển sang gói trả phí sau free trial | ≥ 20% |
| Kết nối thành công | Số cuộc chat được khởi tạo/tháng | 2.000 conversations |
| Chất lượng phản hồi | % Seller trả lời tin nhắn trong 24 giờ | ≥ 60% |
| Trải nghiệm onboarding | % user hoàn tất đăng ký + đăng được tin đầu tiên | ≥ 70% |
| Tin đăng hoạt động | Số listing active trên sàn | 1.000 listings |

---

## 5. Target Users & Personas

### Persona 1: Nguyễn Văn Hùng — Nông dân / Hộ sản xuất nhỏ
- **Vai trò**: Seller
- **Goals**: Bán gạo trực tiếp cho người mua, bỏ qua thương lái, nhận giá tốt hơn
- **Pain Points**: Không có kênh tiếp cận người mua trực tiếp; thông tin thị trường hạn chế; không rành công nghệ phức tạp
- **User Journey**: Nghe giới thiệu → Đăng ký bằng SĐT → Đăng tin sản phẩm → Nhận tin nhắn từ người mua → Thỏa thuận và chốt đơn ngoài sàn

### Persona 2: Trần Thị Mai — Đại lý / Thương nhân gạo
- **Vai trò**: Seller + Buyer
- **Goals**: Mở rộng mạng lưới khách hàng, tìm nguồn hàng ổn định với giá cạnh tranh
- **Pain Points**: Mất thời gian tìm nguồn hàng qua điện thoại và quan hệ cá nhân; khó so sánh giá nhiều nguồn
- **User Journey**: Tìm kiếm theo loại gạo/vùng → So sánh giá các tin đăng → Chat trực tiếp với Seller → Đàm phán → Kết thúc giao dịch ngoài sàn

### Persona 3: Lê Minh Khoa — Doanh nghiệp chế biến / Xuất khẩu gạo
- **Vai trò**: Buyer (quy mô lớn)
- **Goals**: Tìm nguồn gạo nguyên liệu ổn định, số lượng lớn, đủ chứng nhận chất lượng
- **Pain Points**: Khó tìm nhiều nhà cung cấp đủ tiêu chuẩn; quy trình thu mua kéo dài
- **User Journey**: Tìm kiếm theo tiêu chí cụ thể (loại gạo, SL, chứng nhận) → Liên hệ nhiều Seller cùng lúc → So sánh → Thỏa thuận

### Persona 4: Admin / Operator
- **Vai trò**: Quản trị viên hệ thống
- **Goals**: Đảm bảo nội dung lành mạnh, theo dõi tăng trưởng, xử lý vi phạm
- **Pain Points**: Cần công cụ quản lý hiệu quả khi số lượng tin đăng tăng nhanh

---

## 6. Scope

### In Scope (Trong phạm vi — MVP 3 tháng)
- Đăng ký / đăng nhập bằng số điện thoại (OTP)
- Chấp nhận Terms of Service & Privacy Policy khi đăng ký
- Phân quyền Buyer / Seller / Admin
- Đăng tin và quản lý sản phẩm gạo (Seller, tối đa 50 listing active)
- Tìm kiếm, lọc và xem chi tiết sản phẩm (Buyer)
- Chat realtime 1:1 giữa các bên với Inbox độc lập
- Thông báo đẩy (push notification)
- Hệ thống đánh giá và uy tín người bán
- Báo cáo vi phạm (report listing / report user)
- Quản lý subscription (free trial 30 ngày + gói trả phí 30K/tháng)
- Web Admin Dashboard (quản lý user, tin đăng, khiếu nại, báo cáo vi phạm)
- Mobile App (iOS ≥ 14, Android ≥ 8) và Web Admin (Chrome/Edge mới nhất)

### Out of Scope (Ngoài phạm vi MVP)
- Xử lý thanh toán trong sàn (payment gateway)
- Quản lý đơn hàng / logistics / vận chuyển
- Tích hợp xuất hóa đơn điện tử
- Hợp đồng điện tử trong sàn
- Chương trình affiliate/giới thiệu
- Đa ngôn ngữ (ngoài tiếng Việt)
- Offline mode đầy đủ

### Future Considerations (Tương lai)
- Tích hợp thanh toán online (VNPay, MoMo)
- Module quản lý hợp đồng số
- Thị trường giá gạo realtime / báo cáo phân tích
- Mở rộng sang nông sản khác

---

## 7. Functional Requirements

### FR-001: Đăng Ký Tài Khoản Bằng SĐT
- **Priority**: Critical
- **Description**: Người dùng đăng ký tài khoản bằng số điện thoại, xác minh qua OTP, chấp nhận T&C và chọn vai trò Buyer hoặc Seller.
- **User Story**: Là người dùng mới, tôi muốn đăng ký tài khoản bằng SĐT để sử dụng sàn mà không cần email.
- **Acceptance Criteria**:
  - Given: Người dùng mở app lần đầu, When: Nhập SĐT Việt Nam hợp lệ và nhấn "Gửi OTP", Then: OTP 6 số được gửi qua SMS trong ≤ 60 giây
  - Given: Nhận được OTP, When: Nhập đúng OTP trong 5 phút, Then: Chuyển đến màn hình chấp nhận Điều khoản sử dụng và Chính sách bảo mật
  - Given: Màn hình T&C, When: User chưa tích vào checkbox đồng ý, Then: Nút "Tiếp tục" bị disable — không thể bỏ qua bước này
  - Given: User đã tích đồng ý T&C, When: Nhấn "Tiếp tục", Then: Tài khoản được tạo, chuyển đến màn hình chọn vai trò (Buyer/Seller)
  - Given: Nhập OTP sai, When: Sai 3 lần liên tiếp, Then: Khóa yêu cầu OTP trong 15 phút
  - Given: Cùng SĐT, When: Yêu cầu OTP > 3 lần trong 1 giờ, Then: Rate limit kích hoạt, hiện thông báo thử lại sau
  - Given: Tài khoản đã tồn tại, When: Nhập đúng OTP, Then: Đăng nhập thành công (không tạo tài khoản mới)
- **Dependencies**: SMS Gateway (Viettel/Twilio)

### FR-002: Quản Lý Hồ Sơ Người Dùng
- **Priority**: High
- **Description**: Người dùng cập nhật thông tin cá nhân/doanh nghiệp sau khi đăng ký.
- **User Story**: Là Seller, tôi muốn điền đầy đủ thông tin hồ sơ để tăng uy tín với người mua.
- **Acceptance Criteria**:
  - Given: Đã đăng nhập, When: Mở trang hồ sơ, Then: Thấy form gồm tên, SĐT (readonly), địa chỉ, ảnh đại diện, mô tả (Seller: thêm tên tổ chức/HTX)
  - Given: Upload ảnh đại diện, When: File > 5MB hoặc không phải JPG/PNG, Then: Hiện thông báo lỗi cụ thể
  - Given: Lưu hồ sơ thành công, When: Người dùng khác xem profile, Then: Thông tin hiển thị đúng, SĐT hiển thị dạng ẩn 3 số giữa (vd: 0912***789)
- **Dependencies**: FR-001, Cloudinary

### FR-003: Đăng Tin Sản Phẩm (Seller)
- **Priority**: Critical
- **Description**: Seller tạo tin đăng sản phẩm gạo. Tối đa 50 listing active/tài khoản. Tin không cần qua duyệt trước khi hiển thị.
- **User Story**: Là Seller, tôi muốn đăng tin sản phẩm nhanh chóng để tiếp cận người mua trên toàn quốc.
- **Acceptance Criteria**:
  - Given: Seller có tài khoản active (free trial hoặc trả phí), When: Tạo tin đăng mới, Then: Form yêu cầu tối thiểu: tên sản phẩm, loại gạo, tỉnh/thành xuất xứ, số lượng (kg/tấn), giá (VND/kg), ít nhất 1 ảnh
  - Given: Upload ảnh sản phẩm, When: File > 10MB hoặc không phải JPG/PNG/WEBP, Then: Hiện lỗi "File không hợp lệ. Vui lòng dùng ảnh JPG/PNG/WEBP dưới 10MB"
  - Given: Đăng tin, When: Seller đã có 50 listing active, Then: Hiện thông báo "Bạn đã đạt giới hạn 50 tin đăng. Xóa hoặc ẩn bớt tin để đăng tin mới." Tối đa 3 ảnh/tin đăng.
  - Given: Điền đủ thông tin bắt buộc, When: Nhấn "Đăng tin", Then: Tin được tạo ở trạng thái không cần duyệt và Seller nhận thông báo xác nhận
  - Given: Tin đăng không cần duyệt, When: Trạng thái chuyển sang "Active", Then: Tin hiện lên sàn trong ≤ 30 giây và Seller nhận thông báo
  - Given: Seller có gói free trial hết hạn và chưa trả phí, When: Cố gắng đăng tin mới, Then: Hiện prompt yêu cầu gia hạn, không cho đăng tin
  - Given: Seller đang có tin đăng active, When: Subscription hết hạn, Then: Tin đăng bị ẩn khỏi sàn; Seller vẫn thấy chúng trong "Tin đăng của tôi" với trạng thái "Tạm ẩn - Chưa gia hạn"
- **Dependencies**: FR-001, FR-007, Cloudinary

### FR-004: Quản Lý Tin Đăng Cá Nhân (Seller)
- **Priority**: High
- **Description**: Seller xem, chỉnh sửa, ẩn hoặc xóa tin đăng của mình, bao gồm cả listing đang bị tạm ẩn.
- **User Story**: Là Seller, tôi muốn cập nhật giá và số lượng tin đăng khi thị trường thay đổi.
- **Acceptance Criteria**:
  - Given: Seller xem danh sách tin đăng của mình, When: Mở tab "Tin đăng của tôi", Then: Thấy tất cả tin với trạng thái rõ ràng: Active / Tạm ẩn (gia hạn) / Bị từ chối / Đã xóa
  - Given: Tin đang Active, When: Seller chỉnh sửa giá hoặc số lượng, Then: Cập nhật ngay lập tức, không cần duyệt lại
  - Given: Tin đang "Tạm ẩn - Chưa gia hạn", When: Seller gia hạn subscription thành công, Then: Tin tự động chuyển về Active và hiện lên sàn (không cần duyệt lại)
  - Given: Seller xóa tin đăng, When: Xác nhận xóa, Then: Tin bị xóa mềm (soft delete 30 ngày), các cuộc chat liên quan vẫn được giữ lại trong Inbox
- **Dependencies**: FR-003, FR-007

### FR-005: Tìm Kiếm & Khám Phá Sàn
- **Priority**: Critical
- **Description**: Người dùng tìm kiếm sản phẩm theo từ khóa, lọc theo loại gạo, vùng, giá, số lượng.
- **User Story**: Là Buyer, tôi muốn tìm gạo ST25 từ Sóc Trăng với giá dưới 20K/kg để so sánh nhiều nguồn cùng lúc.
- **Acceptance Criteria**:
  - Given: Người dùng ở trang sàn, When: Nhập từ khóa tìm kiếm, Then: Kết quả hiện trong ≤ 1 giây với sản phẩm liên quan nhất lên đầu
  - Given: Có kết quả tìm kiếm, When: Áp filter (loại gạo, tỉnh/thành, khoảng giá, số lượng tối thiểu), Then: Danh sách cập nhật ngay, hiện số kết quả tìm thấy
  - Given: Không có kết quả, When: Tìm kiếm với filter quá hẹp, Then: Hiện gợi ý "Xóa filter để xem thêm kết quả"
  - Given: Người dùng xem trang chi tiết tin đăng, When: Mở tin đăng, Then: Hiện đầy đủ thông tin sản phẩm + thông tin Seller (tên, vùng, rating, tổng số đánh giá) + nút "Chat với người bán"
  - Given: Chỉ tin đăng Active mới hiển thị trên sàn công khai, When: Listing bị ẩn hoặc hết hạn subscription, Then: Không xuất hiện trong kết quả tìm kiếm

### FR-006: Chat Realtime 1:1 & Inbox
- **Priority**: Critical
- **Description**: Người dùng nhắn tin trực tiếp qua Phoenix WebSocket. Có Inbox độc lập để xem tất cả cuộc hội thoại. Seller hết subscription bị giới hạn gửi tin mới.
- **User Story**: Là Buyer, tôi muốn nhắn tin với Seller ngay trên app và xem lại lịch sử các cuộc chat.
- **Acceptance Criteria**:
  - Given: Buyer xem tin đăng của Seller, When: Nhấn "Chat với người bán", Then: Cửa sổ chat mở trong ≤ 2 giây; nếu đã có cuộc chat trước đó thì mở lại cuộc chat cũ (không tạo mới)
  - Given: Hai bên đang trong cuộc chat, When: Một bên gửi tin nhắn, Then: Tin nhắn hiện ở phía bên kia trong ≤ 500ms (kết nối ổn định)
  - Given: Gửi ảnh, When: Chọn file ảnh ≤ 10MB (JPG/PNG), Then: Ảnh upload và hiện trong chat trong ≤ 5 giây
  - Given: Người nhận đã đọc tin nhắn, When: Người gửi nhìn vào chat, Then: Hiện tick xanh "Đã đọc"
  - Given: Người dùng mất kết nối, When: Kết nối lại, Then: Lịch sử chat đồng bộ lại đầy đủ từ MongoDB
  - Given: Người dùng mở tab "Tin nhắn" (Inbox), When: Có nhiều cuộc hội thoại, Then: Hiện danh sách tất cả cuộc chat, sắp xếp theo tin nhắn mới nhất, có thể tìm kiếm theo tên Seller/Buyer
  - Given: Listing của Seller đã bị xóa, When: Buyer mở lại cuộc chat từ Inbox, Then: Chat vẫn hoạt động bình thường; hiện banner "Tin đăng liên quan đã bị xóa"
  - **[Fix C-1]** Given: Seller có subscription hết hạn, When: Cố gắng gửi tin nhắn mới trong bất kỳ cuộc chat nào, Then: Input bị disable, hiện prompt "Gia hạn gói để tiếp tục nhắn tin" với nút "Gia hạn ngay"
- **Dependencies**: Phoenix Framework, MongoDB, Cloudinary

### FR-007: Quản Lý Subscription (Freemium)
- **Priority**: Critical
- **Description**: Seller được dùng thử 30 ngày miễn phí. Sau đó cần trả 30.000 VND/tháng. Gia hạn tự động kích hoạt lại toàn bộ listing bị ẩn.
- **User Story**: Là Seller mới, tôi muốn dùng thử miễn phí 30 ngày để đánh giá hiệu quả trước khi trả tiền.
- **Acceptance Criteria**:
  - Given: Seller đăng ký lần đầu, When: Hoàn tất đăng ký và chọn vai trò Seller, Then: Tự động kích hoạt free trial 30 ngày, hiện banner "Còn X ngày dùng thử miễn phí"
  - Given: Còn 7 ngày hết free trial, When: Seller mở app, Then: Hiện thông báo in-app nhắc gia hạn (không chặn màn hình)
  - Given: Còn 3 ngày và 1 ngày hết free trial, When: Đến mốc, Then: Gửi push notification nhắc gia hạn
  - Given: Free trial hết hạn và Seller chưa thanh toán, When: Bất kỳ hành động Seller nào, Then: Tài khoản Seller chuyển sang chế độ hạn chế: (a) không đăng tin mới, (b) listing hiện có bị tạm ẩn khỏi sàn, (c) không gửi tin nhắn mới
  - Given: Admin xác nhận thanh toán 30K thành công, When: Admin cập nhật trạng thái subscription, Then: Tài khoản kích hoạt lại ngay lập tức; tất cả listing "Tạm ẩn - Chưa gia hạn" tự động chuyển về Active — không cần duyệt lại
  - **[Fix C-2]** Given: Seller gia hạn thành công, When: Thanh toán được xác nhận, Then: Toàn bộ listing đã bị ẩn do hết subscription được tự động kích hoạt lại; Seller nhận thông báo "X tin đăng của bạn đã được kích hoạt lại"
  - Given: Một SĐT, When: Cố đăng ký nhiều tài khoản Seller, Then: Hệ thống chặn tạo tài khoản thứ 2 với cùng SĐT (1 SĐT = 1 tài khoản duy nhất)
- **Dependencies**: FR-001; Thanh toán: Admin xác nhận thủ công giai đoạn MVP

### FR-008: Hệ Thống Thông Báo
- **Priority**: High
- **Description**: Người dùng nhận thông báo realtime (in-app) và push notification cho các sự kiện quan trọng.
- **User Story**: Là Seller, tôi muốn nhận thông báo ngay khi có người nhắn tin để phản hồi kịp thời.
- **Acceptance Criteria**:
  - Given: Có tin nhắn mới, When: Người nhận không ở trong màn hình chat đó, Then: Push notification đến thiết bị trong ≤ 3 giây
  - Given: Tin đăng được duyệt hoặc từ chối, When: Admin xử lý, Then: Seller nhận thông báo in-app và push notification với nội dung cụ thể
  - Given: Free trial còn 7 ngày / 3 ngày / 1 ngày / hết hạn, When: Đến mốc, Then: Seller nhận push notification nhắc gia hạn
  - Given: Người dùng tắt push notification trên thiết bị, When: Có sự kiện mới, Then: Vẫn hiện badge đỏ trên icon app và thông báo in-app khi mở app
- **Dependencies**: Firebase Cloud Messaging, Phoenix Channels

### FR-009: Đánh Giá & Uy Tín Người Bán
- **Priority**: Medium
- **Description**: Buyer đánh giá sao (1-5) và bình luận sau khi có đủ tương tác thực với Seller. Có cơ chế báo cáo đánh giá không hợp lệ.
- **User Story**: Là Buyer, tôi muốn xem rating của Seller để đánh giá độ tin cậy trước khi chat.
- **Acceptance Criteria**:
  - **[Fix AC-1]** Given: Buyer muốn đánh giá Seller, When: Mở profile Seller, Then: Nút "Đánh giá" chỉ hiện nếu Buyer và Seller đã có ít nhất 5 tin nhắn qua lại (tổng cộng từ cả hai phía) — ngăn đánh giá sau 1 tin nhắn
  - Given: Điều kiện đủ tương tác, When: Buyer gửi đánh giá, Then: Chọn sao (1-5) bắt buộc, bình luận tối thiểu 10 ký tự; 1 Buyer chỉ đánh giá 1 Seller 1 lần (không sửa sau khi gửi)
  - Given: Buyer gửi đánh giá thành công, When: Người dùng xem profile Seller, Then: Đánh giá hiện ngay với rating trung bình (1 chữ số thập phân), tổng số đánh giá, và 5 đánh giá gần nhất
  - Given: Seller cho rằng đánh giá vi phạm (spam, sai sự thật), When: Seller nhấn "Báo cáo đánh giá này", Then: Admin nhận yêu cầu xem xét trong hàng đợi, xử lý trong 48 giờ làm việc
- **Dependencies**: FR-006

### FR-010: Web Admin — Quản Lý Người Dùng
- **Priority**: High
- **Description**: Admin xem, tìm kiếm, khóa/mở khóa tài khoản người dùng và quản lý subscription thủ công.
- **User Story**: Là Admin, tôi muốn khóa tài khoản vi phạm ngay lập tức và xác nhận thanh toán subscription.
- **Acceptance Criteria**:
  - Given: Admin mở trang quản lý user, When: Tìm theo SĐT hoặc tên, Then: Kết quả hiện trong ≤ 1 giây với: tên, SĐT, vai trò, trạng thái subscription, số ngày còn lại, ngày đăng ký
  - Given: Admin khóa tài khoản, When: Xác nhận khóa + nhập lý do, Then: Tài khoản bị đăng xuất ngay lập tức, không thể đăng nhập lại
  - Given: Tài khoản bị khóa, When: Chủ tài khoản cố đăng nhập, Then: Hiện thông báo "Tài khoản đã bị tạm khóa. Liên hệ hỗ trợ: [contact]"
  - Given: Admin xác nhận thanh toán subscription của Seller, When: Nhấn "Xác nhận đã nhận tiền", Then: Subscription Seller được kích hoạt/gia hạn thêm 30 ngày ngay lập tức; Seller nhận thông báo

### FR-011: Web Admin — Duyệt Tin Đăng
- **Priority**: Critical
- **Description**: Admin xem xét và duyệt/từ chối tin đăng. Seller bị giới hạn số lần resubmit sau khi bị từ chối nhiều lần.
- **User Story**: Là Admin, tôi muốn duyệt tin đăng nhanh để Seller không phải chờ lâu.
- **Acceptance Criteria**:
  - Given: Có tin đăng mới ở trạng thái Pending, When: Admin mở queue duyệt tin, Then: Danh sách tin đang chờ sắp xếp theo thứ tự thời gian (FIFO), hiện toàn bộ nội dung và ảnh để review
  - Given: Admin duyệt tin, When: Nhấn "Chấp thuận", Then: Tin hiện lên sàn trong ≤ 30 giây và Seller nhận thông báo
  - Given: Admin từ chối tin, When: Nhập lý do từ chối và nhấn "Từ chối", Then: Seller nhận thông báo với lý do cụ thể; Seller có thể chỉnh sửa và submit lại
  - **[Fix AC-2]** Given: Tin đăng bị từ chối, When: Seller đã resubmit 3 lần và bị từ chối cả 3 lần, Then: Tin bị khóa resubmit; Seller thấy thông báo "Tin đăng này đã bị từ chối 3 lần. Liên hệ hỗ trợ để được hỗ trợ." — Admin phải unlock thủ công trước khi Seller submit lần 4
  - Given: Admin cần xử lý nhanh, When: Có > 20 tin đang pending, Then: Hiện badge cảnh báo trên menu Admin

### FR-012: Web Admin — Dashboard & Báo Cáo
- **Priority**: Medium
- **Description**: Admin xem tổng quan số liệu hệ thống theo thời gian thực và xuất báo cáo.
- **User Story**: Là Admin, tôi muốn theo dõi số người dùng mới và subscription hàng ngày.
- **Acceptance Criteria**:
  - Given: Admin truy cập dashboard, When: Trang load, Then: Hiện trong ≤ 3 giây các KPI: tổng users, users mới 7 ngày, listing active, conversations/ngày, Seller đang free trial, Seller trả phí, doanh thu tháng hiện tại
  - Given: Admin xem báo cáo subscription, When: Lọc theo tháng, Then: Hiện số Seller chuyển đổi free → paid, số Seller hủy/không gia hạn, tổng doanh thu
  - Given: Admin cần export dữ liệu, When: Nhấn "Export CSV", Then: File CSV download về trong ≤ 10 giây

### FR-013: Báo Cáo Vi Phạm (Report)
- **Priority**: High
- **Description**: Người dùng báo cáo tin đăng hoặc tài khoản vi phạm. Admin xem xét và xử lý.
- **User Story**: Là Buyer, tôi muốn báo cáo tin đăng gian lận để bảo vệ cộng đồng.
- **Acceptance Criteria**:
  - Given: Người dùng xem tin đăng hoặc profile của người khác, When: Nhấn "Báo cáo", Then: Hiện form chọn lý do (Thông tin sai lệch / Sản phẩm không hợp lệ / Gian lận / Nội dung vi phạm / Khác) và ô nhập mô tả tùy chọn
  - Given: Người dùng submit báo cáo, When: Nhấn "Gửi báo cáo", Then: Báo cáo được ghi nhận, người dùng thấy thông báo "Cảm ơn. Chúng tôi sẽ xem xét trong 48 giờ."
  - Given: Admin mở queue báo cáo vi phạm, When: Xem danh sách, Then: Thấy tất cả báo cáo chưa xử lý kèm link đến tin đăng/profile bị báo cáo, lý do, ngày báo cáo
  - Given: Admin xử lý báo cáo, When: Chọn hành động (Xóa tin / Cảnh cáo / Khóa tài khoản / Bỏ qua), Then: Hành động thực hiện ngay; người báo cáo nhận thông báo "Báo cáo của bạn đã được xử lý"
  - Given: Cùng một người dùng, When: Gửi > 10 báo cáo trong 1 ngày, Then: Rate limit kích hoạt, hiện thông báo "Bạn đã gửi quá nhiều báo cáo hôm nay"
- **Dependencies**: FR-010, FR-011

---

## 8. Non-Functional Requirements

### Performance (Hiệu năng)
- API response time: ≤ 300ms (P95) cho các endpoint thông thường
- Search response time: ≤ 1 giây
- Chat message delivery latency: ≤ 500ms (kết nối ổn định)
- App startup time: ≤ 3 giây (cold start, kết nối WiFi)
- Image upload và hiển thị trong chat: ≤ 5 giây

### Security (Bảo mật)
- Xác thực: JWT + Refresh Token (access token hết hạn sau 15 phút)
- RBAC: phân quyền Buyer / Seller / Admin — API endpoint kiểm tra quyền server-side, không tin client
- OTP brute force: khóa sau 3 lần sai liên tiếp, cooldown 15 phút
- Rate limit OTP: tối đa 3 yêu cầu OTP/SĐT/giờ
- SĐT hiển thị bị ẩn một phần trên giao diện công khai (0912***789)
- HTTPS bắt buộc toàn bộ hệ thống; HTTP redirect về HTTPS
- Input validation và sanitization để chống XSS/SQL Injection trên tất cả form
- Rate limiting trên tất cả public API: tối đa 100 req/phút/IP

### Scalability (Khả năng mở rộng)
- Kiến trúc containerized (Docker + K8s) cho phép scale ngang
- Phoenix WebSocket server xử lý ≥ 5.000 kết nối đồng thời
- Redis cache giảm tải database cho search và danh sách sản phẩm (TTL: 5 phút)
- MongoDB sharding cho lịch sử chat khi data tăng

### Usability (Tính dễ dùng)
- Mobile-first: font ≥ 16px, touch target ≥ 44px, contrast ratio ≥ 4.5:1
- Thời gian từ mở app đến có tin đăng đầu tiên: ≤ 5 phút (bao gồm đăng ký)
- Người dùng ít kinh nghiệm (nông dân lớn tuổi) phải hoàn thành đăng ký và đăng tin đầu tiên mà không cần hỗ trợ

### Reliability (Độ tin cậy)
- Uptime target: ≥ 99.5% (≤ 3.6 giờ downtime/tháng)
- Backup database PostgreSQL hàng ngày (snapshot), giữ 30 ngày
- Lịch sử chat MongoDB: soft delete, không bao giờ xóa vĩnh viễn
- Graceful degradation: nếu Phoenix chat server lỗi, phần còn lại của app (browse, search) vẫn hoạt động bình thường; chat hiện thông báo "Đang kết nối lại..."

### Compatibility (Tương thích)
- **[Fix NFR-1]** Mobile App: iOS ≥ 14, Android ≥ 8 (API level 26)
- **[Fix NFR-3]** Web Admin: Chrome/Edge phiên bản mới nhất (N-1); không yêu cầu hỗ trợ IE
- **[Fix NFR-2]** Offline/kết nối kém: App hiển thị trạng thái "Không có kết nối" rõ ràng; cache màn hình cuối cùng để xem được; tự động reconnect khi có mạng trở lại — không yêu cầu offline mode đầy đủ ở MVP
- **[Fix NFR-4]** Giới hạn nội dung: Tối đa 50 listing active/Seller; tối đa 10 ảnh/listing; ảnh tối đa 10MB/file

---

## 9. UI/UX Requirements

- **Ngôn ngữ**: Tiếng Việt toàn bộ, font đủ lớn cho người lớn tuổi
- **Mobile App navigation**: Tab bar chính — Sàn / Tin nhắn / Đăng tin / Thông báo / Hồ sơ
- **Tìm kiếm**: Thanh tìm kiếm luôn hiển thị ở đầu trang Sàn, filter có thể mở/thu gọn
- **Đăng tin**: Wizard flow 3 bước (Thông tin cơ bản → Hình ảnh → Xem lại & Đăng), có thể lưu nháp
- **Subscription prompt**: Banner nhắc gia hạn không che nội dung chính; không dùng popup chặn toàn màn hình (interstitial)
- **Trạng thái loading**: Skeleton screen thay vì spinner để tránh cảm giác lag
- **Trạng thái lỗi**: Mọi lỗi phải hiện thông báo rõ ràng bằng tiếng Việt, có hướng dẫn tiếp theo

---

## 10. Integration Requirements

| Hệ thống | Mục đích | Ghi chú |
|----------|----------|---------|
| SMS Gateway (Viettel/Twilio) | Gửi OTP xác minh SĐT | Cần fallback nếu 1 provider lỗi |
| Cloudinary | Lưu và phục vụ ảnh sản phẩm, avatar, ảnh trong chat | CDN tự động |
| Firebase Cloud Messaging | Push notification iOS + Android | |
| Phoenix ↔ Golang | Golang quản lý auth; Phoenix dùng JWT để xác thực user khi kết nối WebSocket | API nội bộ |
| Thanh toán subscription | **Giai đoạn MVP**: Chuyển khoản ngân hàng + Admin xác nhận thủ công trong dashboard | Tích hợp VNPay/MoMo ở giai đoạn 2 |

---

## 11. Data Requirements

- **PostgreSQL**: User, Profile, Product Listing, Subscription, Notification history, Rating, Report
- **Redis**: Session cache, Search cache (TTL 5 phút), Rate limit counter, Online status
- **MongoDB**: Messages collection — `{chat_id, sender_id, content, type, timestamp, read_at}` — không xóa vĩnh viễn
- **Privacy**: SĐT chỉ dùng để xác minh, không hiển thị đầy đủ trên giao diện public; Admin có quyền xem lịch sử chat khi xử lý khiếu nại (đã có consent trong T&C)
- **Retention**: Listing đã xóa giữ lại 30 ngày (soft delete); lịch sử chat không giới hạn thời gian; log truy cập tối thiểu 90 ngày
- **Compliance**: Tuân thủ Luật An toàn thông tin mạng Việt Nam (Luật số 86/2015/QH13)

---

## 12. Assumptions & Constraints

### Assumptions (Giả định)
- Người dùng mục tiêu có smartphone (iOS ≥ 14 hoặc Android ≥ 8) và kết nối internet (3G/4G)
- Admin/Operator xử lý duyệt tin trong giờ hành chính (8h-17h), SLA 4 giờ làm việc
- Giai đoạn MVP, thanh toán subscription xác nhận thủ công qua Admin dashboard; không cần tự động hóa
- Người bán chịu hoàn toàn trách nhiệm về tính chính xác của thông tin sản phẩm; sàn chỉ kiểm tra nội dung vi phạm, không kiểm tra chất lượng gạo thực tế
- Tài khoản Apple Developer ($99/năm) và Google Play Developer ($25) đã được chuẩn bị trước launch

### Constraints (Ràng buộc)
- Timeline: 3 tháng — strict MVP, không thêm tính năng ngoài scope
- Giá subscription cố định: 30.000 VND/tháng/Seller — không có gói khác ở MVP
- Tech stack đã xác định: Golang, Phoenix/Elixir, React Native, ReactJS/Next.js, PostgreSQL, Redis, MongoDB — không thay đổi

---

## 13. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Seller không chuyển đổi sang gói trả phí sau free trial | High | High | Theo dõi engagement trong 30 ngày; push notification nhắc nhở tích cực; xem xét gia hạn thêm 7 ngày cho user active nhưng chưa thanh toán |
| Tin đăng chứa thông tin sai lệch / gian lận | Medium | High | Duyệt tin trước khi public; hệ thống báo cáo vi phạm (FR-013); khóa tài khoản vi phạm |
| Phoenix WebSocket scaling khi user tăng đột biến | Low | High | Load testing trước launch; cấu hình auto-scaling trên K8s |
| Chậm trễ duyệt tin đăng làm giảm trải nghiệm Seller | Medium | Medium | SLA 4 giờ làm việc; thông báo tiến trình realtime; alert Admin khi queue > 20 tin |
| Cạnh tranh từ Shopee/Lazada vào phân khúc nông sản | Medium | Medium | Tập trung B2B số lượng lớn — phân khúc không được phục vụ tốt bởi sàn TMĐT thông thường |
| Chi phí SMS OTP tăng cao do spam đăng ký giả | Medium | Low | Rate limit 3 OTP/SĐT/giờ; monitor bất thường và block IP |
| Bảo mật dữ liệu người dùng bị lộ | Low | High | HTTPS toàn bộ, JWT rotation, không log thông tin nhạy cảm, audit log 90 ngày |

---

## 14. Timeline & Milestones

| Milestone | Tuần | Deliverable |
|-----------|------|-------------|
| Sprint 0 — Setup | Tuần 1 | Infra, CI/CD, môi trường dev/staging, Docker setup |
| Backend Core | Tuần 2-4 | Auth (OTP + JWT), User, Product API hoàn chỉnh |
| Marketplace + Search | Tuần 3-5 | Search API, listing public, Redis cache |
| Chat System | Tuần 4-6 | Phoenix chat, MongoDB, Inbox, realtime |
| Subscription Module | Tuần 5-7 | Free trial logic, Admin manual payment confirm |
| Mobile App MVP | Tuần 3-9 | Toàn bộ screens mobile (iOS + Android) |
| Web Admin MVP | Tuần 5-9 | Dashboard, user mgmt, moderation, report queue |
| Notification + Rating | Tuần 7-9 | Firebase push, in-app notification, rating system |
| Integration Testing | Tuần 10-11 | E2E test, bug fixing, performance testing |
| Beta Launch | Tuần 12 | Soft launch nội bộ, thu thập feedback |

---

## 15. Open Questions

> *Các câu hỏi dưới đây cần được quyết định trước khi bắt đầu sprint solutioning (architecture phase).*

| # | Câu hỏi | Impact | Cần quyết định trước |
|---|---------|--------|---------------------|
| Q1 | **Quy trình xác nhận thanh toán thủ công**: Admin xác nhận qua kênh nào? Có cần Seller upload ảnh chuyển khoản không? | FR-007, FR-010 | Tuần 2 |
| Q2 | **Listing expiry**: Tin đăng có tự hết hạn sau N ngày không (kể cả Seller đang trả phí), hay tồn tại vô thời hạn? | FR-003, FR-004 | Tuần 2 |
| Q3 | **Buyer limits**: Buyer có hoàn toàn miễn phí mãi mãi không, hay sẽ có giới hạn (số lần chat, số tin đăng xem/ngày)? | Monetization, FR-007 | Tuần 3 |

---

*PRD v1.1 — Đã áp dụng tất cả fixes từ kết quả Validation. Sẵn sàng bàn giao cho Architect (Winston) để thiết kế kiến trúc hệ thống.*
