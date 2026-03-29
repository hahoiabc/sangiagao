import 'dart:convert';
import 'dart:io';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_callkit_incoming/flutter_callkit_incoming.dart';
import 'package:flutter_callkit_incoming/entities/entities.dart';
import 'api_service.dart';
import 'call_service.dart' show isInCall;

/// Top-level handler required by Firebase for background messages.
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();

  // Show incoming call UI even when app is in background
  if (message.data['type'] == 'incoming_call') {
    final callerName = message.data['caller_name'] ?? 'Người gọi';
    final callType = message.data['call_type'] ?? 'audio';
    final convId = message.data['conversation_id'] ?? '';
    final callId = message.data['call_id'] ?? convId;

    final params = CallKitParams(
      id: callId,
      nameCaller: callerName,
      appName: 'SanGiaGao',
      type: callType == 'video' ? 1 : 0,
      duration: 60000,
      android: const AndroidParams(
        isShowLogo: false,
        ringtonePath: 'system_ringtone_default',
        backgroundColor: '#1a1a2e',
        actionColor: '#4CAF50',
      ),
      ios: const IOSParams(
        iconName: 'AppIcon',
        ringtonePath: 'system_ringtone_default',
      ),
      extra: {
        'conversation_id': convId,
        'call_type': callType,
      },
    );

    FlutterCallkitIncoming.showCallkitIncoming(params);
  }
}

/// Global navigator key — set from main.dart to enable navigation from push
typedef PushNavigateCallback = void Function(String route);

/// Callback to show in-app incoming call overlay (set from app-level)
typedef IncomingCallOverlayCallback = void Function({
  required String callerName,
  required String callType,
  required String conversationId,
  required String callId,
});

class PushNotificationService {
  final ApiService _api;
  final FlutterLocalNotificationsPlugin _localNotifications =
      FlutterLocalNotificationsPlugin();

  static PushNavigateCallback? onNavigate;

  /// Set this to show a custom in-app incoming call overlay instead of CallKit when foreground
  static IncomingCallOverlayCallback? onIncomingCallOverlay;

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

    // Create both notification channels
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
    // Handle incoming call push (data-only, no notification payload)
    if (message.data['type'] == 'incoming_call') {
      if (isInCall) {
        // Already in a call — send busy (handled by signaling, skip CallKit)
        return;
      }
      _showIncomingCall(message.data);
      return;
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

    // #1 + #2 + #3: Group by convId, replace notification (same ID per conv),
    // suppress sound on 2nd+ message
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
        // #4: iOS threadIdentifier for per-conversation grouping
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
    } else if (type == 'incoming_call' && conversationId != null) {
      onNavigate!('/chat/$conversationId');
    } else {
      onNavigate!('/notifications');
    }
  }

  /// Show incoming call UI — prefer in-app overlay if set, else use CallKit
  void _showIncomingCall(Map<String, dynamic> data) {
    final callerName = data['caller_name'] ?? 'Người gọi';
    final callType = data['call_type'] ?? 'audio';
    final convId = data['conversation_id'] ?? '';
    final callId = data['call_id'] ?? convId;

    // Use in-app overlay if available (foreground + callback registered)
    if (onIncomingCallOverlay != null) {
      onIncomingCallOverlay!(
        callerName: callerName,
        callType: callType,
        conversationId: convId,
        callId: callId,
      );
      return;
    }

    final params = CallKitParams(
      id: callId,
      nameCaller: callerName,
      appName: 'SanGiaGao',
      type: callType == 'video' ? 1 : 0,
      duration: 60000, // 60s ring timeout
      android: const AndroidParams(
        isShowLogo: false,
        ringtonePath: 'system_ringtone_default',
        backgroundColor: '#1a1a2e',
        actionColor: '#4CAF50',
      ),
      ios: const IOSParams(
        iconName: 'AppIcon',
        ringtonePath: 'system_ringtone_default',
      ),
      extra: {
        'conversation_id': convId,
        'call_type': callType,
      },
    );

    FlutterCallkitIncoming.showCallkitIncoming(params);
  }

  /// Initialize CallKit event listeners — call from main.dart after push init
  static void initCallKitListeners() {
    FlutterCallkitIncoming.onEvent.listen((CallEvent? event) {
      if (event == null) return;
      final extra = event.body['extra'] as Map<String, dynamic>? ?? {};
      final convId = extra['conversation_id'] as String?;

      switch (event.event) {
        case Event.actionCallAccept:
          if (convId != null && onNavigate != null) {
            onNavigate!('/chat/$convId?call=accept');
          }
          break;
        case Event.actionCallDecline:
        case Event.actionCallEnded:
        case Event.actionCallTimeout:
          // Call ended/declined — no navigation needed
          break;
        default:
          break;
      }
    });
  }
}
