import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/providers.dart';
import '../theme/app_theme.dart';
import 'subscription_gate.dart';

class MainShell extends ConsumerStatefulWidget {
  final Widget child;
  const MainShell({super.key, required this.child});

  @override
  ConsumerState<MainShell> createState() => _MainShellState();
}

class _MainShellState extends ConsumerState<MainShell> {
  int? _daysLeft;
  bool _isActive = false;
  bool _bannerDismissed = false;
  bool _needsSubscription = false;
  bool _subChecked = false;
  Timer? _unreadPollTimer;
  Timer? _permPollTimer;

  @override
  void initState() {
    super.initState();
    final user = ref.read(authProvider).user;
    final isAuth = user != null;

    if (isAuth) {
      _checkSubscription();
      ref.read(unreadCountProvider.notifier).refresh();
      ref.read(permissionProvider.notifier).load();
      _unreadPollTimer = Timer.periodic(
        const Duration(seconds: 10),
        (_) => ref.read(unreadCountProvider.notifier).refresh(),
      );
      _permPollTimer = Timer.periodic(
        const Duration(seconds: 60),
        (_) => ref.read(permissionProvider.notifier).load(),
      );
    } else {
      // Guest: load guest permissions, skip subscription check
      setState(() => _subChecked = true);
      ref.read(permissionProvider.notifier).loadGuest();
      _permPollTimer = Timer.periodic(
        const Duration(seconds: 60),
        (_) => ref.read(permissionProvider.notifier).loadGuest(),
      );
    }
  }

  @override
  void dispose() {
    _unreadPollTimer?.cancel();
    _permPollTimer?.cancel();
    super.dispose();
  }

  static const _privilegedRoles = ['editor', 'admin', 'owner'];

  Future<void> _checkSubscription() async {
    final user = ref.read(authProvider).user;
    if (user == null || _privilegedRoles.contains(user.role)) {
      setState(() => _subChecked = true);
      return;
    }

    setState(() => _needsSubscription = true);

    try {
      final status = await ref.read(apiServiceProvider).getSubscriptionStatus();
      if (mounted) {
        setState(() {
          _daysLeft = status['days_left'] as int?;
          _isActive = status['is_active'] == true;
          _subChecked = true;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _subChecked = true);
    }
  }

  bool get _shouldShowBanner {
    if (_bannerDismissed || !_needsSubscription || !_subChecked) return false;
    if (!_isActive) return true;
    if (_daysLeft != null && _daysLeft! <= 15) return true;
    return false;
  }

  /// Routes allowed without active subscription
  bool _isAllowedRoute(String location) {
    if (location == '/marketplace') return true;
    if (location.startsWith('/profile')) return true;
    if (location.startsWith('/subscription')) return true;
    if (location.startsWith('/notifications')) return true;
    if (location.startsWith('/feedback')) return true;
    if (location.startsWith('/seller')) return true;
    return false;
  }

  bool get _shouldShowGate {
    if (!_subChecked || !_needsSubscription) return false;
    if (_isActive) return false;
    final location = GoRouterState.of(context).uri.path;
    return !_isAllowedRoute(location);
  }

  bool _hasPerm(String key) {
    final user = ref.read(authProvider).user;
    if (user != null && user.role == 'owner') return true;
    return ref.read(permissionProvider.notifier).hasPermission(key);
  }

  @override
  Widget build(BuildContext context) {
    final unreadCount = ref.watch(unreadCountProvider);
    final user = ref.watch(authProvider).user;
    // Watch permissions to rebuild when loaded
    ref.watch(permissionProvider);

    // Build nav destinations based on permissions
    final navItems = <_NavDest>[];
    if (_hasPerm('marketplace.browse')) {
      navItems.add(_NavDest(
        dest: const NavigationDestination(icon: Icon(Icons.storefront), label: 'Sàn gạo'),
        route: '/marketplace',
      ));
    }
    if (_hasPerm('listings.create')) {
      navItems.add(_NavDest(
        dest: const NavigationDestination(icon: Icon(Icons.list_alt), label: 'Tin của tôi'),
        route: '/my-listings',
      ));
    }
    if (_hasPerm('chat.send')) {
      navItems.add(_NavDest(
        dest: NavigationDestination(
          icon: Badge(
            isLabelVisible: unreadCount > 0,
            label: Text(unreadCount > 99 ? '99+' : '$unreadCount'),
            child: const Icon(Icons.chat_bubble_outline),
          ),
          label: 'Tin nhắn',
        ),
        route: '/inbox',
      ));
    }
    if (user != null) {
      navItems.add(_NavDest(
        dest: const NavigationDestination(icon: Icon(Icons.person_outline), label: 'Tài khoản'),
        route: '/profile',
      ));
    } else {
      navItems.add(_NavDest(
        dest: const NavigationDestination(icon: Icon(Icons.login), label: 'Đăng nhập'),
        route: '/login',
      ));
    }

    // Calculate selected index based on visible destinations
    final location = GoRouterState.of(context).uri.path;
    int selectedIndex = 0;
    for (int i = 0; i < navItems.length; i++) {
      final route = navItems[i].route;
      if (location.startsWith(route) ||
          (route == '/my-listings' && location.startsWith('/create-listing')) ||
          (route == '/inbox' && location.startsWith('/chat')) ||
          (route == '/profile' && (location.startsWith('/notifications') || location.startsWith('/subscription') || location.startsWith('/seller')))) {
        selectedIndex = i;
        break;
      }
    }

    // Show gate for restricted pages when subscription expired
    final showGate = _shouldShowGate;

    return Scaffold(
      body: Column(
        children: [
          if (_shouldShowBanner && !showGate)
            MaterialBanner(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              backgroundColor: !_isActive ? AppColors.error.withValues(alpha: 0.08) : AppColors.warning.withValues(alpha: 0.08),
              leading: Icon(
                !_isActive ? Icons.warning : Icons.access_time,
                color: !_isActive ? AppColors.error : AppColors.warning,
              ),
              content: Text(
                !_isActive
                    ? 'Gói dịch vụ đã hết hạn. Tin đăng đã bị tạm ẩn.'
                    : 'Gói dịch vụ còn $_daysLeft ngày. Gia hạn sớm để không bị gián đoạn.',
                style: TextStyle(
                  color: !_isActive ? AppColors.error : AppColors.warning,
                  fontSize: 13,
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => context.push('/subscription'),
                  child: Text(
                    'Xem chi tiết',
                    style: TextStyle(
                      color: !_isActive ? AppColors.error : AppColors.warning,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close, size: 18),
                  onPressed: () => setState(() => _bannerDismissed = true),
                ),
              ],
            ),
          Expanded(
            child: showGate
                ? SubscriptionGate(userName: user?.name ?? user?.phone ?? '')
                : widget.child,
          ),
        ],
      ),
      bottomNavigationBar: navItems.length >= 2
          ? NavigationBar(
              selectedIndex: selectedIndex.clamp(0, navItems.length - 1),
              onDestinationSelected: (i) {
                final route = navItems[i].route;
                context.go(route);
                if (route == '/inbox') {
                  ref.read(unreadCountProvider.notifier).refresh();
                }
              },
              destinations: navItems.map((d) => d.dest).toList(),
            )
          : null,
    );
  }
}

class _NavDest {
  final NavigationDestination dest;
  final String route;
  const _NavDest({required this.dest, required this.route});
}
