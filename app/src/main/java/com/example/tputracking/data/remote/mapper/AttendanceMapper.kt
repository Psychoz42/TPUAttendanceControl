package com.example.tputracking.data.mapper

import com.example.tputracking.data.local.entity.AttendanceEntity
import com.example.tputracking.domain.model.Attendance

fun AttendanceEntity.toDomain(): Attendance {

    return Attendance(
        id = id,
        subject = subject,
        teacher = teacher,
        dateTime = dateTime,
        visited = visited
    )
}
