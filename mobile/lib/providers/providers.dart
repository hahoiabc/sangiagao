import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/api_service.dart';
import '../models/user.dart';

// API Service singleton
final apiServiceProvider = Provider<ApiService>((ref) => ApiService());

// Auth state
enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthState {
  final AuthStatus status;
  final User? user;

  const AuthState({this.status = AuthStatus.unknown, this.user});

  AuthState copyWith({AuthStatus? status, User? user}) =>
      AuthState(status: status ?? this.status, user: user ?? this.user);
}

class AuthNotifier extends StateNotifier<AuthState> {
  final ApiService _api;

  AuthNotifier(this._api) : super(const AuthState()) {
    _checkAuth();
  }

  /// Whether the last auth check found a blocked account.
  bool wasBlocked = false;

  Future<void> _checkAuth() async {
    final token = await _api.getToken();
    if (token != null) {
      try {
        final user = await _api.getMe();
        if (user.isBlocked) {
          wasBlocked = true;
          await _api.logout();
          state = const AuthState(status: AuthStatus.unauthenticated);
          return;
        }
        wasBlocked = false;
        state = AuthState(status: AuthStatus.authenticated, user: user);
      } catch (_) {
        state = const AuthState(status: AuthStatus.unauthenticated);
      }
    } else {
      state = const AuthState(status: AuthStatus.unauthenticated);
    }
  }

  Future<void> sendOTP(String phone) async {
    await _api.sendOTP(phone);
  }

  Future<void> verifyOTP(String phone, String code) async {
    final data = await _api.verifyOTP(phone, code);
    final user = User.fromJson(data['user']);
    state = AuthState(status: AuthStatus.authenticated, user: user);
  }

  Future<void> register(String phone) async {
    await _api.register(phone);
  }

  Future<void> completeRegister({
    required String phone,
    required String code,
    required String name,
    required String password,
    String? province,
    String? ward,
    String? address,
  }) async {
    final data = await _api.completeRegister(
      phone: phone,
      code: code,
      name: name,
      password: password,
      province: province,
      ward: ward,
      address: address,
    );
    final user = User.fromJson(data['user']);
    state = AuthState(status: AuthStatus.authenticated, user: user);
  }

  Future<void> resetPassword(String phone, String code, String newPassword) async {
    await _api.resetPassword(phone, code, newPassword);
  }

  Future<void> loginPassword(String phone, String password) async {
    final data = await _api.loginPassword(phone, password);
    final user = User.fromJson(data['user']);
    state = AuthState(status: AuthStatus.authenticated, user: user);
  }

  Future<void> updateProfile(Map<String, dynamic> data) async {
    final user = await _api.updateProfile(data);
    state = state.copyWith(user: user);
  }

  Future<void> uploadAvatar(String filePath) async {
    final user = await _api.uploadAvatar(filePath);
    state = state.copyWith(user: user);
  }

  Future<void> refreshUser() async {
    final user = await _api.getMe();
    state = state.copyWith(user: user);
  }

  Future<void> deleteAccount() async {
    await _api.deleteAccount();
    state = const AuthState(status: AuthStatus.unauthenticated);
  }

  Future<void> logout() async {
    await _api.logout();
    state = const AuthState(status: AuthStatus.unauthenticated);
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>(
  (ref) => AuthNotifier(ref.read(apiServiceProvider)),
);

// Unread message count for inbox badge
class UnreadCountNotifier extends StateNotifier<int> {
  final ApiService _api;

  UnreadCountNotifier(this._api) : super(0);

  Future<void> refresh() async {
    try {
      state = await _api.getUnreadTotal();
    } catch (_) {
      // ignore
    }
  }
}

final unreadCountProvider = StateNotifierProvider<UnreadCountNotifier, int>(
  (ref) => UnreadCountNotifier(ref.read(apiServiceProvider)),
);

// System Inbox unread count for badge
class InboxUnreadNotifier extends StateNotifier<int> {
  final ApiService _api;

  InboxUnreadNotifier(this._api) : super(0);

  Future<void> refresh() async {
    try {
      state = await _api.getInboxUnreadCount();
    } catch (_) {
      // ignore
    }
  }

  void increment() => state++;
}

final inboxUnreadProvider = StateNotifierProvider<InboxUnreadNotifier, int>(
  (ref) => InboxUnreadNotifier(ref.read(apiServiceProvider)),
);

// Permission state
class PermissionNotifier extends StateNotifier<Map<String, bool>> {
  final ApiService _api;
  DateTime? _lastLoad;
  static const _cacheTTL = Duration(minutes: 10);

  PermissionNotifier(this._api) : super({});

  Future<void> load({bool force = false}) async {
    if (!force && _lastLoad != null && DateTime.now().difference(_lastLoad!) < _cacheTTL) {
      return; // Cache still valid
    }
    try {
      final perms = await _api.getMyPermissions();
      state = perms;
      _lastLoad = DateTime.now();
    } catch (_) {
      // Keep existing state on error
    }
  }

  Future<void> loadGuest({bool force = false}) async {
    if (!force && _lastLoad != null && DateTime.now().difference(_lastLoad!) < _cacheTTL) {
      return;
    }
    try {
      final perms = await _api.getGuestPermissions();
      state = perms;
      _lastLoad = DateTime.now();
    } catch (_) {}
  }

  bool hasPermission(String key) => state[key] == true;

  void clear() {
    state = {};
    _lastLoad = null;
  }
}

final permissionProvider = StateNotifierProvider<PermissionNotifier, Map<String, bool>>(
  (ref) => PermissionNotifier(ref.read(apiServiceProvider)),
);
