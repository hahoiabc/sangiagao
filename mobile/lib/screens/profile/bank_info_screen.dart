import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/providers.dart';

class BankInfoScreen extends ConsumerStatefulWidget {
  const BankInfoScreen({super.key});

  @override
  ConsumerState<BankInfoScreen> createState() => _BankInfoScreenState();
}

class _BankInfoScreenState extends ConsumerState<BankInfoScreen> {
  final _accountCtrl = TextEditingController();
  final _bankCtrl = TextEditingController();
  final _holderCtrl = TextEditingController();
  final _noteCtrl = TextEditingController();
  bool _loading = true;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _accountCtrl.dispose();
    _bankCtrl.dispose();
    _holderCtrl.dispose();
    _noteCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final api = ref.read(apiServiceProvider);
    final existing = await api.getBankInfo();
    if (!mounted) return;
    if (existing != null) {
      _accountCtrl.text = (existing['account_no'] ?? '').toString();
      _bankCtrl.text = (existing['bank_name'] ?? '').toString();
      _holderCtrl.text = (existing['holder_name'] ?? '').toString();
      _noteCtrl.text = (existing['note'] ?? '').toString();
    }
    setState(() => _loading = false);
  }

  Future<void> _save() async {
    final acc = _accountCtrl.text.trim();
    final bank = _bankCtrl.text.trim();
    final holder = _holderCtrl.text.trim();
    if (acc.length < 4 || bank.length < 2 || holder.length < 2) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Vui lòng nhập đầy đủ thông tin')),
      );
      return;
    }
    setState(() => _saving = true);
    try {
      await ref.read(apiServiceProvider).upsertBankInfo(
            accountNo: acc,
            bankName: bank,
            holderName: holder,
            note: _noteCtrl.text.trim().isEmpty ? null : _noteCtrl.text.trim(),
          );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã lưu thông tin tài khoản')),
      );
      Navigator.pop(context, true);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Lưu thất bại, vui lòng thử lại')),
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Tài khoản nhận hoa hồng')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Card(
                  color: Colors.amber.shade50,
                  child: const Padding(
                    padding: EdgeInsets.all(12),
                    child: Text(
                      'Sàn sẽ chuyển hoa hồng vào tài khoản này khi bạn đạt ngưỡng tối thiểu. '
                      'Phí chuyển khoản thực tế (nếu có) sẽ trừ trực tiếp vào số tiền nhận.',
                      style: TextStyle(fontSize: 13),
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _accountCtrl,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Số tài khoản *',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _bankCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Tên ngân hàng (vd: Vietcombank, MB Bank) *',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _holderCtrl,
                  textCapitalization: TextCapitalization.characters,
                  decoration: const InputDecoration(
                    labelText: 'Chủ tài khoản (VIẾT HOA KHÔNG DẤU) *',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _noteCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Ghi chú (tuỳ chọn)',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 20),
                FilledButton(
                  onPressed: _saving ? null : _save,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Text(_saving ? 'Đang lưu…' : 'Lưu thông tin'),
                  ),
                ),
              ],
            ),
    );
  }
}
