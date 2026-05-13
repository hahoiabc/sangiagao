import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/providers.dart';

/// Affiliate Terms & Conditions screen.
/// Modal/page used both for first-time acceptance and read-only review.
class AffTermsScreen extends ConsumerStatefulWidget {
  /// If true, shows the accept checkbox + button. Otherwise read-only.
  final bool requireAccept;
  const AffTermsScreen({super.key, this.requireAccept = false});

  @override
  ConsumerState<AffTermsScreen> createState() => _AffTermsScreenState();
}

class _AffTermsScreenState extends ConsumerState<AffTermsScreen> {
  bool _checked = false;
  bool _saving = false;
  String _currentVersion = '1.0';
  bool _alreadyAccepted = false;

  @override
  void initState() {
    super.initState();
    _loadStatus();
  }

  Future<void> _loadStatus() async {
    try {
      final r = await ref.read(apiServiceProvider).getAffTerms();
      if (!mounted) return;
      setState(() {
        _currentVersion = (r['current_version'] ?? '1.0').toString();
        _alreadyAccepted = r['accepted'] == true;
      });
    } catch (_) {}
  }

  Future<void> _accept() async {
    if (!_checked) return;
    setState(() => _saving = true);
    try {
      await ref.read(apiServiceProvider).acceptAffTerms(_currentVersion);
      if (!mounted) return;
      Navigator.pop(context, true);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Lưu thất bại, thử lại')),
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final showAccept = widget.requireAccept && !_alreadyAccepted;

    return Scaffold(
      appBar: AppBar(title: const Text('Điều khoản đối tác Affiliate')),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Phiên bản $_currentVersion',
                    style: const TextStyle(color: Colors.grey, fontSize: 12),
                  ),
                  const SizedBox(height: 12),
                  _section('1. Quyền lợi',
                      'Đối tác (Aff) nhận hoa hồng theo 3 giai đoạn dựa trên tuổi của người được giới thiệu (Referee):\n'
                      '• Giai đoạn 1: % của doanh thu ròng (xem dashboard cho rule hiện hành)\n'
                      '• Giai đoạn 2: %\n'
                      '• Giai đoạn 3 (vĩnh viễn): %\n\n'
                      'Doanh thu ròng = số tiền Sàn thực nhận sau khi trừ phí nền tảng (Apple 30%, SePay 0%).'),
                  _section('2. Thanh toán',
                      '• Ngưỡng tối thiểu: theo cài đặt hiện tại, tối thiểu 100.000đ. Sàn có thể điều chỉnh theo từng giai đoạn phát triển nền tảng.\n'
                      '• Thời gian đối soát: T+45 ngày sau khi Referee thanh toán.\n'
                      '• Phí chuyển khoản: do Đối tác chịu, trừ trực tiếp từ số tiền payout (thực tế từng lần, tuỳ ngân hàng).\n'
                      '• Đối tác phải cập nhật chính xác thông tin tài khoản nhận tiền. Sàn không chịu trách nhiệm nếu chuyển sai do thông tin sai.'),
                  _section('3. Thuế thu nhập cá nhân',
                      'Đối tác tự kê khai và đóng thuế TNCN theo Luật thuế Việt Nam. '
                      'Sàn không khấu trừ thuế tại nguồn. '
                      'Đối tác chịu hoàn toàn trách nhiệm pháp lý liên quan đến nghĩa vụ thuế.'),
                  _section('4. Không hoàn tiền (clawback)',
                      'Hoa hồng đã ghi nhận và đã thanh toán SẼ KHÔNG bị thu hồi nếu Referee yêu cầu hoàn tiền sau ngày Sàn thanh toán cho Đối tác.\n\n'
                      'Trong thời gian T+45 trước khi thanh toán, nếu Referee được hoàn tiền, hoa hồng tương ứng sẽ bị huỷ.'),
                  _section('5. Hành vi cấm',
                      '• Không tự đăng ký bằng SĐT/email khác để tự nhận hoa hồng (self-referral).\n'
                      '• Không spam, không quảng cáo sai sự thật về Sàn.\n'
                      '• Không hứa hẹn thưởng/quà tặng vượt ngoài chương trình chính thức.\n'
                      '• Vi phạm → tạm khoá tài khoản Đối tác + huỷ hoa hồng chưa thanh toán.'),
                  _section('6. Bảo mật người được giới thiệu',
                      'Thông tin Referee (SĐT, tên) được mask sẵn trong dashboard. '
                      'Đối tác cam kết không lưu, share, hoặc dùng thông tin này cho mục đích khác ngoài chương trình.'),
                  _section('7. Thay đổi điều khoản',
                      'Sàn có thể điều chỉnh % hoa hồng, ngưỡng payout, hoặc thời gian đối soát. '
                      'Sàn báo trước 30 ngày qua app + email trước khi áp dụng. '
                      'Hoa hồng đã ghi nhận trước khi điều khoản mới có hiệu lực vẫn được tính theo điều khoản cũ (snapshot tại thời điểm payment).'),
                  _section('8. Chấm dứt chương trình',
                      'Sàn có thể chấm dứt chương trình Affiliate bất kỳ lúc nào, báo trước 60 ngày. '
                      'Hoa hồng đã ghi nhận đến ngày chấm dứt vẫn được thanh toán đầy đủ.'),
                  _section('9. Pháp luật áp dụng',
                      'Điều khoản này tuân theo Luật Việt Nam. Tranh chấp giải quyết tại Toà án có thẩm quyền nơi đặt trụ sở Sàn Giá Gạo.'),
                  if (_alreadyAccepted)
                    Container(
                      margin: const EdgeInsets.symmetric(vertical: 16),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.green.shade50,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: const Text(
                        '✓ Bạn đã đồng ý điều khoản phiên bản hiện hành.',
                        style: TextStyle(color: Colors.green),
                      ),
                    ),
                ],
              ),
            ),
          ),
          if (showAccept)
            Container(
              padding: const EdgeInsets.all(16),
              decoration: const BoxDecoration(
                border: Border(top: BorderSide(color: Color(0xFFE0E0E0))),
              ),
              child: Column(
                children: [
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    controlAffinity: ListTileControlAffinity.leading,
                    value: _checked,
                    onChanged: (v) => setState(() => _checked = v ?? false),
                    title: const Text('Tôi đã đọc, hiểu, và đồng ý các điều khoản trên', style: TextStyle(fontSize: 13)),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: double.infinity,
                    child: FilledButton(
                      onPressed: (_checked && !_saving) ? _accept : null,
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        child: Text(_saving ? 'Đang lưu…' : 'Đồng ý & Tiếp tục'),
                      ),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _section(String title, String body) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
            const SizedBox(height: 6),
            Text(body, style: const TextStyle(fontSize: 13, height: 1.5)),
          ],
        ),
      );
}
