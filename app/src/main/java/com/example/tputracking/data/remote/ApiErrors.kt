package com.example.tputracking.data.remote

import com.example.tputracking.data.remote.dto.ApiErrorDto
import com.google.gson.Gson
import retrofit2.Response

/** Коды ошибок бэкенда, поле `code` в ответе `{"success": false, ...}`. */
object ApiErrorCode {
    const val INVALID_DATA_FORMAT = "INVALID_DATA_FORMAT"
    const val LOGIN_ERROR = "LOGIN_ERROR"
    const val INVALID_QR = "INVALID_QR"
    const val UNKNOWN_USER = "UNKNOWN_USER"
    const val DB_ERROR = "DB_ERROR"
    const val INTERNAL_ERROR = "INTERNAL_ERROR"
    const val UNAUTHORIZED = "UNAUTHORIZED"
    const val FORBIDDEN = "FORBIDDEN"
    const val TOO_MANY_REQUESTS = "TOO_MANY_REQUESTS"
}

private val gson = Gson()

private const val DEFAULT_RETRY_AFTER_SECONDS = 60

/**
 * Код ошибки из ответа с HTTP-статусом не 2xx.
 * Тело может оказаться не JSON (например, HTML от прокси при 502), тогда код берётся из статуса.
 */
fun Response<*>.errorCode(): String? {
    val parsed = try {
        errorBody()?.string()?.let { gson.fromJson(it, ApiErrorDto::class.java) }
    } catch (e: Exception) {
        null
    }
    return parsed?.code ?: when (code()) {
        401 -> ApiErrorCode.UNAUTHORIZED
        403 -> ApiErrorCode.FORBIDDEN
        429 -> ApiErrorCode.TOO_MANY_REQUESTS
        else -> null
    }
}

/** Сколько секунд ждать после 429: заголовок Retry-After, без него минута. */
fun Response<*>.retryAfterSeconds(): Int =
    headers()["Retry-After"]?.trim()?.toIntOrNull()?.takeIf { it > 0 }
        ?: DEFAULT_RETRY_AFTER_SECONDS

/** Сообщения пользователю выбираются по коду. Текст `message` сервера английский и только для логов. */
object ApiErrorMessages {

    const val NO_CONNECTION = "Нет соединения с сервером"

    fun login(code: String?, retryAfterSeconds: Int): String = when (code) {
        ApiErrorCode.LOGIN_ERROR -> "Неправильная почта или пароль"
        ApiErrorCode.TOO_MANY_REQUESTS ->
            "Слишком много попыток входа. Подождите $retryAfterSeconds с"
        ApiErrorCode.INVALID_DATA_FORMAT -> "Не удалось отправить данные. Обновите приложение"
        else -> "Сервер временно недоступен. Попробуйте позже"
    }

    fun scan(code: String?, retryAfterSeconds: Int): String = when (code) {
        ApiErrorCode.INVALID_QR ->
            "QR-код устарел или занятие уже закрыто. Наведите камеру на актуальный код"
        ApiErrorCode.UNKNOWN_USER -> "Ваш аккаунт не найден в системе. Обратитесь к преподавателю"
        ApiErrorCode.UNAUTHORIZED -> "Сессия истекла. Войдите снова"
        ApiErrorCode.FORBIDDEN -> "Отмечаться могут только студенты"
        ApiErrorCode.TOO_MANY_REQUESTS -> "Слишком много попыток. Подождите $retryAfterSeconds с"
        ApiErrorCode.DB_ERROR,
        ApiErrorCode.INTERNAL_ERROR,
        ApiErrorCode.INVALID_DATA_FORMAT -> "Не удалось отметить посещение. Попробуйте ещё раз"
        else -> "Ошибка сервера. Попробуйте позже"
    }
}
