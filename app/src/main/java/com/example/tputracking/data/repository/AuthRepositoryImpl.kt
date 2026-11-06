package com.example.tputracking.data.repository

import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.local.dao.AttendanceDao
import com.example.tputracking.data.local.security.JwtExpiry
import com.example.tputracking.data.remote.ApiErrorMessages
import com.example.tputracking.data.remote.api.AuthApiService
import com.example.tputracking.data.remote.dto.LoginRequest
import com.example.tputracking.data.remote.dto.LoginResponse
import com.example.tputracking.data.remote.errorCode
import com.example.tputracking.data.remote.retryAfterSeconds
import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.model.User
import com.example.tputracking.domain.repository.AuthRepository
import com.google.gson.JsonParseException
import com.google.gson.stream.MalformedJsonException
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import java.io.IOException

class AuthRepositoryImpl(
    private val api: AuthApiService,
    private val userPreferences: UserPreferences,
    private val attendanceDao: AttendanceDao
) : AuthRepository {

    companion object {
        private const val INCOMPLETE_USER_DATA = "Данные пользователя неполные"
        private const val TEACHER_NOT_ALLOWED =
            "Приложение предназначено для студентов. Преподавателям нужно использовать сайт."
    }

    override suspend fun login(
        email: String,
        password: String
    ): AuthResult {

        try {

            val response = api.login(
                LoginRequest(
                    email = email,
                    password = password
                )
            )

            if (!response.isSuccessful) {
                return AuthResult.Error(
                    ApiErrorMessages.login(response.errorCode(), response.retryAfterSeconds())
                )
            }

            // Сервер отдаёт ошибки входа с HTTP 200, поэтому смотрим на success и code
            val body = response.body()
            if (body == null || body.success == false || body.code != null) {
                return AuthResult.Error(
                    ApiErrorMessages.login(body?.code, response.retryAfterSeconds())
                )
            }

            val user = body.toUserOrNull()
            val token = body.token
            if (user == null || token.isNullOrBlank()) {
                return AuthResult.Error(INCOMPLETE_USER_DATA)
            }

            // Роль проверяем до сохранения сессии: мобилка только для студентов
            when (user.role) {
                User.ROLE_STUDENT -> Unit
                User.ROLE_TEACHER -> return AuthResult.Error(TEACHER_NOT_ALLOWED)
                else -> return AuthResult.Error(INCOMPLETE_USER_DATA)
            }

            // Другой пользователь на этом устройстве не должен видеть чужую историю.
            // Тот же пользователь после истечения сессии сохраняет свою.
            val previousUser = userPreferences.user.first()
            if (previousUser?.username != user.username) {
                attendanceDao.clearAll()
            }

            userPreferences.saveSession(user, token)
            return AuthResult.Success

        } catch (e: MalformedJsonException) {

            // Ответ не JSON (например, HTML от прокси)
            return AuthResult.Error(ApiErrorMessages.login(null, 0))

        } catch (e: JsonParseException) {

            return AuthResult.Error(ApiErrorMessages.login(null, 0))

        } catch (e: IOException) {

            // Не показываем пользователю технические детали исключения (хосты, стек и т.п.)
            return AuthResult.Error(ApiErrorMessages.NO_CONNECTION)

        } catch (e: Exception) {

            return AuthResult.Error("Неизвестная ошибка")
        }
    }

    private fun LoginResponse.toUserOrNull(): User? {
        val fields = listOf(username, fullName, email, group, role)
        if (fields.any { it.isNullOrBlank() }) return null
        return User(
            username = username!!,
            fullName = fullName!!,
            email = email!!,
            group = group!!,
            role = role!!
        )
    }

    override suspend fun logout() {
        // Удаляем профиль, токен и локальную историю посещений,
        // чтобы следующий пользователь устройства их не увидел.
        // Эндпоинта отзыва токена на сервере нет, токен формально действует до exp.
        userPreferences.clear()
        attendanceDao.clearAll()
    }

    override suspend fun isLoggedIn(): Boolean {
        userPreferences.user.first() ?: return false
        val token = userPreferences.token.first() ?: return false
        if (JwtExpiry.isExpired(token)) {
            userPreferences.clearToken()
            return false
        }
        return true
    }

    override fun sessionActive(): Flow<Boolean> {
        return userPreferences.token
            .map { it != null }
            .distinctUntilChanged()
    }

    override suspend fun getExpiredSessionEmail(): String? {
        val user = userPreferences.user.first() ?: return null
        if (userPreferences.token.first() != null) return null
        return user.email.ifBlank { null }
    }

    override fun getUser(): Flow<User?> {
        return userPreferences.user
    }
}
