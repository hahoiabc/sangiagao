import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../widgets/paginated_list_view.dart';

class MyRefereesScreen extends ConsumerWidget {
  const MyRefereesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('Người tôi giới thiệu')),
      body: PaginatedListView<Map<String, dynamic>>(
        padding: const EdgeInsets.all(16),
        fetcher: (page) async =>
            await ref.read(apiServiceProvider).getMyReferees(page: page, limit: 20),
        emptyState: const Padding(
          padding: EdgeInsets.symmetric(vertical: 32),
          child: Text(
            'Chưa có ai đăng ký qua link giới thiệu.\nHãy chia sẻ link để bắt đầu!',
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.grey),
          ),
        ),
        itemBuilder: (context, r, i) => _buildReferee(r),
      ),
    );
  }

  Widget _buildReferee(Map<String, dynamic> r) {
    final fmt = NumberFormat('#,###', 'vi_VN');
    final phone = r['phone'] ?? '';
    final name = r['name'] ?? '';
    final registered = _formatDate(r['registered_at']?.toString() ?? '');
    final subStatus = r['sub_status'] ?? 'none';
    final count = r['commission_count'] ?? 0;
    final total = (r['total_commission'] as num?)?.toInt() ?? 0;
    final paid = (r['paid_commission'] as num?)?.toInt() ?? 0;

    String statusLabel;
    Color statusColor;
    switch (subStatus) {
      case 'active':
        statusLabel = 'Đang dùng';
        statusColor = Colors.green;
        break;
      case 'expired':
        statusLabel = 'Hết hạn';
        statusColor = Colors.grey;
        break;
      default:
        statusLabel = 'Chưa mua';
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
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(phone, style: const TextStyle(fontWeight: FontWeight.w600, fontFamily: 'monospace')),
                      if (name.isNotEmpty)
                        Text(name, style: const TextStyle(fontSize: 13, color: Colors.grey)),
                    ],
                  ),
                ),
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
            const Divider(height: 16),
            _smallRow('Đăng ký', registered),
            _smallRow('Số lượt mua', '$count'),
            _smallRow('Tổng hoa hồng', '${fmt.format(total)} đ'),
            _smallRow('Đã trả', '${fmt.format(paid)} đ', color: Colors.blue),
          ],
        ),
      ),
    );
  }

  Widget _smallRow(String label, String value, {Color? color}) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 2),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label, style: const TextStyle(fontSize: 13, color: Colors.grey)),
            Text(value, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w500, color: color)),
          ],
        ),
      );

  String _formatDate(String iso) {
    try {
      final d = DateTime.parse(iso).toLocal();
      return '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';
    } catch (_) {
      return iso;
    }
  }
}
