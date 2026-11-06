package com.example.tputracking.data.remote.dto.schedule

data class ScheduleLessonDto(
    val id: Long,
    val subject: String,
    val teacher: String,
    val room: String,
    val startTime: String,
    val endTime: String
)