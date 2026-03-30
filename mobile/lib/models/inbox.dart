class InboxMessage {
  final String id;
  final String title;
  final String body;
  final String? imageUrl;
  final bool isPinned;
  final bool isRead;
  final String createdAt;

  InboxMessage({
    required this.id,
    required this.title,
    required this.body,
    this.imageUrl,
    this.isPinned = false,
    this.isRead = false,
    required this.createdAt,
  });

  factory InboxMessage.fromJson(Map<String, dynamic> json) => InboxMessage(
        id: json['id'] as String,
        title: json['title'] as String,
        body: json['body'] as String,
        imageUrl: json['image_url'] as String?,
        isPinned: json['is_pinned'] as bool? ?? false,
        isRead: json['is_read'] as bool? ?? false,
        createdAt: json['created_at'] as String,
      );
}
