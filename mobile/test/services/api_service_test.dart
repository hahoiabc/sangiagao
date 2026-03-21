import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/services/api_service.dart';
import 'package:rice_marketplace/models/listing.dart';
import 'package:rice_marketplace/models/conversation.dart';
import 'package:rice_marketplace/models/rating.dart';

void main() {
  group('PaginatedResult', () {
    test('fromJson parses Listing list correctly', () {
      final json = {
        'data': [
          {
            'id': 'l1',
            'user_id': 'u1',
            'title': 'Gao ST25',
            'quantity_kg': 1000.0,
            'price_per_kg': 25000.0,
            'status': 'active',
            'created_at': '2024-01-15T10:00:00Z',
          },
          {
            'id': 'l2',
            'user_id': 'u2',
            'title': 'Gao Jasmine',
            'quantity_kg': 500.0,
            'price_per_kg': 18000.0,
            'status': 'active',
            'created_at': '2024-01-16T10:00:00Z',
          },
        ],
        'total': 10,
        'page': 1,
        'limit': 20,
      };

      final result = PaginatedResult.fromJson(json, (j) => Listing.fromJson(j));

      expect(result.data.length, 2);
      expect(result.data[0].title, 'Gao ST25');
      expect(result.data[1].title, 'Gao Jasmine');
      expect(result.total, 10);
      expect(result.page, 1);
      expect(result.limit, 20);
    });

    test('fromJson parses empty list', () {
      final json = {
        'data': [],
        'total': 0,
        'page': 1,
        'limit': 20,
      };

      final result = PaginatedResult.fromJson(json, (j) => Listing.fromJson(j));

      expect(result.data, isEmpty);
      expect(result.total, 0);
    });

    test('hasMore returns true when data.length < total', () {
      final result = PaginatedResult<String>(
        data: ['a', 'b'],
        total: 10,
        page: 1,
        limit: 2,
      );
      expect(result.hasMore, true);
    });

    test('hasMore returns false when all data loaded', () {
      final result = PaginatedResult<String>(
        data: ['a', 'b', 'c'],
        total: 3,
        page: 1,
        limit: 20,
      );
      expect(result.hasMore, false);
    });

    test('fromJson parses Conversation list', () {
      final json = {
        'data': [
          {
            'id': 'c1',
            'buyer_id': 'u1',
            'seller_id': 'u2',
            'last_message_at': '2024-03-01T12:00:00Z',
            'created_at': '2024-03-01T10:00:00Z',
          },
        ],
        'total': 1,
        'page': 1,
        'limit': 20,
      };

      final result = PaginatedResult.fromJson(json, (j) => Conversation.fromJson(j));

      expect(result.data.length, 1);
      expect(result.data[0].id, 'c1');
    });

    test('fromJson parses Message list', () {
      final json = {
        'data': [
          {
            'id': 'm1',
            'conversation_id': 'c1',
            'sender_id': 'u1',
            'content': 'Xin chao',
            'created_at': '2024-03-01T12:00:00Z',
          },
        ],
        'total': 1,
        'page': 1,
        'limit': 30,
      };

      final result = PaginatedResult.fromJson(json, (j) => Message.fromJson(j));

      expect(result.data.length, 1);
      expect(result.data[0].content, 'Xin chao');
    });

    test('fromJson parses Rating list', () {
      final json = {
        'data': [
          {
            'id': 'r1',
            'reviewer_id': 'u1',
            'seller_id': 'u2',
            'stars': 5,
            'created_at': '2024-03-01T10:00:00Z',
          },
        ],
        'total': 1,
        'page': 1,
        'limit': 20,
      };

      final result = PaginatedResult.fromJson(json, (j) => Rating.fromJson(j));

      expect(result.data.length, 1);
      expect(result.data[0].stars, 5);
    });

    test('fromJson parses AppNotification list', () {
      final json = {
        'data': [
          {
            'id': 'n1',
            'user_id': 'u1',
            'type': 'message',
            'title': 'Tin nhan moi',
            'body': 'Noi dung',
            'created_at': '2024-03-01T10:00:00Z',
          },
        ],
        'total': 5,
        'page': 1,
        'limit': 20,
      };

      final result = PaginatedResult.fromJson(json, (j) => AppNotification.fromJson(j));

      expect(result.data.length, 1);
      expect(result.data[0].title, 'Tin nhan moi');
      expect(result.total, 5);
      expect(result.hasMore, true);
    });
  });
}
