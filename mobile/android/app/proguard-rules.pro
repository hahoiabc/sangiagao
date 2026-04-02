# Flutter
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.** { *; }
-keep class io.flutter.util.** { *; }
-keep class io.flutter.view.** { *; }
-keep class io.flutter.** { *; }
-keep class io.flutter.plugins.** { *; }

# Dio / OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keep class okhttp3.** { *; }
-keep class okio.** { *; }

# Google Play Core (deferred components)
-dontwarn com.google.android.play.core.splitcompat.SplitCompatApplication
-dontwarn com.google.android.play.core.splitinstall.**
-dontwarn com.google.android.play.core.tasks.**

# WebSocket
-keep class org.java_websocket.** { *; }
-dontwarn org.java_websocket.**

# Keep JSON model classes (Flutter uses dart:convert, but native side needs these)
-keep class com.sangiagao.rice_marketplace.** { *; }

# Kotlin
-dontwarn kotlin.**
-keep class kotlin.** { *; }

# Firebase
-keep class com.google.firebase.** { *; }
-dontwarn com.google.firebase.**

# image_picker
-keep class io.flutter.plugins.imagepicker.** { *; }
-dontwarn io.flutter.plugins.imagepicker.**

# audioplayers
-keep class xyz.luan.audioplayers.** { *; }
-dontwarn xyz.luan.audioplayers.**

# cached_network_image / flutter_cache_manager
-keep class com.baseflow.** { *; }
-dontwarn com.baseflow.**
-keep class io.flutter.plugins.flutter_plugin_android_lifecycle.** { *; }

