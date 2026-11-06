package com.example.tputracking.presentation.scanner

import com.example.tputracking.domain.model.ScanResult

data class QRScannerUiState(

    val isLoading: Boolean = false,

    val isScanning: Boolean = true,

    val scanResult: ScanResult? = null,

    val errorMessage: String? = null
)