import 'package:flutter/material.dart';
import '../models/location.dart';
import '../services/location_service.dart';
import '../theme/app_theme.dart';

class LocationPicker extends StatefulWidget {
  final String? initialProvince;
  final String? initialWard;
  final void Function(String? province, String? ward) onChanged;

  const LocationPicker({
    super.key,
    this.initialProvince,
    this.initialWard,
    required this.onChanged,
  });

  @override
  State<LocationPicker> createState() => _LocationPickerState();
}

class _LocationPickerState extends State<LocationPicker> {
  final _locationService = LocationService();

  List<Province> _provinces = [];
  List<Ward> _wards = [];

  Province? _selectedProvince;
  Ward? _selectedWard;

  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final provinces = await _locationService.getProvinces();
      if (!mounted) return;
      setState(() {
        _provinces = provinces;
        _loading = false;
      });

      if (widget.initialProvince != null && widget.initialProvince!.isNotEmpty) {
        final match = provinces.where((p) => p.name == widget.initialProvince).firstOrNull;
        if (match != null) {
          setState(() => _selectedProvince = match);
          await _loadWards(match.code);
        }
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _loadWards(String provinceCode) async {
    final wards = await _locationService.getWards(provinceCode);
    if (!mounted) return;
    setState(() {
      _wards = wards;
      _selectedWard = null;
    });

    if (widget.initialWard != null && widget.initialWard!.isNotEmpty) {
      final match = wards.where((w) => w.name == widget.initialWard).firstOrNull;
      if (match != null) {
        setState(() => _selectedWard = match);
      }
    }
  }

  void _notifyChanged() {
    widget.onChanged(_selectedProvince?.name, _selectedWard?.name);
  }

  Future<void> _showProvinceSearch() async {
    final result = await showModalBottomSheet<Province>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _SearchSheet<Province>(
        title: 'Chọn Tỉnh/Thành phố',
        items: _provinces,
        getName: (p) => p.name,
        selected: _selectedProvince,
      ),
    );
    if (result != null && result != _selectedProvince) {
      setState(() {
        _selectedProvince = result;
        _selectedWard = null;
        _wards = [];
      });
      _notifyChanged();
      await _loadWards(result.code);
    }
  }

  Future<void> _showWardSearch() async {
    if (_selectedProvince == null) return;
    final result = await showModalBottomSheet<Ward>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _SearchSheet<Ward>(
        title: 'Chọn Phường/Xã',
        items: _wards,
        getName: (w) => w.name,
        selected: _selectedWard,
      ),
    );
    if (result != null) {
      setState(() => _selectedWard = result);
      _notifyChanged();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 16),
        child: Center(child: CircularProgressIndicator()),
      );
    }

    return Column(
      children: [
        _buildSelector(
          icon: Icons.location_city,
          label: 'Tỉnh/Thành phố',
          value: _selectedProvince?.name,
          onTap: _showProvinceSearch,
        ),
        const SizedBox(height: 12),
        _buildSelector(
          icon: Icons.location_on,
          label: 'Phường/Xã',
          value: _selectedWard?.name,
          onTap: _selectedProvince != null ? _showWardSearch : null,
          enabled: _selectedProvince != null,
        ),
      ],
    );
  }

  Widget _buildSelector({
    required IconData icon,
    required String label,
    String? value,
    VoidCallback? onTap,
    bool enabled = true,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(4),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
          prefixIcon: Icon(icon),
          suffixIcon: const Icon(Icons.arrow_drop_down),
          enabled: enabled,
        ),
        child: Text(
          value ?? '',
          style: TextStyle(
            fontSize: 16,
            color: value != null ? null : AppColors.textHint,
          ),
        ),
      ),
    );
  }
}

class _SearchSheet<T> extends StatefulWidget {
  final String title;
  final List<T> items;
  final String Function(T) getName;
  final T? selected;

  const _SearchSheet({
    required this.title,
    required this.items,
    required this.getName,
    this.selected,
  });

  @override
  State<_SearchSheet<T>> createState() => _SearchSheetState<T>();
}

class _SearchSheetState<T> extends State<_SearchSheet<T>> {
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
    const withDiacritics =    'àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ';
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
                    child: Text(
                      widget.title,
                      style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.pop(context),
                  ),
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
                  suffixIcon: _searchCtrl.text.isNotEmpty
                      ? IconButton(
                          icon: const Icon(Icons.clear),
                          onPressed: () {
                            _searchCtrl.clear();
                            _filter('');
                          },
                        )
                      : null,
                ),
                onChanged: _filter,
              ),
            ),
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  '${_filtered.length} kết quả',
                  style: const TextStyle(color: AppColors.textSecondary, fontSize: 13),
                ),
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
