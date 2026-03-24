import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';

class TermsOfServiceScreen extends StatelessWidget {
  const TermsOfServiceScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Điều khoản sử dụng')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Cập nhật lần cuối: 24/03/2026',
                style: TextStyle(fontSize: 13, color: AppColors.textHint)),
            const SizedBox(height: 20),
            ..._sections.map((s) => _buildSection(s.$1, s.$2)),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(String title, String content) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title,
              style: const TextStyle(
                  fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
          const SizedBox(height: 8),
          Text(content,
              style: const TextStyle(
                  fontSize: 14, color: AppColors.textSecondary, height: 1.6)),
        ],
      ),
    );
  }

  static const _sections = <(String, String)>[
    (
      '1. Giới thiệu',
      'Chào mừng bạn đến với SanGiaGao.vn ("Dịch vụ"). Bằng việc sử dụng ứng dụng di động '
          'và website SanGiaGao.vn, bạn đồng ý tuân thủ các điều khoản sử dụng dưới đây. '
          'Vui lòng đọc kỹ trước khi sử dụng.',
    ),
    (
      '2. Tài khoản người dùng',
      '• Bạn phải cung cấp số điện thoại hợp lệ để đăng ký tài khoản\n'
          '• Bạn chịu trách nhiệm bảo mật thông tin đăng nhập của mình\n'
          '• Mỗi người chỉ được sử dụng một tài khoản\n'
          '• Bạn phải từ 18 tuổi trở lên để sử dụng Dịch vụ',
    ),
    (
      '3. Quy tắc đăng tin',
      '• Tin đăng phải chính xác về loại gạo, giá cả, số lượng và chất lượng\n'
          '• Nghiêm cấm đăng tin sai lệch, gian lận hoặc lừa đảo\n'
          '• Hình ảnh sản phẩm phải thực tế, không sử dụng ảnh giả mạo\n'
          '• Không đăng nội dung vi phạm pháp luật, thuần phong mỹ tục\n'
          '• Người bán cần có gói dịch vụ (subscription) để đăng tin',
    ),
    (
      '4. Giao dịch',
      '• SanGiaGao.vn là nền tảng kết nối người mua và người bán gạo\n'
          '• Chúng tôi KHÔNG tham gia trực tiếp vào giao dịch mua bán\n'
          '• Người mua và người bán tự chịu trách nhiệm về chất lượng, giá cả và giao nhận\n'
          '• Chúng tôi không đảm bảo hoàn tiền cho bất kỳ giao dịch nào',
    ),
    (
      '5. Hành vi bị cấm',
      '• Spam, quấy rối, đe dọa người dùng khác\n'
          '• Sử dụng Dịch vụ cho mục đích trái pháp luật\n'
          '• Cố ý phá hoại, can thiệp vào hệ thống\n'
          '• Mạo danh người khác hoặc tổ chức\n'
          '• Thu thập thông tin người dùng trái phép',
    ),
    (
      '6. Quyền và trách nhiệm của chúng tôi',
      '• Chúng tôi có quyền xóa tin đăng vi phạm mà không cần thông báo trước\n'
          '• Chúng tôi có quyền khóa tài khoản vi phạm điều khoản\n'
          '• Chúng tôi nỗ lực duy trì Dịch vụ ổn định nhưng không đảm bảo hoạt động liên tục 100%\n'
          '• Chúng tôi có quyền thay đổi, tạm ngừng hoặc ngừng cung cấp Dịch vụ',
    ),
    (
      '7. Gói dịch vụ & Thanh toán',
      '• Gói dịch vụ cho phép người bán đăng tin trên sàn\n'
          '• Phí dịch vụ được công bố rõ ràng trước khi thanh toán\n'
          '• Phí đã thanh toán không hoàn lại trừ trường hợp lỗi hệ thống\n'
          '• Khi gói dịch vụ hết hạn, tin đăng sẽ bị ẩn cho đến khi gia hạn',
    ),
    (
      '8. Xóa tài khoản',
      '• Bạn có quyền xóa tài khoản bất kỳ lúc nào trong mục Tài khoản\n'
          '• Khi xóa tài khoản, tất cả dữ liệu cá nhân, tin đăng, tin nhắn sẽ bị xóa vĩnh viễn\n'
          '• Hành động xóa không thể hoàn tác',
    ),
    (
      '9. Giới hạn trách nhiệm',
      'SanGiaGao.vn không chịu trách nhiệm về:\n'
          '• Thiệt hại phát sinh từ giao dịch giữa người mua và người bán\n'
          '• Chất lượng sản phẩm không đúng mô tả\n'
          '• Mất mát do lỗi kết nối internet hoặc thiết bị người dùng\n'
          '• Hành vi vi phạm của người dùng khác',
    ),
    (
      '10. Thay đổi điều khoản',
      'Chúng tôi có quyền cập nhật điều khoản này. Mọi thay đổi sẽ được thông báo '
          'qua ứng dụng. Việc tiếp tục sử dụng Dịch vụ đồng nghĩa với việc chấp nhận điều khoản mới.',
    ),
    (
      '11. Liên hệ',
      'Nếu có câu hỏi về điều khoản sử dụng, vui lòng liên hệ:\n'
          '• Website: sangiagao.vn\n'
          '• Điện thoại: 0968660799',
    ),
  ];
}
