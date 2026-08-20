import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

/// Trang giới thiệu NHÀ PHÁT TRIỂN — quảng bá thương hiệu Trường Sơn Connect.
/// Dữ liệu (công ty + sản phẩm) lấy TỪ XA (feed dùng chung mọi app 365):
///   raw.githubusercontent.com/hahoiabc/truongson-connect-brand/main/products.json
/// → thêm/sửa sản phẩm chỉ cần sửa 1 file, mọi app tự cập nhật (có fallback offline).

/// Mã app HIỆN TẠI (để đánh dấu "Đang dùng").
const _kThisApp = 'sangiagao';
const _kFeedUrl =
    'https://raw.githubusercontent.com/hahoiabc/truongson-connect-brand/main/products.json';

// Màu thương hiệu TSC (Concept B).
const _blue = Color(0xFF0D47A1);
const _navy = Color(0xFF0A2540);
const _cyan = Color(0xFF00BCD4);

class _CongTy {
  const _CongTy(
      {required this.ten,
      required this.slogan,
      required this.gioiThieu,
      required this.website,
      required this.email,
      required this.dienThoai});
  final String ten, slogan, gioiThieu, website, email, dienThoai;
  factory _CongTy.fromJson(Map<String, dynamic> j) => _CongTy(
        ten: '${j['ten'] ?? 'Trường Sơn Connect'}',
        slogan: '${j['slogan'] ?? 'Đơn giản — Hiệu quả'}',
        gioiThieu: '${j['gioi_thieu'] ?? ''}',
        website: '${j['website'] ?? 'https://truongsonconnect.vn'}',
        email: '${j['email'] ?? 'truongsonconect2026@gmail.com'}',
        dienThoai: '${j['dien_thoai'] ?? '0968660799'}',
      );
}

class _SanPham {
  const _SanPham(this.id, this.icon, this.ten, this.moTa);
  final String id, icon, ten, moTa;
  bool get laApp => id == _kThisApp;
  factory _SanPham.fromJson(Map<String, dynamic> j) => _SanPham(
        '${j['id'] ?? ''}',
        '${j['icon'] ?? '📦'}',
        '${j['ten'] ?? ''}',
        '${j['mo_ta'] ?? ''}',
      );
}

class _Brand {
  const _Brand(this.congTy, this.sanPham);
  final _CongTy congTy;
  final List<_SanPham> sanPham;
  factory _Brand.fromJson(Map<String, dynamic> j) => _Brand(
        _CongTy.fromJson((j['cong_ty'] as Map?)?.cast<String, dynamic>() ?? {}),
        ((j['san_pham'] as List?) ?? const [])
            .map((e) => _SanPham.fromJson((e as Map).cast<String, dynamic>()))
            .toList(),
      );
}

/// Fallback nhúng sẵn (offline / feed lỗi) — luôn có nội dung để hiện.
const _fallback = _Brand(
  _CongTy(
    ten: 'Trường Sơn Connect',
    slogan: 'Đơn giản — Hiệu quả',
    gioiThieu:
        'Trường Sơn Connect xây dựng các nền tảng kết nối và quản lý cho '
        'doanh nghiệp Việt Nam: sàn giao dịch ngành hàng, quản lý bán hàng cho '
        'chuỗi lớn lẫn cửa hàng nhỏ, quản lý nhà cho thuê, kế toán số liệu doanh nghiệp, và tài chính cá nhân.',
    website: 'https://truongsonconnect.vn',
    email: 'truongsonconect2026@gmail.com',
    dienThoai: '0968660799',
  ),
  [
    _SanPham('nhachothue365', '🏠', 'NhaChoThue 365',
        'Quản lý nhà cho thuê theo giờ, ngày, tháng'),
    _SanPham('sangiagao', '🌾', 'Sàn Giá Gạo', 'Nền tảng giao dịch giá gạo'),
    _SanPham('sell365', '🏪', 'Sell 365', 'Quản lý bán hàng đa chi nhánh'),
    _SanPham('ketoan365', '📊', 'Kế toán 365', 'Quản lý số liệu doanh nghiệp'),
    _SanPham('banhang365', '🛒', 'Bán hàng 365', 'Ứng dụng bán hàng offline'),
    _SanPham('money365', '💰', 'Money 365', 'Quản lý tài chính cá nhân'),
  ],
);

/// Feed từ xa (fallback nếu lỗi). autoDispose để lần mở sau tải mới.
final _brandProvider = FutureProvider.autoDispose<_Brand>((ref) async {
  try {
    final res = await Dio().get<String>(
      _kFeedUrl,
      options: Options(
        responseType: ResponseType.plain,
        receiveTimeout: const Duration(seconds: 6),
        sendTimeout: const Duration(seconds: 6),
      ),
    );
    final data = jsonDecode(res.data ?? '{}') as Map<String, dynamic>;
    final b = _Brand.fromJson(data);
    return b.sanPham.isEmpty ? _fallback : b;
  } catch (_) {
    return _fallback; // offline / feed lỗi → dùng bản nhúng
  }
});

class NhaPhatTrienScreen extends ConsumerWidget {
  const NhaPhatTrienScreen({super.key});

