import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';

class MyPayoutsScreen extends ConsumerStatefulWidget {
  const MyPayoutsScreen({super.key});

  @override
  ConsumerState<MyPayoutsScreen> createState() => _MyPayoutsScreenState();
}

class _MyPayoutsScreenState extends ConsumerState<MyPayoutsScreen> {
  List<Map<String, dynamic>> _payouts = const [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = ref.read(apiServiceProvider);
    try {
      final list = await api.getMyPayouts();
      if (!mounted) return;
      setState(() {
        _payouts = list;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat('#,###', 'vi_VN');
    final totalSent = _payouts
        .where((p) => p['status'] == 'sent')
        .fold<int>(0, (s, p) => s + ((p['total_amount'] as num?)?.toInt() ?? 0));
    final totalPending = _payouts
        .where((p) => p['status'] == 'pending')
        .fold<int>(0, (s, p) => s + ((p['total_amount'] as num?)?.toInt() ?? 0));

    return Scaffold(
      appBar: AppBar(title: const Text('Lịch sử thanh toán')),
      body: RefreshIndicator(
        onRefresh: _load,
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _row('Đã nhận', '${fmt.format(totalSent)} đ', color: Colors.blue),
                          _row('Chờ chuyển', '${fmt.format(totalPending)} đ', color: Colors.orange),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (_payouts.isEmpty)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 32),
                      child: Text(
                        'Chưa có khoản thanh toán nào.\nHoa hồng đạt ngưỡng tối thiểu sẽ được admin tạo payout cho bạn.',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey),
                      ),
                    )
                  else
                    ..._payouts.map(_buildPayout),
                ],
              ),
      ),
    );
  }

  Widget _row(String label, String value, {Color? color}) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [Text(label), Text(value, style: TextStyle(fontWeight: FontWeight.w600, color: color))],
        ),
      );

  Widget _buildPayout(Map<String, dynamic> p) {
    final fmt = NumberFormat('#,###', 'vi_VN');
    final amount = (p['total_amount'] as num?)?.toInt() ?? 0;
    final count = p['record_count'] ?? 0;
    final method = (p['method'] ?? '').toString();
    final status = p['status'] ?? 'pending';
    final createdAt = _formatDate(p['created_at']?.toString() ?? '');
    final sentAt = _formatDate(p['sent_at']?.toString() ?? '');

    String statusLabel;
    Color statusColor;
    switch (status) {
      case 'sent':
        statusLabel = 'Đã chuyển';
        statusColor = Colors.blue;
        break;
      case 'failed':
        statusLabel = 'Thất bại';
        statusColor = Colors.red;
        break;
      default:
        statusLabel = 'Chờ chuyển';
        statusColor = Colors.orange;
    }

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('${fmt.format(amount)} đ', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(statusLabel, style: TextStyle(fontSize: 12, color: statusColor)),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text('$count khoản hoa hồng · ${method.toUpperCase()}', style: const TextStyle(fontSize: 12, color: Colors.grey)),
            const SizedBox(height: 4),
            Text('Tạo: $createdAt', style: const TextStyle(fontSize: 12, color: Colors.grey)),
            if (status == 'sent' && sentAt.isNotEmpty)
              Text('Đã chuyển: $sentAt', style: const TextStyle(fontSize: 12, color: Colors.blue)),
          ],
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
