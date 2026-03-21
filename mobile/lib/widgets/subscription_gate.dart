import 'dart:async';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../theme/app_theme.dart';

class SubscriptionGate extends StatefulWidget {
  final String userName;
  const SubscriptionGate({super.key, required this.userName});

  @override
  State<SubscriptionGate> createState() => _SubscriptionGateState();
}

class _SubscriptionGateState extends State<SubscriptionGate> {
  int _seconds = 15;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_seconds <= 1) {
        timer.cancel();
        if (mounted) context.go('/subscription');
      } else {
        setState(() => _seconds--);
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    color: AppColors.primary.withValues(alpha: 0.1),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(Icons.card_membership, size: 40, color: AppColors.primary),
                ),
                const SizedBox(height: 24),
                Text(
                  'Xin mời ${widget.userName}',
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'gia hạn gói dịch vụ để ủng hộ\nđội ngũ phát triển ứng dụng',
                  style: TextStyle(fontSize: 16, color: AppColors.textSecondary, height: 1.5),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
                // Countdown
                Stack(
                  alignment: Alignment.center,
                  children: [
                    SizedBox(
                      width: 64,
                      height: 64,
                      child: CircularProgressIndicator(
                        value: _seconds / 15,
                        strokeWidth: 4,
                        color: AppColors.primary,
                        backgroundColor: AppColors.primary.withValues(alpha: 0.1),
                      ),
                    ),
                    Text(
                      '$_seconds',
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                        color: AppColors.primary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                FilledButton.icon(
                  onPressed: () => context.go('/subscription'),
                  icon: const Icon(Icons.arrow_forward),
                  label: const Text('Gia hạn ngay'),
                ),
                const SizedBox(height: 12),
                TextButton(
                  onPressed: () => context.go('/marketplace'),
                  child: Text(
                    'Quay lại sàn gạo',
                    style: TextStyle(color: AppColors.textSecondary),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
