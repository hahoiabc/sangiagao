import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/price_board.dart';
import '../../providers/providers.dart';
import '../../widgets/shimmer_loading.dart';
import '../../theme/app_theme.dart';

class PriceBoardScreen extends ConsumerStatefulWidget {
  const PriceBoardScreen({super.key});

  @override
  ConsumerState<PriceBoardScreen> createState() => _PriceBoardScreenState();
}

class _PriceBoardScreenState extends ConsumerState<PriceBoardScreen> {
  PriceBoardResponse? _data;
  bool _loading = true;
  String? _error;

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

    return Scaffold(
      appBar: AppBar(
        title: const Text('SanGiaGao.Com'),
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
                  child: ListView.separated(
                    padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
                    itemCount: _data!.categories.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 24),
                    itemBuilder: (context, index) {
                      final cat = _data!.categories[index];
                      return _buildCategorySection(cat, theme);
                    },
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
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: AppColors.primary.withValues(alpha: 0.08),
            ),
            child: Row(
              children: [
                Icon(
                  _categoryIcons[cat.categoryKey] ?? Icons.category,
                  size: 20,
                  color: AppColors.primary,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    cat.categoryLabel,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      fontSize: 17,
                    ),
                  ),
                ),
              ],
            ),
          ),
          // Column header row
          Container(
            color: theme.colorScheme.surfaceContainerHighest,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                const Expanded(
                  child: Text('Sản phẩm', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                ),
                const Text('Giá thấp nhất', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                const SizedBox(width: 32),
              ],
            ),
          ),
          // Product rows
          ...cat.products.asMap().entries.map((entry) {
                final i = entry.key;
                final product = entry.value;
                final hasSponsor = product.sponsorLogo != null;
                return InkWell(
                  onTap: () => _viewListings(cat.categoryKey, product.productKey),
                  child: Container(
                    decoration: BoxDecoration(
                      color: i.isEven ? null : theme.colorScheme.surfaceContainerLow,
                      border: Border(bottom: BorderSide(color: theme.dividerColor.withValues(alpha: 0.15))),
                    ),
                    constraints: const BoxConstraints(minHeight: 48),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    child: Row(
                      children: [
                        // Sponsor logo (before product name)
                        if (hasSponsor) ...[
                          CachedNetworkImage(
                            imageUrl: product.sponsorLogo!,
                            width: 28,
                            height: 28,
                            fit: BoxFit.contain,
                            errorWidget: (_, __, ___) => const SizedBox.shrink(),
                          ),
                          const SizedBox(width: 10),
                        ],
                        // Product name
                        Expanded(
                          child: Text(
                            product.productLabel,
                            style: const TextStyle(fontSize: 15, height: 1.3),
                          ),
                        ),
                        const SizedBox(width: 12),
                        // Price
                        Text(
                          product.minPrice != null
                              ? '${_priceFormat.format(product.minPrice)}đ/kg'
                              : 'Chưa có giá',
                          style: TextStyle(
                            fontSize: 15,
                            fontWeight: product.minPrice != null ? FontWeight.w600 : FontWeight.normal,
                            color: product.minPrice != null ? AppColors.priceText : AppColors.textHint,
                          ),
                        ),
                        const SizedBox(width: 4),
                        // Arrow icon
                        const Icon(Icons.chevron_right, size: 22, color: AppColors.textHint),
                      ],
                    ),
                  ),
                );
              }),
            ],
          ),
        );
  }
}
