package com.example.tputracking.presentation.home

import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.domain.model.User

data class HomeUiState(

    val isLoading: Boolean = false,

    val isRefreshing: Boolean = false,

    val user: User? = null,

    val attendance: List<Attendance> =
        emptyList(),

    val errorMessage: String? = null
)