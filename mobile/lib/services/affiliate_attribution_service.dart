import 'dart:io';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:play_install_referrer/play_install_referrer.dart';

/// Reads Google Play Install Referrer once per install and caches the resulting
/// referral code in secure storage. Called automatically on first register
/// attempt — user does not need to copy/paste anything.
///
/// iOS: Play Install Referrer is Android-only. iOS attribution falls back to
/// Universal Link handling (separate path, not implemented here yet).
class AffiliateAttributionService {
  static const _storageKey = 'aff_attribution_code';
  static const _storageCheckedKey = 'aff_attribution_checked';
  static const _storage = FlutterSecureStorage();
  static final _codeRegex = RegExp(r'^[A-Z0-9]{4,8}$');

  /// Returns the cached or freshly-resolved referral code (or null).
  /// Safe to call multiple times — Play Install Referrer queries only run
  /// once. After consuming, call [clear] to prevent re-attribution.
  static Future<String?> getCode() async {
    // Cached?
    final cached = await _storage.read(key: _storageKey);
    if (cached != null && cached.isNotEmpty) return cached;

    // Only fetch from Play API once per install (idempotent best-effort).
    final checked = await _storage.read(key: _storageCheckedKey);
    if (checked == 'true') return null;
    await _storage.write(key: _storageCheckedKey, value: 'true');

    if (!Platform.isAndroid) return null;

    try {
      final details = await PlayInstallReferrer.installReferrer;
      final raw = details.installReferrer;
      if (raw == null || raw.isEmpty) return null;
      final code = _extractCode(raw);
      if (code == null) return null;
      await _storage.write(key: _storageKey, value: code);
      return code;
    } catch (_) {
      return null;
    }
  }

  /// Removes cached code — call after successful attribution so a future
  /// fresh-install/sign-in doesn't re-use a stale referrer.
  static Future<void> clear() async {
    await _storage.delete(key: _storageKey);
  }

  /// Manually persist a code captured from outside (vd: deep link, web cookie).
  static Future<void> setCode(String code) async {
    final normalized = _extractCode(code);
    if (normalized != null) {
      await _storage.write(key: _storageKey, value: normalized);
    }
  }

  /// Parse an incoming Universal Link / deep link URI and save the code if
  /// it matches /r/{code}. Returns the extracted code or null.
  ///
  /// Examples that match:
  ///   https://sangiagao.vn/r/HF9D37
  ///   https://sangiagao.vn/cai-app?ref=HF9D37
  static Future<String?> handleDeepLink(Uri uri) async {
    String? candidate;
    final segs = uri.pathSegments;
    if (segs.length >= 2 && segs[0] == 'r') {
      candidate = segs[1];
    } else if (uri.path.contains('cai-app')) {
      candidate = uri.queryParameters['ref'];
    }
    if (candidate == null) return null;
    final code = _extractCode(candidate);
    if (code == null) return null;
    await _storage.write(key: _storageKey, value: code);
    return code;
  }

  /// Parse code from raw referrer string. Accepts:
  ///   - bare code: "HF9D37"
  ///   - querystring: "referrer=HF9D37&utm_source=..."
  ///   - any token containing 4–8 A-Z/0-9 chars
  static String? _extractCode(String raw) {
    final trimmed = raw.trim();
    final upper = trimmed.toUpperCase();
    if (_codeRegex.hasMatch(upper)) return upper;

    // Treat as querystring
    try {
      final params = Uri.splitQueryString(trimmed);
      final candidate = (params['referrer'] ?? params['ref'] ?? '').toUpperCase();
      if (_codeRegex.hasMatch(candidate)) return candidate;
    } catch (_) {}
    return null;
  }
}
