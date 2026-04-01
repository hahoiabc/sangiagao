import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import '../../models/product_catalog.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class _ProductForm {
  RiceProduct? product;
  RiceCategory? category;
  final TextEditingController priceCtrl;
  final TextEditingController quantityCtrl;
  final TextEditingController seasonCtrl;
  final TextEditingController descCtrl;
  final List<String> imageUrls = [];
  final List<String> localPaths = [];
  bool uploading = false;

  _ProductForm()
      : priceCtrl = TextEditingController(),
        quantityCtrl = TextEditingController(),
        seasonCtrl = TextEditingController(),
        descCtrl = TextEditingController();

  void dispose() {
    priceCtrl.dispose();
    quantityCtrl.dispose();
    seasonCtrl.dispose();
    descCtrl.dispose();
  }

  bool get isValid {
    if (product == null) return false;
    final price = double.tryParse(priceCtrl.text.trim());
    final qty = double.tryParse(quantityCtrl.text.trim());
    if (price == null || qty == null) return false;
    if (price <= 5000 || price >= 99000) return false;
    if (qty <= 500 || qty >= 100000000) return false;
    return true;
  }

  String? get validationError {
    if (product == null) return 'Chưa chọn sản phẩm';
    final price = double.tryParse(priceCtrl.text.trim());
    final qty = double.tryParse(quantityCtrl.text.trim());
    if (price == null || price <= 5000 || price >= 99000) {
      return '${product!.label}: Giá phải từ 5,001 đến 98,999 đ/kg';
    }
    if (qty == null || qty <= 500 || qty >= 100000000) {
      return '${product!.label}: Số lượng phải từ 501 đến 99,999,999 kg';
    }
    final season = seasonCtrl.text.trim();
    if (season.isNotEmpty) {
      final parts = season.split('/');
      if (parts.length == 3) {
        final d = int.tryParse(parts[0]) ?? 0;
        final m = int.tryParse(parts[1]) ?? 0;
        final y = int.tryParse(parts[2]) ?? 0;
        final picked = DateTime(y, m, d);
        if (picked.isAfter(DateTime.now())) {
          return '${product!.label}: Mùa vụ phải trước ngày hiện tại';
        }
      }
    }
    return null;
  }

  bool get hasData =>
      product != null ||
      priceCtrl.text.trim().isNotEmpty ||
      quantityCtrl.text.trim().isNotEmpty;

  Map<String, dynamic>? toPayload() {
    if (product == null || category == null) return null;
    final price = double.tryParse(priceCtrl.text.trim());
    final qty = double.tryParse(quantityCtrl.text.trim());
    if (price == null || qty == null || price <= 0 || qty <= 0) return null;
    final map = <String, dynamic>{
      'category': category!.key,
      'rice_type': product!.key,
      'price_per_kg': price,
      'quantity_kg': qty,
    };
    final season = seasonCtrl.text.trim();
    if (season.isNotEmpty) map['harvest_season'] = season;
    final desc = descCtrl.text.trim();
    if (desc.isNotEmpty) map['description'] = desc;
    return map;
  }
}

class CreateListingScreen extends ConsumerStatefulWidget {
  const CreateListingScreen({super.key});

  @override
  ConsumerState<CreateListingScreen> createState() =>
      _CreateListingScreenState();
}

class _CreateListingScreenState extends ConsumerState<CreateListingScreen> {
  bool _loadingCatalog = true;
  bool _submitting = false;
  List<RiceCategory> _catalog = [];
  final List<_ProductForm> _forms = [];
  final _picker = ImagePicker();
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _forms.add(_ProductForm());
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

  @override
  void dispose() {
    for (final f in _forms) {
      f.dispose();
    }
    _scrollController.dispose();
    super.dispose();
  }

