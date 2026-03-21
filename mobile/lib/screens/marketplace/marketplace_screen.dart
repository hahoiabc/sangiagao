import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/listing.dart';
import '../../models/location.dart';
import '../../providers/providers.dart';
import '../../services/location_service.dart';
import '../../theme/app_theme.dart';

class MarketplaceScreen extends ConsumerStatefulWidget {
  final String? initialCategory;
  final String? initialType;
  final String? initialSort;
  const MarketplaceScreen({super.key, this.initialCategory, this.initialType, this.initialSort});

  @override
  ConsumerState<MarketplaceScreen> createState() => _MarketplaceScreenState();
}

class _MarketplaceScreenState extends ConsumerState<MarketplaceScreen> {
  List<Listing> _listings = [];
  bool _loading = true;
  int _page = 1;
  int _total = 0;
  static const _limit = 20;

  // Location filter
  final _locationService = LocationService();
  List<Province> _provinces = [];
  List<Ward> _wards = [];
  Province? _selectedProvince;
  Ward? _selectedWard;

  int get _totalPages => (_total / _limit).ceil();

  @override
  void initState() {
    super.initState();
    _loadProvinces();
    _loadListings();
  }

  Future<void> _loadProvinces() async {
    final provinces = await _locationService.getProvinces();
    if (mounted) setState(() => _provinces = provinces);
  }

  void _goToPage(int page) {
    if (page < 1 || page > _totalPages || page == _page) return;
    setState(() => _page = page);
    _loadListings();
  }

