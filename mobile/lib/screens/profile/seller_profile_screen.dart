import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_rating_bar/flutter_rating_bar.dart';
import 'package:go_router/go_router.dart';
import '../../models/user.dart';
import '../../models/rating.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class SellerProfileScreen extends ConsumerStatefulWidget {
  final String sellerId;
  const SellerProfileScreen({super.key, required this.sellerId});

  @override
  ConsumerState<SellerProfileScreen> createState() => _SellerProfileScreenState();
}

class _SellerProfileScreenState extends ConsumerState<SellerProfileScreen> {
  PublicProfile? _seller;
  RatingSummary? _ratingSummary;
  List<Rating> _ratings = [];
  bool _loading = true;

  // Rating form
  int _newStars = 5;
  final _commentCtrl = TextEditingController();
  bool _submittingRating = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _commentCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final api = ref.read(apiServiceProvider);
      final results = await Future.wait([
        api.getPublicProfile(widget.sellerId),
        api.getRatingSummary(widget.sellerId),
        api.getSellerRatings(widget.sellerId),
      ]);
      setState(() {
        _seller = results[0] as PublicProfile;
        _ratingSummary = results[1] as RatingSummary;
        _ratings = (results[2] as dynamic).data as List<Rating>;
      });
    } catch (e) {
      debugPrint('Load seller profile error: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _submitRating() async {
    final currentUserId = ref.read(authProvider).user?.id;
    if (currentUserId == widget.sellerId) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Bạn không thể đánh giá chính mình')),
      );
      return;
    }
    setState(() => _submittingRating = true);
    try {
      await ref.read(apiServiceProvider).createRating(
        widget.sellerId,
        _newStars,
        _commentCtrl.text.trim(),
      );
      _commentCtrl.clear();
      _newStars = 5;
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đánh giá thành công!')),
        );
      }
      _load();
    } catch (e) {
      if (mounted) {
        String msg = e.toString();
        if (msg.contains('cannot rate yourself')) {
          msg = 'Bạn không thể đánh giá chính mình';
        } else if (msg.contains('already rated')) {
          msg = 'Bạn đã đánh giá người bán này rồi';
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg)),
        );
      }
    } finally {
      if (mounted) setState(() => _submittingRating = false);
    }
  }

  Future<void> _report() async {
    final success = await showDialog<bool>(
      context: context,
      builder: (_) => _ReportUserDialog(
        sellerId: widget.sellerId,
        apiService: ref.read(apiServiceProvider),
      ),
    );
    if (success == true && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã gửi báo cáo thành công'), backgroundColor: AppColors.primary),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (_seller == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: Text('Không tìm thấy người bán')),
      );
    }

    final seller = _seller!;

    return Scaffold(
      appBar: AppBar(
        title: Text(seller.name ?? 'Thành viên'),
        actions: [
          IconButton(icon: const Icon(Icons.flag_outlined, color: AppColors.error), onPressed: _report, tooltip: 'Báo cáo'),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Seller info
            Row(
              children: [
                CircleAvatar(
                  radius: 36,
                  backgroundImage: seller.avatarUrl != null ? CachedNetworkImageProvider(seller.avatarUrl!) : null,
                  child: seller.avatarUrl == null ? const Icon(Icons.person, size: 36) : null,
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(seller.name ?? 'Thành viên', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                      if (seller.phone.isNotEmpty)
                        Row(
                          children: [
                            Icon(Icons.phone, size: 14, color: AppColors.textHint),
                            const SizedBox(width: 4),
                            Text(seller.phone, style: TextStyle(fontSize: 14, color: AppColors.textSecondary)),
                          ],
                        ),
                      if (seller.orgName != null) Text(seller.orgName!, style: TextStyle(color: AppColors.textSecondary)),
                      if (seller.location != null)
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Icon(Icons.location_on, size: 14, color: AppColors.textHint),
                            const SizedBox(width: 4),
                            Expanded(
                              child: Text(seller.location!, style: TextStyle(fontSize: 13, color: AppColors.textHint)),
                            ),
                          ],
                        ),
                    ],
                  ),
                ),
              ],
            ),
            if (seller.description != null) ...[
              const SizedBox(height: 12),
              Text(seller.description!, style: TextStyle(color: AppColors.textSecondary)),
            ],
            const SizedBox(height: 16),

            // Chat button (ẩn khi xem chính mình)
            if (ref.watch(authProvider).user?.id != widget.sellerId)
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: () async {
                    try {
                      final conv = await ref.read(apiServiceProvider).createConversation(widget.sellerId);
                      if (mounted) context.push('/chat/${conv.id}');
                    } catch (e) {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Lỗi: $e')));
                      }
                    }
                  },
                  icon: const Icon(Icons.chat_bubble_outline),
                  label: const Text('Chat với người bán'),
                ),
              ),
            const SizedBox(height: 24),

            // Rating summary
            if (_ratingSummary != null) ...[
              Row(
                children: [
                  const Text('Đánh giá', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(width: 12),
                  RatingBarIndicator(
                    rating: _ratingSummary!.average,
                    itemBuilder: (_, __) => const Icon(Icons.star, color: AppColors.secondary),
                    itemCount: 5,
                    itemSize: 20,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    '${_ratingSummary!.average.toStringAsFixed(1)} (${_ratingSummary!.count})',
                    style: TextStyle(color: AppColors.textSecondary),
                  ),
                ],
              ),
              const SizedBox(height: 12),
            ],

            // Rating form (ẩn khi xem chính mình)
            if (ref.watch(authProvider).user?.id != widget.sellerId)
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Đánh giá người bán', style: TextStyle(fontWeight: FontWeight.bold)),
                      const SizedBox(height: 8),
                      RatingBar.builder(
                        initialRating: 5,
                        minRating: 1,
                        itemBuilder: (_, __) => const Icon(Icons.star, color: AppColors.secondary),
                        itemCount: 5,
                        itemSize: 32,
                        onRatingUpdate: (v) => _newStars = v.toInt(),
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: _commentCtrl,
                        decoration: const InputDecoration(
                          hintText: 'Nhận xét (tùy chọn)',
                          border: OutlineInputBorder(),
                          isDense: true,
                        ),
                        maxLines: 2,
                      ),
                      const SizedBox(height: 8),
                      Align(
                        alignment: Alignment.centerRight,
                        child: FilledButton(
                          onPressed: _submittingRating ? null : _submitRating,
                          child: _submittingRating
                              ? const SizedBox(height: 16, width: 16, child: CircularProgressIndicator(strokeWidth: 2))
                              : const Text('Gửi đánh giá'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            const SizedBox(height: 16),

            // Reviews list
            if (_ratings.isNotEmpty) ...[
              const Text('Nhận xét', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              ..._ratings.map((r) => Card(
                    margin: const EdgeInsets.only(bottom: 8),
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              RatingBarIndicator(
                                rating: r.stars.toDouble(),
                                itemBuilder: (_, __) => const Icon(Icons.star, color: AppColors.secondary),
                                itemCount: 5,
                                itemSize: 16,
                              ),
                              const Spacer(),
                              Text(
                                _formatDate(r.createdAt),
                                style: TextStyle(fontSize: 12, color: AppColors.textHint),
                              ),
                            ],
                          ),
                          if (r.comment != null && r.comment!.isNotEmpty) ...[
                            const SizedBox(height: 6),
                            Text(r.comment!),
                          ],
                        ],
                      ),
                    ),
                  )),
            ],
          ],
        ),
      ),
    );
  }

  String _formatDate(String iso) {
    final dt = DateTime.tryParse(iso);
    if (dt == null) return '';
    return '${dt.day}/${dt.month}/${dt.year}';
  }
}

