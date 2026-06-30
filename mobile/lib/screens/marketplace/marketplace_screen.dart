import 'dart:ui';
import '../../widgets/thumbnail_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/listing.dart';
import '../../models/location.dart';
import '../../providers/providers.dart';
import '../../providers/user_block_provider.dart';
import '../../services/location_service.dart';
import '../../theme/app_theme.dart';
import '../../widgets/paginated_list_view.dart';

class MarketplaceScreen extends ConsumerStatefulWidget {
  final String? initialCategory;
  final String? initialType;
  final String? initialSort;
  const MarketplaceScreen({super.key, this.initialCategory, this.initialType, this.initialSort});

  @override
  ConsumerState<MarketplaceScreen> createState() => _MarketplaceScreenState();
}

class _MarketplaceScreenState extends ConsumerState<MarketplaceScreen> {
  final _listKey = GlobalKey<PaginatedListViewState>();
  int _total = 0;
  String? _firstTitle;

  // Location filter
  final _locationService = LocationService();
  List<Province> _provinces = [];
  List<Ward> _wards = [];
  Province? _selectedProvince;
  Ward? _selectedWard;

  // Content filters — "Chỉ tin có ảnh" và "Tin mới trong X ngày"
  bool _hasPhotoOnly = false;
  int _postedWithinDays = 0; // 0 = không lọc; 1 / 7 / 30 ngày

  @override
  void initState() {
    super.initState();
    _loadProvinces();
  }

  Future<void> _loadProvinces() async {
    final provinces = await _locationService.getProvinces();
    if (mounted) setState(() => _provinces = provinces);
  }

  void _refresh() => _listKey.currentState?.refresh();

  /// Lấy 1 trang tin theo bộ lọc hiện tại + lọc người bị chặn.
  Future<List<Listing>> _fetch(int page) async {
    final api = ref.read(apiServiceProvider);
    final result = await api.searchMarketplace(
      category: widget.initialCategory,
      type: widget.initialType,
      sort: widget.initialSort,
      province: _selectedProvince?.name,
      ward: _selectedWard?.name,
      hasPhoto: _hasPhotoOnly,
      postedWithinDays: _postedWithinDays > 0 ? _postedWithinDays : null,
      page: page,
    );
    if (page == 1 && mounted) {
      setState(() {
        _total = result.total;
        _firstTitle = result.data.isNotEmpty ? result.data.first.title : null;
      });
    }
    final blocked = ref.read(userBlockProvider);
    return result.data.where((l) => !blocked.contains(l.userId)).toList();
  }

  void _onProvinceChanged(Province? province) async {
    setState(() {
      _selectedProvince = province;
      _selectedWard = null;
      _wards = [];
    });
    if (province != null) {
      final wards = await _locationService.getWards(province.code);
      if (mounted) setState(() => _wards = wards);
    }
    _refresh();
  }

  void _onWardChanged(Ward? ward) {
    setState(() => _selectedWard = ward);
    _refresh();
  }

  final _priceFormat = NumberFormat('#,###', 'vi_VN');

  String get _title {
    String name;
    if (widget.initialType != null) {
      name = _firstTitle ?? widget.initialType!;
    } else {
      name = 'Kết quả tìm kiếm';
    }
    return _total > 0 ? '$name ($_total)' : name;
  }

