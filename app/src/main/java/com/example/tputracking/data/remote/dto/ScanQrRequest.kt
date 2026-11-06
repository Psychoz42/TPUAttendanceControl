package com.example.tputracking.data.remote.dto

// Пользователя сервер берёт из токена, username в теле не нужен
data class ScanQrRequest(
    val qrCode: String
)
