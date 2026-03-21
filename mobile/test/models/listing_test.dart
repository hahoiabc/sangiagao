import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/models/listing.dart';

void main() {
  group('Listing', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'l1',
        'user_id': 'u1',
        'title': 'Gao ST25',
        'rice_type': 'ST25',
        'province': 'Soc Trang',
        'district': 'Thanh Tri',  // backend still sends 'district' key
        'quantity_kg': 1000.0,
        'price_per_kg': 25000.0,
        'description': 'Gao ngon nhat the gioi',
        'certifications': 'VietGAP',
        'images': ['img1.jpg', 'img2.jpg'],
        'status': 'active',
        'view_count': 42,
        'created_at': '2024-01-15T10:00:00Z',
      };

      final listing = Listing.fromJson(json);

      expect(listing.id, 'l1');
      expect(listing.userId, 'u1');
      expect(listing.title, 'Gao ST25');
      expect(listing.riceType, 'ST25');
      expect(listing.province, 'Soc Trang');
      expect(listing.ward, 'Thanh Tri');
      expect(listing.quantityKg, 1000.0);
      expect(listing.pricePerKg, 25000.0);
      expect(listing.description, 'Gao ngon nhat the gioi');
      expect(listing.certifications, 'VietGAP');
      expect(listing.images, ['img1.jpg', 'img2.jpg']);
      expect(listing.status, 'active');
      expect(listing.viewCount, 42);
      expect(listing.createdAt, '2024-01-15T10:00:00Z');
    });

    test('fromJson handles nullable fields and defaults', () {
      final json = {
        'id': 'l2',
        'user_id': 'u2',
        'title': 'Gao Jasmine',
        'quantity_kg': 500,
        'price_per_kg': 18000,
        'status': 'hidden_subscription',
        'created_at': '2024-02-01T00:00:00Z',
      };

      final listing = Listing.fromJson(json);

      expect(listing.riceType, isNull);
      expect(listing.province, isNull);
      expect(listing.ward, isNull);
      expect(listing.description, isNull);
      expect(listing.certifications, isNull);
      expect(listing.images, isEmpty);
      expect(listing.viewCount, 0);
    });

    test('fromJson handles int quantities as double', () {
      final json = {
        'id': 'l3',
        'user_id': 'u1',
        'title': 'Gao nep',
        'quantity_kg': 200,
        'price_per_kg': 30000,
        'status': 'active',
        'created_at': '2024-02-01T00:00:00Z',
      };

      final listing = Listing.fromJson(json);
      expect(listing.quantityKg, 200.0);
      expect(listing.pricePerKg, 30000.0);
    });

    test('isActive returns true for active status', () {
      final listing = Listing(
        id: 'l1', userId: 'u1', title: 'Test',
        quantityKg: 100, pricePerKg: 10000,
        status: 'active', createdAt: '2024-01-01T00:00:00Z',
      );
      expect(listing.isActive, true);
    });

    test('isActive returns false for hidden_subscription status', () {
      final listing = Listing(
        id: 'l1', userId: 'u1', title: 'Test',
        quantityKg: 100, pricePerKg: 10000,
        status: 'hidden_subscription', createdAt: '2024-01-01T00:00:00Z',
      );
      expect(listing.isActive, false);
    });

    test('isActive returns false for deleted status', () {
      final listing = Listing(
        id: 'l1', userId: 'u1', title: 'Test',
        quantityKg: 100, pricePerKg: 10000,
        status: 'deleted', createdAt: '2024-01-01T00:00:00Z',
      );
      expect(listing.isActive, false);
    });
  });

  group('ListingDetail', () {
    test('fromJson parses listing and seller', () {
      final json = {
        'id': 'l1',
        'user_id': 'u1',
        'title': 'Gao ST25',
        'quantity_kg': 500.0,
        'price_per_kg': 25000.0,
        'status': 'active',
        'created_at': '2024-01-15T10:00:00Z',
        'seller': {
          'id': 'u1',
          'role': 'seller',
          'name': 'Nguyen Van A',
          'created_at': '2024-01-01T00:00:00Z',
        },
      };

      final detail = ListingDetail.fromJson(json);

      expect(detail.listing.id, 'l1');
      expect(detail.listing.title, 'Gao ST25');
      expect(detail.seller.id, 'u1');
      expect(detail.seller.name, 'Nguyen Van A');
    });
  });
}
