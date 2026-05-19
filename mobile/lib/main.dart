import 'dart:async';
import 'package:app_links/app_links.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'routes/router.dart';
import 'theme/app_theme.dart';
import 'providers/theme_provider.dart';
import 'providers/providers.dart';
import 'services/push_notification_service.dart';
import 'services/affiliate_attribution_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await Firebase.initializeApp();
    FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
  } catch (e) {
    debugPrint('Firebase init failed: $e');
    // App vẫn chạy, chỉ mất push notification
  }

  // Read Play Install Referrer once after fresh install so the code is
  // captured before Google's referrer service expires it. Best-effort; safe
  // to await briefly. Result is cached in secure storage for later use.
  unawaited(AffiliateAttributionService.getCode());

  runApp(const ProviderScope(child: SanGaoApp()));
}

class SanGaoApp extends ConsumerStatefulWidget {
  const SanGaoApp({super.key});

  @override
  ConsumerState<SanGaoApp> createState() => _SanGaoAppState();
}

class _SanGaoAppState extends ConsumerState<SanGaoApp> with WidgetsBindingObserver {
  bool _pushInitialized = false;
  final _appLinks = AppLinks();
  StreamSubscription<Uri>? _linkSub;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _initDeepLinks();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _linkSub?.cancel();
    super.dispose();
  }

  /// Khi app trở lại foreground sau khoảng thời gian background → refresh
  /// total unread count để đồng bộ tin nhắn miss khi app ngủ. Tránh tình trạng
  /// "mở app phải đợi 1 lúc mới thấy tin mới".
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      final auth = ref.read(authProvider);
      if (auth.status == AuthStatus.authenticated) {
        ref.read(unreadCountProvider.notifier).refresh();
        ref.read(inboxUnreadProvider.notifier).refresh();
      }
    }
  }

  /// Listen for Universal Links (iOS) + App Links (Android). Used to
  /// capture affiliate referrer codes when user clicks /r/{code} or
  /// /cai-app?ref={code} and either opens the app fresh or returns to it.
  Future<void> _initDeepLinks() async {
    // Cold start
    try {
      final initial = await _appLinks.getInitialLink();
      if (initial != null) await AffiliateAttributionService.handleDeepLink(initial);
    } catch (_) {}

    // Warm/background open
    _linkSub = _appLinks.uriLinkStream.listen((uri) {
      AffiliateAttributionService.handleDeepLink(uri);
    });
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);
    final themeOption = ref.watch(themeProvider);

    // Wire push notification navigation to GoRouter
    PushNotificationService.onNavigate = (route) {
      ref.read(routerProvider).go(route);
    };

    // Wire system inbox push — increment badge
    PushNotificationService.onSystemInbox = () {
      ref.read(inboxUnreadProvider.notifier).increment();
    };

    // Init push notifications once authenticated
    ref.listen<AuthState>(authProvider, (prev, next) {
      if (next.status == AuthStatus.authenticated && !_pushInitialized) {
        _pushInitialized = true;
        final api = ref.read(apiServiceProvider);
        PushNotificationService(api).init();
      }
    });

    return MaterialApp.router(
      title: 'SanGiaGao.Vn',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.withPrimary(themeOption.primary, themeOption.primaryDark, themeOption.primaryLight),
      routerConfig: router,
    );
  }
}
