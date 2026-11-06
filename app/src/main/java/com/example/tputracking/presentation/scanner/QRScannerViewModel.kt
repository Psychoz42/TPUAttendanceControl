package com.example.tputracking.presentation.scanner

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.tputracking.domain.model.ScanErrorReason
import com.example.tputracking.domain.model.ScanException
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

    // Анализатор присылает код почти на каждом кадре. Отклонённый код повторно не отправляем,
    // иначе один устаревший QR за секунды исчерпает лимит сервера (30 сканов в минуту).
    private var lastRejectedCode: String? = null

    private var hintCounter = 0L

    fun onQrCodeScanned(
        qrCode: String
    ) {

        if (isProcessing || qrCode == lastRejectedCode) return

        isProcessing = true

        viewModelScope.launch {

            _uiState.value =
                _uiState.value.copy(
                    isLoading = true,
                    errorMessage = null
                )

            val result =
                repository.scanQrCode(qrCode)

            result.onSuccess { data ->

                _uiState.value =
                    _uiState.value.copy(
                        isLoading = false,
                        isScanning = false,
                        scanResult = data,
                        hint = null
                    )
            }

            result.onFailure { error ->

                val message = error.message ?: "Произошла неизвестная ошибка"

                when ((error as? ScanException)?.reason) {

                    // QR на экране меняется примерно каждые 10 секунд: код мог смениться
                    // за время запроса. Не блокируем экран, ждём следующий код.
                    ScanErrorReason.INVALID_QR,
                    ScanErrorReason.INVALID_FORMAT -> {

                        lastRejectedCode = qrCode
                        isProcessing = false

                        _uiState.value =
                            _uiState.value.copy(
                                isLoading = false,
                                hint = ScanHint(++hintCounter, message)
                            )
                    }

                    // Токен сброшен, AppNavGraph сам откроет экран входа
                    ScanErrorReason.UNAUTHORIZED -> {

                        _uiState.value =
                            _uiState.value.copy(
                                isLoading = false,
                                isScanning = false
                            )
                    }

                    else -> {

                        _uiState.value =
                            _uiState.value.copy(
                                isLoading = false,
                                isScanning = false, // Stop scanning while dialog is shown
                                errorMessage = message
                            )
                    }
                }
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
