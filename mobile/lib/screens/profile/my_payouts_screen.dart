import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../widgets/paginated_list_view.dart';

class MyPayoutsScreen extends ConsumerStatefulWidget {
  const MyPayoutsScreen({super.key});

  @override
  ConsumerState<MyPayoutsScreen> createState() => _MyPayoutsScreenState();
}

class _MyPayoutsScreenState extends ConsumerState<MyPayoutsScreen> {
  final _fmt = NumberFormat('#,###', 'vi_VN');

  // Tổng dồn theo các trang đã tải (cập nhật khi cuộn-tải-thêm).
  int _totalSent = 0;
  int _totalPending = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Lịch sử thanh toán')),
      body: PaginatedListView<Map<String, dynamic>>(
        padding: const EdgeInsets.all(16),
        fetcher: (page) async {
          final list = await ref.read(apiServiceProvider).getMyPayouts(page: page, limit: 20);
          if (page == 1) {
            _totalSent = 0;
            _totalPending = 0;
          }
          for (final p in list) {
            final amount = (p['total_amount'] as num?)?.toInt() ?? 0;
            if (p['status'] == 'sent') {
              _totalSent += amount;
            } else if (p['status'] == 'pending') {
              _totalPending += amount;
            }
          }
          if (mounted) setState(() {});
          return list;
        },
        header: Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _row('Đã nhận', '${_fmt.format(_totalSent)} đ', color: Colors.blue),
                  _row('Chờ chuyển', '${_fmt.format(_totalPending)} đ', color: Colors.orange),
                ],
              ),
            ),
          ),
        ),
        emptyState: const Padding(
          padding: EdgeInsets.symmetric(vertical: 32),
          child: Text(
            'Chưa có khoản thanh toán nào.\nHoa hồng đạt ngưỡng tối thiểu sẽ được admin tạo payout cho bạn.',
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.grey),
          ),
        ),
        itemBuilder: (context, p, index) => _buildPayout(p),
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
                Text('${_fmt.format(amount)} đ', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
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
