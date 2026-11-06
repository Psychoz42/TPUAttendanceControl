package com.example.tputracking.presentation.scanner

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.FlashlightOff
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import com.example.tputracking.presentation.scanner.components.CameraPreview
import com.example.tputracking.presentation.scanner.components.ScanResultDialog
import com.example.tputracking.presentation.scanner.components.ScannerOverlay
import kotlinx.coroutines.delay

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun QRScannerScreen(
    viewModel: QRScannerViewModel,
    onClose: () -> Unit = {}
) {

    val context = LocalContext.current

    val uiState by viewModel.uiState.collectAsState()

    var showResultDialog by remember { mutableStateOf(false) }

    val snackbarHostState =
        remember {
            SnackbarHostState()
        }

    var hasPermission by remember {

        mutableStateOf(
            ContextCompat.checkSelfPermission(
                context,
                Manifest.permission.CAMERA
            ) == PackageManager.PERMISSION_GRANTED
        )
    }

    val permissionLauncher =
        rememberLauncherForActivityResult(
            contract =
                ActivityResultContracts.RequestPermission()
        ) { granted ->

            hasPermission = granted
        }

    LaunchedEffect(Unit) {

        if (!hasPermission) {

            permissionLauncher.launch(
                Manifest.permission.CAMERA
            )
        }
    }

    LaunchedEffect(uiState.errorMessage) {
        if (uiState.errorMessage != null) {
            showResultDialog = true
        }
    }

    LaunchedEffect(uiState.scanResult) {
        if (uiState.scanResult != null) {
            showResultDialog = true
        }
    }

    BackHandler {
        onClose()
    }

    if (showResultDialog) {
        val isSuccess = uiState.scanResult != null
        ScanResultDialog(
            isSuccess = isSuccess,
            title = if (isSuccess) "Успешно!" else "Ошибка",
            message = if (isSuccess) 
                "Вы успешно отметились на занятии: ${uiState.scanResult?.subject}" 
                else uiState.errorMessage ?: "Произошла неизвестная ошибка",
            onDismiss = {
                showResultDialog = false
                if (isSuccess) {
                    onClose()
                } else {
                    viewModel.resetScanner()
                }
            }
        )
    }

    Box(
        modifier = Modifier.fillMaxSize()
    ) {

        if (hasPermission) {

            CameraPreview(
                modifier = Modifier.fillMaxSize(),

                onQrDetected = { qrCode ->

                    viewModel.onQrCodeScanned(
                        qrCode
                    )
                }
            )
        }

        ScannerOverlay()

        Column(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding()
                .navigationBarsPadding()
                .padding(
                    horizontal = 18.dp,
                    vertical = 12.dp
                ),

            verticalArrangement =
                Arrangement.SpaceBetween
        ) {

            CenterAlignedTopAppBar(

                title = {

                    Text(
                        text = "Контроль посещаемости",
                        color = Color.White
                    )
                },

                navigationIcon = {

                    Surface(
                        shape = RoundedCornerShape(14.dp),
                        color =
                            Color.Black.copy(
                                alpha = 0.3f
                            )
                    ) {

                        IconButton(
                            onClick = onClose
                        ) {

                            Icon(
                                imageVector =
                                    Icons.Outlined.Close,

                                contentDescription = null,

                                tint = Color.White
                            )
                        }
                    }
                },

                actions = {

                    Surface(
                        shape = RoundedCornerShape(14.dp),
                        color =
                            Color.Black.copy(
                                alpha = 0.3f
                            )
                    ) {

                        IconButton(
                            onClick = {}
                        ) {

                            Icon(
                                imageVector =
                                    Icons.Outlined.FlashlightOff,

                                contentDescription = null,

                                tint = Color.White
                            )
                        }
                    }
                },

                colors =
                    TopAppBarDefaults
                        .centerAlignedTopAppBarColors(
                            containerColor =
                                Color.Transparent
                        )
            )

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 100.dp),

                horizontalAlignment =
                    Alignment.CenterHorizontally
            ) {
                /*
                Text(
                    text = "Сканировать QR - код",

                    style =
                        MaterialTheme.typography
                            .headlineSmall,

                    color = Color.White,

                    fontWeight =
                        FontWeight.SemiBold
                )

                Text(
                    text =
                        "Поместите QR - код в центр рамки\nдля его сканирования",

                    style =
                        MaterialTheme.typography
                            .bodyMedium,

                    color =
                        Color.White.copy(
                            alpha = 0.7f
                        )
                )
                
                 */
            }

            Button(
                onClick = onClose,

                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 30.dp),

                shape = RoundedCornerShape(18.dp),

                colors =
                    ButtonDefaults.buttonColors(
                        containerColor = Color.White,
                        contentColor = Color.Black
                    )
            ) {

                Text(
                    text = "Отменить сканирование"
                )
            }
        }

        if (uiState.isLoading) {

            CircularProgressIndicator(
                modifier = Modifier.align(
                    Alignment.Center
                ),
                color = Color.White
            )
        }

        SnackbarHost(
            hostState = snackbarHostState,
            modifier = Modifier.align(
                Alignment.BottomCenter
            )
        )
    }
}