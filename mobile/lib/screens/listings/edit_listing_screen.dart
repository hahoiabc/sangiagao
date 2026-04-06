import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';

import '../../models/listing.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/thumbnail_image.dart';

class EditListingScreen extends ConsumerStatefulWidget {
  final String listingId;
  const EditListingScreen({super.key, required this.listingId});

  @override
  ConsumerState<EditListingScreen> createState() => _EditListingScreenState();
}

class _EditListingScreenState extends ConsumerState<EditListingScreen> {
  bool _loading = true;
  bool _submitting = false;
  bool _uploading = false;
  Listing? _listing;
  List<String> _images = [];
  String? _newLocalPath;

  final _priceCtrl = TextEditingController();
  final _quantityCtrl = TextEditingController();
  final _seasonCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _picker = ImagePicker();

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
          _images = List<String>.from(l.images);
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

  Future<void> _removeImage(String url) async {
    try {
      await ref.read(apiServiceProvider).removeListingImage(widget.listingId, url);
      if (mounted) {
        setState(() => _images.remove(url));
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đã xóa ảnh')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi xóa ảnh: $e')),
        );
      }
    }
  }

  Future<void> _addImage() async {
    if (_images.isNotEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tối đa 1 hình ảnh')),
      );
      return;
    }
    final image = await _picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1280,
      maxHeight: 1280,
      imageQuality: 80,
    );
    if (image == null) return;

    setState(() {
      _uploading = true;
      _newLocalPath = image.path;
    });
    try {
      final api = ref.read(apiServiceProvider);
      final url = await api.uploadImagePresigned(image.path, 'listings');
      await api.addListingImage(widget.listingId, url);
      if (mounted) {
        setState(() {
          _images.add(url);
          _newLocalPath = null;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đã thêm ảnh')),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() => _newLocalPath = null);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi tải ảnh: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _submit() async {
    final price = double.tryParse(_priceCtrl.text.trim());
    final qty = double.tryParse(_quantityCtrl.text.trim());
    if (price == null || price <= 5000 || price >= 99000) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Giá phải từ 5,001 đến 98,999 đ/kg')),
      );
      return;
    }
    if (qty == null || qty <= 500 || qty >= 100000000) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Số lượng phải từ 501 đến 99,999,999 kg')),
      );
      return;
    }
    final season = _seasonCtrl.text.trim();
    if (season.isNotEmpty) {
      final parts = season.split('/');
      if (parts.length == 3) {
        final d = int.tryParse(parts[0]) ?? 0;
        final m = int.tryParse(parts[1]) ?? 0;
        final y = int.tryParse(parts[2]) ?? 0;
        final picked = DateTime(y, m, d);
        if (picked.isAfter(DateTime.now())) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Mùa vụ phải trước ngày hiện tại')),
          );
          return;
        }
      }
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

                      // Images section
                      Text('Hình ảnh (tối đa 1)', style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 10,
                        runSpacing: 10,
                        children: [
                          ..._images.map((url) => Stack(
                            clipBehavior: Clip.none,
                            children: [
                              ClipRRect(
                                borderRadius: BorderRadius.circular(10),
                                child: SizedBox(
                                  width: 80,
                                  height: 80,
                                  child: ThumbnailImage(
                                    imageUrl: url,
                                    fit: BoxFit.cover,
                                    errorWidget: (_, __, ___) => Container(
                                      color: AppColors.divider,
                                      child: Icon(Icons.broken_image, color: AppColors.textHint),
                                    ),
                                  ),
                                ),
                              ),
                              Positioned(
                                top: -6,
                                right: -6,
                                child: GestureDetector(
                                  onTap: () => _removeImage(url),
                                  child: Container(
                                    decoration: const BoxDecoration(color: AppColors.error, shape: BoxShape.circle),
                                    padding: const EdgeInsets.all(3),
                                    child: const Icon(Icons.close, size: 14, color: Colors.white),
                                  ),
                                ),
                              ),
                            ],
                          )),
                          if (_newLocalPath != null)
                            ClipRRect(
                              borderRadius: BorderRadius.circular(10),
                              child: SizedBox(
                                width: 80,
                                height: 80,
                                child: Stack(
                                  children: [
                                    Image.file(File(_newLocalPath!), width: 80, height: 80, fit: BoxFit.cover),
                                    Container(
                                      color: Colors.black38,
                                      child: const Center(
                                        child: SizedBox(width: 22, height: 22, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          if (_images.isEmpty && _newLocalPath == null)
                            GestureDetector(
                              onTap: _uploading ? null : _addImage,
                              child: Container(
                                width: 80,
                                height: 80,
                                decoration: BoxDecoration(
                                  border: Border.all(color: AppColors.border, width: 1.5),
                                  borderRadius: BorderRadius.circular(10),
                                  color: AppColors.surfaceVariant.withValues(alpha: 0.3),
                                ),
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(Icons.add_a_photo_outlined, size: 22, color: AppColors.textHint),
                                    const SizedBox(height: 2),
                                    Text('${_images.length}/1', style: TextStyle(fontSize: 11, color: AppColors.textHint)),
                                  ],
                                ),
                              ),
                            ),
                        ],
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
