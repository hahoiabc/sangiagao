import 'dart:convert';
import 'dart:io';
import 'dart:ui' show VoidCallback;
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

  /// Track last notification time per conversation — for sound suppression
  static final Map<String, DateTime> _lastNotifTime = {};

  /// Sound suppression window: same conversation within 3 seconds → silent
  static const _soundSuppressWindow = Duration(seconds: 3);

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
    // Handle system inbox push — increment badge
    if (message.data['type'] == 'system_inbox') {
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

    _localNotifications.show(
      convId?.hashCode ?? notification.hashCode,
      notification.title,
      _buildNotificationBody(convId, notification.body),
      NotificationDetails(
        android: AndroidNotificationDetails(
          channelToUse.id,
          channelToUse.name,
          channelDescription: channelToUse.description,
          importance: playSound ? Importance.high : Importance.low,
          priority: playSound ? Priority.high : Priority.defaultPriority,
          icon: '@mipmap/ic_launcher',
          groupKey: convId,
          playSound: playSound,
          enableVibration: playSound,
        ),
        iOS: DarwinNotificationDetails(
          threadIdentifier: convId,
          sound: playSound ? 'default' : null,
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
