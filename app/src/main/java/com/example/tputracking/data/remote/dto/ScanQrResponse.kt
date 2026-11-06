package com.example.tputracking.data.remote.dto

// Ошибки сканирования приходят с HTTP 200 в этом же объекте: {"success": false, "code": "...", "message": "..."}
data class ScanQrResponse(
    val success: Boolean? = null,
    val code: String? = null,
    val message: String? = null,
    val subject: String? = null,
    val time: String? = null
)
