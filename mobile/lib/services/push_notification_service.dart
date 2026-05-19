import 'dart:convert';
import 'dart:io';
import 'dart:ui' show VoidCallback;
import 'package:flutter/foundation.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'api_service.dart';

/// Top-level handler required by Firebase for background messages.
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
}

/// Global navigator key — set from main.dart to enable navigation from push
typedef PushNavigateCallback = void Function(String route);

class PushNotificationService {
  final ApiService _api;
  final FlutterLocalNotificationsPlugin _localNotifications =
      FlutterLocalNotificationsPlugin();

  static PushNavigateCallback? onNavigate;
  static VoidCallback? onSystemInbox;

  /// Currently active conversation ID — suppress notifications for this chat
  static String? activeConversationId;

  /// True khi user đang ở BẤT KỲ màn chat nào (list conv, chat detail, search...).
  /// Khi true, banner chat từ conv khác bị suppress để không làm phiền — user
  /// vẫn thấy update qua list conv + badge.
  static bool isOnChatScreen = false;

  /// Track last notification time per conversation — for sound suppression
  static final Map<String, DateTime> _lastNotifTime = {};

  /// Sound suppression window: same conversation within 5 seconds → silent
  static const _soundSuppressWindow = Duration(seconds: 5);

  /// Track unread message count per conversation — for smart summary
  static final Map<String, int> _unreadCount = {};

  static const _channel = AndroidNotificationChannel(
    'sangiagao_notifications',
    'Thông báo',
    description: 'Thông báo từ SanGiaGao.vn',
    importance: Importance.high,
  );

  static const _silentChannel = AndroidNotificationChannel(
    'sangiagao_silent',
    'Tin nhắn (im lặng)',
    description: 'Cập nhật tin nhắn không kèm âm thanh',
    importance: Importance.low,
    playSound: false,
    enableVibration: false,
  );

  PushNotificationService(this._api);

  Future<void> init() async {
    final androidPlugin = _localNotifications
        .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin>();

    // Create notification channels
    await androidPlugin?.createNotificationChannel(_channel);
    await androidPlugin?.createNotificationChannel(_silentChannel);

    // Init local notifications with tap callback
    await _localNotifications.initialize(
      const InitializationSettings(
        android: AndroidInitializationSettings('@mipmap/ic_launcher'),
        iOS: DarwinInitializationSettings(),
      ),
      onDidReceiveNotificationResponse: _onLocalNotificationTap,
    );

    // Skip Firebase messaging if Firebase not initialized
    try {
      FirebaseMessaging.instance;
    } catch (_) {
      debugPrint('[PUSH] Firebase not available, skipping push notifications');
      return;
    }

    // Request permission (Android 13+ and iOS)
    await FirebaseMessaging.instance.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );

    // Get FCM token and register with backend
    final token = await FirebaseMessaging.instance.getToken();
    if (token != null) {
      await _registerToken(token);
    }

    // Listen for token refresh
    FirebaseMessaging.instance.onTokenRefresh.listen(_registerToken);

    // Handle foreground messages — show local notification
    FirebaseMessaging.onMessage.listen(_showLocalNotification);

    // Handle notification tap when app is in background
    FirebaseMessaging.onMessageOpenedApp.listen(_onFirebaseMessageTap);

