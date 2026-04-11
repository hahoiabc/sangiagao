import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class UserGuideScreen extends ConsumerStatefulWidget {
  const UserGuideScreen({super.key});

  @override
  ConsumerState<UserGuideScreen> createState() => _UserGuideScreenState();
}

class _UserGuideScreenState extends ConsumerState<UserGuideScreen> {
  String? _videoUrl;

  @override
  void initState() {
    super.initState();
    ref.read(apiServiceProvider).getGuideVideo().then((url) {
      if (mounted && url.isNotEmpty) setState(() => _videoUrl = url);
    }).catchError((_) {});
  }

  Future<void> _openVideo() async {
    if (_videoUrl == null) return;
    final uri = Uri.parse(_videoUrl!);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Hướng dẫn sử dụng')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('SanGiaGao.vn — Sàn giao dịch gạo Việt Nam',
                style: TextStyle(fontSize: 13, color: AppColors.textHint)),
            const SizedBox(height: 20),
            if (_videoUrl != null) ...[
              InkWell(
                onTap: _openVideo,
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppColors.primary.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                        child: const Icon(Icons.play_arrow, color: Colors.white, size: 24),
                      ),
                      const SizedBox(width: 12),
                      const Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Xem video hướng dẫn',
                                style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                            SizedBox(height: 2),
                            Text('Mở trên YouTube',
                                style: TextStyle(fontSize: 12, color: AppColors.textHint)),
                          ],
                        ),
                      ),
                      const Icon(Icons.open_in_new, size: 18, color: AppColors.textHint),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),
            ],
            ..._sections.map((s) => _buildSection(s.$1, s.$2)),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(String title, String content) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
          const SizedBox(height: 8),
          Text(content, style: const TextStyle(fontSize: 14, color: AppColors.textSecondary, height: 1.6)),
        ],
      ),
    );
  }

  static const _sections = [
    (
      '1. Đăng ký & Đăng nhập',
      'Đăng ký: Mở ứng dụng → bấm "Đăng ký" → nhập SĐT, họ tên, mật khẩu, địa chỉ → nhập mã OTP gửi về qua Zalo hoặc SMS.\n\n'
      'Đăng nhập: Nhập SĐT và mật khẩu → bấm "Đăng nhập".\n\n'
      'Quên mật khẩu: Bấm "Quên mật khẩu" → nhập SĐT → nhận mã OTP → đặt mật khẩu mới.'
    ),
    (
      '2. Xem sàn giao dịch',
      'Duyệt tin đăng: Vào "Sàn giao dịch" để xem tất cả tin đăng bán gạo. Sử dụng bộ lọc để tìm theo phân loại, loại gạo, khu vực, giá.\n\n'
      'Bảng giá: Vào "Bảng giá" để xem giá thấp nhất theo từng loại gạo. Bấm vào tên loại gạo để xem tất cả tin đăng loại đó.'
    ),
    (
      '3. Đăng tin bán gạo',
      'Cần có gói thành viên còn hiệu lực để đăng tin.\n\n'
      'Đăng 1 tin: Vào "Tin đăng của tôi" → "Đăng tin" → chọn loại gạo, nhập giá, số lượng, thêm 1 hình ảnh → bấm "Đăng tin".\n\n'
      'Đăng nhanh: Vào "Đăng nhanh" → chọn danh mục → tick chọn các loại gạo → nhập giá, số lượng → bấm "Đăng X tin".\n\n'
      'Giới hạn: mỗi loại gạo tối đa 3 lần/ngày, mỗi tin 1 hình ảnh.'
    ),
    (
      '4. Nhắn tin (Chat)',
      'Bắt đầu: Xem chi tiết tin đăng → bấm "Chat với người bán" hoặc bấm SĐT để gọi trực tiếp.\n\n'
      'Trong chat: Gửi tin nhắn, hình ảnh, tin nhắn thoại. Chia sẻ link tin đăng. Thu hồi tin nhắn trong 24 giờ. Thả cảm xúc.\n\n'
      'Hộp thư: Vào tab "Tin nhắn" để xem tất cả cuộc trò chuyện. Badge đỏ hiển thị số tin chưa đọc.'
    ),
    (
      '5. Đánh giá người bán',
      'Vào trang người bán → bấm "Đánh giá" → chọn số sao (1-5) và viết nhận xét.\n\n'
      'Mỗi người chỉ được đánh giá 1 lần cho mỗi người bán. Đánh giá giúp cộng đồng nhận biết người bán uy tín.'
    ),
    (
      '6. Báo cáo vi phạm',
      'Nếu phát hiện tin đăng sai lệch, lừa đảo hoặc spam → bấm "Báo cáo tin đăng" → chọn lý do và mô tả.\n\n'
      'Quản trị viên sẽ xem xét và xử lý trong 24 giờ.'
    ),
    (
      '7. Gói thành viên',
      'Cần gói thành viên để: đăng tin, xem chi tiết tin đăng, chat, xem SĐT người bán.\n\n'
      'Đăng ký: Vào "Gói thành viên" → chọn gói (1/3/6/12 tháng) → liên hệ quản trị viên kích hoạt.\n\n'
      'Hết hạn: Tin đăng tạm ẩn (không bị xóa). Gia hạn gói → tin tự động hiển thị lại.'
    ),
    (
      '8. Quản lý tài khoản',
      'Cập nhật: Tên, ảnh đại diện, địa chỉ, mô tả.\n'
      'Đổi mật khẩu: Nhập mật khẩu cũ + mật khẩu mới.\n'
      'Đổi SĐT: Xác nhận bằng mật khẩu trước khi đổi.\n'
      'Thông báo: Xem thông báo hệ thống và tin nhắn từ quản trị viên.\n\n'
      'Tắt thông báo: Vào Cài đặt điện thoại → Ứng dụng → Sàn Giá Gạo → Thông báo → Tắt.'
    ),
    (
      '9. Câu hỏi thường gặp',
      'Không nhận được OTP? Kiểm tra SĐT đúng chưa. Mã gửi qua Zalo (cần có Zalo). Đợi 2 phút rồi thử lại.\n\n'
      'Tin đăng bị ẩn? Gói thành viên hết hạn. Gia hạn để tin tự động hiển thị lại.\n\n'
      'Đăng được bao nhiêu tin? Mỗi loại gạo tối đa 3 lần/ngày, mỗi tin 1 hình.\n\n'
      'Liên hệ quản trị viên? Vào "Góp ý" trong trang tài khoản.\n\n'
      'Thông tin có an toàn? SĐT mã hóa AES-256, mật khẩu hash bcrypt, toàn bộ kết nối HTTPS.'
    ),
  ];
}
