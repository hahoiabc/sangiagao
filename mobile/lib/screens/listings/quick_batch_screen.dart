import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../models/product_catalog.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

/// Tracks per-product input when the tile is expanded.
class _ProductEntry {
  final RiceProduct product;
  final TextEditingController priceCtrl;
  final TextEditingController qtyCtrl;
  final TextEditingController seasonCtrl;
  final TextEditingController descCtrl;
  bool selected;

  _ProductEntry(this.product)
      : priceCtrl = TextEditingController(),
        qtyCtrl = TextEditingController(),
        seasonCtrl = TextEditingController(),
        descCtrl = TextEditingController(),
        selected = false;

  void dispose() {
    priceCtrl.dispose();
    qtyCtrl.dispose();
    seasonCtrl.dispose();
    descCtrl.dispose();
  }

  bool get isValid {
    final p = double.tryParse(priceCtrl.text.trim());
    final q = double.tryParse(qtyCtrl.text.trim());
    return selected && p != null && q != null && p > 0 && q > 0;
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
      final result = await ref.read(apiServiceProvider).batchCreateListings(items);
      final created = result['created'] as List? ?? [];
      final apiErrors = result['errors'] as List? ?? [];

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
              // Season
              TextField(
                controller: entry.seasonCtrl,
                decoration: const InputDecoration(
                  labelText: 'Vụ mùa',
                  border: OutlineInputBorder(),
                  isDense: true,
                ),
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
            ],
          ),
        ),
      ],
    );
  }
}
