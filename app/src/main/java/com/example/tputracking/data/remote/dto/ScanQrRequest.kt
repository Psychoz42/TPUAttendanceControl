package com.example.tputracking.data.remote.dto

data class ScanQrRequest(
    val qrCode: String,
    val username: String
)