class _ReportUserDialog extends StatefulWidget {
  final String sellerId;
  final dynamic apiService;
  const _ReportUserDialog({required this.sellerId, required this.apiService});

  @override
  State<_ReportUserDialog> createState() => _ReportUserDialogState();
}

class _ReportUserDialogState extends State<_ReportUserDialog> {
  String? _selectedReason;
  final _descCtrl = TextEditingController();
  bool _loading = false;
  String? _error;
  static const _reasons = ['Lừa đảo', 'Thông tin sai lệch', 'Spam', 'Khác'];

  @override
  void dispose() {
    _descCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() { _loading = true; _error = null; });
    try {
      await widget.apiService.createReport(
        'user',
        widget.sellerId,
        _selectedReason!,
        description: _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
      );
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) setState(() { _loading = false; _error = e.toString(); });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Báo cáo người bán'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Chọn lý do:', style: TextStyle(fontWeight: FontWeight.w500)),
          const SizedBox(height: 8),
          ..._reasons.map((r) => RadioListTile<String>(
                title: Text(r, style: const TextStyle(fontSize: 14)),
                value: r,
                groupValue: _selectedReason,
                onChanged: _loading ? null : (v) => setState(() => _selectedReason = v),
                dense: true,
                contentPadding: EdgeInsets.zero,
              )),
          const SizedBox(height: 8),
          TextField(
            controller: _descCtrl,
            enabled: !_loading,
            decoration: const InputDecoration(
              hintText: 'Mô tả chi tiết (tùy chọn)',
              border: OutlineInputBorder(),
              isDense: true,
            ),
            maxLines: 3,
            maxLength: 500,
          ),
          if (_error != null) ...[
            const SizedBox(height: 8),
            Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 13)),
          ],
        ],
      ),
      actions: [
        TextButton(onPressed: _loading ? null : () => Navigator.pop(context), child: const Text('Huỷ')),
        FilledButton(
          onPressed: _selectedReason == null || _loading ? null : _submit,
          child: _loading
              ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.surface))
              : const Text('Gửi báo cáo'),
        ),
      ],
    );
  }
}
