import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../models/listing.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class EditListingScreen extends ConsumerStatefulWidget {
  final String listingId;
  const EditListingScreen({super.key, required this.listingId});

  @override
  ConsumerState<EditListingScreen> createState() => _EditListingScreenState();
}

class _EditListingScreenState extends ConsumerState<EditListingScreen> {
  bool _loading = true;
  bool _submitting = false;
  Listing? _listing;

  final _priceCtrl = TextEditingController();
  final _quantityCtrl = TextEditingController();
  final _seasonCtrl = TextEditingController();
  final _descCtrl = TextEditingController();


  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final detail = await ref.read(apiServiceProvider).getListingDetail(widget.listingId);
      final l = detail.listing;
      if (mounted) {
        setState(() {
          _listing = l;
          _priceCtrl.text = l.pricePerKg.toStringAsFixed(0);
          _quantityCtrl.text = l.quantityKg.toStringAsFixed(0);
          _seasonCtrl.text = l.harvestSeason ?? '';
          _descCtrl.text = l.description ?? '';
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi tải tin đăng: $e')),
        );
      }
    }
  }

  Future<void> _submit() async {
    final price = double.tryParse(_priceCtrl.text.trim());
    final qty = double.tryParse(_quantityCtrl.text.trim());
    if (price == null || qty == null || price <= 0 || qty <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Giá và số lượng phải lớn hơn 0')),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final data = <String, dynamic>{
        'price_per_kg': price,
        'quantity_kg': qty,
      };
      final season = _seasonCtrl.text.trim();
      if (season.isNotEmpty) data['harvest_season'] = season;
      final desc = _descCtrl.text.trim();
      if (desc.isNotEmpty) data['description'] = desc;

      await ref.read(apiServiceProvider).updateListing(widget.listingId, data);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đã cập nhật tin đăng')),
        );
        context.pop(true);
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
    _priceCtrl.dispose();
    _quantityCtrl.dispose();
    _seasonCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Sửa tin đăng')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _listing == null
              ? const Center(child: Text('Không tìm thấy tin đăng'))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Listing title (read-only)
                      Text(
                        _listing!.title,
                        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
                      ),
                      if (_listing!.riceType != null)
                        Padding(
                          padding: const EdgeInsets.only(top: 4),
                          child: Text(
                            _listing!.riceType!,
                            style: TextStyle(fontSize: 14, color: AppColors.textSecondary),
                          ),
                        ),
                      const SizedBox(height: 20),

                      // Price
                      TextField(
                        controller: _priceCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Giá (đ/kg)',
                          border: OutlineInputBorder(),
                        ),
                        keyboardType: TextInputType.number,
                      ),
                      const SizedBox(height: 12),

                      // Quantity
                      TextField(
                        controller: _quantityCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Số lượng (kg)',
                          border: OutlineInputBorder(),
                        ),
                        keyboardType: TextInputType.number,
                      ),
                      const SizedBox(height: 12),

                      // Harvest season
                      TextField(
                        controller: _seasonCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Mùa gặt',
                          border: OutlineInputBorder(),
                        ),
                      ),
                      const SizedBox(height: 12),

                      // Description
                      TextField(
                        controller: _descCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Mô tả thêm',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 3,
                      ),
                      const SizedBox(height: 24),

                      // Submit button
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton(
                          onPressed: _submitting ? null : _submit,
                          child: _submitting
                              ? const SizedBox(
                                  height: 20,
                                  width: 20,
                                  child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.surface),
                                )
                              : const Text('Lưu thay đổi'),
                        ),
                      ),
                    ],
                  ),
                ),
    );
  }
}
