import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import '../../models/product_catalog.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

/// Tracks per-product input when the tile is expanded.
const _kMaxImages = 2;

class _ProductEntry {
  final RiceProduct product;
  final TextEditingController priceCtrl;
  final TextEditingController qtyCtrl;
  final TextEditingController seasonCtrl;
  final TextEditingController descCtrl;
  bool selected;
  List<String> imageUrls; // uploaded URLs
  List<String> localPaths; // local file paths for preview
  bool uploading;

  _ProductEntry(this.product)
      : priceCtrl = TextEditingController(),
        qtyCtrl = TextEditingController(),
        seasonCtrl = TextEditingController(),
        descCtrl = TextEditingController(),
        selected = false,
        imageUrls = [],
        localPaths = [],
        uploading = false;

  void dispose() {
    priceCtrl.dispose();
    qtyCtrl.dispose();
    seasonCtrl.dispose();
    descCtrl.dispose();
  }

  bool get isValid {
    final p = double.tryParse(priceCtrl.text.trim());
    final q = double.tryParse(qtyCtrl.text.trim());
    return selected && p != null && q != null && p > 5000 && p < 99000 && q > 500 && q < 100000000;
  }

  String? get validationError {
    final p = double.tryParse(priceCtrl.text.trim());
    final q = double.tryParse(qtyCtrl.text.trim());
    if (p == null || p <= 5000 || p >= 99000) {
      return '${product.label}: Giá phải từ 5,001 đến 98,999 đ/kg';
    }
    if (q == null || q <= 500 || q >= 100000000) {
      return '${product.label}: Số lượng phải từ 501 đến 99,999,999 kg';
    }
    final s = seasonCtrl.text.trim();
    if (s.isNotEmpty) {
      final parts = s.split('/');
      if (parts.length == 3) {
        final d = int.tryParse(parts[0]) ?? 0;
        final m = int.tryParse(parts[1]) ?? 0;
        final y = int.tryParse(parts[2]) ?? 0;
        final picked = DateTime(y, m, d);
        if (picked.isAfter(DateTime.now())) {
          return '${product.label}: Mùa vụ phải trước ngày hiện tại';
        }
      }
    }
    return null;
  }

  Map<String, dynamic>? toPayload(String categoryKey) {
    if (!isValid) return null;
    final map = <String, dynamic>{
      'category': categoryKey,
      'rice_type': product.key,
      'price_per_kg': double.parse(priceCtrl.text.trim()),
      'quantity_kg': double.parse(qtyCtrl.text.trim()),
    };
    final s = seasonCtrl.text.trim();
    if (s.isNotEmpty) map['harvest_season'] = s;
    final d = descCtrl.text.trim();
    if (d.isNotEmpty) map['description'] = d;
    return map;
  }
}

class QuickBatchScreen extends ConsumerStatefulWidget {
  const QuickBatchScreen({super.key});

  @override
  ConsumerState<QuickBatchScreen> createState() => _QuickBatchScreenState();
}

class _QuickBatchScreenState extends ConsumerState<QuickBatchScreen> {
  bool _loadingCatalog = true;
  bool _submitting = false;
  List<RiceCategory> _catalog = [];
  final _picker = ImagePicker();

  // null = show category grid, non-null = show product list for that category
  RiceCategory? _selectedCategory;
  List<_ProductEntry> _entries = [];

  @override
  void initState() {
    super.initState();
    _loadCatalog();
  }

  Future<void> _loadCatalog() async {
    try {
      final catalog = await ref.read(apiServiceProvider).getProductCatalog();
      if (mounted) setState(() { _catalog = catalog; _loadingCatalog = false; });
    } catch (e) {
      if (mounted) setState(() => _loadingCatalog = false);
    }
  }

  void _selectCategory(RiceCategory cat) {
    // Dispose old entries
    for (final e in _entries) {
      e.dispose();
    }
    setState(() {
      _selectedCategory = cat;
      _entries = cat.products.map((p) => _ProductEntry(p)).toList();
    });
  }

  void _backToCategories() {
    for (final e in _entries) {
      e.dispose();
    }
    setState(() {
      _selectedCategory = null;
      _entries = [];
    });
  }

