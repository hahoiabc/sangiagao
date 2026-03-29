import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';
import '../config/env.dart';

/// Manages Phoenix WebSocket connection for call signaling (call:* channel).
class CallSignalingService {
  final String token;
  final String conversationId;

  WebSocketChannel? _channel;
  StreamSubscription? _subscription;
  Timer? _heartbeat;
  int _ref = 0;
  bool _joined = false;

  // Event callbacks
  void Function(Map<String, dynamic>)? onCallInitiate;
  void Function(Map<String, dynamic>)? onCallOffer;
  void Function(Map<String, dynamic>)? onCallAnswer;
  void Function(Map<String, dynamic>)? onIceCandidate;
  void Function(Map<String, dynamic>)? onCallEnd;
  void Function(Map<String, dynamic>)? onCallReject;
  void Function(Map<String, dynamic>)? onCallBusy;
  void Function(Map<String, dynamic>)? onCallReady;
  void Function()? onCallTimeout;
  void Function()? onDisconnected;

  CallSignalingService({required this.token, required this.conversationId});

  String get _topic => 'call:$conversationId';

  void connect() {
    final wsUrl = '${Env.wsBaseUrl}?token=$token';

    try {
      if (!kDebugMode) {
        final uri = Uri.parse(wsUrl);
        _channel = IOWebSocketChannel.connect(
          uri,
          customClient: HttpClient()
            ..badCertificateCallback = (cert, host, port) =>
                host == 'sangiagao.vn' || host == 'www.sangiagao.vn',
        );
      } else {
        _channel = WebSocketChannel.connect(Uri.parse(wsUrl));
      }
    } catch (e) {
      debugPrint('CallSignaling: connect error: $e');
      return;
    }

    _subscription = _channel!.stream.listen(
      _onMessage,
      onError: (e) {
        debugPrint('CallSignaling: WS error: $e');
        onDisconnected?.call();
      },
      onDone: () {
        debugPrint('CallSignaling: WS closed');
        onDisconnected?.call();
      },
    );

    // Heartbeat every 30s
    _heartbeat = Timer.periodic(const Duration(seconds: 30), (_) {
      _send('phoenix', 'heartbeat', {});
    });

    // Join call channel
    _send(_topic, 'phx_join', {});
    _joined = true;
  }

  void initiateCall(String calleeId, String callType) {
    _send(_topic, 'call_initiate', {
      'callee_id': calleeId,
      'call_type': callType,
    });
  }

  void sendOffer(String sdp) {
    _send(_topic, 'call_offer', {'sdp': sdp});
  }

  void sendAnswer(String sdp) {
    _send(_topic, 'call_answer', {'sdp': sdp});
  }

  void sendIceCandidate(Map<String, dynamic> candidate) {
    _send(_topic, 'ice_candidate', {'candidate': candidate});
  }

  void endCall() {
    _send(_topic, 'call_end', {});
  }

  void rejectCall() {
    _send(_topic, 'call_reject', {});
  }

  void sendBusy() {
    _send(_topic, 'call_busy', {});
  }

  void sendReady() {
    _send(_topic, 'call_ready', {});
  }

  void _send(String topic, String event, Map<String, dynamic> payload) {
    if (_channel == null) return;
    final msg = jsonEncode({
      'topic': topic,
      'event': event,
      'payload': payload,
      'ref': '${++_ref}',
    });
    try {
      _channel!.sink.add(msg);
    } catch (e) {
      debugPrint('CallSignaling: send error: $e');
    }
  }

  void _onMessage(dynamic raw) {
    try {
      final data = jsonDecode(raw as String) as Map<String, dynamic>;
      final topic = data['topic'] as String?;
      final event = data['event'] as String?;
      final payload = data['payload'] as Map<String, dynamic>? ?? {};

      if (topic != _topic) return;

      switch (event) {
        case 'call_initiate':
          onCallInitiate?.call(payload);
          break;
        case 'call_offer':
          onCallOffer?.call(payload);
          break;
        case 'call_answer':
          onCallAnswer?.call(payload);
          break;
        case 'ice_candidate':
          onIceCandidate?.call(payload);
          break;
        case 'call_end':
          onCallEnd?.call(payload);
          break;
        case 'call_reject':
          onCallReject?.call(payload);
          break;
        case 'call_busy':
          onCallBusy?.call(payload);
          break;
        case 'call_ready':
          onCallReady?.call(payload);
          break;
        case 'call_timeout':
          onCallTimeout?.call();
          break;
      }
    } catch (e) {
      debugPrint('CallSignaling: parse error: $e');
    }
  }

  void dispose() {
    _heartbeat?.cancel();
    _subscription?.cancel();
    if (_joined) {
      _send(_topic, 'phx_leave', {});
    }
    _channel?.sink.close();
    _channel = null;
    _joined = false;
  }
}
