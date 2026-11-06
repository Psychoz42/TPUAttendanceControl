package com.example.tputracking.data.local.security

import android.util.Base64
import org.json.JSONObject

/**
 * Срок действия JWT из поля `exp` (секунды Unix).
 * Подпись клиент не проверяет и не должен: это делает сервер.
 */
object JwtExpiry {

    /** true, если срок токена вышел. Токен, который не удалось разобрать, считаем действующим: решит сервер. */
    fun isExpired(token: String, nowMillis: Long = System.currentTimeMillis()): Boolean {
        val expSeconds = try {
            val payload = token.split(".").getOrNull(1) ?: return false
            val json = String(
                Base64.decode(payload, Base64.URL_SAFE or Base64.NO_PADDING or Base64.NO_WRAP),
                Charsets.UTF_8
            )
            JSONObject(json).optLong("exp", 0L)
        } catch (e: Exception) {
            0L
        }
        return expSeconds > 0 && expSeconds * 1000 <= nowMillis
    }
}
