import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class FeedbackHistoryScreen extends ConsumerStatefulWidget {
  const FeedbackHistoryScreen({super.key});

  @override
  ConsumerState<FeedbackHistoryScreen> createState() => _FeedbackHistoryScreenState();
}

class _FeedbackHistoryScreenState extends ConsumerState<FeedbackHistoryScreen> {
  List<dynamic> _items = [];
  bool _loading = true;
  String? _error;

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
      final result = await ref.read(apiServiceProvider).getMyFeedbacks();
      if (mounted) setState(() => _items = result);
    } catch (e) {
      if (mounted) setState(() => _error = 'Không thể tải lịch sử góp ý');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dateFormat = DateFormat('dd/MM/yyyy HH:mm');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Lịch sử góp ý'),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(_error!, style: TextStyle(color: AppColors.textSecondary)),
                      const SizedBox(height: 12),
                      FilledButton.tonal(onPressed: _load, child: const Text('Thử lại')),
                    ],
                  ),
                )
              : _items.isEmpty
                  ? const Center(child: Text('Chưa có góp ý nào', style: TextStyle(color: AppColors.textHint)))
                  : RefreshIndicator(
                      onRefresh: _load,
                      child: ListView.builder(
                        padding: const EdgeInsets.all(12),
                        itemCount: _items.length,
                        itemBuilder: (context, index) {
                          final item = _items[index];
                          final content = item['content'] as String? ?? '';
                          final reply = item['reply'] as String?;
                          final createdAt = DateTime.tryParse(item['created_at'] ?? '');
                          final repliedAt = item['replied_at'] != null ? DateTime.tryParse(item['replied_at']) : null;

                          return Card(
                            margin: const EdgeInsets.only(bottom: 12),
                            child: Padding(
                              padding: const EdgeInsets.all(14),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  // User's feedback
                                  Row(
                                    children: [
                                      Icon(Icons.feedback_outlined, size: 16, color: theme.colorScheme.primary),
                                      const SizedBox(width: 8),
                                      Text('Góp ý của bạn', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: theme.colorScheme.primary)),
                                      const Spacer(),
                                      if (createdAt != null)
                                        Text(dateFormat.format(createdAt.toLocal()), style: TextStyle(fontSize: 11, color: AppColors.textHint)),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  Text(content, style: const TextStyle(fontSize: 14)),

                                  // Reply from dev team
                                  if (reply != null) ...[
                                    const Divider(height: 24),
                                    Row(
                                      children: [
                                        Icon(Icons.reply, size: 16, color: AppColors.primaryDark),
                                        const SizedBox(width: 8),
                                        Text('Phản hồi từ nhà phát triển', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primaryDark)),
                                        const Spacer(),
                                        if (repliedAt != null)
                                          Text(dateFormat.format(repliedAt.toLocal()), style: TextStyle(fontSize: 11, color: AppColors.textHint)),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Container(
                                      width: double.infinity,
                                      padding: const EdgeInsets.all(10),
                                      decoration: BoxDecoration(
                                        color: AppColors.primary.withValues(alpha: 0.08),
                                        borderRadius: BorderRadius.circular(8),
                                      ),
                                      child: Text(reply, style: const TextStyle(fontSize: 14)),
                                    ),
                                  ] else ...[
                                    const SizedBox(height: 8),
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                      decoration: BoxDecoration(
                                        color: AppColors.warning.withValues(alpha: 0.08),
                                        borderRadius: BorderRadius.circular(4),
                                      ),
                                      child: Text('Chờ phản hồi', style: TextStyle(fontSize: 11, color: AppColors.warning)),
                                    ),
                                  ],
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
    );
  }
}
