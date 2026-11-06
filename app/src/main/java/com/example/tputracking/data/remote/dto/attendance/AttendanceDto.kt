package com.example.tputracking.data.remote.dto.attendance

data class AttendanceDto(
    val id: Long,
    val subject: String,
    val teacher: String,
    val dateTime: String,
    val visited: Boolean
)