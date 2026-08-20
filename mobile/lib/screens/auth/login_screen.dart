import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _phoneController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _loading = false;
  bool _obscurePassword = true;
  String? _error;

  @override
  void dispose() {
    _phoneController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;

    if (phone.isEmpty || !RegExp(r'^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$').hasMatch(phone)) {
      setState(() { _error = 'Số điện thoại không hợp lệ, vui lòng kiểm tra đầu số'; });
      return;
    }
    if (password.isEmpty) {
      setState(() { _error = 'Vui lòng nhập mật khẩu'; });
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).loginPassword(phone, password);
      if (mounted) context.go('/marketplace');
      return;
    } catch (e) {
      setState(() { _error = _loginErrorMessage(e); });
    } finally {
      if (mounted) setState(() { _loading = false; });
    }
  }

  String _loginErrorMessage(Object e) {
    if (e is! DioException) return 'Đăng nhập thất bại. Thử lại sau.';
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.connectionError:
        return 'Không kết nối được máy chủ. Kiểm tra mạng và thử lại.';
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return 'Mạng chậm, vui lòng thử lại.';
      case DioExceptionType.badCertificate:
        return 'Lỗi chứng chỉ SSL. Cập nhật ứng dụng / kiểm tra ngày giờ máy.';
      case DioExceptionType.cancel:
        return 'Đã huỷ. Thử lại.';
      case DioExceptionType.badResponse:
      case DioExceptionType.unknown:
        final code = e.response?.statusCode ?? 0;
        if (code == 401) return 'Sai số điện thoại hoặc mật khẩu.';
        if (code == 403) return 'Tài khoản đã bị khoá. Liên hệ hỗ trợ.';
        if (code == 429) return 'Thử quá nhiều lần. Đợi vài phút rồi thử lại.';
        if (code >= 500) return 'Lỗi máy chủ, thử lại sau.';
        if (e.response?.data is Map) {
          final m = e.response?.data['error'];
          if (m is String && m.isNotEmpty) return m;
        }
        if (code == 0) return 'Không kết nối được máy chủ. Kiểm tra mạng.';
        return 'Đăng nhập thất bại (mã $code).';
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
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const SizedBox(height: 48),
              // Logo
              Container(
                width: 88,
                height: 88,
                decoration: BoxDecoration(
                  gradient: AppStyles.primaryGradient,
                  borderRadius: BorderRadius.circular(24),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.primary.withValues(alpha: 0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 8),
                    ),
                  ],
                ),
                child: const Icon(Icons.rice_bowl, size: 44, color: Colors.white),
              ),
              const SizedBox(height: 20),
              Text('SanGiaGao.Vn', style: theme.textTheme.headlineMedium),
              const SizedBox(height: 6),
              Text(
                'Đăng nhập tài khoản',
                style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
              ),
              const SizedBox(height: 36),
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
              const SizedBox(height: 16),
              TextField(
                controller: _passwordController,
                obscureText: _obscurePassword,
                decoration: InputDecoration(
                  labelText: 'Mật khẩu',
                  prefixIcon: const Icon(Icons.lock),
                  border: const OutlineInputBorder(),
                  suffixIcon: IconButton(
                    icon: Icon(_obscurePassword ? Icons.visibility_off : Icons.visibility),
                    onPressed: () => setState(() { _obscurePassword = !_obscurePassword; }),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: FilledButton(
                  onPressed: _loading ? null : _login,
                  child: Text(_loading ? 'Đang đăng nhập...' : 'Đăng nhập'),
                ),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.error.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.error_outline, color: AppColors.error, size: 20),
                      const SizedBox(width: 8),
                      Expanded(child: Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 13))),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () => context.push('/forgot-password'),
                  child: const Text('Quên mật khẩu?'),
                ),
              ),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    'Chưa có tài khoản?',
                    style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                  ),
                  TextButton(
                    onPressed: () => context.go('/register'),
                    child: const Text('Đăng ký ngay', style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              TextButton(
                onPressed: () => context.push('/nha-phat-trien'),
                child: Text(
                  'Xem Trường Sơn Connect ›',
                  style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
