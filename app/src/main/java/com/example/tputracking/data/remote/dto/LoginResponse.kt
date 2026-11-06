package com.example.tputracking.data.remote.dto

data class LoginResponse(
    val success: Boolean? = null,
    val message: String? = null,
    val username: String? = null,
    val fullName: String? = null,
    val email: String? = null,
    val group: String? = null
)
