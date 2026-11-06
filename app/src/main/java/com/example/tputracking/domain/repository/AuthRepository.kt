package com.example.tputracking.domain.repository

import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.model.User
import kotlinx.coroutines.flow.Flow

interface AuthRepository {

    suspend fun login(
        email: String,
        password: String
    ): AuthResult

    /** Явный выход: удаляет профиль, токен и локальную историю. */
    suspend fun logout()

    /** Есть профиль и действующий токен. Просроченный токен сбрасывается. */
    suspend fun isLoggedIn(): Boolean

    /** true, пока сохранён токен. Становится false после 401, выхода или истечения срока. */
    fun sessionActive(): Flow<Boolean>

    /** Email пользователя, если сессия истекла (профиль есть, токена нет), иначе null. */
    suspend fun getExpiredSessionEmail(): String?

    fun getUser(): Flow<User?>
}
