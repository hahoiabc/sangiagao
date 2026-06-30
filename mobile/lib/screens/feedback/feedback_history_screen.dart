import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/paginated_list_view.dart';

class FeedbackHistoryScreen extends ConsumerStatefulWidget {
  const FeedbackHistoryScreen({super.key});

  @override
  ConsumerState<FeedbackHistoryScreen> createState() => _FeedbackHistoryScreenState();
}

class _FeedbackHistoryScreenState extends ConsumerState<FeedbackHistoryScreen> {
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dateFormat = DateFormat('dd/MM/yyyy HH:mm');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Lịch sử góp ý'),
      ),
      body: PaginatedListView<dynamic>(
        padding: const EdgeInsets.all(12),
        pageSize: 20,
        fetcher: (page) async =>
            await ref.read(apiServiceProvider).getMyFeedbacks(page: page, limit: 20),
        emptyState: const Center(
          child: Padding(
            padding: EdgeInsets.symmetric(vertical: 48),
            child: Text('Chưa có góp ý nào', style: TextStyle(color: AppColors.textHint)),
          ),
        ),
        itemBuilder: (context, item, index) {
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
    );
  }
}
