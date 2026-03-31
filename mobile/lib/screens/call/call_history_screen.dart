import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';

class CallHistoryScreen extends ConsumerStatefulWidget {
  final String conversationId;
  final String currentUserId;

  const CallHistoryScreen({
    super.key,
    required this.conversationId,
    required this.currentUserId,
  });

  @override
  ConsumerState<CallHistoryScreen> createState() => _CallHistoryScreenState();
}

class _CallHistoryScreenState extends ConsumerState<CallHistoryScreen> {
  List<Map<String, dynamic>> _calls = [];
  bool _loading = true;
  final int _page = 1;
  // ignore: unused_field
  int _total = 0;

  @override
  void initState() {
    super.initState();
    _loadCalls();
  }

  Future<void> _loadCalls() async {
    setState(() => _loading = true);
    try {
      final api = ref.read(apiServiceProvider);
      final result = await api.getCallHistory(widget.conversationId, page: _page);
      final data = result['data'] as List? ?? [];
      setState(() {
        _calls = data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
        _total = result['total'] as int? ?? 0;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  String _statusText(Map<String, dynamic> call) {
    final status = call['status'] as String? ?? '';
    final isOutgoing = call['caller_id'] == widget.currentUserId;
    switch (status) {
      case 'answered':
        return 'Đã nghe - ${_durationText(call)}';
      case 'missed':
        return isOutgoing ? 'Không trả lời' : 'Cuộc gọi nhỡ';
      case 'rejected':
        return isOutgoing ? 'Bị từ chối' : 'Đã từ chối';
      case 'busy':
        return 'Đang bận';
      case 'initiated':
        return 'Không kết nối';
      default:
        return status;
    }
  }

  String _durationText(Map<String, dynamic> call) {
    final seconds = call['duration_seconds'] as int? ?? 0;
    if (seconds == 0) return '';
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '${m}p${s.toString().padLeft(2, '0')}s';
  }

  IconData _statusIcon(Map<String, dynamic> call) {
    final status = call['status'] as String? ?? '';
    final isOutgoing = call['caller_id'] == widget.currentUserId;
    switch (status) {
      case 'answered':
        return Icons.call;
      case 'missed':
        return isOutgoing ? Icons.call_made : Icons.call_missed;
      case 'rejected':
        return Icons.call_end;
      default:
        return Icons.call_missed_outgoing;
    }
  }

  Color _statusColor(Map<String, dynamic> call) {
    final status = call['status'] as String? ?? '';
    switch (status) {
      case 'answered':
        return Colors.green;
      case 'missed':
        return Colors.red;
      case 'rejected':
        return Colors.orange;
      default:
        return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Lịch sử cuộc gọi')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _calls.isEmpty
              ? const Center(child: Text('Chưa có cuộc gọi nào'))
              : ListView.builder(
                  itemCount: _calls.length,
                  itemBuilder: (context, index) {
                    final call = _calls[index];
                    final isOutgoing = call['caller_id'] == widget.currentUserId;
                    final otherName = isOutgoing
                        ? (call['callee_name'] ?? 'Người nhận')
                        : (call['caller_name'] ?? 'Người gọi');
                    final createdAt = DateTime.tryParse(call['created_at'] ?? '');
                    final timeStr = createdAt != null
                        ? DateFormat('dd/MM HH:mm').format(createdAt.toLocal())
                        : '';

                    return ListTile(
                      leading: Icon(
                        _statusIcon(call),
                        color: _statusColor(call),
                      ),
                      title: Row(
                        children: [
                          Icon(
                            isOutgoing ? Icons.call_made : Icons.call_received,
                            size: 14,
                            color: isOutgoing ? Colors.blue : Colors.green,
                          ),
                          const SizedBox(width: 6),
                          Expanded(child: Text(otherName)),
                        ],
                      ),
                      subtitle: Text(_statusText(call)),
                      trailing: Text(
                        timeStr,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    );
                  },
                ),
    );
  }
}
