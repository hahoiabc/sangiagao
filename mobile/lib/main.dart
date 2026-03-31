import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'routes/router.dart';
import 'theme/app_theme.dart';
import 'providers/theme_provider.dart';
import 'providers/providers.dart';
import 'services/push_notification_service.dart';

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
