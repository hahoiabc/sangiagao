import 'dart:io';
import 'dart:math';
import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../config/env.dart';
import '../models/user.dart';
import '../models/listing.dart';
import '../models/conversation.dart';
import '../models/rating.dart';
import '../models/inbox.dart';
import '../models/price_board.dart';
import '../models/product_catalog.dart';

// Cached device ID (generated once per install, persisted in secure storage)
String? _cachedDeviceId;

/// Trusted hostnames for TLS certificate validation.
const _trustedHosts = ['sangiagao.vn', 'www.sangiagao.vn'];

Dio _createDio() {
  final dio = Dio(BaseOptions(
    baseUrl: ApiService.baseUrl,
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 15),
    sendTimeout: const Duration(seconds: 15),
  ));

  // In production, reject certificates for untrusted hosts
  if (!kDebugMode) {
    (dio.httpClientAdapter as IOHttpClientAdapter).createHttpClient = () {
      final client = HttpClient();
      client.badCertificateCallback = (X509Certificate cert, String host, int port) {
        return _trustedHosts.contains(host);
      };
      return client;
    };
  }

  return dio;
}

class ApiService {
  static const String baseUrl = Env.apiBaseUrl;

  final Dio _dio;
  final FlutterSecureStorage _storage;

  ApiService({Dio? dio, FlutterSecureStorage? storage})
      : _dio = dio ?? _createDio(),
        _storage = storage ?? const FlutterSecureStorage() {
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storage.read(key: 'access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        // Send device ID for spam protection
        final deviceId = await _getDeviceId();
        options.headers['X-Device-ID'] = deviceId;
        return handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          final refreshed = await _refreshToken();
          if (refreshed) {
            final opts = error.requestOptions;
            final token = await _storage.read(key: 'access_token');
            opts.headers['Authorization'] = 'Bearer $token';
            final response = await _dio.fetch(opts);
            return handler.resolve(response);
          } else {
            // Refresh failed — clear tokens to force re-login
            await _storage.deleteAll();
          }
        }
        return handler.next(error);
      },
    ));
  }

  /// Get or generate a persistent device ID for spam protection
  Future<String> _getDeviceId() async {
    if (_cachedDeviceId != null) return _cachedDeviceId!;
    var id = await _storage.read(key: 'device_id');
    if (id == null) {
      // Generate UUID v4 using secure random
      final rng = Random.secure();
      final bytes = List<int>.generate(16, (_) => rng.nextInt(256));
      bytes[6] = (bytes[6] & 0x0f) | 0x40; // version 4
      bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant 1
      final hex = bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
      id = 'dev_${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}';
      await _storage.write(key: 'device_id', value: id);
    }
    _cachedDeviceId = id;
    return id;
  }

  Future<bool> _refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken == null) return false;

      final res = await _createDio().post(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      await _storage.write(key: 'access_token', value: res.data['access_token']);
      await _storage.write(key: 'refresh_token', value: res.data['refresh_token']);
      return true;
    } catch (_) {
      return false;
    }
  }

  // --- Auth ---
  Future<Map<String, dynamic>> sendOTP(String phone) async {
    final res = await _dio.post('/auth/send-otp', data: {'phone': phone});
    return res.data;
  }

  Future<Map<String, dynamic>> verifyOTP(String phone, String code) async {
    final res = await _dio.post('/auth/verify-otp', data: {'phone': phone, 'code': code});
    final data = res.data;
    await _storage.write(key: 'access_token', value: data['tokens']['access_token']);
    await _storage.write(key: 'refresh_token', value: data['tokens']['refresh_token']);
    return data;
  }

  Future<Map<String, dynamic>> register(String phone) async {
    final res = await _dio.post('/auth/register', data: {'phone': phone});
    return res.data;
  }

  Future<Map<String, dynamic>> completeRegister({
    required String phone,
    required String code,
    required String name,
    required String password,
    String? province,
    String? ward,
    String? address,
  }) async {
    final res = await _dio.post('/auth/complete-register', data: {
      'phone': phone,
      'code': code,
      'name': name,
      'password': password,
      if (province != null) 'province': province,
      if (ward != null) 'ward': ward,
      if (address != null) 'address': address,
    });
    final data = res.data;
    await _storage.write(key: 'access_token', value: data['tokens']['access_token']);
    await _storage.write(key: 'refresh_token', value: data['tokens']['refresh_token']);
    return data;
  }

  Future<Map<String, dynamic>> loginPassword(String phone, String password) async {
    final res = await _dio.post('/auth/login', data: {'phone': phone, 'password': password});
    final data = res.data;
    await _storage.write(key: 'access_token', value: data['tokens']['access_token']);
    await _storage.write(key: 'refresh_token', value: data['tokens']['refresh_token']);
    return data;
  }

  Future<void> resetPassword(String phone, String code, String newPassword) async {
    await _dio.post('/auth/reset-password', data: {
      'phone': phone,
      'code': code,
      'new_password': newPassword,
    });
  }

  Future<void> logout() async {
    await _storage.deleteAll();
  }

  Future<String?> getToken() => _storage.read(key: 'access_token');

  // --- User ---
  Future<User> getMe() async {
    final res = await _dio.get('/users/me');
    return User.fromJson(res.data);
  }

  Future<User> updateProfile(Map<String, dynamic> data) async {
    final res = await _dio.put('/users/me', data: data);
    return User.fromJson(res.data);
  }

  static const _maxImageBytes = 5 * 1024 * 1024; // 5 MB
  static const _maxAudioBytes = 10 * 1024 * 1024; // 10 MB

  Future<String> uploadImage(String filePath, String folder) async {
    final file = File(filePath);
    if (await file.length() > _maxImageBytes) {
      throw Exception('Ảnh không được vượt quá 5 MB');
    }
    final formData = FormData.fromMap({
      'image': await MultipartFile.fromFile(filePath),
      'folder': folder,
    });
    final res = await _dio.post('/upload/image', data: formData);
    return res.data['url'] as String;
  }

  Future<String> uploadAudio(String filePath) async {
    final file = File(filePath);
    if (await file.length() > _maxAudioBytes) {
      throw Exception('File âm thanh không được vượt quá 10 MB');
    }
    final formData = FormData.fromMap({
      'audio': await MultipartFile.fromFile(filePath, contentType: DioMediaType('audio', 'm4a')),
    });
    final res = await _dio.post('/upload/audio', data: formData);
    return res.data['url'] as String;
  }

  Future<User> uploadAvatar(String filePath) async {
    final url = await uploadImage(filePath, 'avatars');
    final res = await _dio.post('/users/me/avatar', data: {'url': url});
    return User.fromJson(res.data);
  }

  Future<PublicProfile> getPublicProfile(String userId) async {
    final res = await _dio.get('/users/$userId/profile');
    return PublicProfile.fromJson(res.data);
  }

  // --- Marketplace ---
  Future<PriceBoardResponse> getPriceBoard() async {
    final res = await _dio.get('/marketplace/price-board');
    return PriceBoardResponse.fromJson(res.data);
  }

  Future<List<RiceCategory>> getProductCatalog() async {
    final res = await _dio.get('/marketplace/product-catalog');
    return (res.data as List<dynamic>)
        .map((e) => RiceCategory.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<PaginatedResult<Listing>> browseMarketplace({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/marketplace', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => Listing.fromJson(j));
  }

  Future<PaginatedResult<Listing>> searchMarketplace({
    String? q,
    String? category,
    String? type,
    String? province,
    String? ward,
    double? minPrice,
    double? maxPrice,
    double? minQty,
    String? sort,
    int page = 1,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{'page': page, 'limit': limit};
    if (q != null && q.isNotEmpty) params['q'] = q;
    if (category != null) params['category'] = category;
    if (type != null) params['type'] = type;
    if (province != null) params['province'] = province;
    if (ward != null) params['ward'] = ward;
    if (minPrice != null) params['min_price'] = minPrice;
    if (maxPrice != null) params['max_price'] = maxPrice;
    if (minQty != null) params['min_qty'] = minQty;
    if (sort != null) params['sort'] = sort;
    final res = await _dio.get('/marketplace/search', queryParameters: params);
    return PaginatedResult.fromJson(res.data, (j) => Listing.fromJson(j));
  }

  Future<ListingDetail> getListingDetail(String id) async {
    final res = await _dio.get('/marketplace/$id');
    return ListingDetail.fromJson(res.data);
  }

  // --- My Listings ---
  Future<PaginatedResult<Listing>> getMyListings({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/listings/my', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => Listing.fromJson(j));
  }

  Future<Listing> createListing(Map<String, dynamic> data) async {
    final res = await _dio.post('/listings', data: data);
    return Listing.fromJson(res.data);
  }

  Future<Map<String, dynamic>> batchCreateListings(List<Map<String, dynamic>> items) async {
    final res = await _dio.post('/listings/batch', data: items);
    return res.data as Map<String, dynamic>;
  }

  Future<Listing> updateListing(String id, Map<String, dynamic> data) async {
    final res = await _dio.put('/listings/$id', data: data);
    return Listing.fromJson(res.data);
  }

  Future<void> deleteListing(String id) async {
    await _dio.delete('/listings/$id');
  }

  Future<Listing> addListingImage(String listingId, String imageUrl) async {
    final res = await _dio.post('/listings/$listingId/images', data: {'url': imageUrl});
    return Listing.fromJson(res.data);
  }

  // --- Conversations ---
  Future<Conversation> createConversation(String sellerId, {String? listingId}) async {
    final data = <String, dynamic>{'seller_id': sellerId};
    if (listingId != null) data['listing_id'] = listingId;
    final res = await _dio.post('/conversations', data: data);
    return Conversation.fromJson(res.data);
  }

  Future<PaginatedResult<Conversation>> getConversations({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/conversations', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => Conversation.fromJson(j));
  }

  Future<void> markConversationRead(String convId) async {
    await _dio.put('/conversations/$convId/read');
  }

  Future<PaginatedResult<Message>> getMessages(String convId, {int page = 1, int limit = 30}) async {
    final res = await _dio.get('/conversations/$convId/messages', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => Message.fromJson(j));
  }

  Future<Message> sendMessage(String convId, String content, {String type = 'text'}) async {
    final res = await _dio.post('/conversations/$convId/messages', data: {'content': content, 'type': type});
    return Message.fromJson(res.data);
  }

  Future<void> deleteMessage(String convId, String msgId) async {
    await _dio.delete('/conversations/$convId/messages/$msgId');
  }

  Future<Message> recallMessage(String convId, String msgId) async {
    final res = await _dio.put('/conversations/$convId/messages/$msgId/recall');
    return Message.fromJson(res.data);
  }

  Future<void> batchDeleteMessages(String convId, List<String> msgIds) async {
    await _dio.post('/conversations/$convId/messages/batch-delete', data: {'message_ids': msgIds});
  }

  Future<void> batchRecallMessages(String convId, List<String> msgIds) async {
    await _dio.post('/conversations/$convId/messages/batch-recall', data: {'message_ids': msgIds});
  }

  // --- Calls ---
  Future<Map<String, dynamic>> getTurnCredentials() async {
    final res = await _dio.get('/calls/turn-credentials');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> initiateCall(String convId, String calleeId, String callType) async {
    final res = await _dio.post('/conversations/$convId/calls', data: {
      'callee_id': calleeId,
      'call_type': callType,
    });
    return res.data as Map<String, dynamic>;
  }

  Future<void> answerCall(String callId) async {
    await _dio.put('/conversations/calls/$callId/answer');
  }

  Future<void> endCallLog(String callId) async {
    await _dio.put('/conversations/calls/$callId/end');
  }

  Future<void> rejectCall(String callId) async {
    await _dio.put('/conversations/calls/$callId/reject');
  }

  Future<void> missCall(String callId) async {
    await _dio.put('/conversations/calls/$callId/miss');
  }

  Future<Map<String, dynamic>> getCallHistory(String convId, {int page = 1, int limit = 20}) async {
    final res = await _dio.get('/conversations/$convId/calls', queryParameters: {
      'page': page,
      'limit': limit,
    });
    return res.data as Map<String, dynamic>;
  }

  // --- Ratings ---
  Future<PaginatedResult<Rating>> getSellerRatings(String sellerId, {int page = 1, int limit = 20}) async {
    final res = await _dio.get('/users/$sellerId/ratings', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => Rating.fromJson(j));
  }

  Future<RatingSummary> getRatingSummary(String sellerId) async {
    final res = await _dio.get('/users/$sellerId/rating-summary');
    return RatingSummary.fromJson(res.data);
  }

  Future<Rating> createRating(String sellerId, int stars, String comment) async {
    final data = <String, dynamic>{
      'seller_id': sellerId,
      'stars': stars,
    };
    if (comment.isNotEmpty) {
      data['comment'] = comment;
    }
    final res = await _dio.post('/ratings', data: data);
    return Rating.fromJson(res.data);
  }

  // --- Notifications ---
  Future<PaginatedResult<AppNotification>> getNotifications({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/notifications', queryParameters: {'page': page, 'limit': limit});
    return PaginatedResult.fromJson(res.data, (j) => AppNotification.fromJson(j));
  }

  Future<void> markNotificationRead(String id) async {
    await _dio.put('/notifications/$id/read');
  }

  Future<void> registerDevice(String token, String platform) async {
    await _dio.post('/notifications/register-device', data: {
      'token': token,
      'platform': platform,
    });
  }

  // --- System Inbox ---
  Future<({List<InboxMessage> items, int total, int unreadCount})> getInbox({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/inbox', queryParameters: {'page': page, 'limit': limit});
    final data = res.data as Map<String, dynamic>;
    final items = (data['data'] as List).map((j) => InboxMessage.fromJson(j as Map<String, dynamic>)).toList();
    return (items: items, total: data['total'] as int, unreadCount: data['unread_count'] as int? ?? 0);
  }

  Future<InboxMessage> getInboxDetail(String id) async {
    final res = await _dio.get('/inbox/$id');
    return InboxMessage.fromJson(res.data as Map<String, dynamic>);
  }

  Future<void> markInboxRead(String id) async {
    await _dio.put('/inbox/$id/read');
  }

  Future<int> getInboxUnreadCount() async {
    final res = await _dio.get('/inbox/unread-count');
    return (res.data as Map<String, dynamic>)['unread_count'] as int? ?? 0;
  }

  // --- Subscription ---
  Future<Map<String, dynamic>> getSubscriptionStatus() async {
    final res = await _dio.get('/subscription/status');
    return res.data;
  }

  Future<Map<String, dynamic>> getSubscriptionPlans() async {
    final res = await _dio.get('/subscription/plans');
    return res.data;
  }

  Future<Map<String, dynamic>> getSubscriptionHistory({int page = 1, int limit = 20}) async {
    final res = await _dio.get('/subscription/history', queryParameters: {'page': page, 'limit': limit});
    return res.data;
  }

  // --- Feedback ---
  Future<void> createFeedback(String content) async {
    await _dio.post('/feedbacks', data: {'content': content});
  }

  Future<List<dynamic>> getMyFeedbacks({int page = 1, int limit = 50}) async {
    final res = await _dio.get('/feedbacks/my', queryParameters: {'page': page, 'limit': limit});
    return res.data['data'] as List<dynamic>;
  }

  // --- Reports ---
  Future<void> createReport(String targetType, String targetId, String reason, {String? description}) async {
    await _dio.post('/reports', data: {
      'target_type': targetType,
      'target_id': targetId,
      'reason': reason,
      if (description != null) 'description': description,
    });
  }

  // --- Account ---
  Future<void> deleteAccount() async {
    await _dio.delete('/users/me');
    await _storage.deleteAll();
  }

  Future<void> changePasswordAuth(String currentPassword, String newPassword) async {
    await _dio.post('/users/me/password', data: {
      'current_password': currentPassword,
      'new_password': newPassword,
    });
  }

  Future<Map<String, dynamic>> changePhoneAuth(String newPhone, String code) async {
    final response = await _dio.post('/users/me/phone', data: {
      'new_phone': newPhone,
      'code': code,
    });
    return response.data as Map<String, dynamic>;
  }

  // --- Permissions ---
  Future<Map<String, bool>> getMyPermissions() async {
    final response = await _dio.get('/permissions/me');
    final perms = response.data['permissions'] as Map<String, dynamic>?;
    if (perms == null) return {};
    return perms.map((key, value) => MapEntry(key, value == true));
  }

  Future<Map<String, bool>> getGuestPermissions() async {
    final response = await _dio.get('/permissions/guest');
    final perms = response.data['permissions'] as Map<String, dynamic>?;
    if (perms == null) return {};
    return perms.map((key, value) => MapEntry(key, value == true));
  }
}

class PaginatedResult<T> {
  final List<T> data;
  final int total;
  final int page;
  final int limit;

  PaginatedResult({required this.data, required this.total, required this.page, required this.limit});

  factory PaginatedResult.fromJson(Map<String, dynamic> json, T Function(Map<String, dynamic>) fromJson) {
    return PaginatedResult(
      data: (json['data'] as List<dynamic>).map((e) => fromJson(e as Map<String, dynamic>)).toList(),
      total: json['total'] as int,
      page: json['page'] as int,
      limit: json['limit'] as int,
    );
  }

  bool get hasMore => data.length < total;
}
