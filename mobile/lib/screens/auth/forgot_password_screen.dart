import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class ForgotPasswordScreen extends ConsumerStatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  ConsumerState<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends ConsumerState<ForgotPasswordScreen> {
  final _phoneController = TextEditingController();
  final _otpController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();

  bool _loading = false;
  bool _otpStep = false;
  bool _obscurePassword = true;
  bool _obscureConfirm = true;
  String? _error;
  int _cooldown = 0;
  Timer? _cooldownTimer;

  @override
  void dispose() {
    _phoneController.dispose();
    _otpController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
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
    final phone = _phoneController.text.trim();
    if (phone.isEmpty || !RegExp(r'^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$').hasMatch(phone)) {
      setState(() { _error = 'Số điện thoại không hợp lệ, vui lòng kiểm tra đầu số'; });
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).sendOTP(phone);
      _startCooldown();
      setState(() { _otpStep = true; });
    } catch (e) {
      String msg = 'Gửi mã OTP thất bại';
      if (e is DioException && e.response?.data is Map) {
        msg = e.response?.data['error'] ?? msg;
      }
      setState(() { _error = msg; });
    } finally {
      if (mounted) setState(() { _loading = false; });
    }
  }

  Future<void> _resetPassword() async {
    final code = _otpController.text.trim();
    final password = _passwordController.text;
    final confirm = _confirmPasswordController.text;

    if (code.length != 6) {
      setState(() { _error = 'Vui lòng nhập mã OTP 6 số'; });
      return;
    }
    if (password.length < 6) {
      setState(() { _error = 'Mật khẩu phải có ít nhất 6 ký tự'; });
      return;
    }
    if (!RegExp(r'[A-Z]').hasMatch(password)) {
      setState(() { _error = 'Mật khẩu phải có ít nhất 1 chữ hoa'; });
      return;
    }
    if (!RegExp(r'[a-z]').hasMatch(password)) {
      setState(() { _error = 'Mật khẩu phải có ít nhất 1 chữ thường'; });
      return;
    }
    if (!RegExp(r'[^a-zA-Z0-9]').hasMatch(password)) {
      setState(() { _error = 'Mật khẩu phải có ít nhất 1 ký tự đặc biệt'; });
      return;
    }
    if (password != confirm) {
      setState(() { _error = 'Mật khẩu nhập lại không khớp'; });
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).resetPassword(
        _phoneController.text.trim(),
        code,
        password,
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đặt lại mật khẩu thành công')),
        );
        context.go('/login');
      }
    } catch (e) {
      String msg = 'Đặt lại mật khẩu thất bại';
      if (e is DioException && e.response?.data is Map) {
        msg = e.response?.data['error'] ?? msg;
      }
      setState(() { _error = msg; });
    } finally {
      if (mounted) setState(() { _loading = false; });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: 24),
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  gradient: AppStyles.primaryGradient,
                  borderRadius: BorderRadius.circular(22),
                  boxShadow: [BoxShadow(color: AppColors.primary.withValues(alpha: 0.3), blurRadius: 20, offset: const Offset(0, 8))],
                ),
                child: const Icon(Icons.lock_reset, size: 40, color: Colors.white),
              ),
              const SizedBox(height: 16),
              Text(
                _otpStep ? 'Đặt lại mật khẩu' : 'Quên mật khẩu',
                style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                _otpStep
                    ? 'Nhập mã OTP và mật khẩu mới'
                    : 'Nhập số điện thoại để nhận mã OTP',
                style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),

              if (!_otpStep) ...[
                TextField(
                  controller: _phoneController,
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(
                    labelText: 'Số điện thoại',
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
                TextField(
                  controller: _otpController,
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
                const SizedBox(height: 16),
                TextField(
                  controller: _passwordController,
                  obscureText: _obscurePassword,
                  decoration: InputDecoration(
                    labelText: 'Mật khẩu mới',
                    hintText: 'Chữ hoa, chữ thường, ký tự đặc biệt',
                    prefixIcon: const Icon(Icons.lock),
                    border: const OutlineInputBorder(),
                    suffixIcon: IconButton(
                      icon: Icon(_obscurePassword ? Icons.visibility_off : Icons.visibility),
                      onPressed: () => setState(() { _obscurePassword = !_obscurePassword; }),
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _confirmPasswordController,
                  obscureText: _obscureConfirm,
                  decoration: InputDecoration(
                    labelText: 'Nhập lại mật khẩu mới',
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: const OutlineInputBorder(),
                    suffixIcon: IconButton(
                      icon: Icon(_obscureConfirm ? Icons.visibility_off : Icons.visibility),
                      onPressed: () => setState(() { _obscureConfirm = !_obscureConfirm; }),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: _loading ? null : _resetPassword,
                  style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                  child: Text(_loading ? 'Đang xử lý...' : 'Đặt lại mật khẩu'),
                ),
                const SizedBox(height: 8),
                TextButton(
                  onPressed: () => setState(() { _otpStep = false; _otpController.clear(); _error = null; }),
                  child: const Text('Quay lại'),
                ),
              ],

              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(_error!, style: const TextStyle(color: AppColors.error), textAlign: TextAlign.center),
              ],

              const SizedBox(height: 16),
              TextButton(
                onPressed: () => context.go('/login'),
                child: const Text('Quay lại đăng nhập'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
