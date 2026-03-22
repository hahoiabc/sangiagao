import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/models/conversation.dart';

void main() {
  group('Conversation', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'c1',
        'member_id': 'u1',
        'seller_id': 'u2',
        'listing_id': 'l1',
        'last_message_at': '2024-03-01T12:00:00Z',
        'created_at': '2024-03-01T10:00:00Z',
        'other_user': {
          'id': 'u2',
          'role': 'seller',
          'name': 'Nguyen Van B',
          'created_at': '2024-01-01T00:00:00Z',
        },
        'unread_count': 3,
      };

      final conv = Conversation.fromJson(json);

      expect(conv.id, 'c1');
      expect(conv.memberId, 'u1');
      expect(conv.sellerId, 'u2');
      expect(conv.listingId, 'l1');
      expect(conv.lastMessageAt, '2024-03-01T12:00:00Z');
      expect(conv.createdAt, '2024-03-01T10:00:00Z');
      expect(conv.otherUser, isNotNull);
      expect(conv.otherUser!.name, 'Nguyen Van B');
      expect(conv.unreadCount, 3);
    });

    test('fromJson handles null otherUser and defaults', () {
      final json = {
        'id': 'c2',
        'member_id': 'u1',
        'seller_id': 'u2',
        'last_message_at': '2024-03-01T12:00:00Z',
        'created_at': '2024-03-01T10:00:00Z',
      };

      final conv = Conversation.fromJson(json);

      expect(conv.listingId, isNull);
      expect(conv.otherUser, isNull);
      expect(conv.unreadCount, 0);
    });
  });

  group('Message', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'm1',
        'conversation_id': 'c1',
        'sender_id': 'u1',
        'content': 'Xin chao',
        'type': 'text',
        'read_at': '2024-03-01T12:05:00Z',
        'created_at': '2024-03-01T12:00:00Z',
      };

      final msg = Message.fromJson(json);

      expect(msg.id, 'm1');
      expect(msg.conversationId, 'c1');
      expect(msg.senderId, 'u1');
      expect(msg.content, 'Xin chao');
      expect(msg.type, 'text');
      expect(msg.readAt, '2024-03-01T12:05:00Z');
      expect(msg.createdAt, '2024-03-01T12:00:00Z');
    });

    test('fromJson defaults type to text and readAt to null', () {
      final json = {
        'id': 'm2',
        'conversation_id': 'c1',
        'sender_id': 'u2',
        'content': 'Hinh anh',
        'created_at': '2024-03-01T12:01:00Z',
      };

      final msg = Message.fromJson(json);

      expect(msg.type, 'text');
      expect(msg.readAt, isNull);
    });

    test('isRead returns true when readAt is not null', () {
      final msg = Message(
        id: 'm1', conversationId: 'c1', senderId: 'u1',
        content: 'Hello', readAt: '2024-03-01T12:05:00Z',
        createdAt: '2024-03-01T12:00:00Z',
      );
      expect(msg.isRead, true);
    });

    test('isRead returns false when readAt is null', () {
      final msg = Message(
        id: 'm1', conversationId: 'c1', senderId: 'u1',
        content: 'Hello', createdAt: '2024-03-01T12:00:00Z',
      );
      expect(msg.isRead, false);
    });

    test('fromJson parses image type', () {
      final json = {
        'id': 'm3',
        'conversation_id': 'c1',
        'sender_id': 'u1',
        'content': 'https://example.com/img.jpg',
        'type': 'image',
        'created_at': '2024-03-01T12:02:00Z',
      };

      final msg = Message.fromJson(json);
      expect(msg.type, 'image');
    });
  });
}
