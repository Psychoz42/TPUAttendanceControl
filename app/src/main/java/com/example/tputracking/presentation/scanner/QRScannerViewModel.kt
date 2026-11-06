package com.example.tputracking.presentation.scanner

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.tputracking.domain.repository.AttendanceRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class QRScannerViewModel(
    application: Application,
    private val repository: AttendanceRepository
) : AndroidViewModel(application) {

    private val _uiState =
        MutableStateFlow(QRScannerUiState())

    val uiState: StateFlow<QRScannerUiState> =
        _uiState.asStateFlow()

    private var isProcessing = false

    fun onQrCodeScanned(
        qrCode: String
    ) {

        if (isProcessing) return

        isProcessing = true

        viewModelScope.launch {

            _uiState.value =
                _uiState.value.copy(
                    isLoading = true,
                    isScanning = false,
                    errorMessage = null
                )

            val result =
                repository.scanQrCode(qrCode)

            result.onSuccess { data ->

                _uiState.value =
                    _uiState.value.copy(
                        isLoading = false,
                        scanResult = data
                    )
            }

            result.onFailure { error ->

                _uiState.value =
                    _uiState.value.copy(
                        isLoading = false,
                        isScanning = false, // Stop scanning while dialog is shown
                        errorMessage =
                            error.message
                    )
            }
        }
    }

    fun resetScanner() {

        isProcessing = false

        _uiState.value =
            _uiState.value.copy(
                isScanning = true,
                scanResult = null,
                errorMessage = null
            )
    }
}
