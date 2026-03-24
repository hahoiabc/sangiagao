import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

// ============================================
// SanGiaGao - Design System
// ============================================
// Phong cách: Agriculture Premium - sạch, chuyên nghiệp, đáng tin cậy
// Bảng màu lấy cảm hứng từ đồng lúa Việt Nam

class AppColors {
  AppColors._();

  // --- Primary: Xanh lá đậm (lúa non) ---
  static const primary = Color(0xFF2E7D32);
  static const primaryLight = Color(0xFF4CAF50);
  static const primaryDark = Color(0xFF1B5E20);
  static const onPrimary = Colors.white;

  // --- Secondary: Vàng lúa chín ---
  static const secondary = Color(0xFFF9A825);
  static const secondaryLight = Color(0xFFFDD835);
  static const secondaryDark = Color(0xFFF57F17);

  // --- Surfaces ---
  static const background = Color(0xFFF8F9F4); // warm white
  static const surface = Colors.white;
  static const surfaceVariant = Color(0xFFF0F2EB);
  static const cardBg = Colors.white;

  // --- Semantic ---
  static const success = Color(0xFF2E7D32);
  static const warning = Color(0xFFEF6C00);
  static const error = Color(0xFFD32F2F);
  static const info = Color(0xFF1565C0);

  // --- Price ---
  static const priceText = Color(0xFF1B5E20);
  static const priceHighlight = Color(0xFF2E7D32);

  // --- Text ---
  static const textPrimary = Color(0xFF1A1C1E);
  static const textSecondary = Color(0xFF5F6368);
  static const textHint = Color(0xFF9AA0A6);
  static const textOnDark = Colors.white;

  // --- Borders & Dividers ---
  static const border = Color(0xFFE0E3DB);
  static const divider = Color(0xFFEBEDE6);

  // --- Status badges ---
  static const activeGreen = Color(0xFF2E7D32);
  static const hiddenOrange = Color(0xFFEF6C00);
  static const deletedRed = Color(0xFFD32F2F);

  // --- Chat ---
  static const chatBubbleMe = Color(0xFF2E7D32);
  static const chatBubbleOther = Color(0xFFF0F2EB);
  static const onlineGreen = Color(0xFF4CAF50);
  static const offlineGrey = Color(0xFFBDBDBD);
}

class AppTheme {
  AppTheme._();

  static ThemeData withPrimary(Color primary, Color primaryDark, Color primaryLight) {
    // Derive tinted surface colors from primary
    final scaffoldBg = Color.lerp(Colors.white, primary, 0.04)!;
    final containerLow = Color.lerp(Colors.white, primary, 0.03)!;
    final container = Color.lerp(Colors.white, primary, 0.05)!;
    final containerHigh = Color.lerp(Colors.white, primary, 0.07)!;
    final containerHighest = Color.lerp(Colors.white, primary, 0.09)!;
    final borderColor = Color.lerp(Colors.white, primary, 0.12)!;
    final dividerColor = Color.lerp(Colors.white, primary, 0.08)!;

    final colorScheme = ColorScheme(
      brightness: Brightness.light,
      primary: primary,
      onPrimary: AppColors.onPrimary,
      primaryContainer: primaryLight.withValues(alpha: 0.3),
      onPrimaryContainer: primaryDark,
      secondary: AppColors.secondary,
      onSecondary: Colors.black,
      secondaryContainer: const Color(0xFFFFF8E1),
      onSecondaryContainer: const Color(0xFF3E2723),
      tertiary: AppColors.info,
      onTertiary: Colors.white,
      error: AppColors.error,
      onError: Colors.white,
      errorContainer: const Color(0xFFFFDAD6),
      onErrorContainer: const Color(0xFF410002),
      surface: AppColors.surface,
      onSurface: AppColors.textPrimary,
      onSurfaceVariant: AppColors.textSecondary,
      outline: borderColor,
      outlineVariant: dividerColor,
      surfaceContainerHighest: containerHighest,
      surfaceContainerHigh: containerHigh,
      surfaceContainerLow: containerLow,
      surfaceContainer: container,
    );

    return _buildTheme(colorScheme, primary, primaryDark, scaffoldBg: scaffoldBg);
  }

