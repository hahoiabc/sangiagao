import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:url_launcher/url_launcher.dart';

/// Simple "share Sàn Giá Gạo" screen for ALL users (member + aff + others).
/// Link does NOT include an affiliate code — purely brand invite.
class ShareAppScreen extends StatelessWidget {
  const ShareAppScreen({super.key});

  static const String _link = 'https://sangiagao.vn/cai-app';
  static const String _message =
      'Tải SanGiaGao để xem giá Gạo và kết nối với thương nhân\n$_link';

  Future<void> _shareZalo(BuildContext context) async {
    await Clipboard.setData(const ClipboardData(text: _message));
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Đã sao chép. Mở Zalo và dán vào tin nhắn'),
          duration: Duration(seconds: 3),
        ),
      );
    }
    await launchUrl(Uri.parse('https://zalo.me/'), mode: LaunchMode.externalApplication);
  }

  Future<void> _shareFacebook() async {
    final url = Uri.parse(
      'https://www.facebook.com/sharer/sharer.php?u=${Uri.encodeComponent(_link)}',
    );
    await launchUrl(url, mode: LaunchMode.externalApplication);
  }

  void _copyLink(BuildContext context) {
    Clipboard.setData(const ClipboardData(text: _link));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Đã sao chép link'), duration: Duration(seconds: 1)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Giới thiệu bạn bè')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Chia sẻ Sàn Giá Gạo với bạn bè',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade100,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Row(
                      children: [
                        const Expanded(
                          child: Text(
                            _link,
                            style: TextStyle(color: Colors.blue, fontSize: 14),
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.copy, size: 20),
                          tooltip: 'Sao chép',
                          onPressed: () => _copyLink(context),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: FilledButton.icon(
                          style: FilledButton.styleFrom(
                            backgroundColor: const Color(0xFF0068FF),
                          ),
                          icon: const Icon(Icons.chat_bubble_outline, size: 18),
                          label: const Text('Zalo'),
                          onPressed: () => _shareZalo(context),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: FilledButton.icon(
                          style: FilledButton.styleFrom(
                            backgroundColor: const Color(0xFF1877F2),
                          ),
                          icon: const Icon(Icons.facebook, size: 18),
                          label: const Text('Facebook'),
                          onPressed: _shareFacebook,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: OutlinedButton.icon(
                          icon: const Icon(Icons.copy, size: 18),
                          label: const Text('Sao chép'),
                          onPressed: () => _copyLink(context),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
