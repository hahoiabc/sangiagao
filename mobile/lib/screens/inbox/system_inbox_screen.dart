import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../models/inbox.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/paginated_list_view.dart';

class SystemInboxScreen extends ConsumerStatefulWidget {
  const SystemInboxScreen({super.key});

  @override
  ConsumerState<SystemInboxScreen> createState() => _SystemInboxScreenState();
}

class _SystemInboxScreenState extends ConsumerState<SystemInboxScreen> {
  final _listKey = GlobalKey<PaginatedListViewState<InboxMessage>>();

  Future<void> _markRead(InboxMessage item) async {
    if (item.isRead) return;
    try {
      await ref.read(apiServiceProvider).markInboxRead(item.id);
      ref.read(inboxUnreadProvider.notifier).refresh();
      _listKey.currentState?.refresh();
    } catch (_) {}
  }

  String _formatTime(String iso) {
    final dt = DateTime.tryParse(iso)?.toLocal();
    if (dt == null) return '';
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return 'Vừa xong';
    if (diff.inHours < 1) return '${diff.inMinutes} phút trước';
    if (diff.inDays < 1) return '${diff.inHours} giờ trước';
    if (diff.inDays < 7) return '${diff.inDays} ngày trước';
    return '${dt.day}/${dt.month}/${dt.year}';
  }

  @override
  Widget build(BuildContext context) {
    final unreadCount = ref.watch(inboxUnreadProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hộp thư'),
        actions: [
          if (unreadCount > 0)
            Center(
              child: Padding(
                padding: const EdgeInsets.only(right: 16),
                child: Text(
                  '$unreadCount chưa đọc',
                  style: TextStyle(fontSize: 13, color: AppColors.primary),
                ),
              ),
            ),
        ],
      ),
      body: PaginatedListView<InboxMessage>(
        key: _listKey,
        fetcher: (page) async =>
            (await ref.read(apiServiceProvider).getInbox(page: page, limit: 20))
                .items,
        emptyState: const EmptyState(
          icon: Icons.mail_outline,
          title: 'Chưa có thông báo',
          subtitle: 'Thông báo từ hệ thống sẽ hiển thị ở đây',
        ),
        itemBuilder: (context, item, i) {
          return Column(
            children: [
              ListTile(
                tileColor: item.isRead
                    ? null
                    : AppColors.primary.withValues(alpha: 0.06),
                leading: Icon(
                  item.isPinned ? Icons.push_pin : Icons.mail_outline,
                  color: item.isRead ? AppColors.textHint : AppColors.primary,
                ),
                title: Text(
                  item.title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontWeight:
                        item.isRead ? FontWeight.normal : FontWeight.bold,
                  ),
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item.body,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      _formatTime(item.createdAt),
                      style: TextStyle(
                        fontSize: 12,
                        color: AppColors.textHint,
                      ),
                    ),
                  ],
                ),
                onTap: () {
                  _markRead(item);
                  context.push('/system-inbox/${item.id}');
                },
              ),
              const Divider(height: 1),
            ],
          );
        },
      ),
    );
  }
}