  int get _selectedCount => _entries.where((e) => e.selected).length;

  Future<void> _pickImages(_ProductEntry entry) async {
    if (entry.imageUrls.length >= _kMaxImages) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tối đa $_kMaxImages hình ảnh')),
      );
      return;
    }
    final remaining = _kMaxImages - entry.imageUrls.length;
    final images = await _picker.pickMultiImage(
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 95,
      limit: remaining,
    );
    if (images.isEmpty) return;

    setState(() => entry.uploading = true);
    try {
      for (final image in images.take(remaining)) {
        final url = await ref.read(apiServiceProvider).uploadImage(image.path, 'listings');
        if (mounted) {
          setState(() {
            entry.imageUrls.add(url);
            entry.localPaths.add(image.path);
          });
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi tải ảnh: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => entry.uploading = false);
    }
  }

  void _removeImage(_ProductEntry entry, int index) {
    setState(() {
      entry.imageUrls.removeAt(index);
      entry.localPaths.removeAt(index);
    });
  }

  Future<void> _pickDate(_ProductEntry entry) async {
    final now = DateTime.now();
    int initialDay = now.day;
    int initialMonth = now.month;
    int initialYear = now.year;
    if (entry.seasonCtrl.text.isNotEmpty) {
      final parts = entry.seasonCtrl.text.split('/');
      if (parts.length == 3) {
        initialDay = int.tryParse(parts[0]) ?? now.day;
        initialMonth = int.tryParse(parts[1]) ?? now.month;
        initialYear = int.tryParse(parts[2]) ?? now.year;
      }
    }

    int selectedDay = initialDay;
    int selectedMonth = initialMonth;
    int selectedYear = initialYear;

    final result = await showModalBottomSheet<bool>(
      context: context,
      builder: (_) => StatefulBuilder(
        builder: (context, setSheetState) {
          int daysInMonth(int year, int month) => DateTime(year, month + 1, 0).day;
          final maxDay = daysInMonth(selectedYear, selectedMonth);
          if (selectedDay > maxDay) selectedDay = maxDay;

          return SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      const Text('Chọn ngày gặt', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                      const Spacer(),
                      IconButton(icon: const Icon(Icons.close), onPressed: () => Navigator.pop(context)),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: _DateDropdown(
                          label: 'Ngày',
                          value: selectedDay,
                          items: List.generate(maxDay, (i) => i + 1),
                          onChanged: (v) => setSheetState(() => selectedDay = v!),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _DateDropdown(
                          label: 'Tháng',
                          value: selectedMonth,
                          items: List.generate(12, (i) => i + 1),
                          onChanged: (v) => setSheetState(() => selectedMonth = v!),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _DateDropdown(
                          label: 'Năm',
                          value: selectedYear,
                          items: List.generate(6, (i) => now.year - 5 + i),
                          onChanged: (v) => setSheetState(() => selectedYear = v!),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  SizedBox(
                    width: double.infinity,
                    child: FilledButton(
                      onPressed: () => Navigator.pop(context, true),
                      child: const Text('Xác nhận'),
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
    if (result == true && mounted) {
      setState(() {
        entry.seasonCtrl.text = '${selectedDay.toString().padLeft(2, '0')}/${selectedMonth.toString().padLeft(2, '0')}/$selectedYear';
      });
    }
  }

  Future<void> _submit() async {
    final selected = _entries.where((e) => e.selected).toList();
    if (selected.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Vui lòng chọn ít nhất 1 sản phẩm')),
      );
      return;
    }

    final errors = <String>[];
    final items = <Map<String, dynamic>>[];
    for (final e in selected) {
      final vErr = e.validationError;
      if (vErr != null) {
        errors.add(vErr);
        continue;
      }
      final payload = e.toPayload(_selectedCategory!.key);
      if (payload == null) {
        errors.add('${e.product.label}: giá và số lượng phải > 0');
      } else {
        items.add(payload);
      }
    }

    if (errors.isNotEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(errors.join('\n'))),
      );
      return;
    }

    if (items.length > 20) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tối đa 20 sản phẩm mỗi lần đăng')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final api = ref.read(apiServiceProvider);
      final result = await api.batchCreateListings(items);
      final created = result['created'] as List? ?? [];
      final apiErrors = result['errors'] as List? ?? [];

      // Attach images to created listings
      for (int i = 0; i < created.length && i < selected.length; i++) {
        final entry = selected[i];
        if (entry.imageUrls.isEmpty) continue;
        final listingId = (created[i] as Map<String, dynamic>)['id'] as String;
        for (final url in entry.imageUrls) {
          try {
            await api.addListingImage(listingId, url);
          } catch (_) {}
        }
      }

      if (mounted) {
        String msg = 'Đã đăng ${created.length} tin thành công!';
        if (apiErrors.isNotEmpty) {
          msg += '\n${apiErrors.length} lỗi: ${apiErrors.join(', ')}';
        }
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
        if (apiErrors.isEmpty) context.pop(true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  void dispose() {
    for (final e in _entries) {
      e.dispose();
    }
    super.dispose();
  }

  // ─── Category icon mapping ───
  IconData _categoryIcon(String key) {
    switch (key) {
      case 'gao_deo_thom':
        return Icons.grain;
      case 'gao_no_kho':
        return Icons.rice_bowl;
      case 'tam':
        return Icons.grass;
      case 'nep':
        return Icons.spa;
      default:
        return Icons.category;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_selectedCategory == null ? 'Đăng nhanh' : _selectedCategory!.label),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: _selectedCategory == null ? () => context.pop() : _backToCategories,
        ),
      ),
      body: _loadingCatalog
          ? const Center(child: CircularProgressIndicator())
          : _selectedCategory == null
              ? _buildCategoryGrid()
              : _buildProductList(),
    );
  }

  // ─── Category Grid ───
  Widget _buildCategoryGrid() {
    if (_catalog.isEmpty) {
      return Center(
        child: Text('Không có danh mục nào', style: TextStyle(color: AppColors.textHint)),
      );
    }
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Chọn danh mục để đăng nhanh',
            style: TextStyle(fontSize: 15, color: AppColors.textSecondary),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: GridView.builder(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1.4,
              ),
              itemCount: _catalog.length,
              itemBuilder: (_, i) {
                final cat = _catalog[i];
                return Card(
                  clipBehavior: Clip.antiAlias,
                  child: InkWell(
                    onTap: () => _selectCategory(cat),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(_categoryIcon(cat.key), size: 36, color: AppColors.primary),
                          const SizedBox(height: 10),
                          Text(
                            cat.label,
                            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
                            textAlign: TextAlign.center,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '${cat.products.length} sản phẩm',
                            style: TextStyle(fontSize: 12, color: AppColors.textHint),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  // ─── Product List with ExpansionTile ───
  Widget _buildProductList() {
    return Column(
      children: [
        // Selected counter
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          color: AppColors.surfaceVariant,
          child: Row(
            children: [
              Icon(Icons.check_circle_outline, size: 18, color: AppColors.primary),
              const SizedBox(width: 8),
              Text(
                'Đã chọn $_selectedCount / ${_entries.length} sản phẩm',
                style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
              ),
              const Spacer(),
              if (_selectedCount > 0)
                TextButton(
                  onPressed: () {
                    setState(() {
                      for (final e in _entries) {
                        e.selected = false;
                      }
                    });
                  },
                  child: const Text('Bỏ chọn tất cả', style: TextStyle(fontSize: 12)),
                ),
            ],
          ),
        ),
        // Product tiles
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.fromLTRB(0, 0, 0, 100),
            itemCount: _entries.length,
            itemBuilder: (_, i) => _buildProductTile(_entries[i]),
          ),
        ),
        // Submit button
        SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
            child: SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: _submitting || _selectedCount == 0 ? null : _submit,
                child: _submitting
                    ? const SizedBox(
                        height: 20, width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                      )
                    : Text('Đăng $_selectedCount tin'),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildProductTile(_ProductEntry entry) {
    return ExpansionTile(
      leading: Checkbox(
        value: entry.selected,
        onChanged: (v) => setState(() => entry.selected = v ?? false),
      ),
      title: Text(entry.product.label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
      subtitle: entry.selected && entry.priceCtrl.text.isNotEmpty
          ? Text(
              '${NumberFormat('#,###', 'vi_VN').format(double.tryParse(entry.priceCtrl.text) ?? 0)}đ/kg',
              style: TextStyle(fontSize: 12, color: AppColors.priceText),
            )
          : null,
      initiallyExpanded: false,
      onExpansionChanged: (expanded) {
        if (expanded && !entry.selected) {
          setState(() => entry.selected = true);
        }
      },
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
          child: Column(
            children: [
              // Price + Quantity
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: entry.priceCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Giá (đ/kg) *',
                        border: OutlineInputBorder(),
                        isDense: true,
                      ),
                      keyboardType: TextInputType.number,
                      onChanged: (_) => setState(() {}),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: TextField(
                      controller: entry.qtyCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Số lượng (kg) *',
                        border: OutlineInputBorder(),
                        isDense: true,
                      ),
                      keyboardType: TextInputType.number,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              // Season (date picker)
              TextField(
                controller: entry.seasonCtrl,
                readOnly: true,
                decoration: InputDecoration(
                  labelText: 'Mùa gặt',
                  border: const OutlineInputBorder(),
                  isDense: true,
                  suffixIcon: IconButton(
                    icon: const Icon(Icons.calendar_today, size: 18),
                    onPressed: () => _pickDate(entry),
                  ),
                ),
                onTap: () => _pickDate(entry),
              ),
              const SizedBox(height: 10),
              // Description
              TextField(
                controller: entry.descCtrl,
                decoration: const InputDecoration(
                  labelText: 'Mô tả thêm',
                  border: OutlineInputBorder(),
                  isDense: true,
                ),
                maxLines: 2,
                minLines: 1,
              ),
              const SizedBox(height: 10),
              // Images
              Text('Hình ảnh (tối đa $_kMaxImages)', style: TextStyle(fontSize: 12, color: AppColors.textSecondary)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  ...entry.localPaths.asMap().entries.map((e) {
                    return Stack(
                      clipBehavior: Clip.none,
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(8),
                          child: Image.file(
                            File(e.value),
                            width: 72,
                            height: 72,
                            fit: BoxFit.cover,
                          ),
                        ),
                        Positioned(
                          top: -4,
                          right: -4,
                          child: GestureDetector(
                            onTap: () => _removeImage(entry, e.key),
                            child: Container(
                              decoration: const BoxDecoration(color: Colors.black54, shape: BoxShape.circle),
                              padding: const EdgeInsets.all(3),
                              child: const Icon(Icons.close, size: 14, color: Colors.white),
                            ),
                          ),
                        ),
                      ],
                    );
                  }),
                  if (entry.imageUrls.length < _kMaxImages)
                    GestureDetector(
                      onTap: entry.uploading ? null : () => _pickImages(entry),
                      child: Container(
                        width: 72,
                        height: 72,
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.grey.shade300, style: BorderStyle.solid, width: 1.5),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: entry.uploading
                            ? const Center(child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)))
                            : Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(Icons.add_photo_alternate_outlined, size: 24, color: Colors.grey.shade500),
                                  Text('Thêm ảnh', style: TextStyle(fontSize: 10, color: Colors.grey.shade500)),
                                ],
                              ),
                      ),
                    ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _DateDropdown extends StatelessWidget {
  final String label;
  final int value;
  final List<int> items;
  final ValueChanged<int?> onChanged;

  const _DateDropdown({required this.label, required this.value, required this.items, required this.onChanged});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: TextStyle(fontSize: 12, color: AppColors.textSecondary)),
        const SizedBox(height: 4),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            border: Border.all(color: AppColors.border),
            borderRadius: BorderRadius.circular(8),
          ),
          child: DropdownButton<int>(
            value: items.contains(value) ? value : items.last,
            isExpanded: true,
            underline: const SizedBox(),
            items: items.map((v) => DropdownMenuItem(value: v, child: Text('$v'))).toList(),
            onChanged: onChanged,
          ),
        ),
      ],
    );
  }
}