  void _addForm() {
    setState(() => _forms.add(_ProductForm()));
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    });
  }

  void _removeForm(int index) {
    if (_forms.length <= 1) return;
    setState(() {
      _forms[index].dispose();
      _forms.removeAt(index);
    });
  }

  Future<void> _showProductPicker(_ProductForm form) async {
    final result = await showModalBottomSheet<({RiceCategory cat, RiceProduct product})>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _ProductPickerSheet(catalog: _catalog, current: form.product),
    );
    if (result != null && mounted) {
      setState(() {
        form.category = result.cat;
        form.product = result.product;
      });
    }
  }

  Future<void> _pickImage(_ProductForm form) async {
    if (form.imageUrls.length >= 3) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tối đa 3 hình ảnh')),
      );
      return;
    }
    final remaining = 3 - form.imageUrls.length;
    final images = await _picker.pickMultiImage(
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 95,
      limit: remaining,
    );
    if (images.isEmpty) return;

    setState(() => form.uploading = true);
    try {
      for (final image in images.take(remaining)) {
        final url = await ref.read(apiServiceProvider).uploadImage(image.path, 'listings');
        if (mounted) {
          setState(() {
            form.imageUrls.add(url);
            form.localPaths.add(image.path);
          });
        }
      }
      if (mounted) {
        setState(() => form.uploading = false);
      }
    } catch (e) {
      if (mounted) {
        setState(() => form.uploading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi tải ảnh: $e')),
        );
      }
    }
  }

  void _removeImage(_ProductForm form, int index) {
    setState(() {
      form.imageUrls.removeAt(index);
      form.localPaths.removeAt(index);
    });
  }

  Future<void> _pickDate(_ProductForm form) async {
    final now = DateTime.now();
    // Parse existing value if any
    int initialDay = now.day;
    int initialMonth = now.month;
    int initialYear = now.year;
    if (form.seasonCtrl.text.isNotEmpty) {
      final parts = form.seasonCtrl.text.split('/');
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
                      // Day
                      Expanded(
                        child: _DateDropdown(
                          label: 'Ngày',
                          value: selectedDay,
                          items: List.generate(maxDay, (i) => i + 1),
                          onChanged: (v) => setSheetState(() => selectedDay = v!),
                        ),
                      ),
                      const SizedBox(width: 12),
                      // Month
                      Expanded(
                        child: _DateDropdown(
                          label: 'Tháng',
                          value: selectedMonth,
                          items: List.generate(12, (i) => i + 1),
                          onChanged: (v) => setSheetState(() => selectedMonth = v!),
                        ),
                      ),
                      const SizedBox(width: 12),
                      // Year
                      Expanded(
                        child: _DateDropdown(
                          label: 'Năm',
                          value: selectedYear,
                          items: List.generate(now.year - 2000 + 6, (i) => 2000 + i),
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
        form.seasonCtrl.text = '${selectedDay.toString().padLeft(2, '0')}/${selectedMonth.toString().padLeft(2, '0')}/$selectedYear';
      });
    }
  }

  Future<void> _submit() async {
    final items = <Map<String, dynamic>>[];
    final validForms = <_ProductForm>[];
    final errors = <String>[];

    for (int i = 0; i < _forms.length; i++) {
      final form = _forms[i];
      if (!form.hasData) continue;
      if (form.product == null) {
        errors.add('Sản phẩm ${i + 1}: chưa chọn danh mục');
        continue;
      }
      final vErr = form.validationError;
      if (vErr != null) {
        errors.add(vErr);
        continue;
      }
      final payload = form.toPayload();
      if (payload == null) {
        errors.add('${form.product!.label}: giá và số lượng phải > 0');
      } else {
        items.add(payload);
        validForms.add(form);
      }
    }

    if (items.isEmpty && errors.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Vui lòng nhập thông tin ít nhất 1 sản phẩm')),
      );
      return;
    }
    if (errors.isNotEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(errors.join('\n'))),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final api = ref.read(apiServiceProvider);
      final result = await api.batchCreateListings(items);
      final createdList = result['created'] as List? ?? [];
      final apiErrors = result['errors'] as List? ?? [];

      // Attach images
      for (int i = 0; i < createdList.length && i < validForms.length; i++) {
        final form = validForms[i];
        if (form.imageUrls.isEmpty) continue;
        final listingId = (createdList[i] as Map<String, dynamic>)['id'] as String;
        for (final url in form.imageUrls) {
          try {
            await api.addListingImage(listingId, url);
          } catch (_) {}
        }
      }

      if (mounted) {
        String msg = 'Đã đăng ${createdList.length} tin thành công!';
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
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Đăng tin'),
        automaticallyImplyLeading: false,
        actions: [
          TextButton(
            onPressed: _submitting ? null : () => context.pop(),
            child: const Text('Hủy'),
          ),
        ],
      ),
      body: _loadingCatalog
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                Expanded(
                  child: SingleChildScrollView(
                    controller: _scrollController,
                    padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                    child: Column(
                      children: [
                        for (int i = 0; i < _forms.length; i++) ...[
                          if (i > 0) const SizedBox(height: 16),
                          _buildFormCard(i),
                        ],
                        const SizedBox(height: 16),
                        Center(
                          child: TextButton.icon(
                            onPressed: _addForm,
                            icon: const Icon(Icons.add_circle_outline),
                            label: const Text('Thêm sản phẩm'),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                // Submit button
                SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
                    child: SizedBox(
                      width: double.infinity,
                      child: FilledButton(
                        onPressed: _submitting ? null : _submit,
                        child: _submitting
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                              )
                            : Text('Đăng sản phẩm (${_forms.where((f) => f.hasData).length})'),
                      ),
                    ),
                  ),
                ),
              ],
            ),
    );
  }

  Widget _buildFormCard(int index) {
    final form = _forms[index];
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header: title + remove
            Row(
              children: [
                Text(
                  'Sản phẩm ${index + 1}',
                  style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                ),
                const Spacer(),
                if (_forms.length > 1)
                  IconButton(
                    onPressed: () => _removeForm(index),
                    icon: const Icon(Icons.close, size: 20),
                    visualDensity: VisualDensity.compact,
                    color: AppColors.textHint,
                  ),
              ],
            ),
            const SizedBox(height: 12),

            // Product picker
            InkWell(
              onTap: () => _showProductPicker(form),
              borderRadius: BorderRadius.circular(14),
              child: InputDecorator(
                decoration: InputDecoration(
                  labelText: 'Danh mục sản phẩm',
                  border: const OutlineInputBorder(),
                  suffixIcon: const Icon(Icons.arrow_drop_down),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  filled: form.product != null,
                ),
                child: Text(
                  form.product?.label ?? 'Chọn sản phẩm',
                  style: TextStyle(
                    fontSize: 15,
                    color: form.product != null ? AppColors.textPrimary : AppColors.textHint,
                  ),
                ),
              ),
            ),
            if (form.category != null) ...[
              const SizedBox(height: 4),
              Padding(
                padding: const EdgeInsets.only(left: 4),
                child: Text(
                  form.category!.label,
                  style: TextStyle(fontSize: 12, color: AppColors.textSecondary),
                ),
              ),
            ],
            const SizedBox(height: 16),

            // Price + Quantity
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: form.priceCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Giá (đ/kg)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.number,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: form.quantityCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Số lượng (kg)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.number,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),

            // Season
            TextField(
              controller: form.seasonCtrl,
              readOnly: true,
              decoration: InputDecoration(
                labelText: 'Mùa gặt',
                border: const OutlineInputBorder(),
                suffixIcon: IconButton(
                  icon: const Icon(Icons.calendar_today, size: 20),
                  onPressed: () => _pickDate(form),
                ),
              ),
              onTap: () => _pickDate(form),
            ),
            const SizedBox(height: 12),

            // Description
            TextField(
              controller: form.descCtrl,
              decoration: const InputDecoration(
                labelText: 'Mô tả thêm',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
              minLines: 1,
            ),
            const SizedBox(height: 14),

            // Images
            Text('Hình ảnh (tối đa 3)', style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 10,
              runSpacing: 10,
              children: [
                ...form.localPaths.asMap().entries.map((entry) {
                  return Stack(
                    clipBehavior: Clip.none,
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(10),
                        child: Image.file(
                          File(entry.value),
                          width: 72,
                          height: 72,
                          fit: BoxFit.cover,
                        ),
                      ),
                      Positioned(
                        top: -6,
                        right: -6,
                        child: GestureDetector(
                          onTap: () => _removeImage(form, entry.key),
                          child: Container(
                            decoration: const BoxDecoration(color: AppColors.error, shape: BoxShape.circle),
                            padding: const EdgeInsets.all(3),
                            child: const Icon(Icons.close, size: 14, color: Colors.white),
                          ),
                        ),
                      ),
                    ],
                  );
                }),
                if (form.imageUrls.length < 3)
                  GestureDetector(
                    onTap: form.uploading ? null : () => _pickImage(form),
                    child: Container(
                      width: 72,
                      height: 72,
                      decoration: BoxDecoration(
                        border: Border.all(color: AppColors.border, width: 1.5),
                        borderRadius: BorderRadius.circular(10),
                        color: AppColors.surfaceVariant.withValues(alpha: 0.3),
                      ),
                      child: form.uploading
                          ? const Center(child: SizedBox(width: 22, height: 22, child: CircularProgressIndicator(strokeWidth: 2)))
                          : Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.add_a_photo_outlined, size: 22, color: AppColors.textHint),
                                const SizedBox(height: 2),
                                Text('${form.imageUrls.length}/3', style: TextStyle(fontSize: 11, color: AppColors.textHint)),
                              ],
                            ),
                    ),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

