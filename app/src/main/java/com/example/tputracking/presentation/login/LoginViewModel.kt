package com.example.tputracking.presentation.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.repository.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class LoginViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    init {
        // Профиль без токена: сессия истекла, подставляем email
        viewModelScope.launch {
            val email = authRepository.getExpiredSessionEmail() ?: return@launch
            if (_uiState.value.email.isEmpty()) {
                _uiState.value = _uiState.value.copy(
                    email = email,
                    infoMessage = "Сессия истекла, войдите снова"
                )
            }
        }
    }

    fun onEmailChanged(value: String) {
        _uiState.value = _uiState.value.copy(
            email = value,
            emailError = null
        )
    }

    fun onPasswordChanged(value: String) {
        _uiState.value = _uiState.value.copy(
            password = value,
            passwordError = null
        )
    }

    fun togglePasswordVisibility() {
        _uiState.value = _uiState.value.copy(
            isPasswordVisible = !_uiState.value.isPasswordVisible
        )
    }

    fun login() {
        val email = _uiState.value.email.trim()
        // Пароль отправляем ровно как ввёл пользователь: пробелы могут быть его частью
        val password = _uiState.value.password

        var emailError: String? = null
        var passwordError: String? = null

        if (email.isEmpty()) {
            emailError = "Введите email"
        } else if (!email.endsWith("@tpu.ru")) {
            emailError = "Используйте TPU email"
        }

        if (password.isEmpty()) {
            passwordError = "Введите пароль"
        }

        _uiState.value = _uiState.value.copy(
            emailError = emailError,
            passwordError = passwordError
        )

        if (emailError == null && passwordError == null) {
            viewModelScope.launch {
                _uiState.value = _uiState.value.copy(
                    isLoading = true,
                    errorMessage = null,
                    infoMessage = null
                )
                
                val result = authRepository.login(email, password)
                
                when (result) {
                    is AuthResult.Success -> {
                        // Пароль больше не нужен — не держим его в памяти
                        _uiState.value = _uiState.value.copy(
                            password = "",
                            isLoading = false,
                            isAuthorized = true
                        )
                    }
                    is AuthResult.Error -> {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            errorMessage = result.message
                        )
                    }
                    AuthResult.Unauthorized -> {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            errorMessage = "Неверный логин или пароль"
                        )
                    }
                    AuthResult.Loading -> { /* Handled by isLoading = true */ }
                }
            }
        }
    }
}
