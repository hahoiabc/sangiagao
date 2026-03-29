import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:permission_handler/permission_handler.dart';
import 'call_signaling_service.dart';
import 'api_service.dart';

enum CallState { idle, outgoing, incoming, connecting, connected, ended }

/// Global flag — true when any call is active (for busy detection)
bool isInCall = false;

/// Manages WebRTC peer connection and call lifecycle.
class CallService {
  final ApiService api;
  final String token;
  final String conversationId;
  final String currentUserId;
  final String otherUserId;
  final String otherUserName;
  final String callType; // 'audio' or 'video'
  final bool isInitiator;

  CallSignalingService? _signaling;
  RTCPeerConnection? _peerConnection;
  MediaStream? _localStream;

  CallState state = CallState.idle;
  bool isMuted = false;
  bool isSpeaker = false;
  String? callLogId;
  DateTime? _connectedAt;

  // Callbacks
  void Function(CallState)? onStateChanged;
  void Function(int)? onDurationUpdate;
  Timer? _durationTimer;

  CallService({
    required this.api,
    required this.token,
    required this.conversationId,
    required this.currentUserId,
    required this.otherUserId,
    required this.otherUserName,
    this.callType = 'audio',
    this.isInitiator = true,
  });

  int get durationSeconds {
    if (_connectedAt == null) return 0;
    return DateTime.now().difference(_connectedAt!).inSeconds;
  }

  Future<bool> requestPermissions() async {
    final mic = await Permission.microphone.request();
    if (!mic.isGranted) return false;
    if (callType == 'video') {
      final cam = await Permission.camera.request();
      if (!cam.isGranted) return false;
    }
    return true;
  }

  Future<void> start() async {
    if (!await requestPermissions()) {
      _setState(CallState.ended);
      return;
    }

    // Get TURN credentials
    List<Map<String, dynamic>> iceServers = [
      {'urls': 'stun:stun.l.google.com:19302'},
    ];
    try {
      final turnData = await api.getTurnCredentials();
      final servers = turnData['ice_servers'] as List?;
      if (servers != null) {
        iceServers = servers.map((s) => Map<String, dynamic>.from(s as Map)).toList();
      }
    } catch (e) {
      debugPrint('CallService: Failed to get TURN credentials: $e');
    }

    // Create peer connection
    final config = {
      'iceServers': iceServers,
      'sdpSemantics': 'unified-plan',
    };

    _peerConnection = await createPeerConnection(config);

    // Get local audio stream
    final mediaConstraints = <String, dynamic>{
      'audio': true,
      'video': callType == 'video',
    };
    _localStream = await navigator.mediaDevices.getUserMedia(mediaConstraints);

    // Add tracks to peer connection
    for (final track in _localStream!.getTracks()) {
      await _peerConnection!.addTrack(track, _localStream!);
    }

    // ICE candidate handler
    _peerConnection!.onIceCandidate = (candidate) {
      if (candidate.candidate != null) {
        _signaling?.sendIceCandidate({
          'candidate': candidate.candidate,
          'sdpMid': candidate.sdpMid,
          'sdpMLineIndex': candidate.sdpMLineIndex,
        });
      }
    };

    // Connection state
    _peerConnection!.onConnectionState = (state) {
      debugPrint('CallService: connection state: $state');
      if (state == RTCPeerConnectionState.RTCPeerConnectionStateConnected) {
        _onConnected();
      } else if (state == RTCPeerConnectionState.RTCPeerConnectionStateFailed ||
          state == RTCPeerConnectionState.RTCPeerConnectionStateDisconnected) {
        // Give 10s grace period for reconnection
        Future.delayed(const Duration(seconds: 10), () {
          if (_peerConnection?.connectionState ==
              RTCPeerConnectionState.RTCPeerConnectionStateFailed) {
            endCall();
          }
        });
      }
    };

    // Connect signaling
    _signaling = CallSignalingService(
      token: token,
      conversationId: conversationId,
    );
    _setupSignalingCallbacks();
    _signaling!.connect();

    if (isInitiator) {
      _setState(CallState.outgoing);
      // Create call log via API
      try {
        final result = await api.initiateCall(conversationId, otherUserId, callType);
        callLogId = result['id'] as String?;
      } catch (e) {
        debugPrint('CallService: Failed to create call log: $e');
      }
      _signaling!.initiateCall(otherUserId, callType);

      // Create offer
      final offer = await _peerConnection!.createOffer();
      await _peerConnection!.setLocalDescription(offer);
      _signaling!.sendOffer(offer.sdp!);
    } else {
      _setState(CallState.incoming);
    }
  }

