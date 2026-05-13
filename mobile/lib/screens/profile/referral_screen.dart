import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:share_plus/share_plus.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';

class ReferralScreen extends ConsumerStatefulWidget {
  const ReferralScreen({super.key});

  @override
  ConsumerState<ReferralScreen> createState() => _ReferralScreenState();
}

class _ReferralScreenState extends ConsumerState<ReferralScreen> {
  Map<String, dynamic>? _stats;
  List<Map<String, dynamic>> _history = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = ref.read(apiServiceProvider);
    try {
      final stats = await api.getReferralStats();
      final hist = await api.getReferralHistory(limit: 20);
      if (!mounted) return;
      setState(() {
        _stats = stats;
        _history = hist;
        _loading = false;
        _error = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Không tải được dữ liệu';
        _loading = false;
      });
    }
  }

  String get _shareLink {
    final code = _stats?['code'] as String? ?? '';
    return 'https://sangiagao.vn/r/$code';
  }

  String get _shareMessage {
    final code = _stats?['code'] as String? ?? '';
    return 'Tham gia Sàn Giá Gạo qua link giới thiệu của tôi để xem giá gạo realtime '
        'và kết nối trực tiếp với thương lái:\n\n$_shareLink\n\nMã giới thiệu: $code';
  }

  Future<void> _becomeAffiliate() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Trở thành đối tác'),
        content: const Text(
          'Khi bạn giới thiệu bạn bè đăng ký + mua gói, bạn nhận hoa hồng theo quy tắc của Sàn Giá Gạo:\n'
          '\n• Giai đoạn 1 (3 tháng đầu): 50%\n'
          '• Giai đoạn 2 (6 tháng kế): 30%\n'
          '• Giai đoạn 3 (vĩnh viễn): 20%\n'
          '\nTính trên doanh thu ròng (sau phí App Store nếu có). Thanh toán sau 45 ngày kể từ giao dịch.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Đồng ý & Kích hoạt')),
        ],
      ),
    );
    if (confirm != true) return;
    final api = ref.read(apiServiceProvider);
    try {
      await api.becomeAffiliate();
      await ref.read(authProvider.notifier).refreshUser();
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã kích hoạt vai trò đối tác. Chúc may mắn!')),
      );
      _load();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Không thể kích hoạt. Vui lòng thử lại.')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authProvider).user;
    final isAff = user?.role == 'aff';
    final isMember = user?.role == 'member';

    return Scaffold(
      appBar: AppBar(title: const Text('Giới thiệu bạn bè')),
      body: RefreshIndicator(
        onRefresh: _load,
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? Center(child: Text(_error!))
                : ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      if (isMember) _buildBecomeAffiliateBanner(),
                      if (isMember) const SizedBox(height: 12),
                      _buildCodeCard(),
                      const SizedBox(height: 16),
                      _buildStatsCard(),
                      if (isAff) const SizedBox(height: 16),
                      if (isAff) _buildQuickNav(),
                      if (isAff) const SizedBox(height: 16),
                      if (isAff) _buildHistoryHeader(),
                      if (isAff) ..._history.map(_buildHistoryItem),
                      if (isAff && _history.isEmpty)
                        const Padding(
                          padding: EdgeInsets.symmetric(vertical: 24),
                          child: Text(
                            'Chưa có hoa hồng nào. Hãy chia sẻ link để bắt đầu kiếm tiền!',
                            textAlign: TextAlign.center,
                            style: TextStyle(color: Colors.grey),
                          ),
                        ),
                    ],
                  ),
      ),
    );
  }

  Widget _buildBecomeAffiliateBanner() {
    return Card(
      color: Colors.amber.shade50,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.star, color: Colors.amber),
                SizedBox(width: 8),
                Text('Trở thành đối tác chính thức', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
              ],
            ),
            const SizedBox(height: 8),
            const Text(
              'Kích hoạt vai trò Đối tác Aff để xem thống kê chi tiết người bạn giới thiệu, '
              'theo dõi hoa hồng theo từng giai đoạn, và rút tiền khi đạt ngưỡng.',
              style: TextStyle(fontSize: 13, height: 1.4),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                style: FilledButton.styleFrom(backgroundColor: Colors.amber.shade700),
                onPressed: _becomeAffiliate,
                child: const Text('Kích hoạt làm đối tác'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCodeCard() {
    final code = _stats?['code'] as String? ?? '------';
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Mã giới thiệu của bạn', style: TextStyle(color: Colors.grey, fontSize: 13)),
            const SizedBox(height: 4),
            Row(
              children: [
                Expanded(
                  child: Text(
                    code,
                    style: const TextStyle(
                      fontSize: 32,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 4,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.copy),
                  tooltip: 'Sao chép mã',
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: code));
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Đã sao chép mã'), duration: Duration(seconds: 1)),
                    );
                  },
                ),
              ],
            ),
            const SizedBox(height: 8),
            const Divider(),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(child: Text(_shareLink, style: const TextStyle(fontSize: 13, color: Colors.blue))),
                IconButton(
                  icon: const Icon(Icons.copy_outlined, size: 20),
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: _shareLink));
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Đã sao chép link'), duration: Duration(seconds: 1)),
                    );
                  },
                ),
              ],
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                icon: const Icon(Icons.share),
                label: const Text('Chia sẻ ngay'),
                onPressed: () {
                  Share.share(_shareMessage, subject: 'Sàn Giá Gạo - Mã giới thiệu');
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatsCard() {
    final s = _stats ?? const {};
    final fmt = NumberFormat('#,###', 'vi_VN');
    final totalReferrals = s['total_referrals'] ?? 0;
    final active = s['active_referees'] ?? 0;
    final totalEarned = (s['total_earned'] as num?)?.toInt() ?? 0;
    final payable = (s['payable_amount'] as num?)?.toInt() ?? 0;
    final pending = (s['pending_amount'] as num?)?.toInt() ?? 0;
    final paid = (s['paid_amount'] as num?)?.toInt() ?? 0;
    final minPayout = (s['minimum_payout'] as num?)?.toInt() ?? 0;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Thống kê', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
            const SizedBox(height: 12),
            _row('Đã giới thiệu', '$totalReferrals người'),
            _row('Đang hoạt động', '$active người'),
            const Divider(height: 24),
            _row('Tổng hoa hồng', '${fmt.format(totalEarned)} đ'),
            _row('Có thể nhận', '${fmt.format(payable)} đ', color: Colors.green),
            _row('Chờ đối soát (T+45)', '${fmt.format(pending)} đ', color: Colors.orange),
            _row('Đã nhận', '${fmt.format(paid)} đ', color: Colors.blue),
            const Divider(height: 24),
            Text(
              'Ngưỡng thanh toán tối thiểu: ${fmt.format(minPayout)} đ. Khi tích đủ, '
              'admin sẽ liên hệ chuyển khoản.',
              style: const TextStyle(fontSize: 12, color: Colors.grey),
            ),
          ],
        ),
      ),
    );
  }

  Widget _row(String label, String value, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label),
          Text(value, style: TextStyle(fontWeight: FontWeight.w600, color: color)),
        ],
      ),
    );
  }

  Widget _buildQuickNav() {
    return Row(
      children: [
        Expanded(
          child: OutlinedButton.icon(
            icon: const Icon(Icons.people_outline, size: 18),
            label: const Text('Người tôi giới thiệu'),
            onPressed: () => context.push('/referral/referees'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: OutlinedButton.icon(
            icon: const Icon(Icons.account_balance_wallet_outlined, size: 18),
            label: const Text('Lịch sử thanh toán'),
            onPressed: () => context.push('/referral/payouts'),
          ),
        ),
      ],
    );
  }

  Widget _buildHistoryHeader() {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: 8),
      child: Text('Lịch sử hoa hồng', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
    );
  }

  Widget _buildHistoryItem(Map<String, dynamic> rec) {
    final fmt = NumberFormat('#,###', 'vi_VN');
    final amount = (rec['commission_amount'] as num?)?.toInt() ?? 0;
    final stage = rec['stage'] ?? 0;
    final rate = ((rec['rate'] as num?)?.toDouble() ?? 0) * 100;
    final status = rec['status'] ?? '';
    final source = rec['payment_source'] ?? '';
    final createdAt = rec['created_at'] ?? '';

    Color statusColor;
    String statusText;
    switch (status) {
      case 'paid':
        statusColor = Colors.blue;
        statusText = 'Đã nhận';
        break;
      case 'payable':
        statusColor = Colors.green;
        statusText = 'Có thể nhận';
        break;
      case 'cancelled':
        statusColor = Colors.red;
        statusText = 'Đã hủy';
        break;
      default:
        statusColor = Colors.orange;
        statusText = 'Chờ đối soát';
    }

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: ListTile(
        title: Text('${fmt.format(amount)} đ', style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: Text(
          'Giai đoạn $stage (${rate.toStringAsFixed(0)}%) · ${source.toString().toUpperCase()} · ${_formatDate(createdAt.toString())}',
          style: const TextStyle(fontSize: 12),
        ),
        trailing: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          decoration: BoxDecoration(
            color: statusColor.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(4),
          ),
          child: Text(statusText, style: TextStyle(fontSize: 12, color: statusColor)),
        ),
      ),
    );
  }

  String _formatDate(String iso) {
    try {
      final d = DateTime.parse(iso).toLocal();
      return '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';
    } catch (_) {
      return iso;
    }
  }
}