  static const _web = 'https://truongsonconnect.vn';

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Loading/error → hiện fallback ngay (không spinner).
    final b = ref.watch(_brandProvider).valueOrNull ?? _fallback;
    final ct = b.congTy;
    return Scaffold(
      appBar: AppBar(title: const Text('Nhà phát triển')),
      body: ListView(
        padding: EdgeInsets.zero,
        children: [
          _banner(ct),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: Text(ct.gioiThieu,
                style: const TextStyle(fontSize: 15, height: 1.5)),
          ),
          const _MucTieuDe('Sản phẩm'),
          ...b.sanPham.map(_dongSanPham),
          const _MucTieuDe('Liên hệ'),
          _dongLienHe(context, Icons.language, 'Website',
              ct.website.replaceFirst(RegExp(r'^https?://'), ''),
              () => _moUrl(context, ct.website.isEmpty ? _web : ct.website)),
          _dongLienHe(context, Icons.email_outlined, 'Email', ct.email,
              () => _moUrl(context, 'mailto:${ct.email}')),
          _dongLienHe(context, Icons.phone_outlined, 'Điện thoại',
              _dinhDangSdt(ct.dienThoai),
              () => _moUrl(context, 'tel:${ct.dienThoai}')),
          const SizedBox(height: 24),
          Center(
            child: Text('© 2026 ${ct.ten}',
                style: const TextStyle(color: Colors.grey, fontSize: 12)),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  String _dinhDangSdt(String s) => s.length == 10
      ? '${s.substring(0, 4)}.${s.substring(4, 7)}.${s.substring(7)}'
      : s;

  Widget _banner(_CongTy ct) => Container(
        width: double.infinity,
        padding: const EdgeInsets.fromLTRB(20, 28, 20, 28),
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [_navy, _blue, _cyan],
          ),
          borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
        ),
        child: Column(
          children: [
            Container(
              width: 64,
              height: 64,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: Colors.white,
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                      color: Colors.black.withValues(alpha: 0.15),
                      blurRadius: 8,
                      offset: const Offset(0, 3)),
                ],
              ),
              child: const Text('TS',
                  style: TextStyle(
                      fontSize: 26,
                      fontWeight: FontWeight.w900,
                      color: _blue,
                      letterSpacing: 1)),
            ),
            const SizedBox(height: 14),
            RichText(
              textAlign: TextAlign.center,
              text: const TextSpan(
                style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 1,
                    color: Colors.white),
                children: [
                  TextSpan(text: 'TRƯỜNG SƠN '),
                  TextSpan(
                      text: 'CONNECT',
                      style: TextStyle(color: Color(0xFFB2FFFF))),
                ],
              ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.white24,
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(ct.slogan,
                  style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.w600)),
            ),
          ],
        ),
      );

  Widget _dongSanPham(_SanPham p) => Container(
        margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 5),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: p.laApp ? _cyan.withValues(alpha: 0.07) : Colors.transparent,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
              color: p.laApp
                  ? _cyan.withValues(alpha: 0.45)
                  : Colors.grey.withValues(alpha: 0.25)),
        ),
        child: Row(
          children: [
            Container(
              width: 44,
              height: 44,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: Colors.grey.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: Text(p.icon, style: const TextStyle(fontSize: 22)),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(p.ten,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                          fontWeight: FontWeight.w700, fontSize: 15)),
                  if (p.laApp) ...[
                    const SizedBox(height: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                          color: _blue,
                          borderRadius: BorderRadius.circular(10)),
                      child: const Text('Đang dùng',
                          style: TextStyle(
                              fontSize: 10,
                              color: Colors.white,
                              fontWeight: FontWeight.w700)),
                    ),
                  ],
                  const SizedBox(height: 3),
                  Text(p.moTa,
                      style:
                          TextStyle(fontSize: 13, color: Colors.grey.shade600)),
                ],
              ),
            ),
          ],
        ),
      );

  Widget _dongLienHe(BuildContext context, IconData icon, String nhan,
          String giaTri, VoidCallback onTap) =>
      ListTile(
        leading: Icon(icon, color: _blue),
        title:
            Text(nhan, style: const TextStyle(fontSize: 13, color: Colors.grey)),
        subtitle: Text(giaTri,
            style: const TextStyle(
                fontSize: 15, color: _navy, fontWeight: FontWeight.w600)),
        onTap: onTap,
        trailing: const Icon(Icons.open_in_new, size: 18, color: Colors.grey),
      );

  Future<void> _moUrl(BuildContext context, String url) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      final ok = await launchUrl(Uri.parse(url),
          mode: LaunchMode.externalApplication);
      if (ok) return;
    } catch (_) {}
    final text = url.replaceFirst(RegExp(r'^(mailto:|tel:)'), '');
    await Clipboard.setData(ClipboardData(text: text));
    messenger.showSnackBar(SnackBar(content: Text('Đã sao chép: $text')));
  }
}

class _MucTieuDe extends StatelessWidget {
  const _MucTieuDe(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.fromLTRB(16, 20, 16, 6),
        child: Text(text.toUpperCase(),
            style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w700,
                letterSpacing: 0.5,
                color: Colors.grey.shade600)),
      );
}
