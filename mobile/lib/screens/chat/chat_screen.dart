import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart' show Share;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:record/record.dart';
import 'package:audioplayers/audioplayers.dart';
import 'package:path_provider/path_provider.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../config/env.dart';
import '../../models/conversation.dart';
import '../../models/listing.dart';
import '../../models/user.dart';
import '../../providers/providers.dart';
import '../../services/push_notification_service.dart';
import '../../theme/app_theme.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final String conversationId;
  const ChatScreen({super.key, required this.conversationId});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _msgCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  List<Message> _messages = [];
  bool _loading = true;
  bool _uploadingImage = false;
  String? _currentUserId;
  PublicProfile? _otherUser;
  WebSocketChannel? _channel;
  StreamSubscription? _wsSub;
  Timer? _heartbeat;
  Timer? _pollTimer;
  int _ref = 0;
  bool _joined = false;
  int _reconnectAttempts = 0;

  // Multi-select mode
  bool _selectMode = false;
  final Set<String> _selectedIds = {};

  // Audio recording
  final AudioRecorder _recorder = AudioRecorder();
  bool _isRecording = false;
  bool _uploadingAudio = false;
  Timer? _recordTimer;
  int _recordSeconds = 0;

  // Audio playback
  final AudioPlayer _audioPlayer = AudioPlayer();
  String? _playingMsgId;
  Duration _playPosition = Duration.zero;
  Duration _playDuration = Duration.zero;
  StreamSubscription? _positionSub;
  StreamSubscription? _durationSub;
  StreamSubscription? _completeSub;

  // Typing indicator
  bool _otherUserTyping = false;
  Timer? _typingDebounce;
  Timer? _typingTimeout;

  // Listing link cache
  final Map<String, ListingDetail?> _listingCache = {};

  // Reply mode
  Message? _replyingTo;

  // Reaction emojis
  static const _reactionEmojis = ['👍', '❤️', '😂', '😮', '😢', '😡'];

  // Phoenix chat service URL
  static const String _phoenixWsUrl = Env.wsBaseUrl;

  @override
  void initState() {
    super.initState();
    // Track active conversation to suppress push notifications
    PushNotificationService.activeConversationId = widget.conversationId;
    PushNotificationService.clearUnreadForConversation(widget.conversationId);
    _init();
    _positionSub = _audioPlayer.onPositionChanged.listen((pos) {
      if (mounted) setState(() => _playPosition = pos);
    });
    _durationSub = _audioPlayer.onDurationChanged.listen((dur) {
      if (mounted) setState(() => _playDuration = dur);
    });
    _completeSub = _audioPlayer.onPlayerComplete.listen((_) {
      if (mounted) setState(() => _playingMsgId = null);
    });
  }

  Future<void> _init() async {
    final user = ref.read(authProvider).user;
    _currentUserId = user?.id;
    await Future.wait([_loadMessages(), _loadConversation()]);
    // Mark messages as read and refresh badge
    _markReadAndRefreshBadge();
    // Always start polling immediately, WS will stop it if connected
    _startPolling();
    _connectPhoenix();
  }

  Future<void> _loadConversation() async {
    try {
      final result = await ref.read(apiServiceProvider).getConversations(limit: 50);
      final conv = result.data.where((c) => c.id == widget.conversationId).firstOrNull;
      final other = conv?.otherUser;
      if (other != null) {
        setState(() => _otherUser = other);
      }
    } catch (e) {
      debugPrint('Load conversation error: $e');
    }
  }

  Future<void> _loadMessages() async {
    try {
      final result = await ref.read(apiServiceProvider).getMessages(widget.conversationId);
      setState(() {
        _messages = result.data.reversed.toList();
      });
      _scrollToBottom();
    } catch (e) {
      debugPrint('Load messages error: $e');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _connectPhoenix() async {
    try {
      final token = await ref.read(apiServiceProvider).getToken();
      if (token == null) {
        _startPolling();
        return;
      }

      // Pass token as query param — Phoenix UserSocket reads it from connect params
      final wsUrl = Uri.parse('$_phoenixWsUrl?token=$token');
      _channel = IOWebSocketChannel.connect(
        wsUrl,
        customClient: !kDebugMode ? (HttpClient()
          ..badCertificateCallback = (cert, host, port) =>
              ['sangiagao.vn', 'www.sangiagao.vn'].contains(host)) : null,
      );

      _wsSub = _channel!.stream.listen(
        (data) {
          _reconnectAttempts = 0;
          _handlePhoenixMessage(jsonDecode(data as String));
        },
        onError: (e) {
          debugPrint('Phoenix WS error: $e');
          _joined = false;
          _scheduleReconnect();
        },
        onDone: () {
          debugPrint('Phoenix WS closed');
          _joined = false;
          _scheduleReconnect();
        },
      );

      _heartbeat = Timer.periodic(const Duration(seconds: 30), (_) {
        _phoenixSend('phoenix', 'heartbeat', {});
      });

      _phoenixSend('chat:${widget.conversationId}', 'phx_join', {});
    } catch (e) {
      debugPrint('Phoenix connect error: $e');
      _startPolling();
    }
  }

  void _scheduleReconnect() {
    if (!mounted) return;
    _heartbeat?.cancel();
    _wsSub?.cancel();
    _channel?.sink.close();
    _channel = null;

    _startPolling();

    if (_reconnectAttempts < 10) {
      _reconnectAttempts++;
      final delay = Duration(seconds: 2 * _reconnectAttempts);
      Future.delayed(delay, () {
        if (mounted && !_joined) _connectPhoenix();
      });
    }
  }

  void _startPolling() {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 10), (_) => _pollNewMessages());
  }


  Future<void> _pollNewMessages() async {
    if (!mounted) return;
    try {
      final result = await ref.read(apiServiceProvider).getMessages(widget.conversationId);
      final freshMessages = result.data.reversed.toList();
      // Compare by last real message ID (ignore temp_ messages)
      final realLocal = _messages.where((m) => !m.id.startsWith('temp_'));
      final lastLocalId = realLocal.isNotEmpty ? realLocal.last.id : null;
      final lastRemoteId = freshMessages.isNotEmpty ? freshMessages.last.id : null;
      if (lastRemoteId != null && lastLocalId != lastRemoteId) {
        setState(() => _messages = freshMessages);
        _scrollToBottom();
      }
    } catch (e) {
      debugPrint('Poll messages error: $e');
    }
  }

  void _phoenixSend(String topic, String event, Map<String, dynamic> payload) {
    final ch = _channel;
    if (ch == null) return;
    _ref++;
    ch.sink.add(jsonEncode({
      'topic': topic,
      'event': event,
      'payload': payload,
      'ref': '$_ref',
    }));
  }

  void _handlePhoenixMessage(Map<String, dynamic> msg) {
    final event = msg['event'] as String?;
    final topic = msg['topic'] as String?;
    final payload = msg['payload'] as Map<String, dynamic>? ?? {};

    if (event == 'phx_reply' && topic == 'chat:${widget.conversationId}') {
      if (payload['status'] == 'ok') {
        _joined = true;
        // Slow down polling as safety net (WS handles real-time via relay)
        _pollTimer?.cancel();
        _pollTimer = Timer.periodic(const Duration(seconds: 30), (_) => _pollNewMessages());
        debugPrint('Joined Phoenix channel chat:${widget.conversationId}');
      }
      return;
    }

    if (event == 'new_message' && topic == 'chat:${widget.conversationId}') {
      final msgData = payload;
      if (msgData['sender_id'] == _currentUserId) return;

      final message = Message(
        id: msgData['id']?.toString() ?? '',
        conversationId: msgData['conversation_id'] ?? widget.conversationId,
        senderId: msgData['sender_id'] ?? '',
        content: msgData['content'] ?? '',
        type: msgData['type'] ?? 'text',
        readAt: msgData['read_at']?.toString(),
        createdAt: msgData['timestamp']?.toString() ?? DateTime.now().toIso8601String(),
      );
      setState(() => _messages.add(message));
      _scrollToBottom();
      // Mark as read immediately since user is viewing this chat
      _markReadAndRefreshBadge();
      return;
    }

    if (event == 'read_receipt') return;

    if (event == 'reaction' && topic == 'chat:${widget.conversationId}') {
      final msgId = payload['message_id']?.toString();
      final userId = payload['user_id']?.toString();
      final emoji = payload['emoji']?.toString();
      if (msgId != null && userId != null && emoji != null && userId != _currentUserId) {
        // Reload messages to get updated reactions from server
        _pollNewMessages();
      }
      return;
    }

    if (event == 'typing' && topic == 'chat:${widget.conversationId}') {
      final typingUserId = payload['user_id']?.toString();
      if (typingUserId != null && typingUserId != _currentUserId) {
        setState(() => _otherUserTyping = true);
        _typingTimeout?.cancel();
        _typingTimeout = Timer(const Duration(seconds: 3), () {
          if (mounted) setState(() => _otherUserTyping = false);
        });
      }
      return;
    }
  }

  Future<void> _markReadAndRefreshBadge() async {
    try {
      await ref.read(apiServiceProvider).markConversationRead(widget.conversationId);
      ref.read(unreadCountProvider.notifier).refresh();
    } catch (_) {}
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.animateTo(
          _scrollCtrl.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _onTyping() {
    if (!_joined || _channel == null) return;
    if (_typingDebounce?.isActive ?? false) return;
    _phoenixSend('chat:${widget.conversationId}', 'typing', {});
    _typingDebounce = Timer(const Duration(seconds: 2), () {});
  }

  Future<void> _send() async {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty) return;
    _msgCtrl.clear();
    _typingDebounce?.cancel();
    final replyId = _replyingTo?.id;
    setState(() => _replyingTo = null);
    await _sendMessage(text, 'text', replyToId: replyId);
  }

  Future<void> _sendMessage(String content, String type, {String? replyToId}) async {
    try {
      // Send via HTTP API — Go backend handles persistence
      final msg = await ref.read(apiServiceProvider).sendMessage(
        widget.conversationId, content, type: type, replyToId: replyToId,
      );
      setState(() => _messages.add(msg));
      _scrollToBottom();
      // Relay via Phoenix WS so other participant gets it in real-time
      if (_joined && _channel != null) {
        _phoenixSend('chat:${widget.conversationId}', 'relay', {
          'id': msg.id,
          'conversation_id': msg.conversationId,
          'sender_id': msg.senderId,
          'content': msg.content,
          'type': msg.type,
          'reply_to_id': msg.replyToId,
          'created_at': msg.createdAt,
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gửi tin nhắn thất bại: $e')),
        );
      }
    }
  }

  void _showImageSourcePicker() {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt, color: AppColors.info),
              title: const Text('Chụp ảnh'),
              onTap: () {
                Navigator.pop(context);
                _pickAndSendImage(ImageSource.camera);
              },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library, color: AppColors.primary),
              title: const Text('Chọn từ thư viện'),
              onTap: () {
                Navigator.pop(context);
                _pickAndSendImage(ImageSource.gallery);
              },
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _pickAndSendImage(ImageSource source) async {
    final picker = ImagePicker();

    if (source == ImageSource.gallery) {
      // Allow multiple images from gallery (max 10)
      final images = await picker.pickMultiImage(
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 80,
        limit: 10,
      );
      if (images.isEmpty) return;

      setState(() => _uploadingImage = true);
      try {
        for (final image in images) {
          final url = await ref.read(apiServiceProvider).uploadImage(image.path, 'images');
          await _sendMessage(url, 'image');
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Gửi ảnh thất bại: $e')),
          );
        }
      } finally {
        if (mounted) setState(() => _uploadingImage = false);
      }
    } else {
      // Camera: single image
      final image = await picker.pickImage(
        source: source,
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 80,
      );
      if (image == null) return;

      setState(() => _uploadingImage = true);
      try {
        final url = await ref.read(apiServiceProvider).uploadImage(image.path, 'images');
        await _sendMessage(url, 'image');
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Gửi ảnh thất bại: $e')),
          );
        }
      } finally {
        if (mounted) setState(() => _uploadingImage = false);
      }
    }
  }

  // --- Audio recording ---

  Future<void> _startRecording() async {
    var status = await Permission.microphone.request();
    if (!status.isGranted) {
      if (status.isPermanentlyDenied) {
        if (mounted) {
          final shouldOpen = await showDialog<bool>(
            context: context,
            builder: (ctx) => AlertDialog(
              title: const Text('Cần quyền micro'),
              content: const Text('Bạn đã từ chối quyền micro. Vui lòng vào Cài đặt để cấp quyền.'),
              actions: [
                TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Hủy')),
                TextButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Mở Cài đặt')),
              ],
            ),
          );
          if (shouldOpen == true) {
            await openAppSettings();
          }
        }
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cần cấp quyền micro để ghi âm')),
        );
      }
      return;
    }

    final dir = await getTemporaryDirectory();
    final path = '${dir.path}/voice_${DateTime.now().millisecondsSinceEpoch}.m4a';

    await _recorder.start(
      const RecordConfig(encoder: AudioEncoder.aacLc, bitRate: 128000),
      path: path,
    );

    setState(() {
      _isRecording = true;
      _recordSeconds = 0;
    });

    _recordTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      setState(() => _recordSeconds++);
    });
  }

  Future<void> _stopAndSendRecording() async {
    _recordTimer?.cancel();
    final path = await _recorder.stop();

    if (path == null || _recordSeconds < 1) {
      setState(() => _isRecording = false);
      return;
    }

    setState(() {
      _isRecording = false;
      _uploadingAudio = true;
    });

    try {
      final url = await ref.read(apiServiceProvider).uploadAudio(path);
      await _sendMessage(url, 'audio');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gửi âm thanh thất bại: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _uploadingAudio = false);
    }
  }

  Future<void> _cancelRecording() async {
    _recordTimer?.cancel();
    await _recorder.stop();
    setState(() {
      _isRecording = false;
      _recordSeconds = 0;
    });
  }

  // --- Audio playback ---

  Future<void> _togglePlayAudio(Message msg) async {
    if (_playingMsgId == msg.id) {
      await _audioPlayer.stop();
      setState(() => _playingMsgId = null);
    } else {
      await _audioPlayer.stop();
      await _audioPlayer.play(UrlSource(msg.content));
      setState(() => _playingMsgId = msg.id);
    }
  }

  String _formatDuration(Duration d) {
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '$m:$s';
  }

  @override
  void dispose() {
    // Clear active conversation tracking
    if (PushNotificationService.activeConversationId == widget.conversationId) {
      PushNotificationService.activeConversationId = null;
    }
    _heartbeat?.cancel();
    _pollTimer?.cancel();
    _typingDebounce?.cancel();
    _typingTimeout?.cancel();
    _wsSub?.cancel();
    _positionSub?.cancel();
    _durationSub?.cancel();
    _completeSub?.cancel();
    _recordTimer?.cancel();
    _recorder.dispose();
    _audioPlayer.dispose();
    _listingCache.clear();
    if (_joined) {
      _phoenixSend('chat:${widget.conversationId}', 'phx_leave', {});
    }
    _channel?.sink.close();
    _msgCtrl.dispose();
    _scrollCtrl.dispose();
    super.dispose();
  }

  // --- Message actions ---

  bool _canRecall(Message msg) {
    final dt = DateTime.tryParse(msg.createdAt)?.toLocal();
    if (dt == null) return false;
    return DateTime.now().difference(dt).inHours < 24;
  }

  void _showMessageActions(Message msg) {
    if (msg.type == 'recalled') return;

    final isMe = msg.senderId == _currentUserId;
    final canRecall = isMe && _canRecall(msg);
    final isText = msg.type == 'text';

    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Reaction row
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: _reactionEmojis.map((emoji) {
                  final hasReacted = msg.reactions.any((r) => r.emoji == emoji && r.userId == _currentUserId);
                  return GestureDetector(
                    onTap: () {
                      Navigator.pop(context);
                      _toggleReaction(msg, emoji);
                    },
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: hasReacted
                          ? BoxDecoration(
                              color: AppColors.primary.withValues(alpha: 0.15),
                              borderRadius: BorderRadius.circular(20),
                            )
                          : null,
                      child: Text(emoji, style: const TextStyle(fontSize: 24)),
                    ),
                  );
                }).toList(),
              ),
            ),
            const Divider(height: 1),
            ListTile(
              leading: const Icon(Icons.reply, color: AppColors.primary),
              title: const Text('Trả lời'),
              onTap: () {
                Navigator.pop(context);
                setState(() => _replyingTo = msg);
              },
            ),
            if (isText)
              ListTile(
                leading: const Icon(Icons.copy, color: AppColors.info),
                title: const Text('Sao chép'),
                onTap: () {
                  Navigator.pop(context);
                  Clipboard.setData(ClipboardData(text: msg.content));
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Đã sao chép'), duration: Duration(seconds: 1)),
                  );
                },
              ),
            if (isText)
              ListTile(
                leading: const Icon(Icons.share, color: AppColors.primary),
                title: const Text('Chia sẻ'),
                onTap: () {
                  Navigator.pop(context);
                  Share.share(msg.content);
                },
              ),
            if (canRecall)
              ListTile(
                leading: const Icon(Icons.replay, color: AppColors.warning),
                title: const Text('Thu hồi tin nhắn'),
                onTap: () {
                  Navigator.pop(context);
                  _recallMessage(msg);
                },
              ),
            ListTile(
              leading: const Icon(Icons.delete, color: AppColors.error),
              title: const Text('Xóa phía tôi'),
              onTap: () {
                Navigator.pop(context);
                _deleteMessage(msg);
              },
            ),
            ListTile(
              leading: const Icon(Icons.checklist, color: AppColors.info),
              title: const Text('Chọn nhiều hơn'),
              onTap: () {
                Navigator.pop(context);
                _enterSelectMode(msg);
              },
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _toggleReaction(Message msg, String emoji) async {
    try {
      final reactions = await ref.read(apiServiceProvider).toggleReaction(
        widget.conversationId, msg.id, emoji,
      );
      // Broadcast via WS so the other user sees it
      if (_joined && _channel != null) {
        _phoenixSend('chat:${widget.conversationId}', 'reaction', {
          'message_id': msg.id,
          'emoji': emoji,
        });
      }
      setState(() {
        final idx = _messages.indexWhere((m) => m.id == msg.id);
        if (idx != -1) {
          final old = _messages[idx];
          _messages[idx] = Message(
            id: old.id,
            conversationId: old.conversationId,
            senderId: old.senderId,
            content: old.content,
            type: old.type,
            readAt: old.readAt,
            replyToId: old.replyToId,
            replyTo: old.replyTo,
            reactions: reactions,
            createdAt: old.createdAt,
          );
        }
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Thả cảm xúc thất bại: $e')),
        );
      }
    }
  }

  void _enterSelectMode(Message msg) {
    setState(() {
      _selectMode = true;
      _selectedIds.clear();
      _selectedIds.add(msg.id);
    });
  }

  Future<void> _reportUser() async {
    if (_otherUser == null) return;
    final reasons = ['Spam', 'Quấy rối', 'Lừa đảo', 'Nội dung không phù hợp', 'Khác'];
    final reason = await showDialog<String>(
      context: context,
      builder: (ctx) => SimpleDialog(
        title: const Text('Báo cáo người dùng'),
        children: reasons.map((r) => SimpleDialogOption(
          onPressed: () => Navigator.pop(ctx, r),
          child: Text(r),
        )).toList(),
      ),
    );
    if (reason == null || !mounted) return;
    try {
      await ref.read(apiServiceProvider).createReport('user', _otherUser!.id, reason);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Đã gửi báo cáo. Cảm ơn bạn!')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gửi báo cáo thất bại: $e')),
        );
      }
    }
  }

  void _exitSelectMode() {
    setState(() {
      _selectMode = false;
      _selectedIds.clear();
    });
  }

  void _toggleSelect(Message msg) {
    setState(() {
      if (_selectedIds.contains(msg.id)) {
        _selectedIds.remove(msg.id);
      } else {
        _selectedIds.add(msg.id);
      }
    });
  }

  Future<void> _deleteMessage(Message msg) async {
    try {
      await ref.read(apiServiceProvider).deleteMessage(widget.conversationId, msg.id);
      setState(() => _messages.removeWhere((m) => m.id == msg.id));
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Xóa tin nhắn thất bại: $e')),
        );
      }
    }
  }

  Future<void> _recallMessage(Message msg) async {
    try {
      final updated = await ref.read(apiServiceProvider).recallMessage(widget.conversationId, msg.id);
      setState(() {
        final idx = _messages.indexWhere((m) => m.id == msg.id);
        if (idx != -1) _messages[idx] = updated;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Thu hồi tin nhắn thất bại: $e')),
        );
      }
    }
  }

  Future<void> _batchDelete() async {
    if (_selectedIds.isEmpty) return;
    try {
      await ref.read(apiServiceProvider).batchDeleteMessages(
        widget.conversationId, _selectedIds.toList(),
      );
      setState(() {
        _messages.removeWhere((m) => _selectedIds.contains(m.id));
      });
      _exitSelectMode();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Xóa tin nhắn thất bại: $e')),
        );
      }
    }
  }

  Future<void> _batchRecall() async {
    if (_selectedIds.isEmpty) return;
    try {
      await ref.read(apiServiceProvider).batchRecallMessages(
        widget.conversationId, _selectedIds.toList(),
      );
      await _loadMessages();
      _exitSelectMode();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Thu hồi tin nhắn thất bại: $e')),
        );
      }
    }
  }

  bool get _hasRecallableSelected {
    return _messages
        .where((m) => _selectedIds.contains(m.id) && m.senderId == _currentUserId && m.type != 'recalled')
        .any((m) => _canRecall(m));
  }

  // --- Grouping ---

  List<MapEntry<String, List<Message>>> _groupMessagesByDay() {
    final Map<String, List<Message>> groups = {};
    for (final msg in _messages) {
      final dt = DateTime.tryParse(msg.createdAt)?.toLocal();
      final key = dt != null
          ? '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year}'
          : 'unknown';
      groups.putIfAbsent(key, () => []).add(msg);
    }
    return groups.entries.toList();
  }

  // --- Build ---

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: _selectMode
          ? _buildSelectAppBar()
          : AppBar(
              title: Builder(builder: (_) {
                final other = _otherUser;
                return GestureDetector(
                  onTap: other != null
                      ? () => context.push('/seller/${other.id}')
                      : null,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(other?.name ?? 'Trò chuyện', style: const TextStyle(fontSize: 16)),
                      if (other != null)
                        Row(
                          children: [
                            Container(
                              width: 8,
                              height: 8,
                              decoration: BoxDecoration(
                                color: other.isOnline ? AppColors.onlineGreen : AppColors.offlineGrey,
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              other.isOnline
                                  ? 'Online'
                                  : (other.lastSeenText ?? 'Offline'),
                              style: TextStyle(
                                fontSize: 12,
                                color: other.isOnline ? AppColors.onlineGreen : AppColors.offlineGrey,
                              ),
                            ),
                          ],
                        ),
                    ],
                  ),
                );
              }),
              actions: [
                PopupMenuButton<String>(
                  onSelected: (value) {
                    if (value == 'report_user' && _otherUser != null) {
                      _reportUser();
                    }
                  },
                  itemBuilder: (_) => [
                    const PopupMenuItem(value: 'report_user', child: Text('Báo cáo người dùng')),
                  ],
                ),
              ],
            ),
      body: Column(
        children: [
          Expanded(
            child: _loading
                ? const Align(alignment: Alignment.centerLeft, child: Padding(padding: EdgeInsets.only(left: 16), child: CircularProgressIndicator()))
                : _messages.isEmpty
                    ? const Center(child: Text('Chưa có tin nhắn'))
                    : _buildGroupedMessages(),
          ),
          if (_otherUserTyping)
            Align(
              alignment: Alignment.centerLeft,
              child: Padding(
                padding: const EdgeInsets.only(left: 16, bottom: 4),
                child: Text(
                  '${_otherUser?.name ?? 'Đối phương'} đang soạn tin...',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
            ),
          if (_selectMode)
            _buildSelectActionBar()
          else if (_isRecording)
            _buildRecordingBar()
          else
            _buildInputBar(),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildSelectAppBar() {
    return AppBar(
      leading: IconButton(
        icon: const Icon(Icons.close),
        onPressed: _exitSelectMode,
      ),
      title: Text('Đã chọn ${_selectedIds.length}'),
    );
  }

  Widget _buildSelectActionBar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4)],
      ),
      child: SafeArea(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            if (_hasRecallableSelected)
              TextButton.icon(
                onPressed: _batchRecall,
                icon: const Icon(Icons.replay, color: AppColors.warning),
                label: const Text('Thu hồi', style: TextStyle(color: AppColors.warning)),
              ),
            TextButton.icon(
              onPressed: _batchDelete,
              icon: const Icon(Icons.delete, color: AppColors.error),
              label: const Text('Xóa phía tôi', style: TextStyle(color: AppColors.error)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildGroupedMessages() {
    final dayGroups = _groupMessagesByDay();
    return ListView.builder(
      controller: _scrollCtrl,
      padding: const EdgeInsets.all(12),
      itemCount: dayGroups.length,
      itemBuilder: (_, i) => _buildDayChain(dayGroups[i].key, dayGroups[i].value),
    );
  }

  Widget _buildDayChain(String dateLabel, List<Message> dayMessages) {
    if (dayMessages.isEmpty) return const SizedBox.shrink();

    final firstDt = DateTime.tryParse(dayMessages.first.createdAt)?.toLocal();
    final lastDt = DateTime.tryParse(dayMessages.last.createdAt)?.toLocal();

    final startTime = firstDt != null
        ? '${firstDt.hour.toString().padLeft(2, '0')}:${firstDt.minute.toString().padLeft(2, '0')}'
        : '';
    final endTime = lastDt != null
        ? '${lastDt.hour.toString().padLeft(2, '0')}:${lastDt.minute.toString().padLeft(2, '0')}'
        : '';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(
          child: Container(
            margin: const EdgeInsets.only(top: 12, bottom: 8),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            decoration: BoxDecoration(
              color: AppColors.offlineGrey.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Text(
              '$startTime, $dateLabel',
              style: TextStyle(fontSize: 11, color: AppColors.textSecondary),
            ),
          ),
        ),
        for (int i = 0; i < dayMessages.length; i++)
          _buildMessage(dayMessages[i], isFirstInChain: i == 0),
        if (dayMessages.length > 1 && startTime != endTime)
          Center(
            child: Padding(
              padding: const EdgeInsets.only(top: 4, bottom: 8),
              child: Text(
                endTime,
                style: TextStyle(fontSize: 10, color: AppColors.textHint),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildAudioBubble(Message msg, bool isMe) {
    final isPlaying = _playingMsgId == msg.id;
    final progress = _playDuration.inMilliseconds > 0 && isPlaying
        ? _playPosition.inMilliseconds / _playDuration.inMilliseconds
        : 0.0;
    final displayDuration = isPlaying ? _playDuration : const Duration(seconds: 0);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: isMe ? AppColors.chatBubbleMe : AppColors.chatBubbleOther,
        borderRadius: BorderRadius.only(
          topLeft: const Radius.circular(16),
          topRight: const Radius.circular(16),
          bottomLeft: Radius.circular(isMe ? 16 : 4),
          bottomRight: Radius.circular(isMe ? 4 : 16),
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          GestureDetector(
            onTap: _selectMode ? null : () => _togglePlayAudio(msg),
            child: Icon(
              isPlaying ? Icons.stop_circle : Icons.play_circle_fill,
              size: 36,
              color: isMe ? Colors.white : AppColors.primary,
            ),
          ),
          const SizedBox(width: 8),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 120,
                child: LinearProgressIndicator(
                  value: progress,
                  backgroundColor: isMe ? Colors.white30 : AppColors.offlineGrey,
                  valueColor: AlwaysStoppedAnimation(
                    isMe ? Colors.white : AppColors.primary,
                  ),
                  minHeight: 3,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                isPlaying ? _formatDuration(_playPosition) : _formatDuration(displayDuration),
                style: TextStyle(
                  fontSize: 11,
                  color: isMe ? Colors.white70 : AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  String? _extractListingId(Message msg) {
    if (msg.type == 'listing_link' && msg.content.startsWith('listing://')) {
      return msg.content.replaceFirst('listing://', '');
    }
    return null;
  }

  Widget _buildListingLinkBubble(Message msg, bool isMe) {
    final listingId = _extractListingId(msg);
    if (listingId == null) return const SizedBox.shrink();

    // Load listing nếu chưa có trong cache
    if (!_listingCache.containsKey(listingId)) {
      _listingCache[listingId] = null; // mark loading
      ref.read(apiServiceProvider).getListingDetail(listingId).then((detail) {
        if (mounted) setState(() => _listingCache[listingId] = detail);
      }).catchError((_) {});
    }

    final detail = _listingCache[listingId];
    final priceFormat = NumberFormat('#,###', 'vi_VN');

    return GestureDetector(
      onTap: _selectMode ? null : () => context.push('/marketplace/$listingId'),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: isMe ? AppColors.chatBubbleMe.withValues(alpha: 0.9) : AppColors.chatBubbleOther,
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isMe ? 16 : 4),
            bottomRight: Radius.circular(isMe ? 4 : 16),
          ),
          border: Border.all(color: isMe ? Colors.transparent : AppColors.border),
        ),
        child: detail == null
            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
            : Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    detail.listing.title,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: isMe ? Colors.white : Colors.black87,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${priceFormat.format(detail.listing.pricePerKg)}đ/kg',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      color: isMe ? Colors.white : AppColors.primary,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${priceFormat.format(detail.listing.quantityKg)} kg',
                    style: TextStyle(
                      fontSize: 13,
                      color: isMe ? Colors.white70 : AppColors.textSecondary,
                    ),
                  ),
                  if (detail.listing.harvestSeason != null && detail.listing.harvestSeason!.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(
                        'Mùa gặt: ${detail.listing.harvestSeason}',
                        style: TextStyle(
                          fontSize: 13,
                          color: isMe ? Colors.white70 : AppColors.textSecondary,
                        ),
                      ),
                    ),
                ],
              ),
      ),
    );
  }

  Widget _buildMessage(Message msg, {bool isFirstInChain = false}) {
    final isMe = msg.senderId == _currentUserId;
    final isImage = msg.type == 'image';
    final isAudio = msg.type == 'audio';
    final isRecalled = msg.type == 'recalled';
    final isListingLink = msg.type == 'listing_link';
    final showAvatar = !isMe && isFirstInChain;

    final isSelected = _selectedIds.contains(msg.id);

    final bubble = GestureDetector(
      onLongPress: () {
        if (_selectMode) return;
        _showMessageActions(msg);
      },
      onTap: _selectMode ? () => _toggleSelect(msg) : null,
      child: Column(
        crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          if (isRecalled)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: AppColors.chatBubbleOther,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: AppColors.border),
              ),
              child: Text(
                msg.content,
                style: TextStyle(color: AppColors.textHint, fontStyle: FontStyle.italic),
              ),
            )
          else if (isListingLink)
            _buildListingLinkBubble(msg, isMe)
          else if (isAudio)
            Column(
              crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                if (msg.replyTo != null) _buildReplyQuote(msg.replyTo!, isMe),
                _buildAudioBubble(msg, isMe),
              ],
            )
          else if (isImage)
            Column(
              crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                if (msg.replyTo != null) _buildReplyQuote(msg.replyTo!, isMe),
                GestureDetector(
                  onTap: _selectMode ? null : () => _showFullImage(msg.content),
                  child: ClipRRect(
                    borderRadius: BorderRadius.only(
                      topLeft: const Radius.circular(16),
                      topRight: const Radius.circular(16),
                      bottomLeft: Radius.circular(isMe ? 16 : 4),
                      bottomRight: Radius.circular(isMe ? 4 : 16),
                    ),
                    child: CachedNetworkImage(
                      imageUrl: msg.content,
                      width: 200,
                      height: 200,
                      fit: BoxFit.cover,
                      placeholder: (_, __) => Container(
                        width: 200,
                        height: 200,
                        color: AppColors.chatBubbleOther,
                        child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                      ),
                      errorWidget: (_, __, ___) => Container(
                        width: 200,
                        height: 200,
                        color: AppColors.chatBubbleOther,
                        child: const Icon(Icons.broken_image, size: 48, color: AppColors.textHint),
                      ),
                    ),
                  ),
                ),
              ],
            )
          else
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: isMe ? AppColors.chatBubbleMe : AppColors.chatBubbleOther,
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(16),
                  topRight: const Radius.circular(16),
                  bottomLeft: Radius.circular(isMe ? 16 : 4),
                  bottomRight: Radius.circular(isMe ? 4 : 16),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (msg.replyTo != null) ...[
                    _buildReplyQuote(msg.replyTo!, isMe),
                    const SizedBox(height: 6),
                  ],
                  Text(
                    msg.content,
                    style: TextStyle(color: isMe ? Colors.white : AppColors.textPrimary),
                  ),
                ],
              ),
            ),
          // Reactions display
          if (msg.reactions.isNotEmpty)
            _buildReactionsRow(msg),
          if (isMe && msg.isRead)
            Padding(
              padding: const EdgeInsets.only(top: 2, right: 4),
              child: Icon(Icons.done_all, size: 12, color: AppColors.textHint),
            ),
        ],
      ),
    );

    final messageRow = Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        if (_selectMode)
          Padding(
            padding: const EdgeInsets.only(right: 8),
            child: Icon(
              isSelected ? Icons.check_circle : Icons.radio_button_unchecked,
              color: isSelected ? AppColors.primary : AppColors.offlineGrey,
              size: 22,
            ),
          ),
        Flexible(child: bubble),
      ],
    );

    if (!isMe && showAvatar) {
      return Align(
        alignment: Alignment.centerLeft,
        child: GestureDetector(
          onTap: _selectMode ? () => _toggleSelect(msg) : null,
          child: Container(
            margin: const EdgeInsets.only(bottom: 6),
            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.85),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (_selectMode)
                  Padding(
                    padding: const EdgeInsets.only(right: 8, top: 4),
                    child: Icon(
                      isSelected ? Icons.check_circle : Icons.radio_button_unchecked,
                      color: isSelected ? AppColors.primary : AppColors.offlineGrey,
                      size: 22,
                    ),
                  ),
                CircleAvatar(
                  radius: 16,
                  backgroundImage: _otherUser?.avatarUrl != null
                      ? CachedNetworkImageProvider(_otherUser!.avatarUrl!)
                      : null,
                  child: _otherUser?.avatarUrl == null
                      ? Text(
                          (_otherUser?.name ?? '?').isNotEmpty
                              ? (_otherUser!.name ?? '?')[0].toUpperCase()
                              : '?',
                          style: const TextStyle(fontSize: 14),
                        )
                      : null,
                ),
                const SizedBox(width: 8),
                Flexible(child: bubble),
              ],
            ),
          ),
        ),
      );
    }

    if (isMe) {
      return Align(
        alignment: Alignment.centerRight,
        child: GestureDetector(
          onTap: _selectMode ? () => _toggleSelect(msg) : null,
          child: Container(
            margin: const EdgeInsets.only(bottom: 6),
            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.80),
            child: messageRow,
          ),
        ),
      );
    }

    return Align(
      alignment: Alignment.centerLeft,
      child: GestureDetector(
        onTap: _selectMode ? () => _toggleSelect(msg) : null,
        child: Container(
          margin: EdgeInsets.only(bottom: 6, left: _selectMode ? 0 : 40),
          constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.80),
          child: messageRow,
        ),
      ),
    );
  }

  Widget _buildReplyQuote(ReplyMessage reply, bool isMe) {
    String preview;
    switch (reply.type) {
      case 'image':
        preview = '[Hình ảnh]';
        break;
      case 'audio':
        preview = '[Tin nhắn thoại]';
        break;
      case 'listing_link':
        preview = '[Tin đăng]';
        break;
      default:
        preview = reply.content.length > 80 ? '${reply.content.substring(0, 80)}...' : reply.content;
    }
    final isMeSender = reply.senderId == _currentUserId;
    final senderName = isMeSender ? 'Bạn' : (_otherUser?.name ?? 'Đối phương');

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: isMe ? Colors.white.withValues(alpha: 0.15) : AppColors.border.withValues(alpha: 0.3),
        borderRadius: BorderRadius.circular(8),
        border: Border(
          left: BorderSide(
            color: isMe ? Colors.white70 : AppColors.primary,
            width: 3,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            senderName,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: isMe ? Colors.white : AppColors.primary,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            preview,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              fontSize: 12,
              color: isMe ? Colors.white70 : AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReactionsRow(Message msg) {
    // Group reactions by emoji with count
    final Map<String, List<String>> grouped = {};
    for (final r in msg.reactions) {
      grouped.putIfAbsent(r.emoji, () => []).add(r.userId);
    }

    return Padding(
      padding: const EdgeInsets.only(top: 2),
      child: Wrap(
        spacing: 4,
        children: grouped.entries.map((e) {
          final hasMyReaction = e.value.contains(_currentUserId);
          return GestureDetector(
            onTap: () => _toggleReaction(msg, e.key),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: hasMyReaction
                    ? AppColors.primary.withValues(alpha: 0.15)
                    : AppColors.chatBubbleOther,
                borderRadius: BorderRadius.circular(12),
                border: hasMyReaction
                    ? Border.all(color: AppColors.primary.withValues(alpha: 0.4))
                    : null,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(e.key, style: const TextStyle(fontSize: 14)),
                  if (e.value.length > 1) ...[
                    const SizedBox(width: 2),
                    Text('${e.value.length}', style: TextStyle(fontSize: 11, color: AppColors.textSecondary)),
                  ],
                ],
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  void _showFullImage(String url) {
    showDialog(
      context: context,
      builder: (ctx) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.all(16),
        child: Stack(
          children: [
            Center(
              child: InteractiveViewer(
                child: CachedNetworkImage(imageUrl: url, fit: BoxFit.contain),
              ),
            ),
            Positioned(
              top: 0,
              right: 0,
              child: IconButton(
                icon: const Icon(Icons.close, color: Colors.white, size: 30),
                onPressed: () => Navigator.pop(ctx),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRecordingBar() {
    final mm = (_recordSeconds ~/ 60).toString().padLeft(2, '0');
    final ss = (_recordSeconds % 60).toString().padLeft(2, '0');

    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4)],
      ),
      child: SafeArea(
        child: Row(
          children: [
            IconButton(
              icon: const Icon(Icons.delete_outline, color: AppColors.error),
              onPressed: _cancelRecording,
            ),
            const SizedBox(width: 8),
            const Icon(Icons.circle, color: AppColors.error, size: 12),
            const SizedBox(width: 8),
            Text(
              '$mm:$ss',
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
            ),
            const Spacer(),
            IconButton.filled(
              onPressed: _stopAndSendRecording,
              icon: const Icon(Icons.send),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInputBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4)],
      ),
      child: SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Reply preview bar
            if (_replyingTo != null)
              Container(
                padding: const EdgeInsets.fromLTRB(12, 8, 4, 8),
                decoration: BoxDecoration(
                  color: AppColors.chatBubbleOther.withValues(alpha: 0.5),
                  border: Border(left: BorderSide(color: AppColors.primary, width: 3)),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.reply, size: 16, color: AppColors.primary),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            _replyingTo!.senderId == _currentUserId ? 'Bạn' : (_otherUser?.name ?? 'Đối phương'),
                            style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primary),
                          ),
                          Text(
                            _replyingTo!.type == 'image' ? '[Hình ảnh]'
                                : _replyingTo!.type == 'audio' ? '[Tin nhắn thoại]'
                                : _replyingTo!.content.length > 50 ? '${_replyingTo!.content.substring(0, 50)}...' : _replyingTo!.content,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(fontSize: 12, color: AppColors.textSecondary),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close, size: 18),
                      onPressed: () => setState(() => _replyingTo = null),
                      padding: EdgeInsets.zero,
                      constraints: const BoxConstraints(),
                    ),
                  ],
                ),
              ),
            Row(
          children: [
            _uploadingImage
                ? const Padding(
                    padding: EdgeInsets.all(8),
                    child: SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : IconButton(
                    icon: const Icon(Icons.image, color: AppColors.primary),
                    onPressed: _showImageSourcePicker,
                  ),
            _uploadingAudio
                ? const Padding(
                    padding: EdgeInsets.all(8),
                    child: SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : IconButton(
                    icon: const Icon(Icons.mic, color: AppColors.warning),
                    onPressed: _startRecording,
                  ),
            Expanded(
              child: TextField(
                controller: _msgCtrl,
                decoration: const InputDecoration(
                  hintText: 'Nhập tin nhắn...',
                  border: OutlineInputBorder(),
                  contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  isDense: true,
                ),
                textInputAction: TextInputAction.send,
                onChanged: (_) => _onTyping(),
                onSubmitted: (_) => _send(),
              ),
            ),
            const SizedBox(width: 8),
            IconButton.filled(
              onPressed: _send,
              icon: const Icon(Icons.send),
            ),
          ],
        ),
          ],
        ),
      ),
    );
  }
}
