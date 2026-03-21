import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class _CategoryItem {
  final String key;
  final String label;
  final IconData icon;
  final Color color;

  const _CategoryItem({
    required this.key,
    required this.label,
    required this.icon,
    required this.color,
  });
}

const _categories = [
  _CategoryItem(key: 'gao_deo_thom', label: 'Gạo dẻo thơm', icon: Icons.rice_bowl, color: Color(0xFF4CAF50)),
  _CategoryItem(key: 'gao_kho', label: 'Gạo khô', icon: Icons.grass, color: Color(0xFFFF9800)),
  _CategoryItem(key: 'tam_deo_thom', label: 'Tấm dẻo thơm', icon: Icons.grain, color: Color(0xFF2196F3)),
  _CategoryItem(key: 'tam_kho', label: 'Tấm khô', icon: Icons.scatter_plot, color: Color(0xFF9C27B0)),
  _CategoryItem(key: 'nep', label: 'Nếp', icon: Icons.spa, color: Color(0xFFE91E63)),
];

class SelectCategoryScreen extends StatelessWidget {
  const SelectCategoryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chọn loại gạo')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Bạn muốn đăng tin loại gạo nào?',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 20),
            Expanded(
              child: GridView.count(
                crossAxisCount: 2,
                mainAxisSpacing: 16,
                crossAxisSpacing: 16,
                childAspectRatio: 1.1,
                children: _categories.map((cat) {
                  return _CategoryCard(
                    item: cat,
                    onTap: () => context.push('/create-listing/${cat.key}'),
                  );
                }).toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _CategoryCard extends StatelessWidget {
  final _CategoryItem item;
  final VoidCallback onTap;

  const _CategoryCard({required this.item, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: InkWell(
        borderRadius: BorderRadius.circular(16),
        onTap: onTap,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: item.color.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: Icon(item.icon, size: 32, color: item.color),
            ),
            const SizedBox(height: 12),
            Text(
              item.label,
              style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}
