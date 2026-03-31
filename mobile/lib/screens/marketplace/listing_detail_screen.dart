import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/listing.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class ListingDetailScreen extends ConsumerStatefulWidget {
  final String id;
  const ListingDetailScreen({super.key, required this.id});

  @override
  ConsumerState<ListingDetailScreen> createState() => _ListingDetailScreenState();
}

class _ListingDetailScreenState extends ConsumerState<ListingDetailScreen> {
  ListingDetail? _detail;
  bool _loading = true;
  bool _needsLogin = false;
  int _currentImageIndex = 0;
  late PageController _pageController;

  @override
  void initState() {
    super.initState();
    _pageController = PageController();
    _load();
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final detail = await ref.read(apiServiceProvider).getListingDetail(widget.id);
      setState(() => _detail = detail);
    } catch (e) {
      debugPrint('Detail error: $e');
      // Check if 403 (no permission) — guest needs to login
      if (e is DioException && e.response?.statusCode == 403) {
        setState(() => _needsLogin = true);
      }
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _startChat() async {
    final auth = ref.read(authProvider);
    if (auth.status != AuthStatus.authenticated) {
      context.go('/login');
      return;
    }
    try {
      final api = ref.read(apiServiceProvider);
      final conv = await api.createConversation(
        _detail!.seller.id,
        listingId: _detail!.listing.id,
      );
      await _sendListingLinkIfNeeded(api, conv.id);
      if (mounted) context.push('/chat/${conv.id}');
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Lỗi: $e')));
    }
  }

  Future<void> _sendListingLinkIfNeeded(dynamic api, String convId) async {
    try {
      final listingId = _detail!.listing.id;
      final linkText = 'listing://$listingId';
      final today = DateTime.now();
      final todayStr = '${today.year}-${today.month.toString().padLeft(2, '0')}-${today.day.toString().padLeft(2, '0')}';

      final result = await api.getMessages(convId, limit: 50);
      final alreadySent = (result.data as List).any((msg) {
        final msgDate = DateTime.tryParse(msg.createdAt)?.toLocal();
        if (msgDate == null) return false;
        final msgDateStr = '${msgDate.year}-${msgDate.month.toString().padLeft(2, '0')}-${msgDate.day.toString().padLeft(2, '0')}';
        return msgDateStr == todayStr && msg.content.contains(linkText);
      });

      if (!alreadySent) {
        await api.sendMessage(convId, linkText, type: 'listing_link');
      }
    } catch (e) {
      debugPrint('Send listing link error: $e');
    }
  }

  Future<void> _reportListing() async {
    final success = await showDialog<bool>(
      context: context,
      builder: (_) => _ReportDialog(
        targetType: 'listing',
        targetId: widget.id,
        apiService: ref.read(apiServiceProvider),
      ),
    );
    if (success == true && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã gửi báo cáo thành công'), backgroundColor: AppColors.primary),
      );
    }
  }

  final _priceFormat = NumberFormat('#,###', 'vi_VN');

