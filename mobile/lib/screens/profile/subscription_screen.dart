import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class SubscriptionScreen extends ConsumerStatefulWidget {
  const SubscriptionScreen({super.key});

  @override
  ConsumerState<SubscriptionScreen> createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends ConsumerState<SubscriptionScreen> {
  Map<String, dynamic>? _status;
  List<Map<String, dynamic>> _plans = [];
  List<Map<String, dynamic>> _history = [];
  int _historyTotal = 0;
  bool _loading = true;
  String? _error;

  final _currencyFormat = NumberFormat.currency(locale: 'vi_VN', symbol: 'đ');

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiServiceProvider);
      final results = await Future.wait([
        api.getSubscriptionStatus(),
        api.getSubscriptionPlans(),
        api.getSubscriptionHistory(),
      ]);
      if (mounted) {
        final plansData = results[1];
        final historyData = results[2];
        setState(() {
          _status = results[0];
          _plans = (plansData['plans'] as List?)?.cast<Map<String, dynamic>>() ?? [];
          _history = (historyData['data'] as List?)?.cast<Map<String, dynamic>>() ?? [];
          _historyTotal = historyData['total'] ?? 0;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = 'Không thể tải thông tin gói dịch vụ';
          _loading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Gói dịch vụ')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(_error!, style: const TextStyle(color: AppColors.error)),
                      const SizedBox(height: 12),
                      FilledButton(onPressed: _load, child: const Text('Thử lại')),
                    ],
                  ),
                )
              : RefreshIndicator(onRefresh: _load, child: _buildContent()),
    );
  }

  Widget _buildContent() {
    final sub = _status?['subscription'] as Map<String, dynamic>?;
    final isActive = _status?['is_active'] == true;
    final daysLeft = _status?['days_left'] ?? 0;

    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 20, 16, 32),
      children: [
        // Status card
        _buildStatusCard(sub, isActive, daysLeft),
        const SizedBox(height: 24),

        // Plans section
        const Text('Bảng giá gia hạn', style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
        const SizedBox(height: 12),
        _buildPlansGrid(),
        const SizedBox(height: 8),

        // Renewal info
        _buildInfoCard(),

        if (!isActive) ...[
          const SizedBox(height: 16),
          _buildWarningCard(),
        ],

        // History section
        if (_history.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text('Lịch sử gia hạn ($_historyTotal)', style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
          const SizedBox(height: 12),
          ..._history.map(_buildHistoryItem),
        ],
      ],
    );
  }

  Widget _buildStatusCard(Map<String, dynamic>? sub, bool isActive, int daysLeft) {
    String? plan;
    DateTime? expiresAt;

    if (sub != null) {
      plan = sub['plan'] as String?;
      if (sub['expires_at'] != null) {
        expiresAt = DateTime.tryParse(sub['expires_at'].toString());
      }
    }

    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          gradient: LinearGradient(
            colors: isActive
                ? [
                    Theme.of(context).colorScheme.onPrimaryContainer,
                    Theme.of(context).colorScheme.primary,
                    Theme.of(context).colorScheme.primaryContainer.withValues(alpha: 1.0),
                  ]
                : [const Color(0xFFC62828), AppColors.error, const Color(0xFFEF5350)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
        ),
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(
              isActive ? Icons.verified : Icons.cancel,
              size: 48,
              color: Colors.white,
            ),
            const SizedBox(height: 10),
            Text(
              isActive ? 'Đang hoạt động' : 'Đã hết hạn',
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white),
            ),
            if (plan != null) ...[
              const SizedBox(height: 4),
              Text(_planLabel(plan), style: const TextStyle(fontSize: 13, color: Colors.white70)),
            ],
            if (isActive && daysLeft > 0) ...[
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  'Còn $daysLeft ngày',
                  style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Colors.white),
                ),
              ),
            ],
            if (expiresAt != null) ...[
              const SizedBox(height: 8),
              Text(
                'Hạn: ${DateFormat('dd/MM/yyyy').format(expiresAt)}',
                style: const TextStyle(fontSize: 12, color: Colors.white60),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildPlansGrid() {
    if (_plans.isEmpty) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Text('Chưa có gói dịch vụ nào', style: TextStyle(color: AppColors.textSecondary)),
        ),
      );
    }

    // Base price from 1-month plan
    final oneMonthPlan = _plans.firstWhere((p) => p['months'] == 1, orElse: () => _plans.first);
    final basePerMonth = (oneMonthPlan['amount'] as num) / (oneMonthPlan['months'] as num);

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: 10,
        mainAxisSpacing: 10,
        childAspectRatio: 1.1,
      ),
      itemCount: _plans.length,
      itemBuilder: (context, index) {
        final plan = _plans[index];
        final months = plan['months'] as int;
        final amount = (plan['amount'] as num).toInt();
        final label = plan['label'] as String;
        final originalPrice = (basePerMonth * months).round();
        final discount = originalPrice > 0 ? ((1 - amount / originalPrice) * 100).round() : 0;

        return Card(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          child: Stack(
            children: [
              Padding(
                padding: const EdgeInsets.all(14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(label, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    if (discount > 0) ...[
                      const SizedBox(height: 4),
                      Text(
                        _currencyFormat.format(originalPrice),
                        style: TextStyle(
                          fontSize: 12,
                          color: AppColors.textHint,
                          decoration: TextDecoration.lineThrough,
                        ),
                      ),
                    ],
                    const SizedBox(height: 2),
                    Text(
                      _currencyFormat.format(amount),
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: AppColors.primary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      '${_currencyFormat.format((amount / months).round())}/tháng',
                      style: TextStyle(fontSize: 11, color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ),
              if (discount > 0)
                Positioned(
                  top: 8,
                  right: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: AppColors.error,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Text(
                      '-$discount%',
                      style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                    ),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildInfoCard() {
    return Card(
      color: AppColors.info.withValues(alpha: 0.06),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.info_outline, color: AppColors.info, size: 20),
                const SizedBox(width: 8),
                Text('Hướng dẫn gia hạn',
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.info)),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Để gia hạn gói dịch vụ, vui lòng chuyển khoản theo số tiền gói đã chọn và liên hệ quản trị viên để xác nhận kích hoạt.',
              style: TextStyle(color: AppColors.info, fontSize: 13, height: 1.4),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildWarningCard() {
    return Card(
      color: AppColors.warning.withValues(alpha: 0.08),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.warning_amber, color: AppColors.warning, size: 20),
                const SizedBox(width: 8),
                Text('Gói đã hết hạn',
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.warning)),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Tin đăng của bạn đã bị tạm ẩn khỏi sàn. Gia hạn gói để tin đăng được hiển thị trở lại.',
              style: TextStyle(color: AppColors.warning, fontSize: 13, height: 1.4),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHistoryItem(Map<String, dynamic> sub) {
    final dateFormat = DateFormat('dd/MM/yyyy');
    final amount = (sub['amount'] as num?)?.toInt() ?? 0;
    final months = sub['duration_months'] ?? 0;
    final status = sub['status'] as String? ?? '';
    final createdAt = DateTime.tryParse(sub['created_at']?.toString() ?? '');
    final expiresAt = DateTime.tryParse(sub['expires_at']?.toString() ?? '');
    final isActive = status == 'active' && expiresAt != null && expiresAt.isAfter(DateTime.now());

    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: isActive ? AppColors.primary.withValues(alpha: 0.1) : AppColors.textHint.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                isActive ? Icons.check_circle : Icons.history,
                color: isActive ? AppColors.primary : AppColors.textHint,
                size: 22,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Gói $months tháng',
                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    createdAt != null ? dateFormat.format(createdAt) : '-',
                    style: TextStyle(fontSize: 12, color: AppColors.textSecondary),
                  ),
                ],
              ),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  _currencyFormat.format(amount),
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: AppColors.primary),
                ),
                const SizedBox(height: 2),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: isActive
                        ? AppColors.primary.withValues(alpha: 0.1)
                        : AppColors.textHint.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    isActive ? 'Hoạt động' : 'Hết hạn',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w500,
                      color: isActive ? AppColors.primary : AppColors.textHint,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _planLabel(String? plan) {
    switch (plan) {
      case 'free_trial':
        return 'Dùng thử miễn phí';
      case 'monthly':
        return 'Gói tháng';
      default:
        return 'Chưa có gói';
    }
  }
}
