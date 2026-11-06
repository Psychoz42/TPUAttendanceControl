package com.example.tputracking.data.repository

import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.local.dao.AttendanceDao
import com.example.tputracking.data.remote.api.AuthApiService
import com.example.tputracking.data.remote.dto.LoginRequest
import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.model.User
import com.example.tputracking.domain.repository.AuthRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.map
import java.io.IOException

class AuthRepositoryImpl(
    private val api: AuthApiService,
    private val userPreferences: UserPreferences,
    private val attendanceDao: AttendanceDao
) : AuthRepository {

    override suspend fun login(
        email: String,
        password: String
    ): AuthResult {

        return try {

            val response = api.login(
                LoginRequest(
                    email = email,
                    password = password
                )
            )

            if (response.isSuccessful) {

                val body = response.body()

                if (body != null && body.success != false) {

                    // Проверяем, что все нужные поля пришли
                    val username = body.username
                    val fullName = body.fullName
                    val emailVal = body.email
                    val group = body.group

                    if (username != null && fullName != null && emailVal != null && group != null) {
                        userPreferences.saveUser(
                            User(
                                username = username,
                                fullName = fullName,
                                email = emailVal,
                                group = group
                            )
                        )
                        AuthResult.Success
                    } else {
                        AuthResult.Error("Данные пользователя неполные")
                    }

                } else {
                    // Обработка случая {"success": false, "message": "..."}
                    val serverMessage = body?.message ?: ""
                    val errorMessage = if (serverMessage == "user did not found") {
                        "Неправильная почта или пароль."
                    } else {
                        serverMessage.ifEmpty { "Неверный логин или пароль" }
                    }
                    
                    AuthResult.Error(errorMessage)
                }

            } else {

                when (response.code()) {
                    401 -> AuthResult.Unauthorized
                    else -> AuthResult.Error("Ошибка сервера: ${response.code()}")
                }
            }

        } catch (e: IOException) {

            // Не показываем пользователю технические детали исключения (хосты, стек и т.п.)
            AuthResult.Error("Нет соединения с сервером")

        } catch (e: Exception) {

            AuthResult.Error("Неизвестная ошибка")
        }
    }

    override suspend fun logout() {
        // Удаляем и профиль, и локальную историю посещений,
        // чтобы следующий пользователь устройства их не увидел
        userPreferences.clear()
        attendanceDao.clearAll()
    }

    override suspend fun isLoggedIn(): Boolean {
        return userPreferences.user.map { it != null }.firstOrNull() ?: false
    }

    override fun getUser(): Flow<User?> {
        return userPreferences.user
    }
}
