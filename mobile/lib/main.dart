import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'routes/router.dart';
import 'theme/app_theme.dart';
import 'providers/theme_provider.dart';
import 'providers/providers.dart';
import 'services/push_notification_service.dart';
import 'services/call_service.dart';
import 'screens/call/incoming_call_screen.dart';
import 'screens/call/active_call_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp();
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);

  // Register CallKit listener EARLY — before runApp — so cold-start accept
  // events are captured into _pendingAccept buffer immediately.
  PushNotificationService.initCallKitListeners();

  runApp(const ProviderScope(child: SanGaoApp()));
}

class SanGaoApp extends ConsumerStatefulWidget {
  const SanGaoApp({super.key});

  @override
  ConsumerState<SanGaoApp> createState() => _SanGaoAppState();
}

class _SanGaoAppState extends ConsumerState<SanGaoApp> {
  bool _pushInitialized = false;
  final GlobalKey<NavigatorState> _navKey = GlobalKey<NavigatorState>();

  /// Flush any pending CallKit accept that was buffered during cold start.
  /// Called ONLY after auth is confirmed ready.
  void _flushPendingAccept() {
    final pending = PushNotificationService.consumePendingAccept();
    if (pending != null) {
      debugPrint('CallKit: flushing pending accept after auth ready');
      _acceptCallDirect(
        callerName: pending['callerName'] ?? 'Người gọi',
        conversationId: pending['conversationId'] ?? '',
        callId: pending['callId'] ?? '',
        callerId: pending['callerId'] ?? '',
      );
      // After pending is consumed, GoRouter redirect will allow splash → marketplace
      // on next rebuild (when ActiveCallScreen pops, user lands on marketplace)
      ref.read(routerProvider).go('/marketplace');
    }
  }

  /// Flush any pending CallKit decline that was buffered during cold start.
  void _flushPendingDecline() {
    final callId = PushNotificationService.consumePendingDecline();
    if (callId != null && callId.isNotEmpty) {
      debugPrint('CallKit: flushing pending decline after auth ready');
      final api = ref.read(apiServiceProvider);
      api.rejectCall(callId).catchError((_) {});
    }
  }

  /// Accept call directly — bypass GoRouter/ChatScreen entirely
  Future<void> _acceptCallDirect({
    required String callerName,
    required String conversationId,
    required String callId,
    String callerId = '',
  }) async {
    final ctx = _navKey.currentContext;
    if (ctx == null) return;

    final api = ref.read(apiServiceProvider);
    String? token = await api.getToken();
    var user = ref.read(authProvider).user;

    // Safety net: if auth not ready yet (cold start), wait up to 8s
    if (token == null || user == null) {
      debugPrint('CallKit: auth not ready, waiting...');
      for (int i = 0; i < 16; i++) {
        await Future.delayed(const Duration(milliseconds: 500));
        token = await api.getToken();
        user = ref.read(authProvider).user;
        if (token != null && user != null) break;
      }
    }

    if (token == null || user == null) {
      debugPrint('CallKit: auth still not ready after wait, fallback to chat');
      ref.read(routerProvider).go('/chat/$conversationId');
      return;
    }

    // Use callerId from push data directly (avoid fetching 50 conversations)
    String otherUserId = callerId;
    String otherUserName = callerName;

    // Fallback: fetch from conversation list if callerId not provided
    if (otherUserId.isEmpty) {
      try {
        final result = await api.getConversations(limit: 50);
        final conv = result.data.where((c) => c.id == conversationId).firstOrNull;
        otherUserId = conv?.otherUser?.id ?? '';
        otherUserName = conv?.otherUser?.name ?? callerName;
      } catch (e) {
        debugPrint('CallService: Failed to load conversation: $e');
      }
    }

    if (otherUserId.isEmpty) {
      ref.read(routerProvider).go('/chat/$conversationId');
      return;
    }

    final callService = CallService(
      api: api,
      token: token,
      conversationId: conversationId,
      currentUserId: user.id,
      otherUserId: otherUserId,
      otherUserName: otherUserName,
      callType: 'audio',
      isInitiator: false,
    );
    if (callId.isNotEmpty) {
      callService.callLogId = callId;
    }

    // Push ActiveCallScreen IMMEDIATELY — show "Đang kết nối..." right away
    // instead of waiting for start()/acceptCall() while user stares at marketplace
    final navCtx = _navKey.currentContext;
    if (navCtx != null) {
      Navigator.of(navCtx).push(MaterialPageRoute(
        builder: (_) => ActiveCallScreen(callService: callService),
      ));
    }

    // Now run start + accept in background — screen updates via onStateChanged
    await callService.start();
    if (callService.state == CallState.ended) return;
    await callService.acceptCall();
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);
    final themeOption = ref.watch(themeProvider);

