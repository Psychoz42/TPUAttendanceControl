package com.example.tputracking.presentation.history.model

data class AttendanceHistoryUi(
    val id: Long,
    val title: String,
    val dateTime: String,
    val visited: Boolean
)