  Future<void> _loadListings() async {
    if (!mounted) return;
    setState(() => _loading = true);
    try {
      final api = ref.read(apiServiceProvider);
      final result = await api.searchMarketplace(
        category: widget.initialCategory,
        type: widget.initialType,
        sort: widget.initialSort,
        province: _selectedProvince?.name,
        ward: _selectedWard?.name,
        page: _page,
      );
      if (!mounted) return;
      setState(() {
        _listings = result.data;
        _total = result.total;
      });
    } catch (e) {
      debugPrint('Load listings error: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _onProvinceChanged(Province? province) async {
    setState(() {
      _selectedProvince = province;
      _selectedWard = null;
      _wards = [];
      _page = 1;
    });
    if (province != null) {
      final wards = await _locationService.getWards(province.code);
      if (mounted) setState(() => _wards = wards);
    }
    _loadListings();
  }

  void _onWardChanged(Ward? ward) {
    setState(() {
      _selectedWard = ward;
      _page = 1;
    });
    _loadListings();
  }

  final _priceFormat = NumberFormat('#,###', 'vi_VN');

  String get _title {
    String name;
    if (widget.initialType != null) {
      name = (_listings.isNotEmpty) ? _listings.first.title : widget.initialType!;
    } else {
      name = 'Kết quả tìm kiếm';
    }
    return _total > 0 ? '$name ($_total)' : name;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
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
      body: Column(
        children: [
          // Content
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _listings.isEmpty
                    ? const Center(child: Text('Không tìm thấy tin đăng nào'))
                    : RefreshIndicator(
                        onRefresh: _loadListings,
                        child: ListView.separated(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                          itemCount: _listings.length,
                          separatorBuilder: (_, __) => const SizedBox(height: 10),
                          itemBuilder: (context, index) {
                            final listing = _listings[index];
                            final stt = (_page - 1) * _limit + index + 1;
                            return Card(
                              clipBehavior: Clip.antiAlias,
                              margin: EdgeInsets.zero,
                              child: InkWell(
                                onTap: () => context.push('/marketplace/${listing.id}'),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                                  child: Row(
                                    children: [
                                      ClipRRect(
                                        borderRadius: BorderRadius.circular(8),
                                        child: SizedBox(
                                          width: 56,
                                          height: 56,
                                          child: listing.images.isNotEmpty
                                              ? CachedNetworkImage(
                                                  imageUrl: listing.images.first,
                                                  fit: BoxFit.cover,
                                                  placeholder: (_, __) => Container(
                                                    color: AppColors.border,
                                                    child: const Icon(Icons.image, size: 24, color: AppColors.textHint),
                                                  ),
                                                  errorWidget: (_, __, ___) => Container(
                                                    color: AppColors.border,
                                                    child: const Icon(Icons.broken_image, size: 24, color: AppColors.textHint),
                                                  ),
                                                )
                                              : Container(
                                                  color: AppColors.border,
                                                  child: const Icon(Icons.inventory_2_outlined, size: 24, color: AppColors.textHint),
                                                ),
                                        ),
                                      ),
                                      const SizedBox(width: 12),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              '${_priceFormat.format(listing.pricePerKg)}đ/kg',
                                              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.priceText),
                                            ),
                                            const SizedBox(height: 6),
                                            Text(
                                              '${_priceFormat.format(listing.quantityKg)} kg',
                                              style: TextStyle(fontSize: 14, color: AppColors.textSecondary),
                                            ),
                                          ],
                                        ),
                                      ),
                                      const SizedBox(width: 12),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.end,
                                          children: [
                                            if (listing.harvestSeason != null && listing.harvestSeason!.isNotEmpty)
                                              Text(listing.harvestSeason!, style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                                            if (listing.province != null && listing.province!.isNotEmpty) ...[
                                              const SizedBox(height: 4),
                                              Text(
                                                listing.province!,
                                                style: TextStyle(fontSize: 12, color: AppColors.textHint),
                                                overflow: TextOverflow.ellipsis,
                                              ),
                                            ],
                                          ],
                                        ),
                                      ),
                                      const SizedBox(width: 4),
                                      Icon(Icons.chevron_right, size: 22, color: AppColors.textHint),
                                    ],
                                  ),
                                ),
                              ),
                            );
                          },
                        ),
                      ),
          ),
          // Pagination
          if (_totalPages > 1)
            Container(
              padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 12),
              decoration: BoxDecoration(
                color: theme.colorScheme.surface,
                border: Border(top: BorderSide(color: theme.dividerColor)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  SizedBox(
                    height: 28,
                    width: 28,
                    child: IconButton(
                      onPressed: _page > 1 ? () => _goToPage(_page - 1) : null,
                      icon: const Icon(Icons.chevron_left),
                      iconSize: 16,
                      padding: EdgeInsets.zero,
                    ),
                  ),
                  ..._buildPageButtons(),
                  SizedBox(
                    height: 28,
                    width: 28,
                    child: IconButton(
                      onPressed: _page < _totalPages ? () => _goToPage(_page + 1) : null,
                      icon: const Icon(Icons.chevron_right),
                      iconSize: 16,
                      padding: EdgeInsets.zero,
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Future<void> _showFilterSheet() async {
    await showModalBottomSheet(
      context: context,
      builder: (_) => StatefulBuilder(
        builder: (context, setSheetState) => SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Text('Lọc theo địa điểm', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    const Spacer(),
                    if (_selectedProvince != null)
                      TextButton(
                        onPressed: () {
                          setState(() {
                            _selectedProvince = null;
                            _selectedWard = null;
                            _wards = [];
                            _page = 1;
                          });
                          _loadListings();
                          Navigator.pop(context);
                        },
                        child: const Text('Xoá bộ lọc'),
                      ),
                  ],
                ),
                const SizedBox(height: 12),
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
                const SizedBox(height: 8),
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

  List<Widget> _buildPageButtons() {
    final pages = <int>[];
    for (var i = 1; i <= _totalPages; i++) {
      if (i == 1 || i == _totalPages || (i >= _page - 1 && i <= _page + 1)) {
        pages.add(i);
      }
    }
    final widgets = <Widget>[];
    int? prev;
    for (final p in pages) {
      if (prev != null && p - prev > 1) {
        widgets.add(const Padding(
          padding: EdgeInsets.symmetric(horizontal: 2),
          child: Text('...', style: TextStyle(fontSize: 13, color: AppColors.textHint)),
        ));
      }
      final isActive = p == _page;
      widgets.add(
        InkWell(
          onTap: isActive ? null : () => _goToPage(p),
          borderRadius: BorderRadius.circular(6),
          child: Container(
            constraints: const BoxConstraints(minWidth: 26, minHeight: 26),
            alignment: Alignment.center,
            margin: const EdgeInsets.symmetric(horizontal: 2),
            decoration: BoxDecoration(
              color: isActive ? Theme.of(context).colorScheme.primary : null,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              '$p',
              style: TextStyle(
                fontSize: 13,
                fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
                color: isActive ? Colors.white : null,
              ),
            ),
          ),
        ),
      );
      prev = p;
    }
    return widgets;
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
