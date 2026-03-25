export default function PrivacyPolicyPage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <h1 className="text-3xl font-bold mb-8">Chính sách bảo mật</h1>
      <p className="text-sm text-muted-foreground mb-6">Cập nhật lần cuối: 26/03/2026</p>

      <div className="prose prose-sm max-w-none space-y-6 text-foreground">
        <section>
          <h2 className="text-xl font-semibold mb-3">1. Giới thiệu</h2>
          <p>
            SanGiaGao.vn (&quot;chúng tôi&quot;) cam kết bảo vệ quyền riêng tư của bạn. Chính sách này mô tả
            cách chúng tôi thu thập, sử dụng, lưu trữ và bảo vệ thông tin cá nhân khi bạn sử dụng
            ứng dụng di động và website SanGiaGao.vn (&quot;Dịch vụ&quot;).
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">2. Thông tin chúng tôi thu thập</h2>
          <h3 className="text-lg font-medium mb-2">2.1. Thông tin bạn cung cấp</h3>
          <ul className="list-disc pl-6 space-y-1">
            <li>Số điện thoại (dùng để đăng ký và đăng nhập)</li>
            <li>Họ tên, địa chỉ, tỉnh/thành phố (thông tin hồ sơ)</li>
            <li>Ảnh đại diện (tùy chọn)</li>
            <li>Nội dung tin đăng: loại gạo, giá, số lượng, hình ảnh sản phẩm</li>
            <li>Tin nhắn trong hội thoại với người bán/mua</li>
            <li>Phản hồi, báo cáo, đánh giá</li>
          </ul>
          <h3 className="text-lg font-medium mb-2 mt-4">2.2. Thông tin tự động thu thập</h3>
          <ul className="list-disc pl-6 space-y-1">
            <li>Địa chỉ IP, loại thiết bị, hệ điều hành</li>
            <li>Thời gian truy cập và các trang đã xem</li>
            <li>Mã định danh thiết bị (device ID)</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">3. Mục đích sử dụng thông tin</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Cung cấp, duy trì và cải thiện Dịch vụ</li>
            <li>Xác minh tài khoản qua OTP</li>
            <li>Hiển thị tin đăng và kết nối người mua - người bán</li>
            <li>Gửi thông báo liên quan đến tài khoản và giao dịch</li>
            <li>Xử lý báo cáo vi phạm và hỗ trợ khách hàng</li>
            <li>Phân tích sử dụng để cải thiện trải nghiệm</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">4. Chia sẻ thông tin</h2>
          <p>Chúng tôi <strong>không bán</strong> thông tin cá nhân của bạn. Chúng tôi chỉ chia sẻ trong các trường hợp:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Với bên đối tác giao dịch (người mua/bán) khi bạn chủ động liên hệ</li>
            <li>Khi có yêu cầu từ cơ quan pháp luật theo quy định</li>
            <li>Với nhà cung cấp dịch vụ kỹ thuật (hosting, SMS) để vận hành hệ thống</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">5. Thông tin hiển thị công khai</h2>
          <p className="mb-2">
            Để phục vụ mục đích kết nối giao dịch giữa các thành viên, các thông tin sau của bạn
            sẽ được hiển thị công khai trên trang tin đăng và hồ sơ tài khoản:
          </p>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Tên tài khoản</strong> (họ tên đăng ký)</li>
            <li><strong>Số điện thoại</strong></li>
            <li><strong>Xã/Phường</strong> và <strong>Tỉnh/Thành phố</strong></li>
            <li><strong>Tên tổ chức</strong> (nếu có)</li>
          </ul>
          <p className="mt-2">
            Việc hiển thị này giúp các đối tác tiềm năng (người mua, người bán gạo) có thể liên hệ
            trực tiếp, tạo điều kiện giao dịch thuận lợi. Bằng việc đăng ký tài khoản, bạn đồng ý
            với việc hiển thị công khai các thông tin trên.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">6. Lưu trữ và bảo mật</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Dữ liệu được lưu trữ trên máy chủ tại Việt Nam</li>
            <li>Mật khẩu được mã hóa (hash) trước khi lưu</li>
            <li>Token xác thực được lưu an toàn trên thiết bị (Secure Storage)</li>
            <li>Kết nối sử dụng HTTPS mã hóa đầu-cuối</li>
            <li>Chúng tôi giữ thông tin trong thời gian tài khoản còn hoạt động</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">7. Quyền của bạn</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Truy cập:</strong> Xem thông tin cá nhân trong trang Tài khoản</li>
            <li><strong>Chỉnh sửa:</strong> Cập nhật hồ sơ, đổi mật khẩu, đổi số điện thoại</li>
            <li><strong>Xóa:</strong> Yêu cầu xóa tài khoản và dữ liệu liên quan bằng cách liên hệ quản trị viên</li>
            <li><strong>Rút đồng ý:</strong> Ngừng sử dụng Dịch vụ bất kỳ lúc nào</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">8. Quyền truy cập thiết bị</h2>
          <p>Ứng dụng di động có thể yêu cầu các quyền sau:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Camera:</strong> Chụp ảnh sản phẩm khi đăng tin</li>
            <li><strong>Thư viện ảnh:</strong> Chọn ảnh sản phẩm từ thiết bị</li>
            <li><strong>Micro:</strong> Ghi âm tin nhắn thoại</li>
            <li><strong>Internet:</strong> Kết nối đến máy chủ</li>
          </ul>
          <p className="mt-2">Bạn có thể từ chối hoặc thu hồi quyền bất kỳ lúc nào trong cài đặt thiết bị.</p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">9. Cookie và công nghệ theo dõi</h2>
          <p>
            Website sử dụng localStorage để lưu phiên đăng nhập. Chúng tôi không sử dụng cookie
            theo dõi từ bên thứ ba hoặc quảng cáo.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">10. Trẻ em</h2>
          <p>
            Dịch vụ không dành cho người dưới 18 tuổi. Chúng tôi không cố ý thu thập thông tin
            từ trẻ em. Nếu phát hiện, vui lòng liên hệ để chúng tôi xóa ngay.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">11. Thay đổi chính sách</h2>
          <p>
            Chúng tôi có thể cập nhật chính sách này theo thời gian. Mọi thay đổi sẽ được thông báo
            qua ứng dụng hoặc website. Việc tiếp tục sử dụng Dịch vụ sau khi thay đổi đồng nghĩa
            với việc bạn chấp nhận chính sách mới.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">12. Liên hệ</h2>
          <p>
            Nếu có câu hỏi về chính sách bảo mật, vui lòng liên hệ:
          </p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Website: sangiagao.vn</li>
            <li>Điện thoại: 0968660799</li>
          </ul>
        </section>
      </div>
    </div>
  );
}
