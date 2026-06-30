import 'package:flutter/material.dart';

/// Lấy 1 trang dữ liệu (page bắt đầu từ 1). Trả về danh sách item của trang đó.
typedef PageFetcher<T> = Future<List<T>> Function(int page);
typedef PagedItemBuilder<T> = Widget Function(BuildContext context, T item, int index);

/// Danh sách cuộn vô hạn dùng chung: tự tải thêm khi cuộn gần cuối, kéo-để-làm-mới,
/// spinner ở đáy, chống tải trùng. "Còn dữ liệu" suy ra từ "lô trả về đầy pageSize".
///
/// Gọi refresh() qua `GlobalKey<PaginatedListViewState>` để tải lại từ đầu (vd sau khi
/// đánh dấu đã đọc).
class PaginatedListView<T> extends StatefulWidget {
  final PageFetcher<T> fetcher;
  final PagedItemBuilder<T> itemBuilder;

  /// Widget cố định nằm TRÊN danh sách (cuộn cùng list). Vd: thẻ thống kê.
  final Widget? header;

  /// Hiện khi không có item nào.
  final Widget? emptyState;

  final EdgeInsetsGeometry padding;
  final int pageSize;

  const PaginatedListView({
    super.key,
    required this.fetcher,
    required this.itemBuilder,
    this.header,
    this.emptyState,
    this.padding = EdgeInsets.zero,
    this.pageSize = 20,
  });

  @override
  State<PaginatedListView<T>> createState() => PaginatedListViewState<T>();
}

class PaginatedListViewState<T> extends State<PaginatedListView<T>> {
  final List<T> _items = [];
  final ScrollController _scroll = ScrollController();
  int _page = 0;
  bool _loading = true;
  bool _loadingMore = false;
  bool _error = false;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _scroll.addListener(_onScroll);
    _loadFirst();
  }

  @override
  void dispose() {
    _scroll.removeListener(_onScroll);
    _scroll.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (!_scroll.hasClients) return;
    if (_scroll.position.pixels >= _scroll.position.maxScrollExtent - 320) {
      _loadMore();
    }
  }

  Future<void> _loadFirst() async {
    setState(() {
      _loading = true;
      _error = false;
    });
    try {
      final batch = await widget.fetcher(1);
      if (!mounted) return;
      setState(() {
        _items
          ..clear()
          ..addAll(batch);
        _page = 1;
        _hasMore = batch.length >= widget.pageSize;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = true;
      });
    }
  }

  Future<void> _loadMore() async {
    if (_loadingMore || _loading || !_hasMore) return;
    setState(() => _loadingMore = true);
    try {
      final batch = await widget.fetcher(_page + 1);
      if (!mounted) return;
      setState(() {
        _page += 1;
        _items.addAll(batch);
        _hasMore = batch.length >= widget.pageSize;
        _loadingMore = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loadingMore = false);
    }
  }

  /// Tải lại từ trang đầu (dùng sau thao tác làm đổi dữ liệu).
  Future<void> refresh() => _loadFirst();

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return ListView(
        padding: widget.padding,
        children: [
          if (widget.header != null) widget.header!,
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 56),
            child: Center(child: CircularProgressIndicator()),
          ),
        ],
      );
    }

    if (_error && _items.isEmpty) {
      return RefreshIndicator(
        onRefresh: _loadFirst,
        child: ListView(
          padding: widget.padding,
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            if (widget.header != null) widget.header!,
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 48),
              child: Center(
                child: Text('Không tải được dữ liệu.\nKéo xuống để thử lại.',
                    textAlign: TextAlign.center, style: TextStyle(color: Colors.grey)),
              ),
            ),
          ],
        ),
      );
    }

    final showEmpty = _items.isEmpty;
    final hasHeader = widget.header != null;
    final itemCount =
        (hasHeader ? 1 : 0) + (showEmpty ? 1 : _items.length) + (_hasMore && !showEmpty ? 1 : 0);

    return RefreshIndicator(
      onRefresh: _loadFirst,
      child: ListView.builder(
        controller: _scroll,
        padding: widget.padding,
        physics: const AlwaysScrollableScrollPhysics(),
        itemCount: itemCount,
        itemBuilder: (context, index) {
          var idx = index;
          if (hasHeader) {
            if (idx == 0) return widget.header!;
            idx -= 1;
          }
          if (showEmpty) {
            return widget.emptyState ??
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 48),
                  child: Center(
                    child: Text('Chưa có dữ liệu', style: TextStyle(color: Colors.grey)),
                  ),
                );
          }
          if (idx < _items.length) {
            return widget.itemBuilder(context, _items[idx], idx);
          }
          // Spinner tải thêm ở đáy
          return const Padding(
            padding: EdgeInsets.symmetric(vertical: 20),
            child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
          );
        },
      ),
    );
  }
}
