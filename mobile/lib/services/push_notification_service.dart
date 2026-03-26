import 'dart:convert';
import 'dart:io';
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

  /// Currently active conversation ID — suppress notifications for this chat
  static String? activeConversationId;

  static const _channel = AndroidNotificationChannel(
    'sangiagao_notifications',
    'Thông báo',
    description: 'Thông báo từ SanGiaGao.vn',
    importance: Importance.high,
  );

  /// Android notification group key for chat messages
  static const _chatGroupKey = 'sangiagao_chat_messages';

  PushNotificationService(this._api);

  Future<void> init() async {
    // Create Android notification channel
    await _localNotifications
        .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin>()
        ?.createNotificationChannel(_channel);

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
    final notification = message.notification;
    if (notification == null) return;

    // Suppress notification if user is currently viewing this conversation
    final convId = message.data['conversation_id'] as String?;
    if (convId != null && convId == activeConversationId) {
      return;
    }

    // Encode data as payload so we can navigate on tap
    final payload = jsonEncode(message.data);

    _localNotifications.show(
      notification.hashCode,
      notification.title,
      notification.body,
      NotificationDetails(
        android: AndroidNotificationDetails(
          _channel.id,
          _channel.name,
          channelDescription: _channel.description,
          importance: Importance.high,
          priority: Priority.high,
          icon: '@mipmap/ic_launcher',
          groupKey: convId != null ? _chatGroupKey : null,
        ),
        iOS: const DarwinNotificationDetails(),
      ),
      payload: payload,
    );

    // Show summary notification for grouped chat messages (Android)
    if (convId != null) {
      _localNotifications.show(
        _chatGroupKey.hashCode,
        'SanGiaGao',
        'Bạn có tin nhắn mới',
        NotificationDetails(
          android: AndroidNotificationDetails(
            _channel.id,
            _channel.name,
            channelDescription: _channel.description,
            groupKey: _chatGroupKey,
            setAsGroupSummary: true,
            icon: '@mipmap/ic_launcher',
          ),
        ),
      );
    }
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
      onNavigate!('/chat/$conversationId');
    } else {
      onNavigate!('/notifications');
    }
  }
}
