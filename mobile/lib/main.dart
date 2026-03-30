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

  /// Accept call directly — bypass GoRouter/ChatScreen entirely
  Future<void> _acceptCallDirect({
    required String callerName,
    required String conversationId,
    required String callId,
  }) async {
    final ctx = _navKey.currentContext;
    if (ctx == null) return;

    final api = ref.read(apiServiceProvider);
    final token = await api.getToken();
    final user = ref.read(authProvider).user;

    debugPrint('CallService: _acceptCallDirect token=${token != null}, user=${user?.id}');
    if (token == null || user == null) {
      // Not logged in — just navigate to chat as fallback
      ref.read(routerProvider).go('/chat/$conversationId');
      return;
    }

    // Find other user ID from conversation
    String? otherUserId;
    String otherUserName = callerName;
    try {
      final result = await api.getConversations(limit: 50);
      final conv = result.data.where((c) => c.id == conversationId).firstOrNull;
      otherUserId = conv?.otherUser?.id;
      otherUserName = conv?.otherUser?.name ?? callerName;
    } catch (e) {
      debugPrint('CallService: Failed to load conversation: $e');
    }

    if (otherUserId == null) {
      debugPrint('CallService: could not find otherUserId, fallback to chat');
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

    await callService.start();

    if (callService.state == CallState.ended) {
      debugPrint('CallService: start() failed, navigating to chat');
      ref.read(routerProvider).go('/chat/$conversationId');
      return;
    }

    callService.acceptCall();

    if (ctx == _navKey.currentContext) {
      Navigator.of(ctx).push(MaterialPageRoute(
        builder: (_) => ActiveCallScreen(callService: callService),
      ));
    }
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
              // Start call directly — no GoRouter, no ChatScreen dependency
              _acceptCallDirect(
                callerName: callerName,
                conversationId: conversationId,
                callId: callId,
              );
            },
            onReject: () {
              Navigator.of(ctx).pop();
            },
          ),
        ),
      );
    };

    // Wire CallKit accept (background case) — also bypass GoRouter
    PushNotificationService.onAcceptCall = ({
      required String callerName,
      required String conversationId,
      required String callId,
    }) {
      _acceptCallDirect(
        callerName: callerName,
        conversationId: conversationId,
        callId: callId,
      );
    };

    // Init push notifications + CallKit once authenticated
    ref.listen<AuthState>(authProvider, (prev, next) {
      if (next.status == AuthStatus.authenticated && !_pushInitialized) {
        _pushInitialized = true;
        final api = ref.read(apiServiceProvider);
        PushNotificationService(api).init();
        PushNotificationService.initCallKitListeners();
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
