class Rating {
  final String id;
  final String reviewerId;
  final String sellerId;
  final int stars;
  final String? comment;
  final String createdAt;

  Rating({
    required this.id,
    required this.reviewerId,
    required this.sellerId,
    required this.stars,
    this.comment,
    required this.createdAt,
  });

  factory Rating.fromJson(Map<String, dynamic> json) => Rating(
        id: json['id'] as String,
        reviewerId: json['reviewer_id'] as String,
        sellerId: json['seller_id'] as String,
        stars: json['stars'] as int,
        comment: json['comment'] as String?,
        createdAt: json['created_at'] as String,
      );
}

class RatingSummary {
  final double average;
  final int count;

  RatingSummary({required this.average, required this.count});

  factory RatingSummary.fromJson(Map<String, dynamic> json) => RatingSummary(
        average: (json['average'] as num).toDouble(),
        count: json['count'] as int,
      );
}

class AppNotification {
  final String id;
  final String userId;
  final String type;
  final String title;
  final String body;
  final bool isRead;
  final String createdAt;

  AppNotification({
    required this.id,
    required this.userId,
    required this.type,
    required this.title,
    required this.body,
    this.isRead = false,
    required this.createdAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) =>
      AppNotification(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        type: json['type'] as String,
        title: json['title'] as String,
        body: json['body'] as String,
        isRead: json['is_read'] as bool? ?? false,
        createdAt: json['created_at'] as String,
      );
}
