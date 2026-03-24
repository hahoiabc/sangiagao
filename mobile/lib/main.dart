import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'routes/router.dart';
import 'theme/app_theme.dart';
import 'providers/theme_provider.dart';

void main() {
  runApp(const ProviderScope(child: SanGaoApp()));
}

class SanGaoApp extends ConsumerWidget {
  const SanGaoApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    final themeOption = ref.watch(themeProvider);

    return MaterialApp.router(
      title: 'SanGiaGao.Vn',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.withPrimary(themeOption.primary, themeOption.primaryDark, themeOption.primaryLight),
      routerConfig: router,
    );
  }
}
