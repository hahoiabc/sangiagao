import 'dart:async';
import 'package:flutter/material.dart';
import 'package:audioplayers/audioplayers.dart';

/// Full-screen overlay shown when an incoming call arrives while app is in foreground.
class IncomingCallScreen extends StatefulWidget {
  final String callerName;
  final String callType;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  const IncomingCallScreen({
    super.key,
    required this.callerName,
    required this.callType,
    required this.onAccept,
    required this.onReject,
  });

  @override
  State<IncomingCallScreen> createState() => _IncomingCallScreenState();
}

class _IncomingCallScreenState extends State<IncomingCallScreen>
    with SingleTickerProviderStateMixin {
  final AudioPlayer _ringtonePlayer = AudioPlayer();
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;
  Timer? _autoDeclineTimer;

  @override
  void initState() {
    super.initState();

    // Pulse animation for avatar
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat(reverse: true);
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.15).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );

    _playRingtone();

    // Auto-decline after 45 seconds
    _autoDeclineTimer = Timer(const Duration(seconds: 45), () {
      if (mounted) widget.onReject();
    });
  }

  void _playRingtone() {
    _ringtonePlayer.setReleaseMode(ReleaseMode.loop);
    _ringtonePlayer.play(
      AssetSource('sounds/ringtone.wav'),
      volume: 0.8,
    ).catchError((_) {
      // No ringtone file — silent
    });
  }

  @override
  void dispose() {
    _autoDeclineTimer?.cancel();
    _pulseController.dispose();
    _ringtonePlayer.stop();
    _ringtonePlayer.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF1a1a2e),
      body: SafeArea(
        child: Column(
          children: [
            const Spacer(flex: 2),

            // Caller name label
            const Text(
              'Cuộc gọi đến',
              style: TextStyle(fontSize: 16, color: Colors.white70),
            ),
            const SizedBox(height: 16),

            // Pulsing avatar
            ScaleTransition(
              scale: _pulseAnimation,
              child: CircleAvatar(
                radius: 55,
                backgroundColor: Colors.white24,
                child: Text(
                  (widget.callerName.isNotEmpty ? widget.callerName[0] : '?')
                      .toUpperCase(),
                  style: const TextStyle(fontSize: 44, color: Colors.white),
                ),
              ),
            ),
            const SizedBox(height: 20),

            // Caller name
            Text(
              widget.callerName,
              style: const TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              widget.callType == 'video' ? 'Gọi video' : 'Gọi thoại',
              style: const TextStyle(fontSize: 16, color: Colors.white54),
            ),

            const Spacer(flex: 3),

            // Accept / Reject buttons
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 48),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  _buildButton(
                    icon: Icons.call_end,
                    color: Colors.red,
                    label: 'Từ chối',
                    onTap: () {
                      _ringtonePlayer.stop();
                      widget.onReject();
                    },
                  ),
                  _buildButton(
                    icon: Icons.call,
                    color: Colors.green,
                    label: 'Nghe',
                    onTap: () {
                      _ringtonePlayer.stop();
                      widget.onAccept();
                    },
                  ),
                ],
              ),
            ),

            const SizedBox(height: 60),
          ],
        ),
      ),
    );
  }

  Widget _buildButton({
    required IconData icon,
    required Color color,
    required String label,
    required VoidCallback onTap,
  }) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        GestureDetector(
          onTap: onTap,
          child: Container(
            width: 64,
            height: 64,
            decoration: BoxDecoration(shape: BoxShape.circle, color: color),
            child: Icon(icon, color: Colors.white, size: 30),
          ),
        ),
        const SizedBox(height: 8),
        Text(label, style: const TextStyle(color: Colors.white70, fontSize: 13)),
      ],
    );
  }
}
