import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';

/// Simple "share Sàn Giá Gạo" screen for ALL users (member + aff + others).
/// Link does NOT include an affiliate code — purely brand invite.
class ShareAppScreen extends StatelessWidget {
  const ShareAppScreen({super.key});

  static const String _link = 'https://sangiagao.vn';
  static const String _message =
      'Mình đang dùng Sàn Giá Gạo để xem giá gạo realtime và mua bán trực tiếp với thương lái. '
      'Bạn cài thử nhé: $_link';

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
                  const SizedBox(height: 8),
                  const Text(
                    'Gửi link dưới đây để mời bạn bè tham gia. Bạn có thể đăng ký làm Đối tác Affiliate '
                    'để nhận hoa hồng khi giới thiệu thành công.',
                    style: TextStyle(fontSize: 13, height: 1.4),
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
                          onPressed: () {
                            Clipboard.setData(const ClipboardData(text: _link));
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Đã sao chép link'), duration: Duration(seconds: 1)),
                            );
                          },
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),
                  SizedBox(
                    width: double.infinity,
                    child: FilledButton.icon(
                      icon: const Icon(Icons.share),
                      label: const Text('Chia sẻ ngay'),
                      onPressed: () => Share.share(_message, subject: 'Sàn Giá Gạo'),
                    ),
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
