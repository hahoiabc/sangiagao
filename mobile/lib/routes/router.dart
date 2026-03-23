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
import '../screens/chat/inbox_screen.dart';
import '../screens/chat/chat_screen.dart';
import '../screens/profile/profile_screen.dart';
import '../screens/profile/seller_profile_screen.dart';
import '../screens/profile/subscription_screen.dart';
import '../screens/profile/change_password_screen.dart';
import '../screens/profile/change_phone_screen.dart';
import '../screens/notifications/notifications_screen.dart';
import '../screens/feedback/feedback_screen.dart';
import '../screens/feedback/feedback_history_screen.dart';
import '../screens/splash_screen.dart';
import '../widgets/main_shell.dart';

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
          loc.startsWith('/seller');
      if (!isAuth && !isPublic) {
        return '/marketplace';
      }

      if (isAuth && (loc == '/login' || loc == '/register' || loc == '/forgot-password')) {
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
              initialCategory: state.uri.queryParameters['category'],
              initialType: state.uri.queryParameters['type'],
              initialSort: state.uri.queryParameters['sort'],
            ),
          ),
          GoRoute(
            path: '/marketplace/:id',
            builder: (_, state) => ListingDetailScreen(id: state.pathParameters['id']!),
          ),
          GoRoute(path: '/my-listings', builder: (_, __) => const MyListingsScreen()),
          GoRoute(path: '/create-listing', builder: (_, __) => const CreateListingScreen()),
          GoRoute(
            path: '/edit-listing/:id',
            builder: (_, state) => EditListingScreen(listingId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/inbox', builder: (_, __) => const InboxScreen()),
          GoRoute(
            path: '/chat/:id',
            builder: (_, state) => ChatScreen(conversationId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/profile', builder: (_, __) => const ProfileScreen()),
          GoRoute(
            path: '/seller/:id',
            builder: (_, state) => SellerProfileScreen(sellerId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/subscription', builder: (_, __) => const SubscriptionScreen()),
          GoRoute(path: '/change-password', builder: (_, __) => const ChangePasswordScreen()),
          GoRoute(path: '/change-phone', builder: (_, __) => const ChangePhoneScreen()),
          GoRoute(path: '/notifications', builder: (_, __) => const NotificationsScreen()),
          GoRoute(path: '/feedback', builder: (_, __) => const FeedbackScreen()),
          GoRoute(path: '/feedback-history', builder: (_, __) => const FeedbackHistoryScreen()),
        ],
      ),
    ],
  );
});
