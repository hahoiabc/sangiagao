import 'package:flutter_test/flutter_test.dart';
import 'package:rice_marketplace/providers/providers.dart';
import 'package:rice_marketplace/models/user.dart';

void main() {
  group('AuthState', () {
    test('default state has unknown status and null user', () {
      const state = AuthState();
      expect(state.status, AuthStatus.unknown);
      expect(state.user, isNull);
    });

    test('copyWith updates status', () {
      const state = AuthState();
      final updated = state.copyWith(status: AuthStatus.authenticated);
      expect(updated.status, AuthStatus.authenticated);
      expect(updated.user, isNull);
    });

    test('copyWith updates user', () {
      const state = AuthState(status: AuthStatus.authenticated);
      final user = User(
        id: 'u1', phone: '0901234567', role: 'buyer',
        createdAt: '2024-01-01T00:00:00Z',
      );
      final updated = state.copyWith(user: user);
      expect(updated.user, isNotNull);
      expect(updated.user!.id, 'u1');
      expect(updated.status, AuthStatus.authenticated);
    });

    test('copyWith preserves existing values when not specified', () {
      final user = User(
        id: 'u1', phone: '0901234567', role: 'seller',
        name: 'Nguyen Van A', createdAt: '2024-01-01T00:00:00Z',
      );
      final state = AuthState(status: AuthStatus.authenticated, user: user);
      final updated = state.copyWith(status: AuthStatus.unauthenticated);
      expect(updated.status, AuthStatus.unauthenticated);
      expect(updated.user!.id, 'u1'); // user preserved
    });
  });

  group('AuthStatus', () {
    test('has all expected values', () {
      expect(AuthStatus.values.length, 3);
      expect(AuthStatus.values, contains(AuthStatus.unknown));
      expect(AuthStatus.values, contains(AuthStatus.authenticated));
      expect(AuthStatus.values, contains(AuthStatus.unauthenticated));
    });
  });
}
