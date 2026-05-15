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
  final String? bumpedAt;
  final int bumpCount;
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
    this.bumpedAt,
    this.bumpCount = 0,
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
        bumpedAt: json['bumped_at'] as String?,
        bumpCount: json['bump_count'] as int? ?? 0,
        createdAt: json['created_at'] as String,
      );

  Listing copyWith({String? bumpedAt, int? bumpCount}) => Listing(
        id: id,
        userId: userId,
        title: title,
        category: category,
        riceType: riceType,
        province: province,
        ward: ward,
        quantityKg: quantityKg,
        pricePerKg: pricePerKg,
        harvestSeason: harvestSeason,
        description: description,
        certifications: certifications,
        images: images,
        status: status,
        viewCount: viewCount,
        bumpedAt: bumpedAt ?? this.bumpedAt,
        bumpCount: bumpCount ?? this.bumpCount,
        createdAt: createdAt,
      );

  bool get isActive => status == 'active';
}

/// Số phút phải đợi giữa 2 lần "Làm mới tin đăng".
const int bumpCooldownMinutes = 354; // 5h54m
/// Tổng số lần "Làm mới" tối đa cho 1 tin (60 ngày × 4/ngày).
const int bumpLifetimeCap = 240;

class ListingDetail {
  final Listing listing;
  final PublicProfile seller;

  ListingDetail({required this.listing, required this.seller});

  factory ListingDetail.fromJson(Map<String, dynamic> json) => ListingDetail(
        listing: Listing.fromJson(json),
        seller: PublicProfile.fromJson(json['seller'] as Map<String, dynamic>),
      );
}
