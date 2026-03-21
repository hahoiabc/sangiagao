import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';

class RoleScreen extends ConsumerWidget {
  const RoleScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Bạn là ai?', style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 8),
              const Text('Chọn vai trò của bạn trên SanGiaGao.Com', style: TextStyle(color: AppColors.textHint)),
              const SizedBox(height: 32),
              _RoleCard(
                icon: Icons.person,
                title: 'Thành viên',
                subtitle: 'Tham gia sàn giao dịch gạo SanGiaGao.Com',
                onTap: () async {
                  await ref.read(authProvider.notifier).updateProfile({'role': 'member', 'accept_tos': true});
                  if (context.mounted) context.go('/marketplace');
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _RoleCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  const _RoleCard({required this.icon, required this.title, required this.subtitle, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: CircleAvatar(child: Icon(icon)),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.bold)),
        subtitle: Text(subtitle),
        trailing: const Icon(Icons.arrow_forward_ios, size: 16),
        onTap: onTap,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      ),
    );
  }
}
