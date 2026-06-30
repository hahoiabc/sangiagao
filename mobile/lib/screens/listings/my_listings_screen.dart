import '../../widgets/thumbnail_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/listing.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/paginated_list_view.dart';

class MyListingsScreen extends ConsumerStatefulWidget {
  const MyListingsScreen({super.key});

  @override
  ConsumerState<MyListingsScreen> createState() => _MyListingsScreenState();
}

class _MyListingsScreenState extends ConsumerState<MyListingsScreen> {
  final _listKey = GlobalKey<PaginatedListViewState>();

  void _refresh() => _listKey.currentState?.refresh();

  Future<void> _bump(Listing listing) async {
    try {
      final res = await ref.read(apiServiceProvider).bumpListing(listing.id);
      if (!mounted) return;
      final newBumpCount = (res['bump_count'] as int?) ?? listing.bumpCount + 1;
      final remaining = (res['bump_remaining'] as int?) ?? (bumpLifetimeCap - newBumpCount);
      _refresh(); // tải lại để phản ánh ranking mới (tin lên đầu)
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Đã làm mới tin — còn $remaining/$bumpLifetimeCap lần')),
      );
    } on DioException catch (e) {
      if (!mounted) return;
      String msg = 'Không thể làm mới tin';
      if (e.response?.statusCode == 429) {
        msg = 'Vui lòng đợi trước khi làm mới lần nữa';
      } else if (e.response?.statusCode == 410) {
        msg = 'Tin đã đạt giới hạn $bumpLifetimeCap lần làm mới — vui lòng đăng tin mới';
      } else if (e.response?.data is Map && (e.response!.data as Map)['error'] != null) {
        msg = (e.response!.data as Map)['error'].toString();
      }
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Lỗi: $e')),
      );
    }
  }

  /// Tính số phút còn lại tới lần được phép bump tiếp.
  /// Trả null nếu đủ điều kiện bump (chưa từng hoặc đã đủ 5h54m).
  String? _bumpRemaining(Listing l) {
    if (l.bumpedAt == null) return null;
    final last = DateTime.tryParse(l.bumpedAt!);
    if (last == null) return null;
    final nextAllowed = last.add(const Duration(minutes: bumpCooldownMinutes));
    final diff = nextAllowed.difference(DateTime.now());
    if (diff.isNegative) return null;
    final hours = diff.inHours;
    final minutes = diff.inMinutes % 60;
    return hours > 0 ? '${hours}h${minutes > 0 ? "${minutes}p" : ""}' : '${minutes}p';
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
        _refresh();
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
      body: PaginatedListView<Listing>(
        key: _listKey,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        fetcher: (page) async =>
            (await ref.read(apiServiceProvider).getMyListings(page: page, limit: 20)).data,
        emptyState: Center(
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 48),
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
                    _refresh();
                  },
                  icon: const Icon(Icons.add),
                  label: const Text('Đăng tin ngay'),
                ),
              ],
            ),
          ),
        ),
        itemBuilder: (context, l, i) => Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: _buildListingCard(l),
        ),
      ),
      floatingActionButton: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                FilledButton.tonal(
                  onPressed: () async {
                    await context.push('/quick-batch');
                    _refresh();
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
                    _refresh();
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
                    itemBuilder: (_) {
                      final remaining = _bumpRemaining(l);
                      final quotaExhausted = l.bumpCount >= bumpLifetimeCap;
                      final canBump = l.status == 'active' && remaining == null && !quotaExhausted;
                      final bumpLabel = quotaExhausted
                          ? 'Hết lượt làm mới'
                          : remaining != null
                              ? 'Làm mới sau $remaining'
                              : 'Làm mới tin đăng';
                      return [
                        PopupMenuItem(
                          value: 'bump',
                          enabled: canBump,
                          child: Row(
                            children: [
                              Icon(Icons.refresh, size: 18, color: canBump ? AppColors.primary : AppColors.textHint),
                              const SizedBox(width: 8),
                              Text(bumpLabel, style: TextStyle(color: canBump ? AppColors.primary : AppColors.textHint)),
                            ],
                          ),
                        ),
                        const PopupMenuItem(value: 'edit', child: Text('Sửa')),
                        const PopupMenuItem(value: 'delete', child: Text('Xóa', style: TextStyle(color: AppColors.error))),
                      ];
                    },
                    onSelected: (v) {
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        if (v == 'bump') {
                          _bump(l);
                        } else if (v == 'edit') {
                          context.push('/edit-listing/${l.id}').then((result) {
                            if (result == true) _refresh();
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
                      child: SizedBox(
                        width: 88,
                        height: 88,
                        child: ThumbnailImage(
                          imageUrl: l.images[i],
                          fit: BoxFit.cover,
                          errorWidget: (_, __, ___) => Container(
                            color: AppColors.divider,
                            child: Icon(Icons.broken_image, color: AppColors.textHint),
                          ),
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
