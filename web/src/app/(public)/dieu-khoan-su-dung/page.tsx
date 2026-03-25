export default function TermsOfServicePage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <h1 className="text-3xl font-bold mb-8">Điều khoản sử dụng</h1>
      <p className="text-sm text-muted-foreground mb-6">Cập nhật lần cuối: 26/03/2026</p>

      <div className="prose prose-sm max-w-none space-y-6 text-foreground">
        <section>
          <h2 className="text-xl font-semibold mb-3">1. Giới thiệu</h2>
          <p>
            Chào mừng bạn đến với SanGiaGao.vn (&quot;Dịch vụ&quot;). Bằng việc sử dụng ứng dụng di động
            và website SanGiaGao.vn, bạn đồng ý tuân thủ các điều khoản sử dụng dưới đây.
            Vui lòng đọc kỹ trước khi sử dụng.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">2. Tài khoản người dùng</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Bạn phải cung cấp số điện thoại hợp lệ để đăng ký tài khoản</li>
            <li>Bạn chịu trách nhiệm bảo mật thông tin đăng nhập của mình</li>
            <li>Mỗi người chỉ được sử dụng một tài khoản</li>
            <li>Bạn phải từ 18 tuổi trở lên để sử dụng Dịch vụ</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">3. Hiển thị thông tin liên hệ</h2>
          <p className="mb-2">
            Khi đăng ký và sử dụng Dịch vụ, bạn hiểu và đồng ý rằng các thông tin sau sẽ được
            hiển thị công khai trên trang tin đăng và hồ sơ tài khoản của bạn:
          </p>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Tên tài khoản</strong> (họ tên đăng ký)</li>
            <li><strong>Số điện thoại</strong></li>
            <li><strong>Xã/Phường</strong> và <strong>Tỉnh/Thành phố</strong></li>
            <li><strong>Tên tổ chức</strong> (nếu có)</li>
          </ul>
          <p className="mt-2">
            Mục đích hiển thị công khai là để các đối tác tiềm năng (người mua, người bán) có thể
            dễ dàng liên hệ và kết nối trực tiếp, phục vụ giao dịch mua bán gạo hiệu quả hơn.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">4. Quy tắc đăng tin</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Tin đăng phải chính xác về loại gạo, giá cả, số lượng và chất lượng</li>
            <li>Nghiêm cấm đăng tin sai lệch, gian lận hoặc lừa đảo</li>
            <li>Hình ảnh sản phẩm phải thực tế, không sử dụng ảnh giả mạo</li>
            <li>Không đăng nội dung vi phạm pháp luật, thuần phong mỹ tục</li>
            <li>Người bán cần có gói dịch vụ (subscription) để đăng tin</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">5. Giao dịch</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>SanGiaGao.vn là nền tảng kết nối người mua và người bán gạo</li>
            <li>Chúng tôi <strong>KHÔNG</strong> tham gia trực tiếp vào giao dịch mua bán</li>
            <li>Người mua và người bán tự chịu trách nhiệm về chất lượng, giá cả và giao nhận</li>
            <li>Chúng tôi không đảm bảo hoàn tiền cho bất kỳ giao dịch nào</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">6. Hành vi bị cấm</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Spam, quấy rối, đe dọa người dùng khác</li>
            <li>Sử dụng Dịch vụ cho mục đích trái pháp luật</li>
            <li>Cố ý phá hoại, can thiệp vào hệ thống</li>
            <li>Mạo danh người khác hoặc tổ chức</li>
            <li>Thu thập thông tin người dùng trái phép</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">7. Quyền và trách nhiệm của chúng tôi</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Chúng tôi có quyền xóa tin đăng vi phạm mà không cần thông báo trước</li>
            <li>Chúng tôi có quyền khóa tài khoản vi phạm điều khoản</li>
            <li>Chúng tôi nỗ lực duy trì Dịch vụ ổn định nhưng không đảm bảo hoạt động liên tục 100%</li>
            <li>Chúng tôi có quyền thay đổi, tạm ngừng hoặc ngừng cung cấp Dịch vụ</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">8. Gói dịch vụ & Thanh toán</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Gói dịch vụ cho phép người bán đăng tin trên sàn</li>
            <li>Phí dịch vụ được công bố rõ ràng trước khi thanh toán</li>
            <li>Phí đã thanh toán không hoàn lại trừ trường hợp lỗi hệ thống</li>
            <li>Khi gói dịch vụ hết hạn, tin đăng sẽ bị ẩn cho đến khi gia hạn</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">9. Xóa tài khoản</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Bạn có quyền xóa tài khoản bất kỳ lúc nào trong mục Tài khoản</li>
            <li>Khi xóa tài khoản, tất cả dữ liệu cá nhân, tin đăng, tin nhắn sẽ bị xóa vĩnh viễn</li>
            <li>Hành động xóa không thể hoàn tác</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">10. Giới hạn trách nhiệm</h2>
          <p>SanGiaGao.vn không chịu trách nhiệm về:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Thiệt hại phát sinh từ giao dịch giữa người mua và người bán</li>
            <li>Chất lượng sản phẩm không đúng mô tả</li>
            <li>Mất mát do lỗi kết nối internet hoặc thiết bị người dùng</li>
            <li>Hành vi vi phạm của người dùng khác</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">11. Thay đổi điều khoản</h2>
          <p>
            Chúng tôi có quyền cập nhật điều khoản này. Mọi thay đổi sẽ được thông báo
            qua ứng dụng. Việc tiếp tục sử dụng Dịch vụ đồng nghĩa với việc chấp nhận điều khoản mới.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">12. Liên hệ</h2>
          <p>Nếu có câu hỏi về điều khoản sử dụng, vui lòng liên hệ:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Website: sangiagao.vn</li>
            <li>Điện thoại: 0968660799</li>
          </ul>
        </section>
      </div>
    </div>
  );
}
