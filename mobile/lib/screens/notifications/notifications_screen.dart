import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../models/rating.dart';
import '../../providers/providers.dart';
import '../../widgets/empty_state.dart';
import '../../theme/app_theme.dart';
import '../../widgets/shimmer_loading.dart';

class NotificationsScreen extends ConsumerStatefulWidget {
  const NotificationsScreen({super.key});

  @override
  ConsumerState<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends ConsumerState<NotificationsScreen> {
  List<AppNotification> _notifications = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final result = await ref.read(apiServiceProvider).getNotifications();
      setState(() => _notifications = result.data);
    } catch (e) {
      debugPrint('Load notifications error: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _markRead(AppNotification notif) async {
    if (notif.isRead) return;
    try {
      await ref.read(apiServiceProvider).markNotificationRead(notif.id);
      setState(() {
        final idx = _notifications.indexWhere((n) => n.id == notif.id);
        if (idx != -1) {
          _notifications[idx] = AppNotification(
            id: notif.id,
            userId: notif.userId,
            type: notif.type,
            title: notif.title,
            body: notif.body,
            isRead: true,
            createdAt: notif.createdAt,
          );
        }
      });
    } catch (e) {
      debugPrint('Mark read error: $e');
    }
  }

  IconData _iconForType(String type) {
    switch (type) {
      case 'message':
        return Icons.chat_bubble_outline;
      case 'rating':
        return Icons.star_outline;
      case 'subscription':
        return Icons.card_membership;
      case 'report':
        return Icons.flag_outlined;
      default:
        return Icons.notifications_outlined;
    }
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
      appBar: AppBar(title: const Text('Thông báo')),
      body: _loading
          ? const ListSkeleton()
          : _notifications.isEmpty
              ? const EmptyState(
                  icon: Icons.notifications_none,
                  title: 'Chưa có thông báo',
                  subtitle: 'Thông báo mới sẽ hiển thị ở đây',
                )
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView.separated(
                    itemCount: _notifications.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (_, i) {
                      final notif = _notifications[i];
                      return ListTile(
                        leading: Icon(
                          _iconForType(notif.type),
                          color: notif.isRead ? AppColors.textHint : Theme.of(context).colorScheme.primary,
                        ),
                        title: Text(
                          notif.title,
                          style: TextStyle(
                            fontWeight: notif.isRead ? FontWeight.normal : FontWeight.bold,
                          ),
                        ),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(notif.body, maxLines: 2, overflow: TextOverflow.ellipsis),
                            const SizedBox(height: 4),
                            Text(
                              _formatTime(notif.createdAt),
                              style: TextStyle(fontSize: 12, color: AppColors.textHint),
                            ),
                          ],
                        ),
                        tileColor: notif.isRead ? null : Theme.of(context).colorScheme.primary.withValues(alpha: 0.05),
                        onTap: () => _markRead(notif),
                      );
                    },
                  ),
                ),
    );
  }
}
