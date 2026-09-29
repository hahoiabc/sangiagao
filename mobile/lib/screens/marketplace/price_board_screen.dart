import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/price_board.dart';
import '../../providers/providers.dart';
import '../../widgets/shimmer_loading.dart';
import '../../theme/app_theme.dart';
import '../../widgets/marquee_text.dart';

class PriceBoardScreen extends ConsumerStatefulWidget {
  const PriceBoardScreen({super.key});

  @override
  ConsumerState<PriceBoardScreen> createState() => _PriceBoardScreenState();
}

class _PriceBoardScreenState extends ConsumerState<PriceBoardScreen> {
  PriceBoardResponse? _data;
  bool _loading = true;
  String? _error;
  String _slogan = 'Kết nối ngành gạo';
  Color _sloganColor = const Color(0xFF4F46E5);

  final _priceFormat = NumberFormat('#,###', 'vi_VN');

  static const _categoryIcons = <String, IconData>{
    'gao_deo_thom': Icons.rice_bowl,
    'gao_kho': Icons.grass,
    'tam_deo_thom': Icons.grain,
    'tam_kho': Icons.scatter_plot,
    'nep': Icons.spa,
  };

  @override
  void initState() {
    super.initState();
    _load();
    _loadSlogan();
  }

  Future<void> _loadSlogan() async {
    try {
      final api = ref.read(apiServiceProvider);
      final results = await Future.wait([api.getSlogan(), api.getSloganColor()]);
      if (!mounted) return;
      setState(() {
        _slogan = results[0];
        final hex = results[1].replaceFirst('#', '');
        if (hex.length == 6) {
          _sloganColor = Color(int.parse('FF$hex', radix: 16));
        }
      });
    } catch (_) {}
  }

  Future<void> _load() async {
    if (!mounted) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await ref.read(apiServiceProvider).getPriceBoard();
      if (!mounted) return;
      setState(() => _data = result);
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Không thể tải bảng giá');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _viewListings(String categoryKey, String productKey) {
    context.push(Uri(
      path: '/marketplace/search',
      queryParameters: {'category': categoryKey, 'type': productKey, 'sort': 'price_asc'},
    ).toString());
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    final inboxUnread = ref.watch(inboxUnreadProvider);
    final isAuth = ref.watch(authProvider).status == AuthStatus.authenticated;

    return Scaffold(
      appBar: AppBar(
        title: MarqueeText(
          text: _slogan,
          style: TextStyle(
            color: _sloganColor,
            fontSize: 18,
            fontWeight: FontWeight.w600,
          ),
        ),
        actions: [
          if (isAuth)
            IconButton(
              onPressed: () => context.push('/system-inbox'),
              icon: Badge(
                isLabelVisible: inboxUnread > 0,
                label: Text(inboxUnread > 99 ? '99+' : '$inboxUnread'),
                child: const Icon(Icons.mail_outline),
              ),
              tooltip: 'Hộp thư',
            ),
        ],
      ),
      body: _loading
          ? const PriceBoardSkeleton()
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
              : RefreshIndicator(
                  onRefresh: _load,
                  child: Column(
                    children: [
                      // Inbox banner
                      if (isAuth && inboxUnread > 0)
                        GestureDetector(
                          onTap: () => context.push('/system-inbox'),
                          child: Container(
                            width: double.infinity,
                            margin: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                            decoration: BoxDecoration(
                              color: AppColors.primary.withValues(alpha: 0.08),
                              borderRadius: BorderRadius.circular(10),
                              border: Border.all(color: AppColors.primary.withValues(alpha: 0.2)),
                            ),
                            child: Row(
                              children: [
                                Icon(Icons.mail, color: AppColors.primary, size: 20),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: Text(
                                    'Bạn có $inboxUnread thông báo mới',
                                    style: TextStyle(
                                      color: AppColors.primary,
                                      fontWeight: FontWeight.w600,
                                      fontSize: 13,
                                    ),
                                  ),
                                ),
                                Text(
                                  'Xem ngay',
                                  style: TextStyle(
                                    color: AppColors.primary,
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                ),
                                const SizedBox(width: 4),
                                Icon(Icons.chevron_right, color: AppColors.primary, size: 18),
                              ],
                            ),
                          ),
                        ),
                      // Đơn vị giá — ghi 1 LẦN ở đầu (đã bỏ "đ/kg" mỗi dòng để tên+giá hiện đủ)
                      Padding(
                        padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                        child: Row(
                          children: [
                            Icon(Icons.info_outline, size: 14, color: AppColors.textHint),
                            const SizedBox(width: 4),
                            Text('Đơn giá: đồng/kg (đ/kg)',
                                style: TextStyle(fontSize: 12, color: AppColors.textSecondary)),
                          ],
                        ),
                      ),
                      // Price board list
                      Expanded(
                        child: ListView.separated(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
                          itemCount: _data!.categories.length,
                          separatorBuilder: (_, __) => const SizedBox(height: 24),
                          itemBuilder: (context, index) {
                            final cat = _data!.categories[index];
                            return _buildCategorySection(cat, theme);
                          },
                        ),
                      ),
                    ],
                  ),
                ),
    );
  }

