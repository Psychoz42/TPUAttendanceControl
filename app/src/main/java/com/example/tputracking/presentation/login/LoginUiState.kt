package com.example.tputracking.presentation.login
data class LoginUiState(
    val email: String = "",
    val password: String = "",
    val isPasswordVisible: Boolean = false,
    val emailError: String? = null,
    val passwordError: String? = null,
    val isLoading: Boolean = false,
    val isAuthorized: Boolean = false,
    val errorMessage: String? = null
)