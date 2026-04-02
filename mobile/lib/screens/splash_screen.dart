import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/providers.dart';
import '../widgets/marquee_text.dart';

class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _fadeIn;
  bool _minDelayDone = false;
  bool _navigated = false;
  String _slogan = 'Kết nối ngành gạo';

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 1200));
    _fadeIn = CurvedAnimation(parent: _controller, curve: Curves.easeIn);
    _controller.forward();

    ref.read(apiServiceProvider).getSlogan().then((s) {
      if (mounted) setState(() => _slogan = s);
    }).catchError((_) {});

    Future.delayed(const Duration(milliseconds: 2500), () {
      if (mounted) {
        _minDelayDone = true;
        _tryNavigate();
      }
    });
  }

  void _tryNavigate() {
    if (_navigated || !_minDelayDone || !mounted) return;
    final authState = ref.read(authProvider);
    if (authState.status == AuthStatus.unknown) return;

    _navigated = true;

    // Check if user was blocked
    final authNotifier = ref.read(authProvider.notifier);
    if (authNotifier.wasBlocked) {
      authNotifier.wasBlocked = false;
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (_) => AlertDialog(
          title: const Text('Tài khoản bị khóa'),
          content: const Text(
            'Tài khoản của bạn đã bị khóa bởi quản trị viên. '
            'Nếu bạn cho rằng đây là nhầm lẫn, vui lòng liên hệ hỗ trợ.',
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(context).pop();
                context.go('/marketplace');
              },
              child: const Text('Đã hiểu'),
            ),
          ],
        ),
      );
    } else {
      context.go('/marketplace');
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Watch auth state to trigger navigation when auth resolves
    ref.listen<AuthState>(authProvider, (_, __) => _tryNavigate());

    return Scaffold(
      body: Stack(
        fit: StackFit.expand,
        children: [
          Container(color: const Color(0xFF007FFF)),
          Center(
            child: FadeTransition(
              opacity: _fadeIn,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    'SanGiaGao.vn',
                    style: TextStyle(
                      fontSize: 32,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 1.0,
                      shadows: [Shadow(blurRadius: 10, color: Colors.black54)],
                    ),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: 280,
                    child: MarqueeText(
                      text: _slogan,
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w400,
                        color: Colors.white.withValues(alpha: 0.9),
                        letterSpacing: 0.3,
                        shadows: [Shadow(blurRadius: 8, color: Colors.black54)],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
