import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';

class PrivacyPolicyScreen extends StatelessWidget {
  const PrivacyPolicyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chính sách bảo mật')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Cập nhật lần cuối: 23/03/2026',
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
      'SanGiaGao.vn ("chúng tôi") cam kết bảo vệ quyền riêng tư của bạn. '
          'Chính sách này mô tả cách chúng tôi thu thập, sử dụng, lưu trữ và bảo vệ '
          'thông tin cá nhân khi bạn sử dụng ứng dụng di động và website SanGiaGao.vn ("Dịch vụ").',
    ),
    (
      '2. Thông tin chúng tôi thu thập',
      'Thông tin bạn cung cấp:\n'
          '• Số điện thoại (dùng để đăng ký và đăng nhập)\n'
          '• Họ tên, địa chỉ, tỉnh/thành phố (thông tin hồ sơ)\n'
          '• Ảnh đại diện (tùy chọn)\n'
          '• Nội dung tin đăng: loại gạo, giá, số lượng, hình ảnh sản phẩm\n'
          '• Tin nhắn trong hội thoại với người bán/mua\n'
          '• Phản hồi, báo cáo, đánh giá\n\n'
          'Thông tin tự động thu thập:\n'
          '• Địa chỉ IP, loại thiết bị, hệ điều hành\n'
          '• Thời gian truy cập và các trang đã xem\n'
          '• Mã định danh thiết bị (device ID)',
    ),
    (
      '3. Mục đích sử dụng thông tin',
      '• Cung cấp, duy trì và cải thiện Dịch vụ\n'
          '• Xác minh tài khoản qua OTP\n'
          '• Hiển thị tin đăng và kết nối người mua - người bán\n'
          '• Gửi thông báo liên quan đến tài khoản và giao dịch\n'
          '• Xử lý báo cáo vi phạm và hỗ trợ khách hàng\n'
          '• Phân tích sử dụng để cải thiện trải nghiệm',
    ),
    (
      '4. Chia sẻ thông tin',
      'Chúng tôi KHÔNG BÁN thông tin cá nhân của bạn. Chúng tôi chỉ chia sẻ trong các trường hợp:\n'
          '• Với bên đối tác giao dịch (người mua/bán) khi bạn chủ động liên hệ\n'
          '• Khi có yêu cầu từ cơ quan pháp luật theo quy định\n'
          '• Với nhà cung cấp dịch vụ kỹ thuật (hosting, SMS) để vận hành hệ thống',
    ),
    (
      '5. Lưu trữ và bảo mật',
      '• Dữ liệu được lưu trữ trên máy chủ tại Việt Nam\n'
          '• Mật khẩu được mã hóa (hash) trước khi lưu\n'
          '• Token xác thực được lưu an toàn trên thiết bị (Secure Storage)\n'
          '• Kết nối sử dụng HTTPS mã hóa đầu-cuối\n'
          '• Chúng tôi giữ thông tin trong thời gian tài khoản còn hoạt động',
    ),
    (
      '6. Quyền của bạn',
      '• Truy cập: Xem thông tin cá nhân trong trang Tài khoản\n'
          '• Chỉnh sửa: Cập nhật hồ sơ, đổi mật khẩu, đổi số điện thoại\n'
          '• Xóa: Yêu cầu xóa tài khoản và dữ liệu liên quan bằng cách liên hệ quản trị viên\n'
          '• Rút đồng ý: Ngừng sử dụng Dịch vụ bất kỳ lúc nào',
    ),
    (
      '7. Quyền truy cập thiết bị',
      'Ứng dụng di động có thể yêu cầu các quyền sau:\n'
          '• Camera: Chụp ảnh sản phẩm khi đăng tin\n'
          '• Thư viện ảnh: Chọn ảnh sản phẩm từ thiết bị\n'
          '• Micro: Ghi âm tin nhắn thoại\n'
          '• Internet: Kết nối đến máy chủ\n\n'
          'Bạn có thể từ chối hoặc thu hồi quyền bất kỳ lúc nào trong cài đặt thiết bị.',
    ),
    (
      '8. Trẻ em',
      'Dịch vụ không dành cho người dưới 18 tuổi. Chúng tôi không cố ý thu thập '
          'thông tin từ trẻ em. Nếu phát hiện, vui lòng liên hệ để chúng tôi xóa ngay.',
    ),
    (
      '9. Thay đổi chính sách',
      'Chúng tôi có thể cập nhật chính sách này theo thời gian. Mọi thay đổi sẽ được '
          'thông báo qua ứng dụng hoặc website. Việc tiếp tục sử dụng Dịch vụ sau khi thay đổi '
          'đồng nghĩa với việc bạn chấp nhận chính sách mới.',
    ),
    (
      '10. Liên hệ',
      'Nếu có câu hỏi về chính sách bảo mật, vui lòng liên hệ:\n'
          '• Website: sangiagao.vn\n'
          '• Điện thoại: 0968660799',
    ),
  ];
}