  String _timeAgo(String createdAt) {
    final date = DateTime.tryParse(createdAt)?.toLocal();
    if (date == null) return '';
    final diff = DateTime.now().difference(date);
    if (diff.inDays > 30) return '${diff.inDays ~/ 30} tháng trước';
    if (diff.inDays > 0) return '${diff.inDays} ngày trước';
    if (diff.inHours > 0) return '${diff.inHours} giờ trước';
    if (diff.inMinutes > 0) return '${diff.inMinutes} phút trước';
    return 'Vừa xong';
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(appBar: AppBar(), body: const Center(child: CircularProgressIndicator()));
    }
    if (_detail == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  _needsLogin ? Icons.lock_outline : Icons.search_off,
                  size: 64,
                  color: AppColors.textHint,
                ),
                const SizedBox(height: 16),
                Text(
                  _needsLogin
                      ? 'Đăng nhập để xem chi tiết sản phẩm và nhà cung cấp'
                      : 'Không tìm thấy tin đăng',
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 16, color: AppColors.textSecondary),
                ),
                if (_needsLogin) ...[
                  const SizedBox(height: 20),
                  FilledButton.icon(
                    onPressed: () => context.go('/login'),
                    icon: const Icon(Icons.login),
                    label: const Text('Đăng nhập'),
                  ),
                ],
              ],
            ),
          ),
        ),
      );
    }

    final listing = _detail!.listing;
    final seller = _detail!.seller;
    final authUser = ref.watch(authProvider).user;
    final isOwner = authUser?.id == listing.userId;

    return Scaffold(
      backgroundColor: AppColors.background,
      body: CustomScrollView(
        slivers: [
          // ── Image carousel as SliverAppBar ──
          SliverAppBar(
            expandedHeight: 300,
            pinned: true,
            stretch: true,
            backgroundColor: AppColors.surface,
            foregroundColor: AppColors.textPrimary,
            leading: _CircleBackButton(),
            actions: [
              if (!isOwner)
                Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: _CircleIconButton(
                    icon: Icons.flag_outlined,
                    onTap: _reportListing,
                    color: AppColors.error,
                  ),
                ),
            ],
            flexibleSpace: FlexibleSpaceBar(
              background: listing.images.isNotEmpty
                  ? Stack(
                      fit: StackFit.expand,
                      children: [
                        PageView.builder(
                          controller: _pageController,
                          itemCount: listing.images.length,
                          onPageChanged: (i) => setState(() => _currentImageIndex = i),
                          itemBuilder: (_, i) => GestureDetector(
                            onTap: () => _showImageGallery(context, listing.images, initialIndex: i),
                            child: CachedNetworkImage(
                              imageUrl: listing.images[i],
                              fit: BoxFit.cover,
                              placeholder: (_, __) => Container(
                                color: AppColors.surfaceVariant,
                                child: const Center(child: CircularProgressIndicator()),
                              ),
                              errorWidget: (_, __, ___) => Container(
                                color: AppColors.surfaceVariant,
                                child: const Icon(Icons.broken_image, size: 48, color: AppColors.textHint),
                              ),
                            ),
                          ),
                        ),
                        // Gradient overlay for status bar readability
                        Positioned(
                          top: 0,
                          left: 0,
                          right: 0,
                          height: 100,
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              gradient: LinearGradient(
                                begin: Alignment.topCenter,
                                end: Alignment.bottomCenter,
                                colors: [Colors.black38, Colors.transparent],
                              ),
                            ),
                          ),
                        ),
                        // Page indicator
                        if (listing.images.length > 1)
                          Positioned(
                            bottom: 12,
                            left: 0,
                            right: 0,
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: List.generate(
                                listing.images.length,
                                (i) => AnimatedContainer(
                                  duration: const Duration(milliseconds: 250),
                                  margin: const EdgeInsets.symmetric(horizontal: 3),
                                  width: i == _currentImageIndex ? 20 : 6,
                                  height: 6,
                                  decoration: BoxDecoration(
                                    color: i == _currentImageIndex ? Colors.white : Colors.white54,
                                    borderRadius: BorderRadius.circular(3),
                                  ),
                                ),
                              ),
                            ),
                          ),
                        // Photo count badge
                        if (listing.images.length > 1)
                          Positioned(
                            bottom: 12,
                            right: 16,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.black54,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '${_currentImageIndex + 1}/${listing.images.length}',
                                style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w500),
                              ),
                            ),
                          ),
                      ],
                    )
                  : Container(
                      color: AppColors.surfaceVariant,
                      child: const Center(
                        child: Icon(Icons.image_not_supported, size: 64, color: AppColors.textHint),
                      ),
                    ),
            ),
          ),

          SliverToBoxAdapter(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // ── Price card ──
                Container(
                  width: double.infinity,
                  color: AppColors.surface,
                  padding: const EdgeInsets.fromLTRB(20, 16, 20, 16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Price
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(
                            '${_priceFormat.format(listing.pricePerKg)}đ',
                            style: const TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.w700,
                              color: AppColors.priceText,
                              height: 1.1,
                            ),
                          ),
                          const SizedBox(width: 2),
                          const Padding(
                            padding: EdgeInsets.only(bottom: 3),
                            child: Text('/kg', style: TextStyle(fontSize: 15, color: AppColors.textSecondary)),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      // Title
                      Text(
                        listing.title,
                        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                      ),
                      const SizedBox(height: 8),
                      // Time + views
                      Row(
                        children: [
                          Icon(Icons.access_time, size: 14, color: AppColors.textHint),
                          const SizedBox(width: 4),
                          Text(_timeAgo(listing.createdAt), style: const TextStyle(fontSize: 13, color: AppColors.textHint)),
                          const SizedBox(width: 16),
                          Icon(Icons.visibility_outlined, size: 14, color: AppColors.textHint),
                          const SizedBox(width: 4),
                          Text('${listing.viewCount}', style: const TextStyle(fontSize: 13, color: AppColors.textHint)),
                        ],
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 8),

                // ── Detail info section ──
                Container(
                  color: AppColors.surface,
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Thông tin chi tiết',
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                      ),
                      const SizedBox(height: 16),
                      _DetailRow(icon: Icons.scale, label: 'Số lượng', value: '${_priceFormat.format(listing.quantityKg)} kg'),
                      if (listing.harvestSeason != null && listing.harvestSeason!.isNotEmpty)
                        _DetailRow(icon: Icons.calendar_month, label: 'Vụ mùa', value: listing.harvestSeason!),
                      if (listing.certifications != null && listing.certifications!.isNotEmpty)
                        _DetailRow(icon: Icons.verified, label: 'Chứng nhận', value: listing.certifications!, valueColor: AppColors.primary),
                      if (listing.province != null || listing.ward != null)
                        _DetailRow(
                          icon: Icons.location_on_outlined,
                          label: 'Khu vực',
                          value: [listing.ward, listing.province].where((s) => s != null && s.isNotEmpty).join(', '),
                        ),
                      if (seller.phone.isNotEmpty)
                        GestureDetector(
                          onTap: () {
                            Clipboard.setData(ClipboardData(text: seller.phone));
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Đã sao chép số điện thoại'), duration: Duration(seconds: 2)),
                            );
                          },
                          child: _DetailRow(
                            icon: Icons.phone,
                            label: 'Điện thoại',
                            value: seller.phone,
                            trailing: const Icon(Icons.copy, size: 16, color: AppColors.textHint),
                          ),
                        ),
                    ],
                  ),
                ),

                // ── Description section ──
                if (listing.description != null && listing.description!.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Container(
                    width: double.infinity,
                    color: AppColors.surface,
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Mô tả',
                          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          listing.description!,
                          style: const TextStyle(fontSize: 14, color: AppColors.textSecondary, height: 1.6),
                        ),
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 8),

                // ── Seller section ──
                Container(
                  color: AppColors.surface,
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Người đăng',
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                      ),
                      const SizedBox(height: 16),
                      InkWell(
                        onTap: () => context.push('/seller/${seller.id}'),
                        borderRadius: BorderRadius.circular(12),
                        child: Row(
                          children: [
                            // Avatar
                            CircleAvatar(
                              radius: 26,
                              backgroundColor: AppColors.primary.withValues(alpha: 0.12),
                              backgroundImage: seller.avatarUrl != null
                                  ? CachedNetworkImageProvider(seller.avatarUrl!)
                                  : null,
                              child: seller.avatarUrl == null
                                  ? Text(
                                      (seller.name ?? 'U').substring(0, 1).toUpperCase(),
                                      style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppColors.primary),
                                    )
                                  : null,
                            ),
                            const SizedBox(width: 12),
                            // Name & location
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      Flexible(
                                        child: Text(
                                          seller.name ?? 'Thành viên',
                                          style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ),
                                      const SizedBox(width: 6),
                                      Container(
                                        width: 8,
                                        height: 8,
                                        decoration: BoxDecoration(
                                          color: seller.isOnline ? AppColors.onlineGreen : AppColors.offlineGrey,
                                          shape: BoxShape.circle,
                                        ),
                                      ),
                                    ],
                                  ),
                                  if (seller.province != null) ...[
                                    const SizedBox(height: 2),
                                    Text(
                                      [seller.ward, seller.province].where((s) => s != null && s.isNotEmpty).join(', '),
                                      style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ],
                                ],
                              ),
                            ),
                            Icon(Icons.chevron_right, color: AppColors.textHint, size: 22),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),

                // Bottom spacing for the floating button
                const SizedBox(height: 100),
              ],
            ),
          ),
        ],
      ),

      // ── Floating bottom bar ──
      bottomNavigationBar: !isOwner
          ? Container(
              decoration: BoxDecoration(
                color: AppColors.surface,
                boxShadow: [
                  BoxShadow(color: Colors.black.withValues(alpha: 0.08), blurRadius: 8, offset: const Offset(0, -2)),
                ],
              ),
              padding: EdgeInsets.fromLTRB(20, 12, 20, 12 + MediaQuery.of(context).padding.bottom),
              child: Row(
                children: [
                  // Price summary
                  Expanded(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${_priceFormat.format(listing.pricePerKg)}đ/kg',
                          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: AppColors.priceText),
                        ),
                        Text(
                          '${_priceFormat.format(listing.quantityKg)} kg',
                          style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  // Chat button
                  FilledButton.icon(
                    onPressed: _startChat,
                    icon: const Icon(Icons.chat_bubble_outline, size: 18),
                    label: const Text('Chat với người bán', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                    style: FilledButton.styleFrom(
                      backgroundColor: AppColors.primary,
                      foregroundColor: Colors.white,
                      minimumSize: const Size(0, 48),
                      padding: const EdgeInsets.symmetric(horizontal: 24),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                  ),
                ],
              ),
            )
          : null,
    );
  }

  void _showImageGallery(BuildContext context, List<String> images, {int initialIndex = 0}) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => _ImageGalleryScreen(images: images, initialIndex: initialIndex),
      ),
    );
  }
}

