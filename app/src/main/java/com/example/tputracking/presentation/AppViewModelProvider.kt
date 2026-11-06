package com.example.tputracking.presentation

import androidx.lifecycle.ViewModelProvider.AndroidViewModelFactory
import androidx.lifecycle.viewmodel.CreationExtras
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.example.tputracking.TPUTrackingApplication
import com.example.tputracking.presentation.history.HistoryViewModel
import com.example.tputracking.presentation.home.HomeViewModel
import com.example.tputracking.presentation.login.LoginViewModel
import com.example.tputracking.presentation.scanner.QRScannerViewModel
import com.example.tputracking.presentation.splash.SplashViewModel

/**
 * Provides Factory to create instance of ViewModel for the entire TPUTracking app
 */
object AppViewModelProvider {
    val Factory = viewModelFactory {
        // Initializer for SplashViewModel
        initializer {
            SplashViewModel(
                tpuTrackingApplication().container.authRepository
            )
        }

        // Initializer for LoginViewModel
        initializer {
            LoginViewModel(
                tpuTrackingApplication().container.authRepository
            )
        }

        // Initializer for HomeViewModel
        initializer {
            HomeViewModel(
                tpuTrackingApplication().container.authRepository,
                tpuTrackingApplication().container.attendanceRepository
            )
        }

        // Initializer for HistoryViewModel
        initializer {
            HistoryViewModel(
                tpuTrackingApplication().container.attendanceRepository
            )
        }

        // Initializer for QRScannerViewModel
        initializer {
            QRScannerViewModel(
                tpuTrackingApplication(),
                tpuTrackingApplication().container.attendanceRepository
            )
        }
    }
}

/**
 * Extension function to queries for [Application] object and returns an instance of
 * [TPUTrackingApplication].
 */
fun CreationExtras.tpuTrackingApplication(): TPUTrackingApplication =
    (this[AndroidViewModelFactory.APPLICATION_KEY] as TPUTrackingApplication)
