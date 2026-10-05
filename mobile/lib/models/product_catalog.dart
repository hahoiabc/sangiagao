class RiceProduct {
  final String key;
  final String label;
  final String category;

  const RiceProduct({required this.key, required this.label, required this.category});

  factory RiceProduct.fromJson(Map<String, dynamic> json) => RiceProduct(
        key: json['key'] as String,
        label: json['label'] as String,
        category: json['category'] as String,
      );
}

class RiceCategory {
  final String key;
  final String label;
  final String kind; // nong_san | mat_hang
  final String unit; // kg | cái | km | chiếc...
  final bool aggregatePrice;
  final List<RiceProduct> products;

  const RiceCategory({
    required this.key,
    required this.label,
    this.kind = 'nong_san',
    this.unit = 'kg',
    this.aggregatePrice = true,
    required this.products,
  });

  bool get isItem => kind == 'mat_hang';

  factory RiceCategory.fromJson(Map<String, dynamic> json) => RiceCategory(
        key: json['key'] as String,
        label: json['label'] as String,
        kind: json['kind'] as String? ?? 'nong_san',
        unit: json['unit'] as String? ?? 'kg',
        aggregatePrice: json['aggregate_price'] as bool? ?? true,
        products: ((json['products'] as List<dynamic>?) ?? const [])
            .map((e) => RiceProduct.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}