// ── Circular back button for SliverAppBar ──
class _CircleBackButton extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 8),
      child: Center(
        child: SizedBox(
          width: 40,
          height: 40,
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: Colors.white,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(color: Colors.black26, blurRadius: 6, offset: const Offset(0, 2)),
              ],
            ),
            child: Material(
              color: Colors.transparent,
              shape: const CircleBorder(),
              clipBehavior: Clip.antiAlias,
              child: InkWell(
                onTap: () => Navigator.maybePop(context),
                child: const Icon(Icons.arrow_back, color: AppColors.textPrimary, size: 22),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

// ── Circular icon button for SliverAppBar ──
class _CircleIconButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onTap;
  final Color? color;
  const _CircleIconButton({required this.icon, required this.onTap, this.color});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        width: 40,
        height: 40,
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: Colors.white,
            shape: BoxShape.circle,
            boxShadow: [
              BoxShadow(color: Colors.black26, blurRadius: 6, offset: const Offset(0, 2)),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            shape: const CircleBorder(),
            clipBehavior: Clip.antiAlias,
            child: InkWell(
              onTap: onTap,
              child: Icon(icon, color: color ?? AppColors.textPrimary, size: 22),
            ),
          ),
        ),
      ),
    );
  }
}

// ── Detail row with icon ──
class _DetailRow extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final Color? valueColor;
  final Widget? trailing;
  const _DetailRow({required this.icon, required this.label, required this.value, this.valueColor, this.trailing});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 18, color: AppColors.textHint),
          const SizedBox(width: 10),
          SizedBox(
            width: 80,
            child: Text(label, style: const TextStyle(fontSize: 14, color: AppColors.textSecondary)),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              value,
              style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500, color: valueColor ?? AppColors.textPrimary),
            ),
          ),
          if (trailing != null) ...[
            const SizedBox(width: 8),
            trailing!,
          ],
        ],
      ),
    );
  }
}

