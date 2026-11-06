package com.example.tputracking.domain.model

data class Attendance(
    val id: Long,
    val subject: String,
    val teacher: String,
    val dateTime: String,
    val visited: Boolean
)