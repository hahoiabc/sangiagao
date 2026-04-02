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

class _MarqueeTextState extends State<MarqueeText> with SingleTickerProviderStateMixin {
  late final ScrollController _scrollController;
  double _textWidth = 0;
  double _containerWidth = 0;
  bool _needsScroll = false;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    WidgetsBinding.instance.addPostFrameCallback((_) => _startScroll());
  }

  void _startScroll() async {
    if (!mounted) return;

    // Measure text
    final textPainter = TextPainter(
      text: TextSpan(text: widget.text, style: widget.style),
      maxLines: 1,
      textDirection: TextDirection.ltr,
    )..layout();
    _textWidth = textPainter.width;

    await Future.delayed(const Duration(milliseconds: 100));
    if (!mounted) return;

    _containerWidth = _scrollController.position.viewportDimension;
    _needsScroll = _textWidth > _containerWidth;

    if (!_needsScroll) return;

    _animate();
  }

  void _animate() async {
    while (mounted && _needsScroll) {
      final maxScroll = _scrollController.position.maxScrollExtent;
      if (maxScroll <= 0) break;

      final duration = Duration(
        milliseconds: (maxScroll / widget.velocity * 1000).toInt(),
      );

      await _scrollController.animateTo(
        maxScroll,
        duration: duration,
        curve: Curves.linear,
      );

      if (!mounted) break;
      await Future.delayed(const Duration(seconds: 1));
      if (!mounted) break;

      _scrollController.jumpTo(0);
      await Future.delayed(const Duration(seconds: 1));
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
          Text(widget.text, style: widget.style, maxLines: 1),
          if (_needsScroll) ...[
            SizedBox(width: _containerWidth * 0.5),
            Text(widget.text, style: widget.style, maxLines: 1),
          ],
        ],
      ),
    );
  }
}
