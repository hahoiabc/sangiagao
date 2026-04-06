import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class ChangePhoneScreen extends ConsumerStatefulWidget {
  const ChangePhoneScreen({super.key});

  @override
  ConsumerState<ChangePhoneScreen> createState() => _ChangePhoneScreenState();
}

class _ChangePhoneScreenState extends ConsumerState<ChangePhoneScreen> {
  final _phoneCtrl = TextEditingController();
  final _otpCtrl = TextEditingController();
  bool _loading = false;
  bool _otpStep = false;
  String? _error;
  int _cooldown = 0;
  Timer? _cooldownTimer;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _otpCtrl.dispose();
    _cooldownTimer?.cancel();
    super.dispose();
  }

  void _startCooldown() {
    _cooldown = 60;
    _cooldownTimer?.cancel();
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (mounted) {
        setState(() { _cooldown--; });
        if (_cooldown <= 0) t.cancel();
      } else {
        t.cancel();
      }
    });
  }

  Future<void> _sendOTP() async {
    final phone = _phoneCtrl.text.trim();
    if (phone.isEmpty || !RegExp(r'^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$').hasMatch(phone)) {
      setState(() => _error = 'Số điện thoại không hợp lệ');
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(apiServiceProvider).sendOTP(phone);
      _startCooldown();
      setState(() => _otpStep = true);
    } catch (e) {
      String msg = 'Gửi mã OTP thất bại';
      if (e is DioException && e.response?.data is Map) {
        msg = e.response?.data['error'] ?? msg;
      }
      setState(() => _error = msg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _changePhone() async {
    final code = _otpCtrl.text.trim();
    if (code.length != 6) {
      setState(() => _error = 'Vui lòng nhập mã OTP 6 số');
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(apiServiceProvider).changePhoneAuth(_phoneCtrl.text.trim(), code);
      // Refresh user data
      await ref.read(authProvider.notifier).refreshUser();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đổi số điện thoại thành công')),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      String msg = 'Đổi số điện thoại thất bại';
      if (e is DioException && e.response?.data is Map) {
        msg = e.response?.data['error'] ?? msg;
      }
      setState(() => _error = msg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authProvider).user;

    return Scaffold(
      appBar: AppBar(title: const Text('Đổi số điện thoại')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (user != null) ...[
              Text(
                'SĐT hiện tại: ${user.phone}',
                style: TextStyle(color: AppColors.textSecondary),
              ),
              const SizedBox(height: 20),
            ],
            if (!_otpStep) ...[
              TextField(
                controller: _phoneCtrl,
                keyboardType: TextInputType.phone,
                decoration: const InputDecoration(
                  labelText: 'Số điện thoại mới',
                  hintText: '0901234567',
                  prefixIcon: Icon(Icons.phone),
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 24),
              FilledButton(
                onPressed: (_loading || _cooldown > 0) ? null : _sendOTP,
                style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                child: Text(_loading ? 'Đang gửi mã OTP...' : _cooldown > 0 ? 'Gửi lại sau $_cooldown giây' : 'Gửi mã OTP'),
              ),
            ] else ...[
              Text(
                'Mã OTP đã gửi đến ${_phoneCtrl.text.trim()}',
                style: TextStyle(color: AppColors.textSecondary),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _otpCtrl,
                keyboardType: TextInputType.number,
                maxLength: 6,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 24, letterSpacing: 8),
                decoration: const InputDecoration(
                  labelText: 'Mã OTP',
                  hintText: '000000',
                  prefixIcon: Icon(Icons.sms),
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 24),
              FilledButton(
                onPressed: _loading ? null : _changePhone,
                style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                child: Text(_loading ? 'Đang xử lý...' : 'Xác nhận đổi SĐT'),
              ),
              const SizedBox(height: 8),
              TextButton(
                onPressed: () => setState(() { _otpStep = false; _otpCtrl.clear(); _error = null; }),
                child: const Text('Quay lại'),
              ),
            ],
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(_error!, style: const TextStyle(color: AppColors.error), textAlign: TextAlign.center),
            ],
          ],
        ),
      ),
    );
  }
}
