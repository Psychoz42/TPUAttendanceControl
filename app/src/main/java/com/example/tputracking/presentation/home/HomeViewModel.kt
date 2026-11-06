package com.example.tputracking.presentation.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.paging.PagingData
import androidx.paging.cachedIn
import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.domain.repository.AttendanceRepository
import com.example.tputracking.domain.repository.AuthRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class HomeViewModel(
    private val authRepository: AuthRepository,
    private val attendanceRepository: AttendanceRepository
) : ViewModel() {

    private val _uiState =
        MutableStateFlow(HomeUiState())

    val uiState: StateFlow<HomeUiState> =
        _uiState.asStateFlow()

    val attendancePaging: Flow<PagingData<Attendance>> =
        attendanceRepository
            .getAttendancePaging()
            .cachedIn(viewModelScope)

    init {
        observeUser()
    }

    private fun observeUser() {
        viewModelScope.launch {
            authRepository.getUser().collectLatest { user ->
                _uiState.value = _uiState.value.copy(
                    user = user
                )
            }
        }
    }
}
