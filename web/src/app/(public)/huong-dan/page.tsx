import type { Metadata } from "next";
import GuideVideo from "./guide-video";

export const metadata: Metadata = {
  title: "Hướng dẫn sử dụng | SanGiaGao.Vn",
  description: "Hướng dẫn chi tiết cách sử dụng Sàn Giá Gạo - nền tảng giao dịch gạo Việt Nam",
  alternates: { canonical: "https://sangiagao.vn/huong-dan" },
};

export default function UserGuidePage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <h1 className="text-3xl font-bold mb-2">Hướng dẫn sử dụng</h1>
      <p className="text-sm text-muted-foreground mb-4">SanGiaGao.vn — Sàn giao dịch gạo Việt Nam</p>
      <GuideVideo />

      <div className="prose prose-sm max-w-none space-y-8 text-foreground">

        {/* 1. Getting Started */}
        <section>
          <h2 className="text-xl font-semibold mb-3">1. Đăng ký & Đăng nhập</h2>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium mb-1">Đăng ký tài khoản mới</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Mở ứng dụng SanGiaGao hoặc truy cập <strong>sangiagao.vn</strong></li>
                <li>Bấm <strong>&quot;Đăng ký&quot;</strong></li>
                <li>Nhập số điện thoại, họ tên, mật khẩu và địa chỉ (tỉnh/thành, xã/phường)</li>
                <li>Nhập mã OTP gửi về điện thoại qua Zalo hoặc SMS</li>
                <li>Đăng ký thành công — bạn có thể bắt đầu sử dụng ngay</li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Đăng nhập</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Nhập số điện thoại và mật khẩu đã đăng ký</li>
                <li>Bấm <strong>&quot;Đăng nhập&quot;</strong></li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Quên mật khẩu</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Bấm <strong>&quot;Quên mật khẩu&quot;</strong> tại trang đăng nhập</li>
                <li>Nhập số điện thoại → nhận mã OTP</li>
                <li>Nhập mã OTP và mật khẩu mới</li>
              </ol>
            </div>
          </div>
        </section>

        {/* 2. Marketplace */}
        <section>
          <h2 className="text-xl font-semibold mb-3">2. Xem sàn giao dịch</h2>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium mb-1">Duyệt tin đăng</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Vào <strong>&quot;Sàn giao dịch&quot;</strong> để xem tất cả tin đăng bán gạo</li>
                <li>Sử dụng <strong>bộ lọc</strong> để tìm theo phân loại, loại gạo, khu vực, giá</li>
                <li>Bấm vào tin đăng để xem chi tiết: hình ảnh, giá, số lượng, thông tin người bán</li>
              </ul>
            </div>
            <div>
              <h3 className="font-medium mb-1">Bảng giá gạo</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Vào <strong>&quot;Bảng giá&quot;</strong> để xem giá thấp nhất theo từng loại gạo</li>
                <li>Bấm vào tên loại gạo (VD: ST 25) để xem tất cả tin đăng loại đó</li>
              </ul>
            </div>
          </div>
        </section>

        {/* 3. Create Listing */}
        <section>
          <h2 className="text-xl font-semibold mb-3">3. Đăng tin bán gạo</h2>
          <p className="mb-2 text-muted-foreground">
            Bạn cần có <strong>gói thành viên</strong> còn hiệu lực để đăng tin.
          </p>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium mb-1">Đăng 1 tin</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Vào <strong>&quot;Tin đăng của tôi&quot;</strong> → bấm <strong>&quot;Đăng tin&quot;</strong></li>
                <li>Chọn phân loại gạo và loại gạo</li>
                <li>Nhập số lượng (kg) và giá (đ/kg)</li>
                <li>Chọn vụ mùa (không bắt buộc)</li>
                <li>Thêm mô tả và <strong>1 hình ảnh</strong> sản phẩm</li>
                <li>Bấm <strong>&quot;Đăng tin&quot;</strong></li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Đăng nhanh nhiều tin</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Vào <strong>&quot;Tin đăng của tôi&quot;</strong> → bấm <strong>&quot;Đăng nhanh&quot;</strong></li>
                <li>Chọn danh mục gạo (VD: Gạo thơm, Nếp...)</li>
                <li>Tick chọn các loại gạo muốn đăng</li>
                <li>Nhập giá và số lượng cho từng loại</li>
                <li>Bấm <strong>&quot;Đăng X tin&quot;</strong> — tất cả được tạo cùng lúc</li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Quản lý tin đăng</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li><strong>Sửa:</strong> Bấm icon bút chì trên tin đăng để chỉnh sửa giá, số lượng</li>
                <li><strong>Xóa:</strong> Bấm icon thùng rác để xóa tin</li>
                <li>Giới hạn: mỗi loại gạo tối đa <strong>3 lần/ngày</strong>, mỗi tin <strong>1 hình ảnh</strong></li>
              </ul>
            </div>
          </div>
        </section>

        {/* 4. Chat */}
        <section>
          <h2 className="text-xl font-semibold mb-3">4. Nhắn tin (Chat)</h2>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium mb-1">Bắt đầu cuộc trò chuyện</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Xem chi tiết tin đăng → bấm <strong>&quot;Chat với người bán&quot;</strong></li>
                <li>Hoặc bấm vào số điện thoại người bán để gọi trực tiếp</li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Trong cuộc trò chuyện</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Gửi tin nhắn văn bản, hình ảnh, tin nhắn thoại</li>
                <li>Chia sẻ link tin đăng cho người mua</li>
                <li><strong>Thu hồi tin nhắn</strong> trong vòng 24 giờ</li>
                <li><strong>Xóa tin nhắn</strong> (chỉ xóa phía bạn)</li>
                <li>Thả cảm xúc (reaction) trên tin nhắn</li>
              </ul>
            </div>
            <div>
              <h3 className="font-medium mb-1">Hộp thư</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Vào tab <strong>&quot;Tin nhắn&quot;</strong> để xem tất cả cuộc trò chuyện</li>
                <li>Badge đỏ hiển thị số tin chưa đọc</li>
                <li>Vuốt trái để xóa cuộc trò chuyện (trên app)</li>
              </ul>
            </div>
          </div>
        </section>

        {/* 5. Rating */}
        <section>
          <h2 className="text-xl font-semibold mb-3">5. Đánh giá người bán</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Vào trang người bán → bấm <strong>&quot;Đánh giá&quot;</strong></li>
            <li>Chọn số sao (1-5) và viết nhận xét</li>
            <li>Mỗi người chỉ được đánh giá 1 lần cho mỗi người bán</li>
            <li>Đánh giá giúp cộng đồng nhận biết người bán uy tín</li>
          </ul>
        </section>

        {/* 6. Report */}
        <section>
          <h2 className="text-xl font-semibold mb-3">6. Báo cáo vi phạm</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Nếu phát hiện tin đăng sai lệch, lừa đảo hoặc spam</li>
            <li>Bấm <strong>&quot;Báo cáo tin đăng&quot;</strong> trong trang chi tiết</li>
            <li>Chọn lý do và mô tả chi tiết</li>
            <li>Quản trị viên sẽ xem xét và xử lý trong 24 giờ</li>
          </ul>
        </section>

        {/* 7. Subscription */}
        <section>
          <h2 className="text-xl font-semibold mb-3">7. Gói thành viên</h2>
          <div className="space-y-3">
            <div>
              <h3 className="font-medium mb-1">Tại sao cần gói thành viên?</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Đăng tin bán gạo</li>
                <li>Xem chi tiết tin đăng của người khác</li>
                <li>Chat với người mua/bán</li>
                <li>Xem thông tin liên hệ người bán</li>
              </ul>
            </div>
            <div>
              <h3 className="font-medium mb-1">Đăng ký gói</h3>
              <ol className="list-decimal pl-6 space-y-1">
                <li>Vào <strong>&quot;Gói thành viên&quot;</strong> từ menu tài khoản</li>
                <li>Chọn gói phù hợp (1 tháng, 3 tháng, 6 tháng, 12 tháng)</li>
                <li>Liên hệ quản trị viên để kích hoạt gói</li>
              </ol>
            </div>
            <div>
              <h3 className="font-medium mb-1">Khi gói hết hạn</h3>
              <ul className="list-disc pl-6 space-y-1">
                <li>Tin đăng của bạn sẽ tạm ẩn (không bị xóa)</li>
                <li>Gia hạn gói → tin đăng tự động hiển thị lại</li>
                <li>Bạn vẫn xem được sàn giao dịch nhưng không xem chi tiết</li>
              </ul>
            </div>
          </div>
        </section>

        {/* 8. Profile */}
        <section>
          <h2 className="text-xl font-semibold mb-3">8. Quản lý tài khoản</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Cập nhật thông tin:</strong> Tên, ảnh đại diện, địa chỉ, mô tả</li>
            <li><strong>Đổi mật khẩu:</strong> Nhập mật khẩu cũ + mật khẩu mới</li>
            <li><strong>Đổi số điện thoại:</strong> Xác nhận bằng mật khẩu trước khi đổi</li>
            <li><strong>Thông báo:</strong> Xem thông báo hệ thống và tin nhắn từ quản trị viên</li>
            <li><strong>Tắt thông báo:</strong> Vào Cài đặt điện thoại → Ứng dụng → Sàn Giá Gạo → Thông báo → Tắt</li>
          </ul>
        </section>

        {/* 9. FAQ */}
        <section>
          <h2 className="text-xl font-semibold mb-3">9. Câu hỏi thường gặp</h2>
          <div className="space-y-4">
            <div>
              <p className="font-medium">Tôi không nhận được mã OTP?</p>
              <p className="text-muted-foreground">Kiểm tra số điện thoại đúng chưa. Mã OTP gửi qua Zalo (cần có Zalo). Nếu vẫn không nhận được, đợi 2 phút rồi thử lại.</p>
            </div>
            <div>
              <p className="font-medium">Tin đăng bị ẩn là sao?</p>
              <p className="text-muted-foreground">Tin đăng tạm ẩn khi gói thành viên hết hạn. Gia hạn gói để tin tự động hiển thị lại.</p>
            </div>
            <div>
              <p className="font-medium">Tôi có thể đăng bao nhiêu tin?</p>
              <p className="text-muted-foreground">Mỗi loại gạo tối đa 3 lần/ngày. Mỗi tin 1 hình ảnh.</p>
            </div>
            <div>
              <p className="font-medium">Làm sao liên hệ quản trị viên?</p>
              <p className="text-muted-foreground">Vào mục <strong>&quot;Góp ý&quot;</strong> trong trang tài khoản để gửi phản hồi trực tiếp.</p>
            </div>
            <div>
              <p className="font-medium">Thông tin của tôi có an toàn không?</p>
              <p className="text-muted-foreground">Số điện thoại được mã hóa AES-256. Mật khẩu được hash bcrypt. Toàn bộ kết nối sử dụng HTTPS.</p>
            </div>
          </div>
        </section>

        {/* Contact */}
        <section className="rounded-lg border p-4 bg-muted/30">
          <h2 className="text-xl font-semibold mb-2">Liên hệ hỗ trợ</h2>
          <p>Nếu cần hỗ trợ thêm, vui lòng liên hệ:</p>
          <ul className="list-disc pl-6 space-y-1 mt-2">
            <li>Website: <strong>sangiagao.vn</strong></li>
            <li>Góp ý trong ứng dụng: <strong>Tài khoản → Góp ý</strong></li>
          </ul>
        </section>
      </div>
    </div>
  );
}