// ── Full-screen image gallery ──
class _ImageGalleryScreen extends StatefulWidget {
  final List<String> images;
  final int initialIndex;
  const _ImageGalleryScreen({required this.images, this.initialIndex = 0});

  @override
  State<_ImageGalleryScreen> createState() => _ImageGalleryScreenState();
}

class _ImageGalleryScreenState extends State<_ImageGalleryScreen> {
  late int _current;
  late PageController _controller;

  @override
  void initState() {
    super.initState();
    _current = widget.initialIndex;
    _controller = PageController(initialPage: widget.initialIndex);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final topPadding = MediaQuery.of(context).padding.top;
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          // Image viewer
          PageView.builder(
            controller: _controller,
            itemCount: widget.images.length,
            onPageChanged: (i) => setState(() => _current = i),
            itemBuilder: (_, i) => InteractiveViewer(
              child: Center(
                child: CachedNetworkImage(
                  imageUrl: widget.images[i],
                  fit: BoxFit.contain,
                  placeholder: (_, __) => const Center(child: CircularProgressIndicator(color: Colors.white)),
                  errorWidget: (_, __, ___) => const Icon(Icons.broken_image, color: Colors.white54, size: 60),
                ),
              ),
            ),
          ),
          // Top bar: back button + counter
          Positioned(
            top: topPadding + 8,
            left: 12,
            right: 12,
            child: Row(
              children: [
                // Back button
                SizedBox(
                  width: 44,
                  height: 44,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: Colors.white,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(color: Colors.black38, blurRadius: 8, offset: const Offset(0, 2)),
                      ],
                    ),
                    child: Material(
                      color: Colors.transparent,
                      shape: const CircleBorder(),
                      clipBehavior: Clip.antiAlias,
                      child: InkWell(
                        onTap: () => Navigator.maybePop(context),
                        child: const Icon(Icons.close, color: AppColors.textPrimary, size: 24),
                      ),
                    ),
                  ),
                ),
                const Spacer(),
                // Counter
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: Colors.black54,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text(
                    '${_current + 1} / ${widget.images.length}',
                    style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w500),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ── Report dialog ──
class _ReportDialog extends StatefulWidget {
  final String targetType;
  final String targetId;
  final dynamic apiService;
  const _ReportDialog({required this.targetType, required this.targetId, required this.apiService});

  @override
  State<_ReportDialog> createState() => _ReportDialogState();
}

class _ReportDialogState extends State<_ReportDialog> {
  String? _selectedReason;
  final _descCtrl = TextEditingController();
  bool _loading = false;
  String? _error;

  List<String> get _reasons => widget.targetType == 'listing'
      ? ['Thông tin sai lệch', 'Hàng giả/kém chất lượng', 'Lừa đảo', 'Spam', 'Khác']
      : ['Lừa đảo', 'Thông tin sai lệch', 'Spam', 'Khác'];

  @override
  void dispose() {
    _descCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() { _loading = true; _error = null; });
    try {
      await widget.apiService.createReport(
        widget.targetType,
        widget.targetId,
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
      title: Text(widget.targetType == 'listing' ? 'Báo cáo tin đăng' : 'Báo cáo người dùng'),
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
