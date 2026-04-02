import 'user.dart';

/// Convert image URL to thumbnail URL by inserting "thumb_" prefix on filename.
/// e.g. ".../listings/uuid.jpg" → ".../listings/thumb_uuid.jpg"
/// Falls back to original URL if format is unexpected.
String toThumbnailUrl(String url) {
  final lastSlash = url.lastIndexOf('/');
  if (lastSlash < 0) return url;
  final base = url.substring(0, lastSlash + 1);
  final filename = url.substring(lastSlash + 1);
  final dotIdx = filename.lastIndexOf('.');
  if (dotIdx < 0) return '${base}thumb_$filename.jpg';
  final name = filename.substring(0, dotIdx);
  return '${base}thumb_$name.jpg';
}

class Listing {
  final String id;
  final String userId;
  final String title;
  final String? category;
  final String? riceType;
  final String? province;
  final String? ward;
  final double quantityKg;
  final double pricePerKg;
  final String? harvestSeason;
  final String? description;
  final String? certifications;
  final List<String> images;
  final String status;
  final int viewCount;
  final String createdAt;

  Listing({
    required this.id,
    required this.userId,
    required this.title,
    this.category,
    this.riceType,
    this.province,
    this.ward,
    required this.quantityKg,
    required this.pricePerKg,
    this.harvestSeason,
    this.description,
    this.certifications,
    this.images = const [],
    required this.status,
    this.viewCount = 0,
    required this.createdAt,
  });

  factory Listing.fromJson(Map<String, dynamic> json) => Listing(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        title: json['title'] as String,
        category: json['category'] as String?,
        riceType: json['rice_type'] as String?,
        province: json['province'] as String?,
        ward: json['ward'] as String?,
        quantityKg: (json['quantity_kg'] as num).toDouble(),
        pricePerKg: (json['price_per_kg'] as num).toDouble(),
        harvestSeason: json['harvest_season'] as String?,
        description: json['description'] as String?,
        certifications: json['certifications'] as String?,
        images: (json['images'] as List<dynamic>?)?.cast<String>() ?? [],
        status: json['status'] as String,
        viewCount: json['view_count'] as int? ?? 0,
        createdAt: json['created_at'] as String,
      );

  bool get isActive => status == 'active';
}

class ListingDetail {
  final Listing listing;
  final PublicProfile seller;

  ListingDetail({required this.listing, required this.seller});

  factory ListingDetail.fromJson(Map<String, dynamic> json) => ListingDetail(
        listing: Listing.fromJson(json),
        seller: PublicProfile.fromJson(json['seller'] as Map<String, dynamic>),
      );
}
