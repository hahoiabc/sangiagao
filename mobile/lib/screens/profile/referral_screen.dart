import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
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

  @override
  Widget build(BuildContext context) {
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
                      _buildCodeCard(),
                      const SizedBox(height: 16),
                      _buildStatsCard(),
                      const SizedBox(height: 16),
                      _buildHistoryHeader(),
                      ..._history.map(_buildHistoryItem),
                      if (_history.isEmpty)
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
