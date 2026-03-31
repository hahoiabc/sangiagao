package com.sangiagao.rice_marketplace

import android.os.Bundle
import android.util.Log
import com.google.firebase.messaging.RemoteMessage
import com.hiennv.flutter_callkit_incoming.CallkitIncomingBroadcastReceiver
import com.hiennv.flutter_callkit_incoming.CallkitConstants
import com.hiennv.flutter_callkit_incoming.Data
import io.flutter.plugins.firebase.messaging.FlutterFirebaseMessagingService

/**
 * Native Android FCM handler for incoming calls.
 *
 * Why: Flutter's Dart background handler does NOT fire when FCM sends
 * notification+data hybrid messages (Android shows system notification instead).
 * And data-only messages may be blocked by OEM battery optimization.
 *
 * This service intercepts FCM at the native Android layer and shows
 * CallKit full-screen notification directly — no Dart isolate needed.
 * For all other message types, delegates to Flutter's default handler.
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
            // Handle incoming call NATIVELY — bypass Dart isolate entirely
            showCallKitNative(data)
            // Also let Dart handler run as backup (won't duplicate — same call ID)
        }

        // Delegate ALL messages to Flutter's handler (for Dart onMessage/onBackgroundMessage)
        super.onMessageReceived(remoteMessage)
    }

    /**
     * Show CallKit incoming call notification using the plugin's native API.
     * This creates a full-screen notification with accept/decline buttons,
     * ringtone, and wake-lock — identical to calling showCallkitIncoming() from Dart.
     */
    private fun showCallKitNative(data: Map<String, String>) {
        try {
            val callerName = data["caller_name"] ?: "Người gọi"
            val callType = data["call_type"] ?: "audio"
            val convId = data["conversation_id"] ?: ""
            val callId = data["call_id"] ?: convId
            val callerId = data["caller_id"] ?: ""

            Log.d(TAG, "Showing CallKit natively: caller=$callerName, callId=$callId")

            // Build the same data structure that FlutterCallkitIncoming expects
            val callData = Bundle().apply {
                putString(CallkitConstants.EXTRA_CALLKIT_ID, callId)
                putString(CallkitConstants.EXTRA_CALLKIT_NAME_CALLER, callerName)
                putString(CallkitConstants.EXTRA_CALLKIT_APP_NAME, "SanGiaGao")
                putInt(CallkitConstants.EXTRA_CALLKIT_TYPE, if (callType == "video") 1 else 0)
                putLong(CallkitConstants.EXTRA_CALLKIT_DURATION, 60000L)

                // Android-specific params
                putBoolean(CallkitConstants.EXTRA_CALLKIT_IS_SHOW_LOGO, false)
                putBoolean(CallkitConstants.EXTRA_CALLKIT_IS_SHOW_FULL_LOCKED_SCREEN, true)
                putString(CallkitConstants.EXTRA_CALLKIT_RINGTONE_PATH, "system_ringtone_default")
                putString(CallkitConstants.EXTRA_CALLKIT_BACKGROUND_COLOR, "#1a1a2e")
                putString(CallkitConstants.EXTRA_CALLKIT_ACTION_COLOR, "#4CAF50")

                // Extra data for Dart callback (accept/decline handlers)
                val extra = HashMap<String, Any>()
                extra["conversation_id"] = convId
                extra["call_type"] = callType
                extra["caller_id"] = callerId
                extra["caller_name"] = callerName
                putSerializable(CallkitConstants.EXTRA_CALLKIT_EXTRA, extra)
            }

            // Send broadcast to CallkitIncomingBroadcastReceiver — same mechanism as Dart plugin
            val intent = CallkitIncomingBroadcastReceiver.getIntentIncoming(
                applicationContext,
                callData
            )
            sendBroadcast(intent)

            Log.d(TAG, "CallKit broadcast sent successfully")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to show CallKit natively", e)
        }
    }
}
