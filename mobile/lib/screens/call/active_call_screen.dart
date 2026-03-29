import 'dart:async';
import 'package:flutter/material.dart';
import '../../services/call_service.dart';

class ActiveCallScreen extends StatefulWidget {
  final CallService callService;

  const ActiveCallScreen({super.key, required this.callService});

  @override
  State<ActiveCallScreen> createState() => _ActiveCallScreenState();
}

class _ActiveCallScreenState extends State<ActiveCallScreen> {
  int _duration = 0;
  CallState _state = CallState.idle;

  CallService get _call => widget.callService;

  @override
  void initState() {
    super.initState();
    _state = _call.state;

    _call.onStateChanged = (state) {
      if (!mounted) return;
      setState(() => _state = state);
      if (state == CallState.ended) {
        Future.delayed(const Duration(seconds: 1), () {
          if (mounted) Navigator.of(context).pop();
        });
      }
    };

    _call.onDurationUpdate = (seconds) {
      if (mounted) setState(() => _duration = seconds);
    };
  }

  @override
  void dispose() {
    _call.dispose();
    super.dispose();
  }

  String _formatDuration(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
  }

  String get _statusText {
    switch (_state) {
      case CallState.outgoing:
        return 'Đang gọi...';
      case CallState.incoming:
        return 'Cuộc gọi đến...';
      case CallState.connecting:
        return 'Đang kết nối...';
      case CallState.connected:
        return _formatDuration(_duration);
      case CallState.ended:
        return 'Đã kết thúc';
      default:
        return '';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF1a1a2e),
      body: SafeArea(
        child: Column(
          children: [
            const Spacer(flex: 2),

            // Avatar circle
            CircleAvatar(
              radius: 50,
              backgroundColor: Colors.white24,
              child: Text(
                (_call.otherUserName.isNotEmpty ? _call.otherUserName[0] : '?').toUpperCase(),
                style: const TextStyle(fontSize: 40, color: Colors.white),
              ),
            ),
            const SizedBox(height: 16),

            // Name
            Text(
              _call.otherUserName,
              style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.white),
            ),
            const SizedBox(height: 8),

            // Status / Duration
            Text(
              _statusText,
              style: TextStyle(
                fontSize: 16,
                color: _state == CallState.connected ? Colors.greenAccent : Colors.white70,
              ),
            ),

            const Spacer(flex: 3),

            // Controls
            if (_state == CallState.incoming) _buildIncomingControls(),
            if (_state != CallState.incoming && _state != CallState.ended) _buildActiveControls(),

            const SizedBox(height: 48),
          ],
        ),
      ),
    );
  }

  Widget _buildIncomingControls() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        // Reject
        _buildCircleButton(
          icon: Icons.call_end,
          color: Colors.red,
          label: 'Từ chối',
          onTap: () => _call.rejectCall(),
        ),
        // Accept
        _buildCircleButton(
          icon: Icons.call,
          color: Colors.green,
          label: 'Nghe',
          onTap: () {
            // This should have remote SDP from signaling
            // For now, start will handle it
            _call.start();
          },
        ),
      ],
    );
  }

  Widget _buildActiveControls() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        // Mute
        _buildCircleButton(
          icon: _call.isMuted ? Icons.mic_off : Icons.mic,
          color: _call.isMuted ? Colors.orange : Colors.white24,
          label: _call.isMuted ? 'Bật mic' : 'Tắt mic',
          onTap: () {
            _call.toggleMute();
            setState(() {});
          },
        ),
        // End call
        _buildCircleButton(
          icon: Icons.call_end,
          color: Colors.red,
          size: 70,
          label: 'Kết thúc',
          onTap: () => _call.endCall(),
        ),
        // Speaker
        _buildCircleButton(
          icon: _call.isSpeaker ? Icons.volume_up : Icons.volume_down,
          color: _call.isSpeaker ? Colors.blue : Colors.white24,
          label: _call.isSpeaker ? 'Tai nghe' : 'Loa ngoài',
          onTap: () {
            _call.toggleSpeaker();
            setState(() {});
          },
        ),
      ],
    );
  }

  Widget _buildCircleButton({
    required IconData icon,
    required Color color,
    required String label,
    required VoidCallback onTap,
    double size = 56,
  }) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        GestureDetector(
          onTap: onTap,
          child: Container(
            width: size,
            height: size,
            decoration: BoxDecoration(shape: BoxShape.circle, color: color),
            child: Icon(icon, color: Colors.white, size: size * 0.45),
          ),
        ),
        const SizedBox(height: 8),
        Text(label, style: const TextStyle(color: Colors.white70, fontSize: 12)),
      ],
    );
  }
}
