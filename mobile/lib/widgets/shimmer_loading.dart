import 'package:flutter/material.dart';
import 'package:shimmer/shimmer.dart';
import '../theme/app_theme.dart';

/// A shimmer placeholder box used to build loading skeletons.
class ShimmerBox extends StatelessWidget {
  final double width;
  final double height;
  final double borderRadius;

  const ShimmerBox({
    super.key,
    this.width = double.infinity,
    required this.height,
    this.borderRadius = 6,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    );
  }
}

/// Wraps children in a Shimmer effect.
class ShimmerWrap extends StatelessWidget {
  final Widget child;

  const ShimmerWrap({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Shimmer.fromColors(
      baseColor: isDark ? AppColors.textSecondary : AppColors.border,
      highlightColor: isDark ? AppColors.textHint : AppColors.divider,
      child: child,
    );
  }
}

/// Skeleton for the price board screen (5 category sections).
class PriceBoardSkeleton extends StatelessWidget {
  const PriceBoardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerWrap(
      child: ListView(
        padding: const EdgeInsets.fromLTRB(12, 8, 12, 24),
        physics: const NeverScrollableScrollPhysics(),
        children: List.generate(3, (catIndex) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(4, 16, 4, 8),
                child: Center(child: ShimmerBox(width: 140, height: 20)),
              ),
              Card(
                clipBehavior: Clip.antiAlias,
                child: Column(
                  children: [
                    // Header
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      child: Row(
                        children: [
                          const SizedBox(width: 36),
                          Expanded(flex: 4, child: ShimmerBox(height: 14, width: 70)),
                          const SizedBox(width: 8),
                          Expanded(flex: 3, child: ShimmerBox(height: 14, width: 80)),
                          const SizedBox(width: 64),
                        ],
                      ),
                    ),
                    // Rows
                    ...List.generate(catIndex == 0 ? 8 : 6, (i) {
                      return Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        child: Row(
                          children: [
                            const SizedBox(width: 36),
                            Expanded(flex: 4, child: ShimmerBox(height: 14)),
                            const SizedBox(width: 8),
                            Expanded(flex: 3, child: ShimmerBox(height: 14)),
                            const SizedBox(width: 64),
                          ],
                        ),
                      );
                    }),
                  ],
                ),
              ),
            ],
          );
        }),
      ),
    );
  }
}

/// Skeleton for a list of items (marketplace, inbox, listings).
class ListSkeleton extends StatelessWidget {
  final int itemCount;

  const ListSkeleton({super.key, this.itemCount = 6});

  @override
  Widget build(BuildContext context) {
    return ShimmerWrap(
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        physics: const NeverScrollableScrollPhysics(),
        itemCount: itemCount,
        itemBuilder: (_, __) => Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: Row(
            children: [
              ShimmerBox(width: 56, height: 56, borderRadius: 8),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    ShimmerBox(height: 14, width: 180),
                    const SizedBox(height: 8),
                    ShimmerBox(height: 12, width: 120),
                    const SizedBox(height: 6),
                    ShimmerBox(height: 12, width: 80),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
