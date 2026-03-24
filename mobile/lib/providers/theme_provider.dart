import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class ThemeOption {
  final String key;
  final String label;
  final Color primary;
  final Color primaryDark;
  final Color primaryLight;

  const ThemeOption({
    required this.key,
    required this.label,
    required this.primary,
    required this.primaryDark,
    required this.primaryLight,
  });
}

const themeOptions = <ThemeOption>[
  ThemeOption(key: 'green', label: 'Xanh lá', primary: Color(0xFF2E7D32), primaryDark: Color(0xFF1B5E20), primaryLight: Color(0xFF4CAF50)),
  ThemeOption(key: 'teal', label: 'Xanh ngọc', primary: Color(0xFF339999), primaryDark: Color(0xFF267373), primaryLight: Color(0xFF4DB3B3)),
  ThemeOption(key: 'blue', label: 'Xanh dương', primary: Color(0xFF3399FF), primaryDark: Color(0xFF2673BF), primaryLight: Color(0xFF66B3FF)),
  ThemeOption(key: 'mint', label: 'Teal', primary: Color(0xFF33CC99), primaryDark: Color(0xFF269973), primaryLight: Color(0xFF66D9B3)),
  ThemeOption(key: 'gray', label: 'Xám đậm', primary: Color(0xFF444444), primaryDark: Color(0xFF333333), primaryLight: Color(0xFF666666)),
];

const _storageKey = 'sgg_theme_color';

class ThemeNotifier extends StateNotifier<ThemeOption> {
  static const _storage = FlutterSecureStorage();

  ThemeNotifier() : super(themeOptions[0]) {
    _load();
  }

  Future<void> _load() async {
    final key = await _storage.read(key: _storageKey) ?? 'green';
    final option = themeOptions.firstWhere((t) => t.key == key, orElse: () => themeOptions[0]);
    state = option;
  }

  Future<void> setTheme(String key) async {
    final option = themeOptions.firstWhere((t) => t.key == key, orElse: () => themeOptions[0]);
    state = option;
    await _storage.write(key: _storageKey, value: key);
  }
}

final themeProvider = StateNotifierProvider<ThemeNotifier, ThemeOption>(
  (ref) => ThemeNotifier(),
);
