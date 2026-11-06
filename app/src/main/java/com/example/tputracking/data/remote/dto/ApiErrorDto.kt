package com.example.tputracking.data.remote.dto

// DTO должны лежать в этом пакете: только его сохраняет правило R8 в proguard-rules.pro
data class ApiErrorDto(
    val success: Boolean? = null,
    val code: String? = null,
    val message: String? = null
)
