import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../models/inbox.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/shimmer_loading.dart';

class SystemInboxScreen extends ConsumerStatefulWidget {
  const SystemInboxScreen({super.key});

  @override
  ConsumerState<SystemInboxScreen> createState() => _SystemInboxScreenState();
}

class _SystemInboxScreenState extends ConsumerState<SystemInboxScreen> {
  List<InboxMessage> _items = [];
  bool _loading = true;
  int _unreadCount = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final result = await ref.read(apiServiceProvider).getInbox();
      if (mounted) {
        setState(() {
          _items = result.items;
          _unreadCount = result.unreadCount;
        });
      }
    } catch (e) {
      debugPrint('Load inbox error: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _markRead(int index) async {
    final item = _items[index];
    if (item.isRead) return;
    try {
      await ref.read(apiServiceProvider).markInboxRead(item.id);
      setState(() {
        _items[index] = InboxMessage(
          id: item.id,
          title: item.title,
          body: item.body,
          imageUrl: item.imageUrl,
          isPinned: item.isPinned,
          isRead: true,
          createdAt: item.createdAt,
        );
        if (_unreadCount > 0) _unreadCount--;
      });
      ref.read(inboxUnreadProvider.notifier).refresh();
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
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hộp thư'),
        actions: [
          if (_unreadCount > 0)
            Center(
              child: Padding(
                padding: const EdgeInsets.only(right: 16),
                child: Text(
                  '$_unreadCount chưa đọc',
                  style: TextStyle(fontSize: 13, color: AppColors.primary),
                ),
              ),
            ),
        ],
      ),
      body: _loading
          ? const ListSkeleton()
          : _items.isEmpty
              ? const EmptyState(
                  icon: Icons.mail_outline,
                  title: 'Chưa có thông báo',
                  subtitle: 'Thông báo từ hệ thống sẽ hiển thị ở đây',
                )
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView.separated(
                    itemCount: _items.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (_, i) {
                      final item = _items[i];
                      return ListTile(
                        tileColor: item.isRead
                            ? null
                            : AppColors.primary.withValues(alpha: 0.06),
                        leading: Icon(
                          item.isPinned ? Icons.push_pin : Icons.mail_outline,
                          color: item.isRead
                              ? AppColors.textHint
                              : AppColors.primary,
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
                          _markRead(i);
                          context.push('/system-inbox/${item.id}');
                        },
                      );
                    },
                  ),
                ),
    );
  }
}
