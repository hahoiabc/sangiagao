import 'dart:async';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../models/conversation.dart';
import '../../providers/providers.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/shimmer_loading.dart';
import '../../theme/app_theme.dart';

class InboxScreen extends ConsumerStatefulWidget {
  const InboxScreen({super.key});

  @override
  ConsumerState<InboxScreen> createState() => _InboxScreenState();
}

class _InboxScreenState extends ConsumerState<InboxScreen> {
  List<Conversation> _conversations = [];
  bool _loading = true;
  Timer? _pollTimer;

  @override
  void initState() {
    super.initState();
    _load();
    _pollTimer = Timer.periodic(const Duration(seconds: 15), (_) => _poll());
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final result = await ref.read(apiServiceProvider).getConversations();
      setState(() => _conversations = result.data);
    } catch (e) {
      debugPrint('Load conversations error: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _poll() async {
    if (!mounted) return;
    try {
      final result = await ref.read(apiServiceProvider).getConversations();
      if (mounted) setState(() => _conversations = result.data);
    } catch (_) {}
  }

  String _formatTime(String iso) {
    final dt = DateTime.tryParse(iso);
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
      appBar: AppBar(title: const Text('Tin nhắn')),
      body: _loading
          ? const ListSkeleton()
          : _conversations.isEmpty
              ? const EmptyState(
                  icon: Icons.chat_bubble_outline,
                  title: 'Chưa có cuộc trò chuyện',
                  subtitle: 'Khi bạn liên hệ với người bán, tin nhắn sẽ hiển thị ở đây',
                )
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView.separated(
                    itemCount: _conversations.length,
                    separatorBuilder: (_, __) => const Divider(height: 1, indent: 72),
                    itemBuilder: (_, i) {
                      final conv = _conversations[i];
                      final other = conv.otherUser;
                      final hasUnread = conv.unreadCount > 0;
                      final isOnline = other?.isOnline ?? false;

                      return ListTile(
                        tileColor: hasUnread
                            ? AppColors.primary.withValues(alpha: 0.06)
                            : null,
                        leading: Stack(
                          clipBehavior: Clip.none,
                          children: [
                            CircleAvatar(
                              backgroundImage: other?.avatarUrl != null
                                  ? CachedNetworkImageProvider(other!.avatarUrl!)
                                  : null,
                              child: other?.avatarUrl == null
                                  ? const Icon(Icons.person)
                                  : null,
                            ),
                            Positioned(
                              bottom: 0,
                              right: 0,
                              child: Container(
                                width: 12,
                                height: 12,
                                decoration: BoxDecoration(
                                  color: isOnline ? AppColors.onlineGreen : AppColors.offlineGrey,
                                  shape: BoxShape.circle,
                                  border: Border.all(color: Colors.white, width: 2),
                                ),
                              ),
                            ),
                            if (hasUnread)
                              Positioned(
                                top: -4,
                                right: -4,
                                child: Container(
                                  padding: const EdgeInsets.all(4),
                                  decoration: const BoxDecoration(
                                    color: AppColors.error,
                                    shape: BoxShape.circle,
                                  ),
                                  constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
                                  child: Text(
                                    '${conv.unreadCount}',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                              ),
                          ],
                        ),
                        title: Text(
                          other?.name ?? other?.id ?? 'Người dùng',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontWeight: hasUnread ? FontWeight.bold : FontWeight.normal,
                            color: hasUnread ? AppColors.primary : null,
                          ),
                        ),
                        subtitle: Text(
                          _formatTime(conv.lastMessageAt),
                          style: TextStyle(
                            fontSize: 12,
                            color: hasUnread ? AppColors.primary : AppColors.textSecondary,
                          ),
                        ),
                        trailing: hasUnread
                            ? Container(
                                padding: const EdgeInsets.all(6),
                                decoration: const BoxDecoration(
                                  color: AppColors.primary,
                                  shape: BoxShape.circle,
                                ),
                                child: Text(
                                  '${conv.unreadCount}',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              )
                            : null,
                        onTap: () => context.push('/chat/${conv.id}'),
                      );
                    },
                  ),
                ),
    );
  }
}
