package com.example.tputracking.data.remote

import com.example.tputracking.data.local.UserPreferences
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

/**
 * Добавляет `Authorization: Bearer <token>` ко всем запросам `/api/v1/attendance/`.
 * На `/auth/login` заголовок не нужен. Ответ 401 сбрасывает токен: приложение откроет экран входа.
 */
class AuthInterceptor(
    private val userPreferences: UserPreferences
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        if (!request.url.encodedPath.startsWith("/api/v1/attendance/")) {
            return chain.proceed(request)
        }

        val token = runBlocking { userPreferences.token.first() }
        val authorized = if (token != null) {
            request.newBuilder().header("Authorization", "Bearer $token").build()
        } else {
            request
        }

        val response = chain.proceed(authorized)
        if (response.code == 401) {
            runBlocking { userPreferences.clearToken() }
        }
        return response
    }
}