// ---- Product Picker Bottom Sheet ----

class _ProductPickerSheet extends StatefulWidget {
  final List<RiceCategory> catalog;
  final RiceProduct? current;

  const _ProductPickerSheet({required this.catalog, this.current});

  @override
  State<_ProductPickerSheet> createState() => _ProductPickerSheetState();
}

class _ProductPickerSheetState extends State<_ProductPickerSheet> {
  final _searchCtrl = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  static String _normalize(String str) {
    const withDiacritics = 'àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ';
    const withoutDiacritics = 'aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd';
    var result = str.toLowerCase();
    for (int i = 0; i < withDiacritics.length; i++) {
      result = result.replaceAll(withDiacritics[i], withoutDiacritics[i]);
    }
    return result;
  }

  @override
  Widget build(BuildContext context) {
    final normalizedQuery = _normalize(_query);

    // Filter categories and products
    final filteredCategories = <RiceCategory>[];
    for (final cat in widget.catalog) {
      if (_query.isEmpty) {
        filteredCategories.add(cat);
      } else {
        final matchedProducts = cat.products
            .where((p) => _normalize(p.label).contains(normalizedQuery))
            .toList();
        if (matchedProducts.isNotEmpty) {
          filteredCategories.add(RiceCategory(key: cat.key, label: cat.label, products: matchedProducts));
        }
      }
    }

    return DraggableScrollableSheet(
      initialChildSize: 0.75,
      minChildSize: 0.5,
      maxChildSize: 0.95,
      expand: false,
      builder: (context, scrollController) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: Row(
              children: [
                const Expanded(
                  child: Text('Chọn sản phẩm', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
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
                hintText: 'Tìm sản phẩm...',
                prefixIcon: const Icon(Icons.search),
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                contentPadding: const EdgeInsets.symmetric(horizontal: 12),
              ),
              onChanged: (v) => setState(() => _query = v.trim()),
            ),
          ),
          const SizedBox(height: 8),
          Expanded(
            child: filteredCategories.isEmpty
                ? Center(child: Text('Không tìm thấy sản phẩm', style: TextStyle(color: AppColors.textHint)))
                : ListView.builder(
                    controller: scrollController,
                    itemCount: filteredCategories.length,
                    itemBuilder: (context, catIndex) {
                      final cat = filteredCategories[catIndex];
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                            child: Text(
                              cat.label,
                              style: TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                                color: AppColors.textSecondary,
                                letterSpacing: 0.3,
                              ),
                            ),
                          ),
                          ...cat.products.map((product) {
                            final isSelected = widget.current?.key == product.key;
                            return ListTile(
                              title: Text(product.label),
                              trailing: isSelected
                                  ? const Icon(Icons.check, color: AppColors.primary)
                                  : null,
                              selected: isSelected,
                              contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                              onTap: () => Navigator.pop(context, (cat: cat, product: product)),
                            );
                          }),
                          if (catIndex < filteredCategories.length - 1)
                            const Divider(indent: 16, endIndent: 16),
                        ],
                      );
                    },
                  ),
          ),
        ],
      ),
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