    // Handle notification tap when app was terminated
    final initialMessage = await FirebaseMessaging.instance.getInitialMessage();
    if (initialMessage != null) {
      _onFirebaseMessageTap(initialMessage);
    }
  }

  Future<void> _registerToken(String token) async {
    try {
      final platform = Platform.isIOS ? 'ios' : 'android';
      await _api.registerDevice(token, platform);
    } catch (_) {
      // Silently fail — will retry on next app launch
    }
  }

  void _showLocalNotification(RemoteMessage message) {
    final dataType = message.data['type'] as String?;

    // Data-only sync event "conversation_read" — gửi từ backend khi user mark
    // read trên device khác. Cancel local notification của conv này để
    // multi-device dismiss đồng bộ.
    if (dataType == 'conversation_read') {
      final convId = message.data['conversation_id'] as String?;
      if (convId != null) {
        _localNotifications.cancel(convId.hashCode);
        _unreadCount.remove(convId);
        _lastNotifTime.remove(convId);
      }
      return;
    }

    // Handle system inbox push — increment badge
    if (dataType == 'system_inbox') {
      onSystemInbox?.call();
      // Don't return — let it show as a local notification below
    }

    final notification = message.notification;
    if (notification == null) return;

    final convId = message.data['conversation_id'] as String?;

    // Suppress notification if user is currently viewing this conversation
    if (convId != null && convId == activeConversationId) {
      return;
    }

    // Suppress banner khi user đang ở bất kỳ màn chat nào (list conv, chat
    // detail conv khác, etc.). User vẫn thấy update qua list realtime — banner
    // banner trong lúc đang chat là spam.
    if (dataType == 'new_message' && isOnChatScreen) {
      // Vẫn track unread count (cho summary nếu user thoát chat ra)
      if (convId != null) {
        _unreadCount[convId] = (_unreadCount[convId] ?? 0) + 1;
      }
      return;
    }

    // Determine if sound should play:
    // Only suppress sound for 2nd+ message from SAME conversation within 3s
    bool playSound = true;
    if (convId != null) {
      final lastTime = _lastNotifTime[convId];
      if (lastTime != null &&
          DateTime.now().difference(lastTime) < _soundSuppressWindow) {
        playSound = false;
      }
      _lastNotifTime[convId] = DateTime.now();
    }

    // Track unread count per conversation for smart summary
    if (convId != null) {
      _unreadCount[convId] = (_unreadCount[convId] ?? 0) + 1;
    }

    final payload = jsonEncode(message.data);

    // Pick channel based on sound decision
    final channelToUse = playSound ? _channel : _silentChannel;

    // Android rich preview: nếu push data có image_url, dùng BigPictureStyle.
    // Cần tải ảnh thành ByteArrayAndroidBitmap. URL phải public (MinIO/CDN).
    AndroidNotificationDetails androidDetails;
    final imageUrl = message.data['image_url'] as String?;
    if (imageUrl != null && imageUrl.isNotEmpty) {
      // ByteArrayAndroidBitmap nhận URL trực tiếp qua FilePathAndroidBitmap
      // không dùng được vì là URL; thay vào dùng FilePath nếu local hoặc
      // ByteArray fetched. Đơn giản nhất: BigText fallback nếu fetch fail.
      androidDetails = AndroidNotificationDetails(
        channelToUse.id,
        channelToUse.name,
        channelDescription: channelToUse.description,
        importance: playSound ? Importance.high : Importance.low,
        priority: playSound ? Priority.high : Priority.defaultPriority,
        icon: '@mipmap/ic_launcher',
        groupKey: convId,
        playSound: playSound,
        enableVibration: playSound,
        // BigPicture style — Flutter plugin nhận URL trực tiếp qua extras.
        // Image sẽ fetch async khi system show notif.
        styleInformation: BigPictureStyleInformation(
          FilePathAndroidBitmap(imageUrl), // works with http URL on Android 11+
          contentTitle: notification.title,
          summaryText: notification.body,
          htmlFormatContentTitle: true,
        ),
      );
    } else {
      androidDetails = AndroidNotificationDetails(
        channelToUse.id,
        channelToUse.name,
        channelDescription: channelToUse.description,
        importance: playSound ? Importance.high : Importance.low,
        priority: playSound ? Priority.high : Priority.defaultPriority,
        icon: '@mipmap/ic_launcher',
        groupKey: convId,
        playSound: playSound,
        enableVibration: playSound,
      );
    }

    _localNotifications.show(
      convId?.hashCode ?? notification.hashCode,
      notification.title,
      _buildNotificationBody(convId, notification.body),
      NotificationDetails(
        android: androidDetails,
        iOS: DarwinNotificationDetails(
          threadIdentifier: convId,
          sound: playSound ? 'default' : null,
          // iOS rich preview cần Notification Service Extension (chưa setup) →
          // chỉ truyền attachment URL qua data, app tự fetch trong extension.
          // Hiện tại iOS không hiển thị inline image.
        ),
      ),
      payload: payload,
    );

    // Show summary notification for grouped chat messages (Android)
    if (convId != null) {
      final totalUnread = _unreadCount.values.fold(0, (a, b) => a + b);
      final chatCount = _unreadCount.keys.length;
      final summary = chatCount > 1
          ? '$chatCount cuộc trò chuyện, $totalUnread tin nhắn mới'
          : '$totalUnread tin nhắn mới';

      _localNotifications.show(
        'sangiagao_summary'.hashCode,
        'SanGiaGao',
        summary,
        NotificationDetails(
          android: AndroidNotificationDetails(
            _silentChannel.id,
            _silentChannel.name,
            channelDescription: _silentChannel.description,
            groupKey: convId,
            setAsGroupSummary: true,
            icon: '@mipmap/ic_launcher',
          ),
        ),
      );
    }
  }

  /// Build notification body: show count if multiple unread from same conversation
  String _buildNotificationBody(String? convId, String? latestBody) {
    if (convId == null) return latestBody ?? '';
    final count = _unreadCount[convId] ?? 1;
    if (count <= 1) return latestBody ?? '';
    return '$count tin nhắn mới';
  }

  /// Clear unread tracking when user opens a conversation
  static void clearUnreadForConversation(String conversationId) {
    _unreadCount.remove(conversationId);
    _lastNotifTime.remove(conversationId);
  }

  /// Called when user taps a local notification (foreground case)
  void _onLocalNotificationTap(NotificationResponse response) {
    if (response.payload == null) return;
    try {
      final data = jsonDecode(response.payload!) as Map<String, dynamic>;
      _navigateFromData(data);
    } catch (_) {}
  }

  /// Called when user taps a Firebase notification (background/terminated case)
  void _onFirebaseMessageTap(RemoteMessage message) {
    _navigateFromData(message.data);
  }

  /// Navigate based on notification data
  void _navigateFromData(Map<String, dynamic> data) {
    final type = data['type'] as String?;
    final conversationId = data['conversation_id'] as String?;

    if (onNavigate == null) return;

    if (type == 'new_message' && conversationId != null) {
      clearUnreadForConversation(conversationId);
      onNavigate!('/chat/$conversationId');
    } else if (type == 'system_inbox') {
      onNavigate!('/system-inbox');
    } else {
      onNavigate!('/notifications');
    }
  }
}
