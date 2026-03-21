import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/models/rating.dart';

void main() {
  group('Rating', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'r1',
        'reviewer_id': 'u1',
        'seller_id': 'u2',
        'stars': 5,
        'comment': 'Gao rat ngon',
        'created_at': '2024-03-01T10:00:00Z',
      };

      final rating = Rating.fromJson(json);

      expect(rating.id, 'r1');
      expect(rating.reviewerId, 'u1');
      expect(rating.sellerId, 'u2');
      expect(rating.stars, 5);
      expect(rating.comment, 'Gao rat ngon');
      expect(rating.createdAt, '2024-03-01T10:00:00Z');
    });

    test('fromJson handles null comment', () {
      final json = {
        'id': 'r2',
        'reviewer_id': 'u3',
        'seller_id': 'u2',
        'stars': 3,
        'comment': null,
        'created_at': '2024-03-02T10:00:00Z',
      };

      final rating = Rating.fromJson(json);
      expect(rating.comment, isNull);
    });
  });

  group('RatingSummary', () {
    test('fromJson parses correctly', () {
      final json = {'average': 4.5, 'count': 12};

      final summary = RatingSummary.fromJson(json);

      expect(summary.average, 4.5);
      expect(summary.count, 12);
    });

    test('fromJson handles int average', () {
      final json = {'average': 4, 'count': 1};

      final summary = RatingSummary.fromJson(json);
      expect(summary.average, 4.0);
    });

    test('fromJson handles zero values', () {
      final json = {'average': 0, 'count': 0};

      final summary = RatingSummary.fromJson(json);
      expect(summary.average, 0.0);
      expect(summary.count, 0);
    });
  });

  group('AppNotification', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'n1',
        'user_id': 'u1',
        'type': 'message',
        'title': 'Tin nhan moi',
        'body': 'Ban co tin nhan moi tu Nguyen Van A',
        'is_read': true,
        'created_at': '2024-03-01T10:00:00Z',
      };

      final notif = AppNotification.fromJson(json);

      expect(notif.id, 'n1');
      expect(notif.userId, 'u1');
      expect(notif.type, 'message');
      expect(notif.title, 'Tin nhan moi');
      expect(notif.body, 'Ban co tin nhan moi tu Nguyen Van A');
      expect(notif.isRead, true);
      expect(notif.createdAt, '2024-03-01T10:00:00Z');
    });

    test('fromJson defaults isRead to false', () {
      final json = {
        'id': 'n2',
        'user_id': 'u1',
        'type': 'subscription',
        'title': 'Goi sap het han',
        'body': 'Goi dich vu cua ban se het han trong 3 ngay',
        'created_at': '2024-03-01T10:00:00Z',
      };

      final notif = AppNotification.fromJson(json);
      expect(notif.isRead, false);
    });

    test('fromJson handles null is_read', () {
      final json = {
        'id': 'n3',
        'user_id': 'u1',
        'type': 'rating',
        'title': 'Danh gia moi',
        'body': 'Ban nhan duoc danh gia moi',
        'is_read': null,
        'created_at': '2024-03-01T10:00:00Z',
      };

      final notif = AppNotification.fromJson(json);
      expect(notif.isRead, false);
    });
  });
}