  static ThemeData get light {
    const colorScheme = ColorScheme(
      brightness: Brightness.light,
      primary: AppColors.primary,
      onPrimary: AppColors.onPrimary,
      primaryContainer: Color(0xFFB9F6CA),
      onPrimaryContainer: Color(0xFF002106),
      secondary: AppColors.secondary,
      onSecondary: Colors.black,
      secondaryContainer: Color(0xFFFFF8E1),
      onSecondaryContainer: Color(0xFF3E2723),
      tertiary: AppColors.info,
      onTertiary: Colors.white,
      error: AppColors.error,
      onError: Colors.white,
      errorContainer: Color(0xFFFFDAD6),
      onErrorContainer: Color(0xFF410002),
      surface: AppColors.surface,
      onSurface: AppColors.textPrimary,
      onSurfaceVariant: AppColors.textSecondary,
      outline: AppColors.border,
      outlineVariant: AppColors.divider,
      surfaceContainerHighest: Color(0xFFE8EBE0),
      surfaceContainerHigh: Color(0xFFEDF0E5),
      surfaceContainerLow: Color(0xFFF5F7F0),
      surfaceContainer: Color(0xFFF0F2EB),
    );

    return _buildTheme(colorScheme, AppColors.primary, AppColors.primaryDark);
  }

  static ThemeData _buildTheme(ColorScheme colorScheme, Color primary, Color primaryDark, {Color? scaffoldBg}) {
    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: scaffoldBg ?? AppColors.background,
      fontFamily: null, // system font for Vietnamese

      // --- AppBar ---
      appBarTheme: AppBarTheme(
        elevation: 0,
        scrolledUnderElevation: 1,
        centerTitle: true,
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        surfaceTintColor: Colors.transparent,
        shadowColor: AppColors.border,
        titleTextStyle: TextStyle(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: AppColors.textPrimary,
          letterSpacing: -0.3,
        ),
        iconTheme: IconThemeData(color: AppColors.textPrimary, size: 22),
        systemOverlayStyle: SystemUiOverlayStyle(
          statusBarColor: Colors.transparent,
          statusBarIconBrightness: Brightness.dark,
          statusBarBrightness: Brightness.light,
        ),
      ),

      // --- Card ---
      cardTheme: CardThemeData(
        elevation: 0,
        color: AppColors.cardBg,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: AppColors.divider, width: 1),
        ),
        margin: EdgeInsets.zero,
        clipBehavior: Clip.antiAlias,
      ),

      // --- NavigationBar ---
      navigationBarTheme: NavigationBarThemeData(
        elevation: 0,
        height: 68,
        backgroundColor: AppColors.surface,
        surfaceTintColor: Colors.transparent,
        shadowColor: AppColors.border,
        indicatorColor: primary.withValues(alpha: 0.12),
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: primary,
              letterSpacing: 0.2,
            );
          }
          return const TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: AppColors.textHint,
            letterSpacing: 0.2,
          );
        }),
        iconTheme: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return IconThemeData(color: primary, size: 24);
          }
          return const IconThemeData(color: AppColors.textHint, size: 24);
        }),
      ),

      // --- Buttons ---
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: primary,
          foregroundColor: AppColors.onPrimary,
          minimumSize: const Size(double.infinity, 52),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, letterSpacing: 0.3),
          elevation: 0,
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: primary,
          minimumSize: const Size(double.infinity, 52),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          side: BorderSide(color: primary, width: 1.5),
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, letterSpacing: 0.3),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: primary,
          textStyle: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        ),
      ),

      // --- Input ---
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.surfaceVariant.withValues(alpha: 0.5),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: primary, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.error, width: 2),
        ),
        labelStyle: const TextStyle(color: AppColors.textSecondary, fontSize: 14),
        hintStyle: const TextStyle(color: AppColors.textHint, fontSize: 14),
        prefixIconColor: AppColors.textHint,
        suffixIconColor: AppColors.textHint,
        floatingLabelStyle: TextStyle(color: primary, fontWeight: FontWeight.w500),
      ),

      // --- Chip ---
      chipTheme: ChipThemeData(
        backgroundColor: AppColors.surfaceVariant,
        selectedColor: primary.withValues(alpha: 0.12),
        labelStyle: const TextStyle(fontSize: 13, color: AppColors.textPrimary),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        side: const BorderSide(color: AppColors.border),
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      ),

      // --- Dialog ---
      dialogTheme: DialogThemeData(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        surfaceTintColor: Colors.transparent,
        backgroundColor: AppColors.surface,
        titleTextStyle: const TextStyle(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: AppColors.textPrimary,
        ),
      ),

      // --- BottomSheet ---
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: AppColors.surface,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
        showDragHandle: true,
        dragHandleColor: AppColors.border,
      ),

      // --- SnackBar ---
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        backgroundColor: AppColors.textPrimary,
        contentTextStyle: const TextStyle(color: Colors.white, fontSize: 14),
      ),

      // --- ListTile ---
      listTileTheme: const ListTileThemeData(
        contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        minVerticalPadding: 12,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
        ),
        titleTextStyle: TextStyle(
          fontSize: 15,
          fontWeight: FontWeight.w500,
          color: AppColors.textPrimary,
        ),
        subtitleTextStyle: TextStyle(
          fontSize: 13,
          color: AppColors.textSecondary,
        ),
      ),

      // --- Divider ---
      dividerTheme: const DividerThemeData(
        color: AppColors.divider,
        thickness: 1,
        space: 1,
      ),

      // --- Badge ---
      badgeTheme: const BadgeThemeData(
        backgroundColor: AppColors.error,
        textColor: Colors.white,
        smallSize: 8,
        largeSize: 18,
        textStyle: TextStyle(fontSize: 11, fontWeight: FontWeight.w600),
      ),

      // --- Text ---
      textTheme: const TextTheme(
        headlineLarge: TextStyle(fontSize: 28, fontWeight: FontWeight.w700, letterSpacing: -0.5, color: AppColors.textPrimary),
        headlineMedium: TextStyle(fontSize: 24, fontWeight: FontWeight.w700, letterSpacing: -0.3, color: AppColors.textPrimary),
        headlineSmall: TextStyle(fontSize: 20, fontWeight: FontWeight.w600, letterSpacing: -0.3, color: AppColors.textPrimary),
        titleLarge: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
        titleMedium: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
        titleSmall: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
        bodyLarge: TextStyle(fontSize: 16, fontWeight: FontWeight.w400, color: AppColors.textPrimary, height: 1.5),
        bodyMedium: TextStyle(fontSize: 14, fontWeight: FontWeight.w400, color: AppColors.textPrimary, height: 1.5),
        bodySmall: TextStyle(fontSize: 12, fontWeight: FontWeight.w400, color: AppColors.textSecondary, height: 1.4),
        labelLarge: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
        labelMedium: TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: AppColors.textSecondary),
        labelSmall: TextStyle(fontSize: 11, fontWeight: FontWeight.w500, color: AppColors.textHint, letterSpacing: 0.3),
      ),

      // --- FloatingActionButton ---
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: primary,
        foregroundColor: AppColors.onPrimary,
        elevation: 3,
        shape: const CircleBorder(),
      ),

      // --- TabBar ---
      tabBarTheme: TabBarThemeData(
        labelColor: primary,
        unselectedLabelColor: AppColors.textHint,
        indicatorColor: primary,
        labelStyle: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        unselectedLabelStyle: const TextStyle(fontSize: 14, fontWeight: FontWeight.w400),
      ),

      // --- ProgressIndicator ---
      progressIndicatorTheme: ProgressIndicatorThemeData(
        color: primary,
        linearTrackColor: AppColors.divider,
      ),

      // --- Switch ---
      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return primary;
          return AppColors.textHint;
        }),
        trackColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return primary.withValues(alpha: 0.3);
          return AppColors.border;
        }),
      ),
    );
  }
}

