import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/listing.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/shimmer_loading.dart';

class MyListingsScreen extends ConsumerStatefulWidget {
  const MyListingsScreen({super.key});

  @override
  ConsumerState<MyListingsScreen> createState() => _MyListingsScreenState();
}

class _MyListingsScreenState extends ConsumerState<MyListingsScreen> {
  List<Listing> _listings = [];
  bool _loading = true;
  int _page = 1;
  int _total = 0;
  static const _pageSize = 20;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load({int page = 1}) async {
    setState(() => _loading = true);
    try {
      final result = await ref.read(apiServiceProvider).getMyListings(page: page, limit: _pageSize);
      if (mounted) setState(() { _listings = result.data; _total = result.total; _page = page; });
    } catch (e) {
      debugPrint('My listings error: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _delete(String id) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Xóa tin đăng'),
        content: const Text('Bạn có chắc muốn xóa tin đăng này?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Hủy')),
          TextButton(onPressed: () => Navigator.pop(dialogContext, true), child: const Text('Xóa', style: TextStyle(color: AppColors.error))),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      await ref.read(apiServiceProvider).deleteListing(id);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đã xóa tin đăng')),
        );
        _load();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi xóa: $e')),
        );
      }
    }
  }

  final _priceFormat = NumberFormat('#,###', 'vi_VN');
  final _dateFormat = DateFormat('dd/MM/yyyy HH:mm');

  String _statusLabel(String status) {
    switch (status) {
      case 'active':
        return 'Đang hiển thị';
      case 'hidden':
        return 'Đã ẩn';
      case 'deleted':
        return 'Đã xóa';
      default:
        return status;
    }
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'active':
        return AppColors.activeGreen;
      case 'hidden':
        return AppColors.hiddenOrange;
      case 'deleted':
        return AppColors.deletedRed;
      default:
        return AppColors.textHint;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Tin đăng của tôi')),
      body: _loading
          ? const ListSkeleton()
          : _listings.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.inventory_2_outlined, size: 64, color: AppColors.textHint),
                      const SizedBox(height: 12),
                      Text('Chưa có tin đăng nào', style: TextStyle(fontSize: 16, color: AppColors.textSecondary)),
                      const SizedBox(height: 8),
                      FilledButton.icon(
                        onPressed: () async {
                          await context.push('/create-listing');
                          _load();
                        },
                        icon: const Icon(Icons.add),
                        label: const Text('Đăng tin ngay'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: () => _load(page: _page),
                  child: ListView.separated(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                    itemCount: _listings.length + (_total > _pageSize ? 1 : 0),
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (_, i) {
                      if (i < _listings.length) return _buildListingCard(_listings[i]);
                      // Pagination row
                      final totalPages = (_total + _pageSize - 1) ~/ _pageSize;
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            IconButton(
                              onPressed: _page > 1 ? () => _load(page: _page - 1) : null,
                              icon: const Icon(Icons.chevron_left),
                            ),
                            Text('Trang $_page / $totalPages', style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                            IconButton(
                              onPressed: _page < totalPages ? () => _load(page: _page + 1) : null,
                              icon: const Icon(Icons.chevron_right),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
      floatingActionButton: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                FilledButton.tonal(
                  onPressed: () async {
                    await context.push('/quick-batch');
                    _load();
                  },
                  style: FilledButton.styleFrom(
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                    minimumSize: Size.zero,
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    elevation: 4,
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.flash_on, size: 16),
                      SizedBox(width: 4),
                      Text('Đăng nhanh', style: TextStyle(fontSize: 14)),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                FilledButton(
                  onPressed: () async {
                    await context.push('/create-listing');
                    _load();
                  },
                  style: FilledButton.styleFrom(
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    minimumSize: Size.zero,
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    elevation: 4,
                  ),
                  child: const Text('Đăng tin', style: TextStyle(fontSize: 14)),
                ),
              ],
            ),
    );
  }

  Widget _buildListingCard(Listing l) {
    final createdAt = DateTime.tryParse(l.createdAt);
    final statusColor = _statusColor(l.status);

    return Card(
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => context.push('/marketplace/${l.id}'),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header: title + status + menu
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 14, 4, 0),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          l.title,
                          style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(
                            color: statusColor.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            _statusLabel(l.status),
                            style: TextStyle(fontSize: 11, color: statusColor, fontWeight: FontWeight.w500),
                          ),
                        ),
                      ],
                    ),
                  ),
                  PopupMenuButton(
                    itemBuilder: (_) => [
                      const PopupMenuItem(value: 'edit', child: Text('Sửa')),
                      const PopupMenuItem(value: 'delete', child: Text('Xóa', style: TextStyle(color: AppColors.error))),
                    ],
                    onSelected: (v) {
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        if (v == 'edit') {
                          context.push('/edit-listing/${l.id}').then((result) {
                            if (result == true) _load();
                          });
                        } else if (v == 'delete') {
                          _delete(l.id);
                        }
                      });
                    },
                  ),
                ],
              ),
            ),

            // Image row
            if (l.images.isNotEmpty) ...[
              const SizedBox(height: 8),
              SizedBox(
                height: 88,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: l.images.length,
                  itemBuilder: (_, i) => Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(10),
                      child: CachedNetworkImage(
                        imageUrl: l.images[i],
                        width: 88,
                        height: 88,
                        fit: BoxFit.cover,
                        errorWidget: (_, __, ___) => Container(
                          width: 88,
                          height: 88,
                          color: AppColors.divider,
                          child: Icon(Icons.broken_image, color: AppColors.textHint),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ],

            // Details
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 14),
              child: Column(
                children: [
                  // Price + Quantity row
                  Row(
                    children: [
                      Icon(Icons.monetization_on_outlined, size: 16, color: AppColors.priceText),
                      const SizedBox(width: 4),
                      Text(
                        '${_priceFormat.format(l.pricePerKg)}đ/kg',
                        style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.priceText),
                      ),
                      const SizedBox(width: 16),
                      Icon(Icons.inventory_outlined, size: 16, color: AppColors.textSecondary),
                      const SizedBox(width: 4),
                      Text(
                        '${_priceFormat.format(l.quantityKg)} kg',
                        style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),

                  // Harvest season + Views + Date
                  Row(
                    children: [
                      if (l.harvestSeason != null && l.harvestSeason!.isNotEmpty) ...[
                        Icon(Icons.grass, size: 14, color: AppColors.textHint),
                        const SizedBox(width: 4),
                        Text(l.harvestSeason!, style: TextStyle(fontSize: 12, color: AppColors.textSecondary)),
                        const SizedBox(width: 12),
                      ],
                      Icon(Icons.visibility_outlined, size: 14, color: AppColors.textHint),
                      const SizedBox(width: 4),
                      Text('${l.viewCount}', style: TextStyle(fontSize: 12, color: AppColors.textSecondary)),
                      const Spacer(),
                      if (createdAt != null)
                        Text(
                          _dateFormat.format(createdAt.toLocal()),
                          style: TextStyle(fontSize: 11, color: AppColors.textHint),
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
