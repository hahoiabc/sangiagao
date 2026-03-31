import 'user.dart';

class Conversation {
  final String id;
  final String memberId;
  final String sellerId;
  final String? listingId;
  final String lastMessageAt;
  final String createdAt;
  final PublicProfile? otherUser;
  final int unreadCount;

  Conversation({
    required this.id,
    required this.memberId,
    required this.sellerId,
    this.listingId,
    required this.lastMessageAt,
    required this.createdAt,
    this.otherUser,
    this.unreadCount = 0,
  });

  factory Conversation.fromJson(Map<String, dynamic> json) => Conversation(
        id: json['id'] as String,
        memberId: json['member_id'] as String,
        sellerId: json['seller_id'] as String,
        listingId: json['listing_id'] as String?,
        lastMessageAt: json['last_message_at'] as String,
        createdAt: json['created_at'] as String,
        otherUser: json['other_user'] != null
            ? PublicProfile.fromJson(json['other_user'] as Map<String, dynamic>)
            : null,
        unreadCount: json['unread_count'] as int? ?? 0,
      );
}

class Message {
  final String id;
  final String conversationId;
  final String senderId;
  final String content;
  final String type;
  final String? readAt;
  final String? replyToId;
  final ReplyMessage? replyTo;
  final List<MessageReaction> reactions;
  final String createdAt;

  Message({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.content,
    this.type = 'text',
    this.readAt,
    this.replyToId,
    this.replyTo,
    this.reactions = const [],
    required this.createdAt,
  });

  factory Message.fromJson(Map<String, dynamic> json) => Message(
        id: json['id'] as String,
        conversationId: json['conversation_id'] as String,
        senderId: json['sender_id'] as String,
        content: json['content'] as String,
        type: json['type'] as String? ?? 'text',
        readAt: json['read_at'] as String?,
        replyToId: json['reply_to_id'] as String?,
        replyTo: json['reply_to'] != null
            ? ReplyMessage.fromJson(json['reply_to'] as Map<String, dynamic>)
            : null,
        reactions: (json['reactions'] as List<dynamic>?)
                ?.map((r) => MessageReaction.fromJson(r as Map<String, dynamic>))
                .toList() ??
            [],
        createdAt: json['created_at'] as String,
      );

  bool get isRead => readAt != null;
}

class ReplyMessage {
  final String id;
  final String senderId;
  final String content;
  final String type;

  ReplyMessage({
    required this.id,
    required this.senderId,
    required this.content,
    required this.type,
  });

  factory ReplyMessage.fromJson(Map<String, dynamic> json) => ReplyMessage(
        id: json['id'] as String,
        senderId: json['sender_id'] as String,
        content: json['content'] as String,
        type: json['type'] as String? ?? 'text',
      );
}

class MessageReaction {
  final String id;
  final String messageId;
  final String userId;
  final String emoji;

  MessageReaction({
    required this.id,
    required this.messageId,
    required this.userId,
    required this.emoji,
  });

  factory MessageReaction.fromJson(Map<String, dynamic> json) => MessageReaction(
        id: json['id'] as String,
        messageId: json['message_id'] as String,
        userId: json['user_id'] as String,
        emoji: json['emoji'] as String,
      );
}
