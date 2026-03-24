/// Environment configuration loaded via --dart-define at build time.
///
/// Usage:
///   flutter run --dart-define=API_BASE_URL=http://localhost:8080/api/v1
///   flutter build apk --dart-define=API_BASE_URL=https://sangiagao.vn/api/v1
class Env {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://sangiagao.vn/api/v1',
  );

  static const String wsBaseUrl = String.fromEnvironment(
    'WS_BASE_URL',
    defaultValue: 'wss://sangiagao.vn/socket/websocket',
  );
}
