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
  String? _remoteOfferSdp;
  Completer<String>? _offerCompleter;
  bool _remoteDescriptionSet = false;
  bool _gracePeriodActive = false;
  int _iceRestartAttempts = 0;
  static const _maxIceRestarts = 2;

  // Buffer ICE candidates received before remote SDP is set
  final List<RTCIceCandidate> _pendingCandidates = [];

  // Callbacks
  void Function(CallState)? onStateChanged;
  void Function(int)? onDurationUpdate;
  /// Called when call ends with (status, durationSeconds) for chat log
  void Function(String status, int duration)? onCallEnded;
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
    debugPrint('CallService: start() isInitiator=$isInitiator');

    if (!await requestPermissions()) {
      debugPrint('CallService: permissions denied');
      _setState(CallState.ended);
      return;
    }
    debugPrint('CallService: permissions granted');

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
      debugPrint('CallService: TURN servers: ${iceServers.length}');
    } catch (e) {
      debugPrint('CallService: Failed to get TURN credentials: $e');
    }

    // Create peer connection
    final config = {
      'iceServers': iceServers,
      'sdpSemantics': 'unified-plan',
    };

    _peerConnection = await createPeerConnection(config);
    debugPrint('CallService: peer connection created');

    // Get local audio stream
    final mediaConstraints = <String, dynamic>{
      'audio': true,
      'video': callType == 'video',
    };
    _localStream = await navigator.mediaDevices.getUserMedia(mediaConstraints);
    debugPrint('CallService: local stream ready, tracks=${_localStream!.getTracks().length}');

    // Add tracks to peer connection
    for (final track in _localStream!.getTracks()) {
      await _peerConnection!.addTrack(track, _localStream!);
    }

    // ICE candidate handler
    _peerConnection!.onIceCandidate = (candidate) {
      if (candidate.candidate != null) {
        debugPrint('CallService: sending ICE candidate');
        _signaling?.sendIceCandidate({
          'candidate': candidate.candidate,
          'sdpMid': candidate.sdpMid,
          'sdpMLineIndex': candidate.sdpMLineIndex,
        });
      }
    };

    // Connection state
    _peerConnection!.onConnectionState = (rtcState) {
      debugPrint('CallService: connection state: $rtcState');
      if (rtcState == RTCPeerConnectionState.RTCPeerConnectionStateConnected) {
        _onConnected();
        _iceRestartAttempts = 0;
      } else if (rtcState == RTCPeerConnectionState.RTCPeerConnectionStateFailed) {
        _attemptIceRestart();
      } else if (rtcState == RTCPeerConnectionState.RTCPeerConnectionStateDisconnected) {
        if (!_gracePeriodActive) {
          _gracePeriodActive = true;
          Future.delayed(const Duration(seconds: 5), () {
            _gracePeriodActive = false;
            final currentState = _peerConnection?.connectionState;
            if (currentState == RTCPeerConnectionState.RTCPeerConnectionStateFailed ||
                currentState == RTCPeerConnectionState.RTCPeerConnectionStateDisconnected) {
              _attemptIceRestart();
            }
          });
        }
      }
    };

    // Connect signaling — wait for join confirmation
    _signaling = CallSignalingService(
      token: token,
      conversationId: conversationId,
    );
    _setupSignalingCallbacks();

    final joinCompleter = Completer<bool>();
    final origDisconnected = _signaling!.onDisconnected;
    _signaling!.onDisconnected = () {
      debugPrint('CallService: signaling disconnected during start');
      if (!joinCompleter.isCompleted) joinCompleter.complete(false);
      origDisconnected?.call();
    };

    // Override join confirmation to know when we're connected
    _signaling!.onJoined = () {
      debugPrint('CallService: signaling channel joined');
      if (!joinCompleter.isCompleted) joinCompleter.complete(true);
    };

    _signaling!.connect();

    // Wait for channel join (max 10s)
    bool joined = false;
    try {
      joined = await joinCompleter.future.timeout(const Duration(seconds: 10));
    } on TimeoutException {
      debugPrint('CallService: signaling join timeout');
    }

    if (!joined) {
      debugPrint('CallService: FAILED to join signaling channel');
      _cleanup('signaling_failed');
      return;
    }

    // Restore original disconnect handler
    _signaling?.onDisconnected = () {
      if (state != CallState.ended) {
        debugPrint('CallService: signaling disconnected during call');
        _cleanup('disconnected');
      }
    };

    if (isInitiator) {
      _setState(CallState.outgoing);
      // Create call log via API (this also sends FCM push to callee)
      try {
        final result = await api.initiateCall(conversationId, otherUserId, callType);
        callLogId = result['id'] as String?;
        debugPrint('CallService: call initiated, callLogId=$callLogId');
      } catch (e) {
        debugPrint('CallService: Failed to create call log: $e');
      }
      _signaling!.initiateCall(otherUserId, callType);
      debugPrint('CallService: call_initiate sent, waiting for call_ready...');
    } else {
      _setState(CallState.incoming);
      _signaling!.sendReady();
      debugPrint('CallService: callee ready, call_ready sent');
    }
  }

  Future<void> acceptCall([String? remoteSdp]) async {
    String? sdp = remoteSdp ?? _remoteOfferSdp;

    // If offer hasn't arrived yet, wait for it (up to 15 seconds)
    if (sdp == null) {
      debugPrint('CallService: Offer not yet received, waiting...');
      _setState(CallState.connecting);
      _offerCompleter = Completer<String>();
      try {
        sdp = await _offerCompleter!.future.timeout(const Duration(seconds: 15));
      } on TimeoutException {
        debugPrint('CallService: Timed out waiting for offer SDP');
        _cleanup('no_sdp');
        return;
      } catch (e) {
        debugPrint('CallService: Offer wait cancelled: $e');
        return;
      }
    }

    if (_peerConnection == null) {
      debugPrint('CallService: No peer connection available');
      _cleanup('no_peer_connection');
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

    try {
      // Set remote description (offer)
      await _peerConnection!.setRemoteDescription(
        RTCSessionDescription(sdp, 'offer'),
      );
      _remoteDescriptionSet = true;

      // Flush buffered ICE candidates
      await _flushPendingCandidates();

      // Create answer
      final answer = await _peerConnection!.createAnswer();
      await _peerConnection!.setLocalDescription(answer);
      _signaling!.sendAnswer(answer.sdp!);
    } catch (e) {
      debugPrint('CallService: acceptCall SDP error: $e');
      _cleanup('sdp_error');
    }
  }

  /// Attempt ICE restart to recover from network changes
  Future<void> _attemptIceRestart() async {
    if (_iceRestartAttempts >= _maxIceRestarts || _peerConnection == null) {
      debugPrint('CallService: ICE restart exhausted ($_iceRestartAttempts/$_maxIceRestarts), ending call');
      endCall();
      return;
    }
    _iceRestartAttempts++;
    debugPrint('CallService: ICE restart attempt $_iceRestartAttempts/$_maxIceRestarts');
    try {
      final offer = await _peerConnection!.createOffer({'iceRestart': true});
      await _peerConnection!.setLocalDescription(offer);
      _signaling?.sendOffer(offer.sdp!);
    } catch (e) {
      debugPrint('CallService: ICE restart failed: $e');
      endCall();
    }
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
      track.enableSpeakerphone(isSpeaker);
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
    _signaling!.onCallReady = (payload) async {
      // Callee joined and is ready — now create and send the offer
      if (isInitiator && _peerConnection != null) {
        debugPrint('CallService: callee ready, creating offer');
        try {
          final offer = await _peerConnection!.createOffer();
          await _peerConnection!.setLocalDescription(offer);
          _signaling?.sendOffer(offer.sdp!);
        } catch (e) {
          debugPrint('CallService: create offer error: $e');
          _cleanup('offer_error');
        }
      }
    };

    _signaling!.onCallOffer = (payload) async {
      final sdp = payload['sdp'] as String?;
      if (sdp != null) {
        _remoteOfferSdp = sdp;
        debugPrint('CallService: received remote offer SDP');
        // If acceptCall() is waiting for the offer, complete it
        if (_offerCompleter != null && !_offerCompleter!.isCompleted) {
          _offerCompleter!.complete(sdp);
        }
      }
    };

    _signaling!.onCallAnswer = (payload) async {
      final sdp = payload['sdp'] as String?;
      if (sdp != null && _peerConnection != null) {
        _setState(CallState.connecting);
        try {
          await _peerConnection!.setRemoteDescription(
            RTCSessionDescription(sdp, 'answer'),
          );
          _remoteDescriptionSet = true;
          await _flushPendingCandidates();
        } catch (e) {
          debugPrint('CallService: setRemoteDescription error: $e');
        }
      }
    };

    _signaling!.onIceCandidate = (payload) async {
      final candidate = payload['candidate'] as Map<String, dynamic>?;
      if (candidate != null) {
        final iceCandidate = RTCIceCandidate(
          candidate['candidate'] as String?,
          candidate['sdpMid'] as String?,
          candidate['sdpMLineIndex'] as int?,
        );
        if (_remoteDescriptionSet && _peerConnection != null) {
          await _peerConnection!.addCandidate(iceCandidate);
        } else {
          // Buffer until remote SDP is set
          _pendingCandidates.add(iceCandidate);
        }
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

  /// Flush buffered ICE candidates after remote description is set
  Future<void> _flushPendingCandidates() async {
    if (_peerConnection == null) return;
    for (final candidate in _pendingCandidates) {
      try {
        await _peerConnection!.addCandidate(candidate);
      } catch (e) {
        debugPrint('CallService: flush candidate error: $e');
      }
    }
    _pendingCandidates.clear();
  }

  void _cleanup(String reason) {
    debugPrint('CallService: cleanup — $reason');

    // Cancel any pending offer wait
    if (_offerCompleter != null && !_offerCompleter!.isCompleted) {
      _offerCompleter!.completeError('cleanup: $reason');
    }
    _offerCompleter = null;

    // Determine call status for log
    String callStatus;
    if (state == CallState.connected) {
      callStatus = 'answered';
    } else if (reason == 'timeout') {
      callStatus = 'missed';
    } else if (reason == 'rejected' || reason == 'rejected by remote') {
      callStatus = 'rejected';
    } else if (reason == 'busy') {
      callStatus = 'busy';
    } else {
      callStatus = 'ended';
    }

    // Update call log via API based on reason
    if (callLogId != null) {
      if (state == CallState.connected) {
        api.endCallLog(callLogId!).catchError((_) {});
      } else if (reason == 'timeout') {
        api.missCall(callLogId!).catchError((_) {});
      } else if (reason == 'disconnected') {
        api.endCallLog(callLogId!).catchError((_) {});
      }
    }

    final duration = durationSeconds;

    _durationTimer?.cancel();
    _pendingCandidates.clear();
    _localStream?.getTracks().forEach((track) => track.stop());
    _localStream?.dispose();
    _localStream = null;
    _peerConnection?.close();
    _peerConnection = null;
    _signaling?.dispose();
    _signaling = null;

    _setState(CallState.ended);

    // Notify chat screen to add call log message
    onCallEnded?.call(callStatus, duration);
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
