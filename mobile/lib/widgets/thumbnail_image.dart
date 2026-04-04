import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import '../models/listing.dart';

/// Displays thumbnail image with fallback to original if thumb 404.
class ThumbnailImage extends StatefulWidget {
  final String imageUrl;
  final BoxFit fit;
  final Color? color;
  final BlendMode? colorBlendMode;
  final Widget Function(BuildContext, String)? placeholder;
  final Widget Function(BuildContext, String, dynamic)? errorWidget;

  const ThumbnailImage({
    super.key,
    required this.imageUrl,
    this.fit = BoxFit.cover,
    this.color,
    this.colorBlendMode,
    this.placeholder,
    this.errorWidget,
  });

  @override
  State<ThumbnailImage> createState() => _ThumbnailImageState();
}

class _ThumbnailImageState extends State<ThumbnailImage> {
  late String _currentUrl;
  bool _triedFallback = false;

  @override
  void initState() {
    super.initState();
    _currentUrl = toThumbnailUrl(widget.imageUrl);
  }

  @override
  void didUpdateWidget(ThumbnailImage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.imageUrl != widget.imageUrl) {
      _currentUrl = toThumbnailUrl(widget.imageUrl);
      _triedFallback = false;
    }
  }

  void _onError() {
    if (!_triedFallback && _currentUrl != widget.imageUrl) {
      setState(() {
        _currentUrl = widget.imageUrl;
        _triedFallback = true;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return CachedNetworkImage(
      key: ValueKey(_currentUrl),
      imageUrl: _currentUrl,
      fit: widget.fit,
      color: widget.color,
      colorBlendMode: widget.colorBlendMode,
      placeholder: widget.placeholder,
      errorWidget: (ctx, url, err) {
        if (!_triedFallback && _currentUrl != widget.imageUrl) {
          WidgetsBinding.instance.addPostFrameCallback((_) => _onError());
          return widget.placeholder?.call(ctx, url) ?? const SizedBox.shrink();
        }
        return widget.errorWidget?.call(ctx, url, err) ??
            Container(
              color: Colors.grey[200],
              child: const Center(child: Icon(Icons.broken_image, color: Colors.grey)),
            );
      },
    );
  }
}
