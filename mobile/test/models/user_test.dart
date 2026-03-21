import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/models/user.dart';

void main() {
  group('User', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'u1',
        'phone': '0901234567',
        'role': 'seller',
        'name': 'Nguyen Van A',
        'avatar_url': 'https://example.com/avatar.jpg',
        'province': 'An Giang',
        'description': 'Mo ta',
        'org_name': 'Cong ty ABC',
        'is_blocked': false,
        'accepted_tos_at': '2024-01-01T00:00:00Z',
        'created_at': '2024-01-01T00:00:00Z',
      };

      final user = User.fromJson(json);

      expect(user.id, 'u1');
      expect(user.phone, '0901234567');
      expect(user.role, 'seller');
      expect(user.name, 'Nguyen Van A');
      expect(user.avatarUrl, 'https://example.com/avatar.jpg');
      expect(user.province, 'An Giang');
      expect(user.description, 'Mo ta');
      expect(user.orgName, 'Cong ty ABC');
      expect(user.isBlocked, false);
      expect(user.acceptedTosAt, '2024-01-01T00:00:00Z');
      expect(user.createdAt, '2024-01-01T00:00:00Z');
    });

    test('fromJson handles nullable fields as null', () {
      final json = {
        'id': 'u2',
        'phone': '0909999999',
        'role': 'buyer',
        'created_at': '2024-06-01T00:00:00Z',
      };

      final user = User.fromJson(json);

      expect(user.name, isNull);
      expect(user.avatarUrl, isNull);
      expect(user.province, isNull);
      expect(user.description, isNull);
      expect(user.orgName, isNull);
      expect(user.acceptedTosAt, isNull);
      expect(user.isBlocked, false); // default
    });

    test('hasAcceptedTos returns true when acceptedTosAt is not null', () {
      final user = User(
        id: 'u1', phone: '0901234567', role: 'buyer',
        acceptedTosAt: '2024-01-01T00:00:00Z', createdAt: '2024-01-01T00:00:00Z',
      );
      expect(user.hasAcceptedTos, true);
    });

    test('hasAcceptedTos returns false when acceptedTosAt is null', () {
      final user = User(
        id: 'u1', phone: '0901234567', role: 'buyer',
        createdAt: '2024-01-01T00:00:00Z',
      );
      expect(user.hasAcceptedTos, false);
    });

    test('isSeller returns true for seller role', () {
      final user = User(
        id: 'u1', phone: '0901234567', role: 'seller',
        createdAt: '2024-01-01T00:00:00Z',
      );
      expect(user.isSeller, true);
      expect(user.isBuyer, false);
    });

    test('isBuyer returns true for buyer role', () {
      final user = User(
        id: 'u1', phone: '0901234567', role: 'buyer',
        createdAt: '2024-01-01T00:00:00Z',
      );
      expect(user.isBuyer, true);
      expect(user.isSeller, false);
    });

    test('is_blocked defaults to false when missing', () {
      final json = {
        'id': 'u1',
        'phone': '0901234567',
        'role': 'buyer',
        'created_at': '2024-01-01T00:00:00Z',
      };
      final user = User.fromJson(json);
      expect(user.isBlocked, false);
    });

    test('is_blocked parses true', () {
      final json = {
        'id': 'u1',
        'phone': '0901234567',
        'role': 'buyer',
        'is_blocked': true,
        'created_at': '2024-01-01T00:00:00Z',
      };
      final user = User.fromJson(json);
      expect(user.isBlocked, true);
    });
  });

  group('PublicProfile', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'u1',
        'role': 'seller',
        'name': 'Nguyen Van B',
        'avatar_url': 'https://example.com/av.jpg',
        'province': 'Can Tho',
        'description': 'Gioi thieu',
        'org_name': 'HTX Gao',
        'created_at': '2024-03-01T00:00:00Z',
      };

      final profile = PublicProfile.fromJson(json);

      expect(profile.id, 'u1');
      expect(profile.role, 'seller');
      expect(profile.name, 'Nguyen Van B');
      expect(profile.avatarUrl, 'https://example.com/av.jpg');
      expect(profile.province, 'Can Tho');
      expect(profile.description, 'Gioi thieu');
      expect(profile.orgName, 'HTX Gao');
      expect(profile.createdAt, '2024-03-01T00:00:00Z');
    });

    test('fromJson handles nullable fields', () {
      final json = {
        'id': 'u2',
        'role': 'buyer',
        'created_at': '2024-03-01T00:00:00Z',
      };

      final profile = PublicProfile.fromJson(json);

      expect(profile.name, isNull);
      expect(profile.avatarUrl, isNull);
      expect(profile.province, isNull);
      expect(profile.description, isNull);
      expect(profile.orgName, isNull);
    });
  });
}
