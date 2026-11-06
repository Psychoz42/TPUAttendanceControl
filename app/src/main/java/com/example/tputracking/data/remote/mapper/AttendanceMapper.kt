package com.example.tputracking.data.mapper

import com.example.tputracking.data.local.entity.AttendanceEntity
import com.example.tputracking.data.remote.dto.attendance.AttendanceDto
import com.example.tputracking.domain.model.Attendance

fun AttendanceDto.toEntity(): AttendanceEntity {

    return AttendanceEntity(
        id = id,
        subject = subject,
        teacher = teacher,
        dateTime = dateTime,
        visited = visited
    )
}

fun AttendanceEntity.toDomain(): Attendance {

    return Attendance(
        id = id,
        subject = subject,
        teacher = teacher,
        dateTime = dateTime,
        visited = visited
    )
}