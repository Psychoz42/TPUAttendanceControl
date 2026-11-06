package com.example.tputracking.domain.repository

import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.model.User
import kotlinx.coroutines.flow.Flow

interface AuthRepository {

    suspend fun login(
        email: String,
        password: String
    ): AuthResult

    suspend fun logout()

    suspend fun isLoggedIn(): Boolean

    fun getUser(): Flow<User?>
}
