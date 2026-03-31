package com.sangiagao.rice_marketplace

import android.os.Bundle
import android.util.Log
import com.google.firebase.messaging.RemoteMessage
import com.hiennv.flutter_callkit_incoming.CallkitConstants
import com.hiennv.flutter_callkit_incoming.CallkitNotificationManager
import com.hiennv.flutter_callkit_incoming.CallkitSoundPlayerManager
import io.flutter.plugins.firebase.messaging.FlutterFirebaseMessagingService

/**
 * Native Android FCM handler for incoming calls.
 *
 * Problem: When app is killed, FlutterCallkitIncomingPlugin singleton is null.
 * The BroadcastReceiver depends on it → callkitNotificationManager is null → no UI.
 *
 * Solution: Create CallkitNotificationManager DIRECTLY (bypassing plugin singleton)
 * and call showIncomingNotification() with the call data Bundle.
 * This shows full-screen CallKit notification without needing Dart or plugin init.
 */
class CallFirebaseService : FlutterFirebaseMessagingService() {

    companion object {
        private const val TAG = "CallFirebaseService"
    }

    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        val data = remoteMessage.data
        val type = data["type"]

        Log.d(TAG, "FCM received: type=$type, data=$data")

        if (type == "incoming_call") {
            showCallKitNative(data)
        }

        // Delegate to Flutter's handler for Dart onMessage/onBackgroundMessage
        super.onMessageReceived(remoteMessage)
    }

    /**
     * Show CallKit notification by creating CallkitNotificationManager directly.
     * Bypasses the plugin singleton — works even when app is killed.
     */
    private fun showCallKitNative(data: Map<String, String>) {
        try {
            val callerName = data["caller_name"] ?: "Người gọi"
            val callType = data["call_type"] ?: "audio"
            val convId = data["conversation_id"] ?: ""
            val callId = data["call_id"] ?: convId
            val callerId = data["caller_id"] ?: ""

            Log.d(TAG, "Showing CallKit natively: caller=$callerName, callId=$callId")

            // Build the Bundle that CallkitNotificationManager expects
            val callData = Bundle().apply {
                putString(CallkitConstants.EXTRA_CALLKIT_ID, callId)
                putString(CallkitConstants.EXTRA_CALLKIT_NAME_CALLER, callerName)
                putString(CallkitConstants.EXTRA_CALLKIT_APP_NAME, "SanGiaGao")
                putInt(CallkitConstants.EXTRA_CALLKIT_TYPE, if (callType == "video") 1 else 0)
                putLong(CallkitConstants.EXTRA_CALLKIT_DURATION, 60000L)

                // Android display params
                putBoolean(CallkitConstants.EXTRA_CALLKIT_IS_SHOW_LOGO, false)
                putBoolean(CallkitConstants.EXTRA_CALLKIT_IS_SHOW_FULL_LOCKED_SCREEN, true)
                putString(CallkitConstants.EXTRA_CALLKIT_RINGTONE_PATH, "system_ringtone_default")
                putString(CallkitConstants.EXTRA_CALLKIT_BACKGROUND_COLOR, "#1a1a2e")
                putString(CallkitConstants.EXTRA_CALLKIT_ACTION_COLOR, "#4CAF50")

                // Extra data for Dart callbacks (accept/decline handlers read this)
                val extra = HashMap<String, Any>()
                extra["conversation_id"] = convId
                extra["call_type"] = callType
                extra["caller_id"] = callerId
                extra["caller_name"] = callerName
                putSerializable(CallkitConstants.EXTRA_CALLKIT_EXTRA, extra)
            }

            // Create notification manager DIRECTLY — bypass plugin singleton
            val soundPlayer = CallkitSoundPlayerManager(applicationContext)
            val notifManager = CallkitNotificationManager(applicationContext, soundPlayer)

            // This creates a full-screen notification with accept/decline buttons,
            // ringtone, vibration, and wake-lock — same as Dart showCallkitIncoming()
            notifManager.showIncomingNotification(callData)

            Log.d(TAG, "CallKit notification shown successfully via direct manager")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to show CallKit natively", e)
        }
    }
}