    // Wire push notification navigation to GoRouter — always read fresh router
    PushNotificationService.onNavigate = (route) {
      ref.read(routerProvider).go(route);
    };

    // Wire in-app incoming call overlay for foreground
    PushNotificationService.onIncomingCallOverlay = ({
      required String callerName,
      required String callType,
      required String conversationId,
      required String callId,
      required String callerId,
    }) {
      final ctx = _navKey.currentContext;
      if (ctx == null) return;
      Navigator.of(ctx).push(
        MaterialPageRoute(
          fullscreenDialog: true,
          builder: (_) => IncomingCallScreen(
            callerName: callerName,
            callType: callType,
            onAccept: () {
              Navigator.of(ctx).pop();
              _acceptCallDirect(
                callerName: callerName,
                conversationId: conversationId,
                callId: callId,
                callerId: callerId,
              );
            },
            onReject: () {
              Navigator.of(ctx).pop();
              // Reject via API so caller (A) gets notified immediately
              if (callId.isNotEmpty) {
                ref.read(apiServiceProvider).rejectCall(callId).catchError((_) {});
              }
            },
          ),
        ),
      );
    };

    // Wire CallKit accept (background case) — bypass GoRouter
    // Note: pending accept is flushed separately in auth listener, NOT here
    PushNotificationService.onAcceptCall = ({
      required String callerName,
      required String conversationId,
      required String callId,
      required String callerId,
    }) {
      _acceptCallDirect(
        callerName: callerName,
        conversationId: conversationId,
        callId: callId,
        callerId: callerId,
      );
    };

    // Wire call rejected (callee busy/rejected via API push)
    PushNotificationService.onCallRejected = () {
      activeCallService?.endCall();
    };

    // Wire CallKit decline — B declined from CallKit UI, reject via API
    PushNotificationService.onDeclineCall = (callId) {
      final api = ref.read(apiServiceProvider);
      api.rejectCall(callId).catchError((_) {});
    };

    // Wire system inbox push — increment badge
    PushNotificationService.onSystemInbox = () {
      ref.read(inboxUnreadProvider.notifier).increment();
    };

    // Init push notifications + CallKit once authenticated
    ref.listen<AuthState>(authProvider, (prev, next) {
      if (next.status == AuthStatus.authenticated && !_pushInitialized) {
        _pushInitialized = true;
        final api = ref.read(apiServiceProvider);
        PushNotificationService(api).init();
        // initCallKitListeners() already called in main() for cold-start support

        // Flush any pending CallKit accept/decline that arrived during cold start
        _flushPendingAccept();
        _flushPendingDecline();
      }
    });

    return MaterialApp.router(
      title: 'SanGiaGao.Vn',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.withPrimary(themeOption.primary, themeOption.primaryDark, themeOption.primaryLight),
      routerConfig: router,
      builder: (context, child) {
        return Navigator(
          key: _navKey,
          onGenerateRoute: (_) => MaterialPageRoute(
            builder: (_) => child ?? const SizedBox.shrink(),
          ),
        );
      },
    );
  }
}
