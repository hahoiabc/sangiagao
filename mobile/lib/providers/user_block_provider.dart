import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/api_service.dart';
import 'providers.dart';

/// Cache of user IDs that the current user has blocked.
/// Filtering uses this set client-side to hide listings/chats from blocked users.
/// Apple Guideline 1.2 requires "remove from user's feed instantly".
class UserBlockNotifier extends StateNotifier<Set<String>> {
  UserBlockNotifier(this._api) : super(const {}) {
    refresh();
  }

  final ApiService _api;

  Future<void> refresh() async {
    try {
      final list = await _api.listBlocks();
      state = list
          .map((m) => m['blocked_id'] as String?)
          .whereType<String>()
          .toSet();
    } catch (_) {
      // Keep last known state on failure.
    }
  }

  Future<void> block(String userId, {String? reason}) async {
    await _api.blockUser(userId, reason: reason);
    state = {...state, userId};
  }

  Future<void> unblock(String userId) async {
    await _api.unblockUser(userId);
    final next = {...state}..remove(userId);
    state = next;
  }

  bool isBlocked(String userId) => state.contains(userId);
}

final userBlockProvider =
    StateNotifierProvider<UserBlockNotifier, Set<String>>((ref) {
  return UserBlockNotifier(ref.read(apiServiceProvider));
});
