package com.example.tputracking.presentation.scanner

import com.example.tputracking.domain.model.ScanResult

data class QRScannerUiState(

    val isLoading: Boolean = false,

    val isScanning: Boolean = true,

    val scanResult: ScanResult? = null,

    // Ошибка, которую показываем блокирующим диалогом
    val errorMessage: String? = null,

    // Неблокирующая подсказка (устаревший или посторонний QR): сканирование продолжается
    val hint: ScanHint? = null
)

/** id меняется при каждой подсказке, чтобы одинаковый текст показался снова. */
data class ScanHint(
    val id: Long,
    val message: String
)
