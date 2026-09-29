class PriceBoardEntry {
  final String productKey;
  final String productLabel;
  final double? minPrice;
  final int listingCount;
  final String? sponsorLogo;
  final String? imageUrl; // ảnh tin rẻ nhất của loại này

  const PriceBoardEntry({
    required this.productKey,
    required this.productLabel,
    this.minPrice,
    this.listingCount = 0,
    this.sponsorLogo,
    this.imageUrl,
  });

  factory PriceBoardEntry.fromJson(Map<String, dynamic> json) => PriceBoardEntry(
        productKey: json['product_key'] as String,
        productLabel: json['product_label'] as String,
        minPrice: json['min_price'] != null ? (json['min_price'] as num).toDouble() : null,
        listingCount: json['listing_count'] as int? ?? 0,
        sponsorLogo: json['sponsor_logo'] as String?,
        imageUrl: json['image_url'] as String?,
      );
}

class PriceBoardCategory {
  final String categoryKey;
  final String categoryLabel;
  final List<PriceBoardEntry> products;

  const PriceBoardCategory({
    required this.categoryKey,
    required this.categoryLabel,
    required this.products,
  });

  factory PriceBoardCategory.fromJson(Map<String, dynamic> json) => PriceBoardCategory(
        categoryKey: json['category_key'] as String,
        categoryLabel: json['category_label'] as String,
        products: (json['products'] as List<dynamic>)
            .map((e) => PriceBoardEntry.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class PriceBoardResponse {
  final List<PriceBoardCategory> categories;
  final String updatedAt;

  const PriceBoardResponse({required this.categories, required this.updatedAt});

  factory PriceBoardResponse.fromJson(Map<String, dynamic> json) => PriceBoardResponse(
        categories: (json['categories'] as List<dynamic>)
            .map((e) => PriceBoardCategory.fromJson(e as Map<String, dynamic>))
            .toList(),
        updatedAt: json['updated_at'] as String,
      );
}
