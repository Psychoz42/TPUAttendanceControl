package com.example.tputracking.presentation.history

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.paging.cachedIn
import com.example.tputracking.domain.repository.AttendanceRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class HistoryViewModel(
    private val repository: AttendanceRepository
) : ViewModel() {

    private val _uiState =
        MutableStateFlow(
            HistoryUiState()
        )

    val uiState: StateFlow<HistoryUiState> =
        _uiState.asStateFlow()

    val attendancePaging =
        repository
            .getAttendancePaging()
            .cachedIn(viewModelScope)
}
