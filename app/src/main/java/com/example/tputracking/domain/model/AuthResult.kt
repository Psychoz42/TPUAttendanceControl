package com.example.tputracking.domain.model

sealed class AuthResult {

    data object Loading : AuthResult()

    data object Success : AuthResult()

    data class Error(
        val message: String
    ) : AuthResult()

    data object Unauthorized : AuthResult()
}