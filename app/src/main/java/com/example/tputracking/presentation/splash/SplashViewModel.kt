package com.example.tputracking.presentation.splash

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.tputracking.domain.repository.AuthRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class SplashViewModel(
    private val repository: AuthRepository
) : ViewModel() {

    private val _uiState =
        MutableStateFlow(
            SplashUiState()
        )

    val uiState: StateFlow<SplashUiState> =
        _uiState.asStateFlow()

    /** null, пока состояние не прочитано; false после 401, выхода или истечения токена. */
    val sessionActive: StateFlow<Boolean?> =
        repository.sessionActive()
            .stateIn(viewModelScope, SharingStarted.Eagerly, null)

    init {
        checkSession()
    }

    private fun checkSession() {

        viewModelScope.launch {

            delay(1000)

            val isLoggedIn =
                repository.isLoggedIn()

            _uiState.value =
                SplashUiState(
                    isLoading = false,
                    isAuthorized = isLoggedIn
                )
        }
    }
}
