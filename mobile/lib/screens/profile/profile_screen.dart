import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import '../../providers/providers.dart';
import '../../providers/theme_provider.dart';
import '../../widgets/location_picker.dart';
import '../../theme/app_theme.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  bool _editing = false;
  bool _uploadingAvatar = false;
  final _nameCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _orgCtrl = TextEditingController();
  final _addressCtrl = TextEditingController();

  String? _province;
  String? _ward;

  @override
  void initState() {
    super.initState();
    final user = ref.read(authProvider).user;
    if (user != null) {
      _nameCtrl.text = user.name ?? '';
      _descCtrl.text = user.description ?? '';
      _orgCtrl.text = user.orgName ?? '';
      _addressCtrl.text = user.address ?? '';
      _province = user.province;
      _ward = user.ward;
    }
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _descCtrl.dispose();
    _orgCtrl.dispose();
    _addressCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickAndUploadAvatar() async {
    final picker = ImagePicker();
    final image = await picker.pickImage(source: ImageSource.gallery, maxWidth: 512, maxHeight: 512, imageQuality: 80);
    if (image == null) return;

    setState(() => _uploadingAvatar = true);
    try {
      await ref.read(authProvider.notifier).uploadAvatar(image.path);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cập nhật ảnh đại diện thành công')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi tải ảnh: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _uploadingAvatar = false);
    }
  }

  Future<void> _save() async {
    final name = _nameCtrl.text.trim();
    final address = _addressCtrl.text.trim();
    if (name.length < 4 || name.length > 60) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tên phải có từ 4 đến 60 ký tự')),
      );
      return;
    }
    if (address.isNotEmpty && (address.length < 6 || address.length > 80)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Địa chỉ chi tiết phải có từ 6 đến 80 ký tự')),
      );
      return;
    }
    try {
      await ref.read(authProvider.notifier).updateProfile({
        'name': _nameCtrl.text.trim(),
        'province': _province ?? '',
        'ward': _ward ?? '',
        'address': _addressCtrl.text.trim(),
        'description': _descCtrl.text.trim(),
        'org_name': _orgCtrl.text.trim(),
      });
      setState(() => _editing = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cập nhật thành công')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi: $e')),
        );
      }
    }
  }

  Future<void> _deleteAccount() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Xóa tài khoản'),
        content: const Text(
          'Bạn có chắc chắn muốn xóa tài khoản? '
          'Hành động này không thể hoàn tác. Tất cả dữ liệu của bạn sẽ bị xóa vĩnh viễn.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Huỷ')),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('Xóa tài khoản'),
          ),
        ],
      ),
    );
    if (confirm == true) {
      try {
        await ref.read(authProvider.notifier).deleteAccount();
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Lỗi: $e')),
          );
        }
      }
    }
  }

  void _showThemePicker(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Chọn màu chủ đạo', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: themeOptions.map((t) {
                final current = ref.watch(themeProvider);
                final isSelected = current.key == t.key;
                return GestureDetector(
                  onTap: () {
                    ref.read(themeProvider.notifier).setTheme(t.key);
                    Navigator.pop(ctx);
                  },
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 48,
                        height: 48,
                        decoration: BoxDecoration(
                          color: t.primary,
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: isSelected ? AppColors.textPrimary : Colors.transparent,
                            width: 3,
                          ),
                        ),
                        child: isSelected ? const Icon(Icons.check, color: Colors.white, size: 24) : null,
                      ),
                      const SizedBox(height: 4),
                      Text(t.label, style: const TextStyle(fontSize: 11)),
                    ],
                  ),
                );
              }).toList(),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _logout() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Đăng xuất'),
        content: const Text('Bạn có chắc muốn đăng xuất?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Huỷ')),
          TextButton(onPressed: () => Navigator.pop(dialogContext, true), child: const Text('Đăng xuất')),
        ],
      ),
    );
    if (confirm == true) {
      await ref.read(authProvider.notifier).logout();
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);
    final user = authState.user;

    if (user == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Tài khoản')),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('Bạn chưa đăng nhập'),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () => context.go('/login'),
                child: const Text('Đăng nhập'),
              ),
            ],
          ),
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tài khoản'),
        actions: [
          if (!_editing)
            IconButton(icon: const Icon(Icons.edit), onPressed: () => setState(() => _editing = true)),
          IconButton(icon: const Icon(Icons.notifications_outlined), onPressed: () => context.push('/notifications')),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            // Avatar with upload
            GestureDetector(
              onTap: _uploadingAvatar ? null : _pickAndUploadAvatar,
              child: Stack(
                children: [
                  CircleAvatar(
                    radius: 48,
                    backgroundImage: user.avatarUrl != null ? CachedNetworkImageProvider(user.avatarUrl!) : null,
                    child: user.avatarUrl == null ? const Icon(Icons.person, size: 48) : null,
                  ),
                  Positioned(
                    bottom: 0,
                    right: 0,
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: AppColors.primary,
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white, width: 2),
                      ),
                      child: _uploadingAvatar
                          ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                          : const Icon(Icons.camera_alt, size: 16, color: Colors.white),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            Text(user.name ?? user.phone, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, height: 1.3)),
            const SizedBox(height: 4),
            Text('Thành viên', style: TextStyle(color: AppColors.textSecondary, height: 1.3)),
            const SizedBox(height: 20),
            if (_editing) ...[
              TextField(
                controller: _nameCtrl,
                maxLength: 60,
                decoration: const InputDecoration(labelText: 'Tên hiển thị (4-60 ký tự)', border: OutlineInputBorder()),
              ),
              const SizedBox(height: 12),
              // Location picker
              LocationPicker(
                initialProvince: user.province,
                initialWard: user.ward,
                onChanged: (province, ward) {
                  _province = province;
                  _ward = ward;
                },
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _addressCtrl,
                maxLength: 80,
                decoration: const InputDecoration(
                  labelText: 'Địa chỉ chi tiết (6-80 ký tự)',
                  hintText: 'VD: 123 Nguyễn Huệ, Quận 1',
                  prefixIcon: Icon(Icons.home),
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _descCtrl,
                decoration: const InputDecoration(labelText: 'Giới thiệu', border: OutlineInputBorder()),
                maxLines: 3,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _orgCtrl,
                decoration: const InputDecoration(labelText: 'Tên tổ chức/doanh nghiệp (nếu có)', border: OutlineInputBorder()),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => setState(() => _editing = false),
                      child: const Text('Huỷ'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton(onPressed: _save, child: const Text('Lưu')),
                  ),
                ],
              ),
            ] else ...[
              _infoTile(Icons.phone, 'Số điện thoại', user.phone),
              if (user.province != null) _infoTile(Icons.location_city, 'Tỉnh/TP', user.province!),
              if (user.ward != null) _infoTile(Icons.location_on, 'Phường/Xã', user.ward!),
              if (user.address != null) _infoTile(Icons.home, 'Địa chỉ chi tiết', user.address!),
              if (user.orgName != null) _infoTile(Icons.business, 'Tổ chức', user.orgName!),
              if (user.description != null) _infoTile(Icons.info_outline, 'Giới thiệu', user.description!),
              const Divider(height: 28),
              if (!['editor', 'admin', 'owner'].contains(user.role))
                ListTile(
                  leading: const Icon(Icons.card_membership),
                  title: const Text('Gói dịch vụ & Gia hạn'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.push('/subscription'),
                ),
              ListTile(
                leading: const Icon(Icons.lock_outline),
                title: const Text('Đổi mật khẩu'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/change-password'),
              ),
              ListTile(
                leading: const Icon(Icons.phone),
                title: const Text('Đổi số điện thoại'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/change-phone'),
              ),
              ListTile(
                leading: const Icon(Icons.feedback_outlined),
                title: const Text('Góp ý cho nhà phát triển'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/feedback'),
              ),
              ListTile(
                leading: const Icon(Icons.history),
                title: const Text('Lịch sử góp ý & phản hồi'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/feedback-history'),
              ),
              // Theme color picker
              ListTile(
                leading: Icon(Icons.palette_outlined, color: ref.watch(themeProvider).primary),
                title: const Text('Màu chủ đạo'),
                subtitle: Text(ref.watch(themeProvider).label),
                trailing: Container(
                  width: 24,
                  height: 24,
                  decoration: BoxDecoration(
                    color: ref.watch(themeProvider).primary,
                    shape: BoxShape.circle,
                  ),
                ),
                onTap: () => _showThemePicker(context, ref),
              ),
              const Divider(),
              ListTile(
                leading: const Icon(Icons.description_outlined),
                title: const Text('Điều khoản sử dụng'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/terms-of-service'),
              ),
              ListTile(
                leading: const Icon(Icons.privacy_tip_outlined),
                title: const Text('Chính sách bảo mật'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/privacy-policy'),
              ),
              ListTile(
                leading: const Icon(Icons.help_outline),
                title: const Text('Hướng dẫn sử dụng'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/user-guide'),
              ),
              ListTile(
                leading: const Icon(Icons.logout, color: AppColors.error),
                title: const Text('Đăng xuất', style: TextStyle(color: AppColors.error)),
                onTap: _logout,
              ),
              ListTile(
                leading: const Icon(Icons.delete_forever, color: AppColors.error),
                title: const Text('Xóa tài khoản', style: TextStyle(color: AppColors.error)),
                onTap: _deleteAccount,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _infoTile(IconData icon, String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Icon(icon, size: 22, color: AppColors.textSecondary),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label, style: TextStyle(fontSize: 12, color: AppColors.textHint, height: 1.4)),
                const SizedBox(height: 2),
                Text(value, style: const TextStyle(fontSize: 15, height: 1.4)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
