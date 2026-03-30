import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../models/inbox.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class InboxDetailScreen extends ConsumerStatefulWidget {
  final String id;
  const InboxDetailScreen({super.key, required this.id});

  @override
  ConsumerState<InboxDetailScreen> createState() => _InboxDetailScreenState();
}

class _InboxDetailScreenState extends ConsumerState<InboxDetailScreen> {
  InboxMessage? _message;
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final msg = await ref.read(apiServiceProvider).getInboxDetail(widget.id);
      if (mounted) setState(() => _message = msg);
      // Auto mark read — server already marks on GET /inbox/:id
      ref.read(inboxUnreadProvider.notifier).refresh();
    } catch (e) {
      if (mounted) setState(() => _error = 'Không thể tải thông báo');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  String _formatDate(String iso) {
    final dt = DateTime.tryParse(iso)?.toLocal();
    if (dt == null) return '';
    return '${dt.day}/${dt.month}/${dt.year} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chi tiết thông báo')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!, style: TextStyle(color: AppColors.error)))
              : _message == null
                  ? const Center(child: Text('Không tìm thấy'))
                  : SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (_message!.isPinned)
                            Container(
                              margin: const EdgeInsets.only(bottom: 8),
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: AppColors.warning.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(Icons.push_pin, size: 14, color: AppColors.warning),
                                  const SizedBox(width: 4),
                                  Text('Ghim', style: TextStyle(fontSize: 12, color: AppColors.warning, fontWeight: FontWeight.w600)),
                                ],
                              ),
                            ),
                          Text(
                            _message!.title,
                            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _formatDate(_message!.createdAt),
                            style: TextStyle(fontSize: 13, color: AppColors.textHint),
                          ),
                          const SizedBox(height: 16),
                          const Divider(),
                          const SizedBox(height: 16),
                          if (_message!.imageUrl != null && _message!.imageUrl!.isNotEmpty) ...[
                            ClipRRect(
                              borderRadius: BorderRadius.circular(8),
                              child: CachedNetworkImage(
                                imageUrl: _message!.imageUrl!,
                                width: double.infinity,
                                fit: BoxFit.cover,
                                placeholder: (_, __) => const SizedBox(
                                  height: 200,
                                  child: Center(child: CircularProgressIndicator()),
                                ),
                                errorWidget: (_, __, ___) => const SizedBox.shrink(),
                              ),
                            ),
                            const SizedBox(height: 16),
                          ],
                          Text(
                            _message!.body,
                            style: const TextStyle(fontSize: 15, height: 1.6),
                          ),
                        ],
                      ),
                    ),
    );
  }
}