  Widget _buildCategorySection(PriceBoardCategory cat, ThemeData theme) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        children: [
          // Category header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  theme.colorScheme.onPrimaryContainer,
                  theme.colorScheme.primary,
                ],
                begin: Alignment.centerLeft,
                end: Alignment.centerRight,
              ),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    _categoryIcons[cat.categoryKey] ?? Icons.category,
                    size: 18,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    cat.categoryLabel,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                      color: Colors.white,
                      letterSpacing: 0.3,
                    ),
                  ),
                ),
                Text(
                  '${cat.products.length} SP',
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.white.withValues(alpha: 0.7),
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 4),
          // Product rows
          ...cat.products.asMap().entries.map((entry) {
                final i = entry.key;
                final product = entry.value;
                final hasSponsor = product.sponsorLogo != null;
                return InkWell(
                  onTap: () => _viewListings(cat.categoryKey, product.productKey),
                  child: Container(
                    margin: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: i.isEven ? null : theme.colorScheme.surfaceContainerLow,
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: theme.colorScheme.outline, width: 1.2),
                    ),
                    height: 48, // cố định (bằng minHeight cũ) → chặn chiều cao cho stretch hợp lệ
                    clipBehavior: Clip.antiAlias, // để ảnh bo theo góc dòng
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.stretch, // ảnh cao BẰNG dòng
                      children: [
                        // Ảnh tin RẺ NHẤT — cao bằng dòng, sát mép trái (không tăng chiều cao dòng)
                        _PriceThumb(imageUrl: product.imageUrl),
                        Expanded(
                          child: Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                            child: Row(
                              children: [
                                // Sponsor logo (before product name)
                                if (hasSponsor) ...[
                                  CachedNetworkImage(
                                    imageUrl: product.sponsorLogo!,
                                    width: 22,
                                    height: 22,
                                    fit: BoxFit.contain,
                                    errorWidget: (_, __, ___) => const SizedBox.shrink(),
                                  ),
                                  const SizedBox(width: 8),
                                ],
                                // Product name — tự thu vừa khít, KHÔNG cắt "..."
                                Expanded(
                                  child: FittedBox(
                                    fit: BoxFit.scaleDown,
                                    alignment: Alignment.centerLeft,
                                    child: Text(
                                      product.productLabel,
                                      maxLines: 1,
                                      softWrap: false,
                                      style: const TextStyle(fontSize: 15, height: 1.3),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 10),
                                // Price — SỐ THUẦN (đơn vị đ/kg ghi ở đầu bảng); FittedBox chống tràn
                                FittedBox(
                                  fit: BoxFit.scaleDown,
                                  child: Text(
                                    product.minPrice != null
                                        ? _priceFormat.format(product.minPrice)
                                        : 'Chưa có giá',
                                    maxLines: 1,
                                    style: TextStyle(
                                      fontSize: 15,
                                      fontWeight: product.minPrice != null ? FontWeight.w600 : FontWeight.normal,
                                      color: product.minPrice != null ? AppColors.priceText : AppColors.textHint,
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 4),
                                // Arrow icon
                                const Icon(Icons.chevron_right, size: 22, color: AppColors.textHint),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              }),
          const SizedBox(height: 4),
            ],
          ),
        );
  }
}

/// Ảnh loại gạo ở đầu dòng bảng giá — cao BẰNG dòng, sát mép trái, phủ (cover).
/// Không tăng chiều cao dòng (chỉ chiếm phần trái). Thiếu ảnh → placeholder hạt gạo.
class _PriceThumb extends StatelessWidget {
  final String? imageUrl;
  const _PriceThumb({this.imageUrl});

  Widget _placeholder() => Container(
        width: 46,
        alignment: Alignment.center,
        color: AppColors.primary.withValues(alpha: 0.06),
        child: Icon(Icons.grain, size: 22, color: AppColors.textHint.withValues(alpha: 0.5)),
      );

  @override
  Widget build(BuildContext context) {
    if (imageUrl == null || imageUrl!.isEmpty) return _placeholder();
    return SizedBox(
      width: 46,
      child: CachedNetworkImage(
        imageUrl: imageUrl!,
        fit: BoxFit.cover,
        placeholder: (_, __) => Container(color: AppColors.primary.withValues(alpha: 0.05)),
        errorWidget: (_, __, ___) => _placeholder(),
      ),
    );
  }
}
