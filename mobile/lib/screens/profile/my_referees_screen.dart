import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';

class MyRefereesScreen extends ConsumerStatefulWidget {
  const MyRefereesScreen({super.key});

  @override
  ConsumerState<MyRefereesScreen> createState() => _MyRefereesScreenState();
}

class _MyRefereesScreenState extends ConsumerState<MyRefereesScreen> {
  List<Map<String, dynamic>> _referees = const [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = ref.read(apiServiceProvider);
    try {
      final list = await api.getMyReferees();
      if (!mounted) return;
      setState(() {
        _referees = list;
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
    final active = _referees.where((r) => r['sub_status'] == 'active').length;
    final totalEarned = _referees.fold<int>(
      0,
      (s, r) => s + ((r['total_commission'] as num?)?.toInt() ?? 0),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Người tôi giới thiệu')),
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
                          _row('Tổng người được giới thiệu', '${_referees.length}'),
                          _row('Đang hoạt động', '$active'),
                          _row('Tổng hoa hồng ghi nhận', '${fmt.format(totalEarned)} đ'),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (_referees.isEmpty)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 32),
                      child: Text(
                        'Chưa có ai đăng ký qua link giới thiệu.\nHãy chia sẻ link để bắt đầu!',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey),
                      ),
                    )
                  else
                    ..._referees.map(_buildReferee),
                ],
              ),
      ),
    );
  }

  Widget _row(String label, String value) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [Text(label), Text(value, style: const TextStyle(fontWeight: FontWeight.w600))],
        ),
      );

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
