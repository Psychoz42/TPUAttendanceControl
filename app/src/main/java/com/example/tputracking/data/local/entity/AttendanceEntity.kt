package com.example.tputracking.data.local.entity

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(
    tableName = "attendance"
)
data class AttendanceEntity(

    @PrimaryKey
    val id: Long,

    val subject: String,

    val teacher: String,

    val dateTime: String,

    val visited: Boolean
)