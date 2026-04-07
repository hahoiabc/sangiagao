import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/providers.dart';
import '../screens/auth/login_screen.dart';
import '../screens/auth/register_screen.dart';
import '../screens/auth/role_screen.dart';
import '../screens/auth/forgot_password_screen.dart';
import '../screens/marketplace/price_board_screen.dart';
import '../screens/marketplace/marketplace_screen.dart';
import '../screens/marketplace/listing_detail_screen.dart';
import '../screens/listings/my_listings_screen.dart';
import '../screens/listings/create_listing_screen.dart';
import '../screens/listings/edit_listing_screen.dart';
import '../screens/listings/quick_batch_screen.dart';
import '../screens/chat/inbox_screen.dart';
import '../screens/chat/chat_screen.dart';
import '../screens/profile/profile_screen.dart';
import '../screens/profile/seller_profile_screen.dart';
import '../screens/profile/subscription_screen.dart';
import '../screens/profile/change_password_screen.dart';
import '../screens/profile/change_phone_screen.dart';
import '../screens/profile/privacy_policy_screen.dart';
import '../screens/profile/user_guide_screen.dart';
import '../screens/profile/terms_of_service_screen.dart';
import '../screens/inbox/system_inbox_screen.dart';
import '../screens/inbox/inbox_detail_screen.dart';
import '../screens/notifications/notifications_screen.dart';
import '../screens/feedback/feedback_screen.dart';
import '../screens/feedback/feedback_history_screen.dart';
import '../screens/splash_screen.dart';
import '../widgets/main_shell.dart';

final _uuidRegex = RegExp(r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$');
bool _isValidId(String? id) => id != null && _uuidRegex.hasMatch(id);

// Sanitize deep link query params: only allow alphanumeric, underscore, hyphen
String? _sanitizeParam(String? value) {
  if (value == null) return null;
  final sanitized = value.replaceAll(RegExp(r'[^a-zA-Z0-9_\-]'), '');
  return sanitized.isEmpty ? null : sanitized;
}

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/splash',
    redirect: (context, state) {
      final isAuth = authState.status == AuthStatus.authenticated;
      final loc = state.matchedLocation;
      final publicRoutes = ['/splash', '/login', '/register', '/forgot-password'];

      if (authState.status == AuthStatus.unknown) return null;

      final isPublic = publicRoutes.contains(loc) ||
          loc.startsWith('/marketplace') ||
          loc.startsWith('/seller') ||
          loc == '/privacy-policy' ||
          loc == '/terms-of-service' ||
          loc == '/user-guide';
      if (!isAuth && !isPublic) {
        return '/marketplace';
      }

      if (isAuth && (loc == '/login' || loc == '/register' || loc == '/forgot-password' || loc == '/splash')) {
        return '/marketplace';
      }

      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, __) => const SplashScreen()),
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      GoRoute(path: '/register', builder: (_, __) => const RegisterScreen()),
      GoRoute(path: '/forgot-password', builder: (_, __) => const ForgotPasswordScreen()),
      GoRoute(path: '/select-role', builder: (_, __) => const RoleScreen()),
      ShellRoute(
        builder: (_, __, child) => MainShell(child: child),
        routes: [
          GoRoute(path: '/marketplace', builder: (_, __) => const PriceBoardScreen()),
          GoRoute(
            path: '/marketplace/search',
            builder: (_, state) => MarketplaceScreen(
              initialCategory: _sanitizeParam(state.uri.queryParameters['category']),
              initialType: _sanitizeParam(state.uri.queryParameters['type']),
              initialSort: _sanitizeParam(state.uri.queryParameters['sort']),
            ),
          ),
          GoRoute(
            path: '/marketplace/:id',
            redirect: (_, state) => _isValidId(state.pathParameters['id']) ? null : '/marketplace',
            builder: (_, state) => ListingDetailScreen(id: state.pathParameters['id']!),
          ),
          GoRoute(path: '/my-listings', builder: (_, __) => const MyListingsScreen()),
          GoRoute(path: '/create-listing', builder: (_, __) => const CreateListingScreen()),
          GoRoute(path: '/quick-batch', builder: (_, __) => const QuickBatchScreen()),
          GoRoute(
            path: '/edit-listing/:id',
            redirect: (_, state) => _isValidId(state.pathParameters['id']) ? null : '/my-listings',
            builder: (_, state) => EditListingScreen(listingId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/inbox', builder: (_, __) => const InboxScreen()),
          GoRoute(
            path: '/chat/:id',
            redirect: (_, state) => _isValidId(state.pathParameters['id']) ? null : '/inbox',
            builder: (_, state) => ChatScreen(
              conversationId: state.pathParameters['id']!,
            ),
          ),
          GoRoute(path: '/profile', builder: (_, __) => const ProfileScreen()),
          GoRoute(
            path: '/seller/:id',
            redirect: (_, state) => _isValidId(state.pathParameters['id']) ? null : '/marketplace',
            builder: (_, state) => SellerProfileScreen(sellerId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/subscription', builder: (_, __) => const SubscriptionScreen()),
          GoRoute(path: '/change-password', builder: (_, __) => const ChangePasswordScreen()),
          GoRoute(path: '/change-phone', builder: (_, __) => const ChangePhoneScreen()),
          GoRoute(path: '/system-inbox', builder: (_, __) => const SystemInboxScreen()),
          GoRoute(
            path: '/system-inbox/:id',
            redirect: (_, state) => _isValidId(state.pathParameters['id']) ? null : '/system-inbox',
            builder: (_, state) => InboxDetailScreen(id: state.pathParameters['id']!),
          ),
          GoRoute(path: '/notifications', builder: (_, __) => const NotificationsScreen()),
          GoRoute(path: '/feedback', builder: (_, __) => const FeedbackScreen()),
          GoRoute(path: '/feedback-history', builder: (_, __) => const FeedbackHistoryScreen()),
          GoRoute(path: '/privacy-policy', builder: (_, __) => const PrivacyPolicyScreen()),
          GoRoute(path: '/terms-of-service', builder: (_, __) => const TermsOfServiceScreen()),
          GoRoute(path: '/user-guide', builder: (_, __) => const UserGuideScreen()),
        ],
      ),
    ],
  );
});
