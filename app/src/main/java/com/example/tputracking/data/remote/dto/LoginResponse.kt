package com.example.tputracking.data.remote.dto

// Ошибки входа приходят с HTTP 200 в этом же объекте: {"success": false, "code": "...", "message": "..."}
data class LoginResponse(
    val success: Boolean? = null,
    val code: String? = null,
    val message: String? = null,
    val username: String? = null,
    val fullName: String? = null,
    val email: String? = null,
    val group: String? = null,
    val role: String? = null,
    val token: String? = null
)
