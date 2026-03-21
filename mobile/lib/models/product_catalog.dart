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
  final List<RiceProduct> products;

  const RiceCategory({required this.key, required this.label, required this.products});

  factory RiceCategory.fromJson(Map<String, dynamic> json) => RiceCategory(
        key: json['key'] as String,
        label: json['label'] as String,
        products: (json['products'] as List<dynamic>)
            .map((e) => RiceProduct.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}