  @override
  Widget build(BuildContext context) {
    // Khi danh sách người bị chặn đổi → tải lại để ẩn ngay (Apple GL 1.2).
    ref.listen(userBlockProvider, (_, __) => _refresh());
    return Scaffold(
      appBar: AppBar(
        title: Text(_title),
        actions: [
          Stack(
            children: [
              IconButton(
                icon: const Icon(Icons.filter_list),
                onPressed: _showFilterSheet,
              ),
              if (_selectedProvince != null)
                Positioned(
                  right: 8,
                  top: 8,
                  child: Container(
                    width: 8,
                    height: 8,
                    decoration: const BoxDecoration(color: AppColors.error, shape: BoxShape.circle),
                  ),
                ),
            ],
          ),
        ],
      ),
      body: PaginatedListView<Listing>(
        key: _listKey,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        fetcher: _fetch,
        emptyState: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.search_off, size: 48, color: Colors.grey.shade400),
                const SizedBox(height: 12),
                const Text('Không tìm thấy tin đăng nào', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w500)),
                const SizedBox(height: 4),
                Text('Thử thay đổi từ khóa hoặc bộ lọc', style: TextStyle(fontSize: 13, color: Colors.grey.shade500)),
              ],
            ),
          ),
        ),
        itemBuilder: (context, listing, index) => Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: Card(
                              clipBehavior: Clip.antiAlias,
                              margin: EdgeInsets.zero,
                              child: InkWell(
                                onTap: () => context.push('/marketplace/${listing.id}'),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Image section — 3/4 of card
                                    AspectRatio(
                                      aspectRatio: 4 / 3,
                                      child: listing.images.isNotEmpty
                                          ? ClipRect(
                                              child: Stack(
                                                fit: StackFit.expand,
                                                children: [
                                                  // Blur background
                                                  ImageFiltered(
                                                    imageFilter: ImageFilter.blur(sigmaX: 20, sigmaY: 20),
                                                    child: ThumbnailImage(
                                                      imageUrl: listing.images.first,
                                                      fit: BoxFit.cover,
                                                      color: Colors.black.withValues(alpha: 0.3),
                                                      colorBlendMode: BlendMode.darken,
                                                    ),
                                                  ),
                                                  // Main image
                                                  ThumbnailImage(
                                                    imageUrl: listing.images.first,
                                                    fit: BoxFit.contain,
                                                    placeholder: (_, __) => Container(
                                                      color: AppColors.border,
                                                      child: const Center(child: Icon(Icons.image, size: 40, color: AppColors.textHint)),
                                                    ),
                                                    errorWidget: (_, __, ___) => Container(
                                                      color: AppColors.border,
                                                      child: const Center(child: Icon(Icons.broken_image, size: 40, color: AppColors.textHint)),
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            )
                                          : Container(
                                              color: AppColors.border,
                                              child: const Center(child: Icon(Icons.inventory_2_outlined, size: 40, color: AppColors.textHint)),
                                            ),
                                    ),
                                    // Text section — 1/4 of card
                                    Padding(
                                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                                      child: Row(
                                        children: [
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment: CrossAxisAlignment.start,
                                              children: [
                                                Text(
                                                  listing.title,
                                                  style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                                                  maxLines: 1,
                                                  overflow: TextOverflow.ellipsis,
                                                ),
                                                const SizedBox(height: 4),
                                                Row(
                                                  children: [
                                                    Text(
                                                      '${_priceFormat.format(listing.pricePerKg)}đ/kg',
                                                      style: TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: AppColors.priceText),
                                                    ),
                                                    const SizedBox(width: 12),
                                                    Text(
                                                      '${_priceFormat.format(listing.quantityKg)} kg',
                                                      style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
                                                    ),
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                          if (listing.province != null && listing.province!.isNotEmpty)
                                            Flexible(
                                              flex: 0,
                                              child: Padding(
                                                padding: const EdgeInsets.only(left: 8),
                                                child: Row(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    Icon(Icons.location_on_outlined, size: 14, color: AppColors.textHint),
                                                    const SizedBox(width: 2),
                                                    Flexible(
                                                      child: Text(
                                                        listing.province!,
                                                        style: TextStyle(fontSize: 12, color: AppColors.textHint),
                                                        overflow: TextOverflow.ellipsis,
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
          ),
        ),
      );
  }

  Future<void> _showFilterSheet() async {
    final hasAnyFilter = _selectedProvince != null || _hasPhotoOnly || _postedWithinDays > 0;
    await showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (_) => StatefulBuilder(
        builder: (context, setSheetState) => SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Text('Lọc tin đăng', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    const Spacer(),
                    if (hasAnyFilter)
                      TextButton(
                        onPressed: () {
                          setState(() {
                            _selectedProvince = null;
                            _selectedWard = null;
                            _wards = [];
                            _hasPhotoOnly = false;
                            _postedWithinDays = 0;
                          });
                          _refresh();
                          Navigator.pop(context);
                        },
                        child: const Text('Xoá bộ lọc'),
                      ),
                  ],
                ),
                const SizedBox(height: 12),

                // Location group
                Text('Địa điểm', style: TextStyle(fontSize: 13, color: AppColors.textSecondary, fontWeight: FontWeight.w600)),
                const SizedBox(height: 8),
                _FilterChip(
                  label: _selectedProvince?.name ?? 'Chọn Tỉnh/Thành phố',
                  isActive: _selectedProvince != null,
                  onTap: () async {
                    Navigator.pop(context);
                    await _showProvinceSheet();
                  },
                ),
                const SizedBox(height: 8),
                _FilterChip(
                  label: _selectedWard?.name ?? 'Chọn Xã/Phường',
                  isActive: _selectedWard != null,
                  enabled: _selectedProvince != null,
                  onTap: _selectedProvince != null
                      ? () async {
                          Navigator.pop(context);
                          await _showWardSheet();
                        }
                      : null,
                ),

                const SizedBox(height: 16),
                Text('Nội dung tin', style: TextStyle(fontSize: 13, color: AppColors.textSecondary, fontWeight: FontWeight.w600)),
                const SizedBox(height: 8),

                // Chỉ tin có ảnh
                InkWell(
                  onTap: () {
                    setSheetState(() => _hasPhotoOnly = !_hasPhotoOnly);
                    setState(() {});
                  },
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 6),
                    child: Row(
                      children: [
                        Icon(
                          _hasPhotoOnly ? Icons.check_box : Icons.check_box_outline_blank,
                          color: _hasPhotoOnly ? AppColors.primary : AppColors.textHint,
                          size: 22,
                        ),
                        const SizedBox(width: 12),
                        const Text('Chỉ hiển thị tin có ảnh', style: TextStyle(fontSize: 14)),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 12),

                // Thời gian đăng (chip group)
                Text('Thời gian đăng', style: TextStyle(fontSize: 13, color: AppColors.textSecondary, fontWeight: FontWeight.w600)),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _DayChip(label: 'Tất cả',    value: 0,  selected: _postedWithinDays == 0,
                      onTap: () { setSheetState(() => _postedWithinDays = 0); setState(() {}); }),
                    _DayChip(label: 'Trong 24h', value: 1,  selected: _postedWithinDays == 1,
                      onTap: () { setSheetState(() => _postedWithinDays = 1); setState(() {}); }),
                    _DayChip(label: 'Trong 7 ngày', value: 7, selected: _postedWithinDays == 7,
                      onTap: () { setSheetState(() => _postedWithinDays = 7); setState(() {}); }),
                    _DayChip(label: 'Trong 30 ngày', value: 30, selected: _postedWithinDays == 30,
                      onTap: () { setSheetState(() => _postedWithinDays = 30); setState(() {}); }),
                  ],
                ),

                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed: () {
                      Navigator.pop(context);
                      _refresh();
                    },
                    child: const Text('Áp dụng'),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _showProvinceSheet() async {
    final result = await showModalBottomSheet<Province>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _LocationSearchSheet<Province>(
        title: 'Chọn Tỉnh/Thành phố',
        items: _provinces,
        getName: (p) => p.name,
        selected: _selectedProvince,
      ),
    );
    if (result != null) _onProvinceChanged(result);
  }

  Future<void> _showWardSheet() async {
    final result = await showModalBottomSheet<Ward>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _LocationSearchSheet<Ward>(
        title: 'Chọn Xã/Phường',
        items: _wards,
        getName: (w) => w.name,
        selected: _selectedWard,
      ),
    );
    if (result != null) _onWardChanged(result);
  }

}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool isActive;
  final bool enabled;
  final VoidCallback? onTap;

  const _FilterChip({
    required this.label,
    this.isActive = false,
    this.enabled = true,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        decoration: BoxDecoration(
          color: isActive ? theme.colorScheme.primaryContainer : theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isActive ? theme.colorScheme.primary : AppColors.border,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.location_on_outlined,
              size: 16,
              color: enabled ? (isActive ? theme.colorScheme.primary : AppColors.textSecondary) : AppColors.textHint,
            ),
            const SizedBox(width: 4),
            Flexible(
              child: Text(
                label,
                style: TextStyle(
                  fontSize: 13,
                  color: enabled ? (isActive ? theme.colorScheme.primary : AppColors.textSecondary) : AppColors.textHint,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 2),
            Icon(Icons.arrow_drop_down, size: 18, color: enabled ? AppColors.textSecondary : AppColors.textHint),
          ],
        ),
      ),
    );
  }
}

class _DayChip extends StatelessWidget {
  final String label;
  final int value;
  final bool selected;
  final VoidCallback onTap;
  const _DayChip({required this.label, required this.value, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: selected ? theme.colorScheme.primaryContainer : Colors.transparent,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: selected ? theme.colorScheme.primary : AppColors.border,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            color: selected ? theme.colorScheme.primary : AppColors.textSecondary,
            fontWeight: selected ? FontWeight.w600 : FontWeight.normal,
          ),
        ),
      ),
    );
  }
}

class _LocationSearchSheet<T> extends StatefulWidget {
  final String title;
  final List<T> items;
  final String Function(T) getName;
  final T? selected;

  const _LocationSearchSheet({
    required this.title,
    required this.items,
    required this.getName,
    this.selected,
  });

  @override
  State<_LocationSearchSheet<T>> createState() => _LocationSearchSheetState<T>();
}

class _LocationSearchSheetState<T> extends State<_LocationSearchSheet<T>> {
  final _searchCtrl = TextEditingController();
  List<T> _filtered = [];

  @override
  void initState() {
    super.initState();
    _filtered = widget.items;
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  void _filter(String query) {
    if (query.isEmpty) {
      setState(() => _filtered = widget.items);
      return;
    }
    final q = _removeDiacritics(query.toLowerCase());
    setState(() {
      _filtered = widget.items
          .where((item) => _removeDiacritics(widget.getName(item).toLowerCase()).contains(q))
          .toList();
    });
  }

  static String _removeDiacritics(String str) {
    const withDiacritics = 'àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ';
    const withoutDiacritics = 'aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd';
    var result = str;
    for (int i = 0; i < withDiacritics.length; i++) {
      result = result.replaceAll(withDiacritics[i], withoutDiacritics[i]);
    }
    return result;
  }

  @override
  Widget build(BuildContext context) {
    final bottomInset = MediaQuery.of(context).viewInsets.bottom;
    return Padding(
      padding: EdgeInsets.only(bottom: bottomInset),
      child: DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.4,
        maxChildSize: 0.9,
        expand: false,
        builder: (context, scrollController) => Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: Row(
                children: [
                  Expanded(
                    child: Text(widget.title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  ),
                  IconButton(icon: const Icon(Icons.close), onPressed: () => Navigator.pop(context)),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: TextField(
                controller: _searchCtrl,
                autofocus: true,
                decoration: InputDecoration(
                  hintText: 'Tìm kiếm...',
                  prefixIcon: const Icon(Icons.search),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12),
                ),
                onChanged: _filter,
              ),
            ),
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text('${_filtered.length} kết quả', style: TextStyle(color: AppColors.textSecondary, fontSize: 13)),
              ),
            ),
            const Divider(),
            Expanded(
              child: ListView.builder(
                controller: scrollController,
                itemCount: _filtered.length,
                itemBuilder: (context, index) {
                  final item = _filtered[index];
                  final name = widget.getName(item);
                  final isSelected = widget.selected != null && item == widget.selected;
                  return ListTile(
                    title: Text(name),
                    trailing: isSelected ? const Icon(Icons.check, color: AppColors.primary) : null,
                    selected: isSelected,
                    onTap: () => Navigator.pop(context, item),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
