package com.example.tputracking.domain.model

data class User(
    val username: String,
    val fullName: String,
    val email: String,
    val group: String,
    val role: String
) {
    companion object {
        const val ROLE_STUDENT = "student"
        const val ROLE_TEACHER = "teacher"
    }
}