  Future<void> acceptCall(String remoteSdp) async {
    if (!await requestPermissions()) {
      rejectCall();
      return;
    }

    _setState(CallState.connecting);

    // Answer call via API
    if (callLogId != null) {
      try {
        await api.answerCall(callLogId!);
      } catch (e) {
        debugPrint('CallService: Failed to answer call log: $e');
      }
    }

    // Set remote description (offer)
    await _peerConnection!.setRemoteDescription(
      RTCSessionDescription(remoteSdp, 'offer'),
    );

    // Create answer
    final answer = await _peerConnection!.createAnswer();
    await _peerConnection!.setLocalDescription(answer);
    _signaling!.sendAnswer(answer.sdp!);
  }

  void _onConnected() {
    _setState(CallState.connected);
    _connectedAt = DateTime.now();
    _durationTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      onDurationUpdate?.call(durationSeconds);
    });
  }

  void toggleMute() {
    isMuted = !isMuted;
    _localStream?.getAudioTracks().forEach((track) {
      track.enabled = !isMuted;
    });
  }

  void toggleSpeaker() {
    isSpeaker = !isSpeaker;
    _localStream?.getAudioTracks().forEach((track) {
      Helper.setSpeakerphoneOn(isSpeaker);
    });
  }

  void endCall() {
    _signaling?.endCall();
    _cleanup('ended');
  }

  void rejectCall() {
    _signaling?.rejectCall();
    if (callLogId != null) {
      api.rejectCall(callLogId!).catchError((_) {});
    }
    _cleanup('rejected');
  }

  void _setupSignalingCallbacks() {
    _signaling!.onCallAnswer = (payload) async {
      final sdp = payload['sdp'] as String?;
      if (sdp != null && _peerConnection != null) {
        _setState(CallState.connecting);
        await _peerConnection!.setRemoteDescription(
          RTCSessionDescription(sdp, 'answer'),
        );
      }
    };

    _signaling!.onIceCandidate = (payload) async {
      final candidate = payload['candidate'] as Map<String, dynamic>?;
      if (candidate != null && _peerConnection != null) {
        await _peerConnection!.addCandidate(RTCIceCandidate(
          candidate['candidate'] as String?,
          candidate['sdpMid'] as String?,
          candidate['sdpMLineIndex'] as int?,
        ));
      }
    };

    _signaling!.onCallEnd = (_) {
      _cleanup('ended by remote');
    };

    _signaling!.onCallReject = (_) {
      _cleanup('rejected by remote');
    };

    _signaling!.onCallBusy = (_) {
      _cleanup('busy');
    };

    _signaling!.onCallTimeout = () {
      _cleanup('timeout');
    };

    _signaling!.onDisconnected = () {
      if (state != CallState.ended) {
        _cleanup('disconnected');
      }
    };
  }

  void _cleanup(String reason) {
    debugPrint('CallService: cleanup — $reason');

    // Update call log via API based on reason
    if (callLogId != null) {
      if (state == CallState.connected) {
        api.endCallLog(callLogId!).catchError((_) {});
      } else if (reason == 'timeout' || reason == 'disconnected') {
        // Missed call — no answer within timeout
        api.endCallLog(callLogId!).catchError((_) {});
      }
    }

    _durationTimer?.cancel();
    _localStream?.getTracks().forEach((track) => track.stop());
    _localStream?.dispose();
    _localStream = null;
    _peerConnection?.close();
    _peerConnection = null;
    _signaling?.dispose();
    _signaling = null;

    _setState(CallState.ended);
  }

  void _setState(CallState newState) {
    state = newState;
    isInCall = newState != CallState.idle && newState != CallState.ended;
    onStateChanged?.call(newState);
  }

  void dispose() {
    if (state != CallState.ended) {
      _cleanup('disposed');
    }
  }
}
