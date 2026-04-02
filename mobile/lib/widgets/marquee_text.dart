import 'package:flutter/material.dart';

class MarqueeText extends StatefulWidget {
  final String text;
  final TextStyle? style;
  final double velocity; // pixels per second

  const MarqueeText({
    super.key,
    required this.text,
    this.style,
    this.velocity = 40,
  });

  @override
  State<MarqueeText> createState() => _MarqueeTextState();
}

class _MarqueeTextState extends State<MarqueeText> {
  late final ScrollController _scrollController;
  bool _initialized = false;
  double _singleTextWidth = 0;
  double _gap = 0;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    WidgetsBinding.instance.addPostFrameCallback((_) => _init());
  }

  void _init() async {
    if (!mounted) return;

    final textPainter = TextPainter(
      text: TextSpan(text: widget.text, style: widget.style),
      maxLines: 1,
      textDirection: TextDirection.ltr,
    )..layout();
    _singleTextWidth = textPainter.width;

    await Future.delayed(const Duration(milliseconds: 100));
    if (!mounted) return;

    final viewportWidth = _scrollController.position.viewportDimension;
    // Gap = full screen width so text exits completely before next one enters
    _gap = viewportWidth;

    setState(() => _initialized = true);

    await Future.delayed(const Duration(milliseconds: 50));
    if (!mounted) return;

    _animate();
  }

  void _animate() async {
    while (mounted) {
      // Scroll = leading gap + text width = text enters from right, exits left completely
      final scrollDistance = _gap + _singleTextWidth;
      if (scrollDistance <= 0) break;

      final duration = Duration(
        milliseconds: (scrollDistance / widget.velocity * 1000).toInt(),
      );

      await _scrollController.animateTo(
        scrollDistance,
        duration: duration,
        curve: Curves.linear,
      );

      if (!mounted) break;
      _scrollController.jumpTo(0);
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      controller: _scrollController,
      scrollDirection: Axis.horizontal,
      physics: const NeverScrollableScrollPhysics(),
      child: Row(
        children: [
          if (_initialized) SizedBox(width: _gap), // leading gap: text starts off-screen right
          Text(widget.text, style: widget.style, maxLines: 1),
          if (_initialized) SizedBox(width: _gap), // trailing gap: text fully exits left
        ],
      ),
    );
  }
}
