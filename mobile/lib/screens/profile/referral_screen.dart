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
  Map<String, dynamic>? _bankInfo;
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
      Map<String, dynamic>? bank;
      try {
        bank = await api.getBankInfo();
      } catch (_) {}
      if (!mounted) return;
      setState(() {
        _stats = stats;
        _history = hist;
        _bankInfo = bank;
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
    return 'Tải SanGiaGao để xem giá Gạo và kết nối với thương nhân\n'
        '$_shareLink\nMã giới thiệu: $code';
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
                      // Aff-only dashboard. Members are redirected to /referral/join
                      // from the Profile menu instead (see profile_screen.dart).
                      if (!isAff)
                        Padding(
                          padding: const EdgeInsets.all(24),
                          child: Column(
                            children: [
                              const Icon(Icons.lock_outline, size: 48, color: Colors.grey),
                              const SizedBox(height: 12),
                              const Text(
                                'Cần đăng ký làm Đối tác Affiliate để xem thông tin hoa hồng.',
                                textAlign: TextAlign.center,
                                style: TextStyle(color: Colors.grey),
                              ),
                              const SizedBox(height: 12),
                              if (isMember)
                                FilledButton(
                                  onPressed: () => context.push('/referral/join'),
                                  child: const Text('Đăng ký làm Đối tác'),
                                ),
                            ],
                          ),
                        ),
                      if (isAff && _needBankInfoBanner()) _buildBankInfoBanner(),
                      if (isAff && _needBankInfoBanner()) const SizedBox(height: 12),
                      if (isAff) _buildCodeCard(),
                      if (isAff) const SizedBox(height: 16),
                      if (isAff) _buildStatsCard(),
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

  bool _needBankInfoBanner() {
    final s = _stats;
    if (s == null) return false;
    final payable = (s['payable_amount'] as num?)?.toInt() ?? 0;
    final min = (s['minimum_payout'] as num?)?.toInt() ?? 100000;
    final hasInfo = _bankInfo != null && (_bankInfo!['account_no'] ?? '').toString().isNotEmpty;
    return payable >= min && !hasInfo;
  }

  Widget _buildBankInfoBanner() {
    final fmt = NumberFormat('#,###', 'vi_VN');
    final payable = (_stats?['payable_amount'] as num?)?.toInt() ?? 0;
    return Card(
      color: Colors.blue.shade50,
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.account_balance, color: Colors.blue, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Bạn có ${fmt.format(payable)} đ có thể nhận!',
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 6),
            const Text(
              'Cập nhật thông tin tài khoản ngân hàng để Sàn chuyển hoa hồng cho bạn. '
              'Phí chuyển khoản (nếu có) sẽ trừ trực tiếp từ số tiền nhận.',
              style: TextStyle(fontSize: 12, height: 1.4),
            ),
            const SizedBox(height: 10),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: () async {
                  final updated = await context.push<bool>('/referral/bank-info');
                  if (updated == true && mounted) _load();
                },
                child: const Text('Cập nhật tài khoản nhận hoa hồng'),
              ),
            ),
          ],
        ),
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
