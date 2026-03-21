class User {
  final String id;
  final String phone;
  final String role;
  final String? name;
  final String? avatarUrl;
  final String? province;
  final String? ward;
  final String? address;
  final String? description;
  final String? orgName;
  final bool isBlocked;
  final String? acceptedTosAt;
  final String createdAt;

  User({
    required this.id,
    required this.phone,
    required this.role,
    this.name,
    this.avatarUrl,
    this.province,
    this.ward,
    this.address,
    this.description,
    this.orgName,
    this.isBlocked = false,
    this.acceptedTosAt,
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] as String,
        phone: json['phone'] as String,
        role: json['role'] as String,
        name: json['name'] as String?,
        avatarUrl: json['avatar_url'] as String?,
        province: json['province'] as String?,
        ward: json['ward'] as String?,
        address: json['address'] as String?,
        description: json['description'] as String?,
        orgName: json['org_name'] as String?,
        isBlocked: json['is_blocked'] as bool? ?? false,
        acceptedTosAt: json['accepted_tos_at'] as String?,
        createdAt: json['created_at'] as String,
      );

  bool get hasAcceptedTos => acceptedTosAt != null;
  bool get isMember => role == 'member';
}

class PublicProfile {
  final String id;
  final String phone;
  final String role;
  final String? name;
  final String? avatarUrl;
  final String? province;
  final String? ward;
  final String? description;
  final String? orgName;
  final bool isOnline;
  final String createdAt;

  PublicProfile({
    required this.id,
    required this.phone,
    required this.role,
    this.name,
    this.avatarUrl,
    this.province,
    this.ward,
    this.description,
    this.orgName,
    this.isOnline = false,
    required this.createdAt,
  });

  factory PublicProfile.fromJson(Map<String, dynamic> json) => PublicProfile(
        id: json['id'] as String,
        phone: json['phone'] as String? ?? '',
        role: json['role'] as String,
        name: json['name'] as String?,
        avatarUrl: json['avatar_url'] as String?,
        province: json['province'] as String?,
        ward: json['ward'] as String?,
        description: json['description'] as String?,
        orgName: json['org_name'] as String?,
        isOnline: json['is_online'] as bool? ?? false,
        createdAt: json['created_at'] as String,
      );

  /// Build location string from province/ward
  String? get location {
    final parts = <String>[];
    if (ward != null && ward!.isNotEmpty) parts.add(ward!);
    if (province != null && province!.isNotEmpty) parts.add(province!);
    return parts.isEmpty ? null : parts.join(', ');
  }
}