// --- Reusable style helpers ---
class AppStyles {
  AppStyles._();

  // Price text style
  static TextStyle price({double size = 16, FontWeight weight = FontWeight.w700}) => TextStyle(
        fontSize: size,
        fontWeight: weight,
        color: AppColors.priceText,
        letterSpacing: -0.3,
      );

  // Status badge decoration
  static BoxDecoration statusBadge(String status) {
    Color bg;
    switch (status) {
      case 'active':
        bg = AppColors.activeGreen;
      case 'hidden':
        bg = AppColors.hiddenOrange;
      case 'deleted':
      case 'expired':
        bg = AppColors.deletedRed;
      default:
        bg = AppColors.textHint;
    }
    return BoxDecoration(
      color: bg.withValues(alpha: 0.1),
      borderRadius: BorderRadius.circular(8),
    );
  }

  static Color statusColor(String status) {
    switch (status) {
      case 'active':
        return AppColors.activeGreen;
      case 'hidden':
        return AppColors.hiddenOrange;
      case 'deleted':
      case 'expired':
        return AppColors.deletedRed;
      default:
        return AppColors.textHint;
    }
  }

  // Section header
  static TextStyle sectionHeader = const TextStyle(
    fontSize: 16,
    fontWeight: FontWeight.w700,
    color: AppColors.textPrimary,
    letterSpacing: -0.2,
  );

  // Subtle card shadow
  static List<BoxShadow> get cardShadow => [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.04),
          blurRadius: 8,
          offset: const Offset(0, 2),
        ),
      ];

  // Gradient for subscription/premium sections
  static LinearGradient get primaryGradient => const LinearGradient(
        colors: [AppColors.primaryDark, AppColors.primary, AppColors.primaryLight],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      );

  /// Dynamic gradient that follows the current theme
  static LinearGradient primaryGradientOf(ColorScheme cs) => LinearGradient(
        colors: [cs.onPrimaryContainer, cs.primary, cs.primaryContainer.withValues(alpha: 1.0)],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      );

  static LinearGradient get warningGradient => const LinearGradient(
        colors: [Color(0xFFE65100), AppColors.warning, Color(0xFFFFA726)],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      );
}
