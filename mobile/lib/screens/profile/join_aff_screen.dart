import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import 'aff_terms_screen.dart';

/// Member-facing landing page to opt into the Affiliate program.
/// After successful activation, user becomes 'aff' and is redirected to
/// the full dashboard at /referral.
class JoinAffScreen extends ConsumerStatefulWidget {
  const JoinAffScreen({super.key});

  @override
  ConsumerState<JoinAffScreen> createState() => _JoinAffScreenState();
}

class _JoinAffScreenState extends ConsumerState<JoinAffScreen> {
  bool _activating = false;
  int _stage1Days = 90;
  double _stage1Pct = 0.5;
  int _stage2Days = 180;
  double _stage2Pct = 0.3;
  double _stage3Pct = 0.2;
  int _minimumPayout = 100000;

  @override
  void initState() {
    super.initState();
    _loadRule();
  }

  Future<void> _loadRule() async {
    try {
      final r = await ref.read(apiServiceProvider).getAffTerms();
      if (!mounted) return;
      final rule = (r['rule'] as Map?) ?? const {};
      setState(() {
        _stage1Days = (rule['stage1_days'] as num?)?.toInt() ?? _stage1Days;
        _stage1Pct = (rule['stage1_pct'] as num?)?.toDouble() ?? _stage1Pct;
        _stage2Days = (rule['stage2_days'] as num?)?.toInt() ?? _stage2Days;
        _stage2Pct = (rule['stage2_pct'] as num?)?.toDouble() ?? _stage2Pct;
        _stage3Pct = (rule['stage3_pct'] as num?)?.toDouble() ?? _stage3Pct;
        _minimumPayout = (rule['minimum_payout'] as num?)?.toInt() ?? _minimumPayout;
      });
    } catch (_) {}
  }

  String _pct(double v) => '${(v * 100).toStringAsFixed(0)}%';

  Future<void> _activate() async {
    // Show T&C and require explicit accept before activating
    final accepted = await Navigator.of(context).push<bool>(
      MaterialPageRoute(builder: (_) => const AffTermsScreen(requireAccept: true)),
    );
    if (accepted != true) return;
    setState(() => _activating = true);
    try {
      await ref.read(apiServiceProvider).becomeAffiliate();
      await ref.read(authProvider.notifier).refreshUser();
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã kích hoạt vai trò đối tác. Chúc may mắn!')),
      );
      // Replace this screen with the dashboard
      context.go('/referral');
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Không thể kích hoạt. Vui lòng thử lại.')),
      );
    } finally {
      if (mounted) setState(() => _activating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final role = ref.watch(authProvider).user?.role;
    final isAlreadyAff = role == 'aff';

    return Scaffold(
      appBar: AppBar(title: const Text('Đăng ký làm Đối tác')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            color: Colors.amber.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(Icons.star, color: Colors.amber, size: 24),
                      SizedBox(width: 8),
                      Text('Chương trình Đối tác Affiliate',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  const Text(
                    'Nhận hoa hồng mỗi khi bạn bạn giới thiệu đăng ký + mua gói thành viên Sàn Giá Gạo.',
                    style: TextStyle(fontSize: 13, height: 1.4),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          _benefit('Hoa hồng theo lần thanh toán',
              'Lần đầu tiên: ${_pct(_stage1Pct)} doanh thu ròng\n'
              'Lần thứ 2: ${_pct(_stage2Pct)}\n'
              'Từ lần thứ 3 (vĩnh viễn): ${_pct(_stage3Pct)}'),
          _benefit('Mã giới thiệu riêng',
              'Mỗi đối tác có 1 mã + link riêng. Sàn tự động ghi nhận hoa hồng khi referee đăng ký qua link.'),
          _benefit('Theo dõi minh bạch',
              'Xem danh sách người được giới thiệu, status gói, hoa hồng từng kỳ, lịch sử thanh toán.'),
          _benefit('Rút tiền linh hoạt',
              'Đạt ngưỡng tối thiểu (hiện hành ${(_minimumPayout / 1000).toStringAsFixed(0)}.000đ) → Sàn chuyển khoản ngân hàng. Phí CK thực tế trừ trực tiếp từ payout.'),
          const SizedBox(height: 16),
          if (isAlreadyAff)
            const Card(
              color: Color(0xFFE8F5E9),
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Text('✓ Bạn đã là Đối tác Affiliate.', style: TextStyle(color: Colors.green)),
              ),
            )
          else
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                style: FilledButton.styleFrom(
                  backgroundColor: Colors.amber.shade700,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
                onPressed: _activating ? null : _activate,
                child: Text(
                  _activating ? 'Đang kích hoạt…' : 'Đọc điều khoản & Kích hoạt',
                  style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _benefit(String title, String body) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 6),
            Text(body, style: const TextStyle(fontSize: 13, height: 1.4, color: Colors.black87)),
          ],
        ),
      ),
    );
  }
}
