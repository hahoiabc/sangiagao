import 'dart:async';
import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:in_app_purchase/in_app_purchase.dart';

import 'api_service.dart';
import '../providers/providers.dart';

/// IAP subscription product IDs registered in both App Store Connect (iOS)
/// and Play Console (Android). Same IDs across stores for consistency.
const Set<String> kAppleSubscriptionIds = {
  'com.sangiagao.premium.1m',
  'com.sangiagao.premium.3m',
  'com.sangiagao.premium.6m',
  'com.sangiagao.premium.12m',
};

/// Alias for clarity — same IDs reused for Google Play.
const Set<String> kIAPSubscriptionIds = kAppleSubscriptionIds;

class IAPState {
  final bool available;
  final bool loading;
  final List<ProductDetails> products;
  final bool processing;
  final String? error;

  const IAPState({
    required this.available,
    required this.loading,
    required this.products,
    this.processing = false,
    this.error,
  });

  IAPState copyWith({
    bool? available,
    bool? loading,
    List<ProductDetails>? products,
    bool? processing,
    String? error,
    bool clearError = false,
  }) =>
      IAPState(
        available: available ?? this.available,
        loading: loading ?? this.loading,
        products: products ?? this.products,
        processing: processing ?? this.processing,
        error: clearError ? null : (error ?? this.error),
      );

  static const initial = IAPState(available: false, loading: true, products: []);
}

class IAPService extends StateNotifier<IAPState> {
  IAPService(this._api) : super(IAPState.initial) {
    if (Platform.isIOS) {
      _init();
    } else {
      state = state.copyWith(loading: false, available: false);
    }
  }

  final ApiService _api;
  final InAppPurchase _iap = InAppPurchase.instance;
  StreamSubscription<List<PurchaseDetails>>? _purchaseSub;

  /// Last verification result returned by backend (used by UI to show success).
  Map<String, dynamic>? lastVerifyResult;

  Future<void> _init() async {
    final available = await _iap.isAvailable();
    if (!available) {
      state = state.copyWith(loading: false, available: false);
      return;
    }

    _purchaseSub = _iap.purchaseStream.listen(
      _onPurchaseUpdate,
      onError: (e) => state = state.copyWith(error: 'Lỗi giao dịch: $e', processing: false),
    );

    await loadProducts();
  }

  Future<void> loadProducts() async {
    state = state.copyWith(loading: true, clearError: true);
    final response = await _iap.queryProductDetails(kIAPSubscriptionIds);
    if (response.notFoundIDs.isNotEmpty) {
      debugPrint('IAP product IDs not found: ${response.notFoundIDs}');
    }
    final sorted = response.productDetails.toList()
      ..sort((a, b) => _monthOrder(a.id).compareTo(_monthOrder(b.id)));
    state = state.copyWith(
      loading: false,
      available: true,
      products: sorted,
    );
  }

  int _monthOrder(String productId) {
    if (productId.endsWith('.1m')) return 1;
    if (productId.endsWith('.3m')) return 3;
    if (productId.endsWith('.6m')) return 6;
    if (productId.endsWith('.12m')) return 12;
    return 99;
  }

  /// Trigger StoreKit purchase prompt for the given product.
  /// Result delivered asynchronously via _onPurchaseUpdate.
  Future<void> buy(ProductDetails product) async {
    state = state.copyWith(processing: true, clearError: true);
    final purchaseParam = PurchaseParam(productDetails: product);
    try {
      // For auto-renewable subscriptions Apple uses non-consumable buy method.
      await _iap.buyNonConsumable(purchaseParam: purchaseParam);
    } catch (e) {
      state = state.copyWith(processing: false, error: 'Không khởi tạo được giao dịch: $e');
    }
  }

  /// Apple-required Restore Purchases button. Triggers re-emission of past
  /// non-consumable + subscription purchases via _onPurchaseUpdate.
  Future<void> restore() async {
    state = state.copyWith(processing: true, clearError: true);
    try {
      await _iap.restorePurchases();
    } catch (e) {
      state = state.copyWith(processing: false, error: 'Khôi phục thất bại: $e');
    }
  }

  Future<void> _onPurchaseUpdate(List<PurchaseDetails> purchases) async {
    for (final p in purchases) {
      switch (p.status) {
        case PurchaseStatus.pending:
          state = state.copyWith(processing: true);
          break;
        case PurchaseStatus.purchased:
        case PurchaseStatus.restored:
          await _verifyAndComplete(p);
          break;
        case PurchaseStatus.error:
          state = state.copyWith(
            processing: false,
            error: p.error?.message ?? 'Giao dịch thất bại',
          );
          if (p.pendingCompletePurchase) {
            await _iap.completePurchase(p);
          }
          break;
        case PurchaseStatus.canceled:
          state = state.copyWith(processing: false);
          if (p.pendingCompletePurchase) {
            await _iap.completePurchase(p);
          }
          break;
      }
    }
  }

  Future<void> _verifyAndComplete(PurchaseDetails purchase) async {
    try {
      if (Platform.isAndroid) {
        // Google: serverVerificationData = purchase token
        final token = purchase.verificationData.serverVerificationData;
        if (token.isEmpty) {
          state = state.copyWith(processing: false, error: 'Thiếu purchase token');
          return;
        }
        lastVerifyResult = await _api.verifyGoogleIAP(
          productId: purchase.productID,
          purchaseToken: token,
        );
      } else {
        // Apple: use StoreKit transaction id
        final txID = purchase.purchaseID;
        if (txID == null || txID.isEmpty) {
          state = state.copyWith(processing: false, error: 'Thiếu transaction id');
          return;
        }
        lastVerifyResult = await _api.verifyAppleIAP(txID);
      }
      state = state.copyWith(processing: false, clearError: true);
    } catch (e) {
      state = state.copyWith(
        processing: false,
        error: 'Xác minh giao dịch thất bại: $e',
      );
    } finally {
      if (purchase.pendingCompletePurchase) {
        await _iap.completePurchase(purchase);
      }
    }
  }

  @override
  void dispose() {
    _purchaseSub?.cancel();
    super.dispose();
  }
}

final iapServiceProvider = StateNotifierProvider<IAPService, IAPState>((ref) {
  return IAPService(ref.read(apiServiceProvider));
});
