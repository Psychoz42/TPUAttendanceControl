package com.example.tputracking.domain.model

enum class ScanErrorReason {
    /** В QR не код посещения (посторонний QR) */
    INVALID_FORMAT,

    /** Код устарел или занятие закрыто: QR на экране меняется, нужно просто сканировать дальше */
    INVALID_QR,

    /** Токен истёк или неверный */
    UNAUTHORIZED,

    OTHER
}

class ScanException(
    val reason: ScanErrorReason,
    message: String
) : Exception(message)
