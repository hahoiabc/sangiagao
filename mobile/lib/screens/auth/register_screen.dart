import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../widgets/location_picker.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _phoneController = TextEditingController();
  final _otpController = TextEditingController();
  final _nameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _addressController = TextEditingController();

  String? _province;
  String? _ward;

  bool _loading = false;
  bool _obscurePassword = true;
  bool _obscureConfirm = true;
  bool _step2 = false; // false = phone step, true = OTP + details step
  bool _acceptedTOS = false;
  String? _error;
  int _cooldown = 0;
  Timer? _cooldownTimer;

  @override
  void dispose() {
    _phoneController.dispose();
    _otpController.dispose();
    _nameController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _addressController.dispose();
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

  // Step 1: Validate phone and send OTP
  Future<void> _sendOTP() async {
    final phone = _phoneController.text.trim();
    if (phone.isEmpty || !RegExp(r'^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$').hasMatch(phone)) {
      setState(() { _error = 'Số điện thoại không hợp lệ, vui lòng kiểm tra đầu số'; });
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).register(phone);
      _startCooldown();
      setState(() { _step2 = true; });
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

  // Step 2: Validate all fields + verify OTP + complete register
  String? _validateStep2() {
    final code = _otpController.text.trim();
    if (code.length != 6) return 'Vui lòng nhập mã OTP 6 số';

    final name = _nameController.text.trim();
    if (name.length < 4) return 'Tên phải có ít nhất 4 ký tự';
    if (name.length > 60) return 'Tên không được quá 60 ký tự';

    final password = _passwordController.text;
    if (password.length < 6) return 'Mật khẩu phải có ít nhất 6 ký tự';
    if (!RegExp(r'[A-Z]').hasMatch(password)) return 'Mật khẩu phải có ít nhất 1 chữ hoa';
    if (!RegExp(r'[a-z]').hasMatch(password)) return 'Mật khẩu phải có ít nhất 1 chữ thường';
    if (!RegExp(r'[^a-zA-Z0-9]').hasMatch(password)) return 'Mật khẩu phải có ít nhất 1 ký tự đặc biệt';

    final confirm = _confirmPasswordController.text;
    if (password != confirm) return 'Mật khẩu nhập lại không khớp';

    final address = _addressController.text.trim();
    if (address.isNotEmpty) {
      if (address.length < 6) return 'Địa chỉ chi tiết phải có ít nhất 6 ký tự';
      if (address.length > 80) return 'Địa chỉ chi tiết không được quá 80 ký tự';
    }

    if (!_acceptedTOS) return 'Vui lòng đồng ý điều khoản sử dụng';

    return null;
  }

  Future<void> _verifyAndRegister() async {
    final validationError = _validateStep2();
    if (validationError != null) {
      setState(() { _error = validationError; });
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      await ref.read(authProvider.notifier).completeRegister(
        phone: _phoneController.text.trim(),
        code: _otpController.text.trim(),
        name: _nameController.text.trim(),
        password: _passwordController.text,
        province: _province,
        ward: _ward,
        address: _addressController.text.trim().isEmpty ? null : _addressController.text.trim(),
      );
    } catch (e) {
      String msg = 'Đăng ký thất bại';
      if (e is DioException && e.response?.data is Map) {
        msg = e.response?.data['error'] ?? msg;
      }
      setState(() { _error = msg; });
    } finally {
      if (mounted) setState(() { _loading = false; });
    }
  }

  void _showTerms() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => DraggableScrollableSheet(
        initialChildSize: 0.85,
        maxChildSize: 0.95,
        minChildSize: 0.5,
        expand: false,
        builder: (_, scrollCtrl) => Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Điều khoản sử dụng',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            Expanded(
              child: ListView(
                controller: scrollCtrl,
                padding: const EdgeInsets.all(16),
                children: const [
                  Text(
                    'ĐIỀU KHOẢN SỬ DỤNG SANGIAGAO.VN',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '1. GIỚI THIỆU',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'SanGiaGao.Vn là nền tảng công nghệ kết nối người sản xuất, thương nhân '
                    'và người mua trong ngành gạo Việt Nam. SanGiaGao.Vn là công cụ hỗ trợ '
                    'giúp các thành viên kết nối thuận tiện và nhanh chóng, không trực tiếp '
                    'tham gia vào các giao dịch mua bán giữa các bên.',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '2. TRÁCH NHIỆM CỦA THÀNH VIÊN',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'Mỗi thành viên tham gia SanGiaGao.Vn phải tự chịu trách nhiệm hoàn toàn '
                    'cho mọi quyết định giao dịch của mình, bao gồm nhưng không giới hạn:\n\n'
                    '- Tính chính xác của thông tin sản phẩm đăng tải\n'
                    '- Chất lượng hàng hóa và dịch vụ\n'
                    '- Việc thỏa thuận giá cả, số lượng và điều kiện giao hàng\n'
                    '- Thanh toán và các nghĩa vụ tài chính phát sinh\n'
                    '- Tuân thủ pháp luật Việt Nam trong quá trình giao dịch',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '3. VAI TRÒ CỦA SANGIAGAO.VN',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'SanGiaGao.Vn cam kết:\n\n'
                    '- Cung cấp nền tảng kết nối minh bạch và công bằng\n'
                    '- Hỗ trợ cung cấp thông tin trong khả năng của sàn khi có yêu cầu từ thành viên\n'
                    '- Duy trì môi trường giao dịch lành mạnh thông qua hệ thống đánh giá và báo cáo vi phạm\n'
                    '- Bảo mật thông tin cá nhân của thành viên theo quy định pháp luật\n\n'
                    'SanGiaGao.Vn không chịu trách nhiệm cho bất kỳ tranh chấp, tổn thất hoặc thiệt hại '
                    'phát sinh từ giao dịch giữa các thành viên.',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '4. GÓI DỊCH VỤ',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    '- Thành viên được dùng thử miễn phí 30 ngày kể từ ngày đăng ký\n'
                    '- Sau thời gian dùng thử, phí dịch vụ sẽ được tính theo các gói đăng ký của thành viên\n'
                    '- Khi hết hạn gói dịch vụ, tin đăng sẽ bị tạm ẩn cho đến khi gia hạn',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '5. NỘI DUNG BỊ CẤM',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    '- Đăng thông tin sai lệch, gian lận\n'
                    '- Sử dụng sàn cho mục đích bất hợp pháp\n'
                    '- Quấy rối, đe dọa hoặc xúc phạm thành viên khác\n'
                    '- Spam, đăng tin trùng lặp hoặc không liên quan đến gạo/nông sản',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    '6. XỬ LÝ VI PHẠM',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'SanGiaGao.Vn có quyền cảnh cáo, tạm khóa hoặc xóa vĩnh viễn tài khoản vi phạm '
                    'điều khoản sử dụng mà không cần thông báo trước.',
                    style: TextStyle(height: 1.5),
                  ),
                  SizedBox(height: 16),
                  Text(
                    'Bằng việc tích chọn "Đồng ý điều khoản", bạn xác nhận đã đọc, hiểu và chấp nhận '
                    'toàn bộ các điều khoản trên.',
                    style: TextStyle(fontStyle: FontStyle.italic, height: 1.5),
                  ),
                  SizedBox(height: 24),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: FilledButton(
                onPressed: () {
                  setState(() => _acceptedTOS = true);
                  Navigator.pop(ctx);
                },
                style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                child: const Text('Đồng ý điều khoản'),
              ),
            ),
          ],
        ),
      ),
    );
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
                child: const Icon(Icons.person_add, size: 40, color: Colors.white),
              ),
              const SizedBox(height: 16),
              Text(
                _step2 ? 'Xác minh và hoàn tất' : 'Đăng ký tài khoản',
                style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                _step2
                    ? 'Nhập mã OTP đã gửi đến ${_phoneController.text} và điền thông tin'
                    : 'Nhập số điện thoại để đăng ký',
                style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),

              if (!_step2) ...[
                // Step 1: Phone only
                TextField(
                  controller: _phoneController,
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(
                    labelText: 'Số điện thoại *',
                    hintText: '0901234567',
                    prefixIcon: Icon(Icons.phone),
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),

                FilledButton(
                  onPressed: (_loading || _cooldown > 0) ? null : _sendOTP,
                  style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                  child: Text(_loading ? 'Đang gửi mã OTP...' : _cooldown > 0 ? 'Gửi lại sau $_cooldown giây' : 'Tiếp tục'),
                ),
              ] else ...[
                // Step 2: OTP + all details

                // OTP
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

                // Name
                TextField(
                  controller: _nameController,
                  textCapitalization: TextCapitalization.words,
                  maxLength: 60,
                  decoration: const InputDecoration(
                    labelText: 'Họ và tên * (4-60 ký tự)',
                    hintText: 'VD: Nguyễn Văn A',
                    prefixIcon: Icon(Icons.person),
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),

                // Password
                TextField(
                  controller: _passwordController,
                  obscureText: _obscurePassword,
                  decoration: InputDecoration(
                    labelText: 'Mật khẩu *',
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

                // Confirm Password
                TextField(
                  controller: _confirmPasswordController,
                  obscureText: _obscureConfirm,
                  decoration: InputDecoration(
                    labelText: 'Nhập lại mật khẩu *',
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: const OutlineInputBorder(),
                    suffixIcon: IconButton(
                      icon: Icon(_obscureConfirm ? Icons.visibility_off : Icons.visibility),
                      onPressed: () => setState(() { _obscureConfirm = !_obscureConfirm; }),
                    ),
                  ),
                ),
                const SizedBox(height: 16),

                // Location picker (Province -> Ward)
                LocationPicker(
                  onChanged: (province, ward) {
                    _province = province;
                    _ward = ward;
                  },
                ),
                const SizedBox(height: 16),

                // Address
                TextField(
                  controller: _addressController,
                  maxLength: 80,
                  decoration: const InputDecoration(
                    labelText: 'Địa chỉ chi tiết (6-80 ký tự)',
                    hintText: 'VD: 123 Nguyễn Huệ, Quận 1',
                    prefixIcon: Icon(Icons.home),
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),

                // T&C checkbox
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: 24,
                      height: 24,
                      child: Checkbox(
                        value: _acceptedTOS,
                        onChanged: (v) => setState(() => _acceptedTOS = v ?? false),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: GestureDetector(
                        onTap: () => _showTerms(),
                        child: RichText(
                          text: TextSpan(
                            style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                            children: const [
                              TextSpan(text: 'Tôi đã đọc và đồng ý với '),
                              TextSpan(
                                text: 'Điều khoản sử dụng',
                                style: TextStyle(
                                  color: AppColors.primary,
                                  fontWeight: FontWeight.bold,
                                  decoration: TextDecoration.underline,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                FilledButton(
                  onPressed: _loading || !_acceptedTOS ? null : _verifyAndRegister,
                  style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
                  child: Text(_loading ? 'Đang xử lý...' : 'Đăng ký'),
                ),
                const SizedBox(height: 8),
                TextButton(
                  onPressed: () => setState(() { _step2 = false; _otpController.clear(); _error = null; }),
                  child: const Text('Quay lại'),
                ),
              ],

              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(_error!, style: const TextStyle(color: AppColors.error), textAlign: TextAlign.center),
              ],

              const SizedBox(height: 16),
              if (!_step2)
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text('Đã có tài khoản?', style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
                    TextButton(
                      onPressed: () => context.go('/login'),
                      child: const Text('Đăng nhập', style: TextStyle(fontWeight: FontWeight.bold)),
                    ),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }
}
