package com.example.tputracking.presentation.scanner

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.provider.Settings
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
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
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
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
            context.hasCameraPermission()
        )
    }

    // Пользователь запретил камеру и попросил больше не спрашивать: остаются только настройки
    var permanentlyDenied by remember { mutableStateOf(false) }

    var cameraError by remember { mutableStateOf(false) }

    val permissionLauncher =
        rememberLauncherForActivityResult(
            contract =
                ActivityResultContracts.RequestPermission()
        ) { granted ->

            hasPermission = granted

            permanentlyDenied = !granted &&
                    context.findActivity()?.let {
                        !ActivityCompat.shouldShowRequestPermissionRationale(
                            it,
                            Manifest.permission.CAMERA
                        )
                    } == true
        }

    LaunchedEffect(Unit) {

        if (!hasPermission) {

            permissionLauncher.launch(
                Manifest.permission.CAMERA
            )
        }
    }

    // Разрешение могли выдать в настройках, пока приложение было свёрнуто
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        hasPermission = context.hasCameraPermission()
    }

    LaunchedEffect(uiState.hint) {
        uiState.hint?.let {
            snackbarHostState.showSnackbar(it.message)
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
                },

                onCameraError = {
                    cameraError = true
                }
            )
        }

        ScannerOverlay()

        when {
            !hasPermission -> CameraMessage(
                text = "Чтобы отметиться, приложению нужен доступ к камере",
                buttonText = if (permanentlyDenied) "Открыть настройки" else "Разрешить доступ",
                onButtonClick = {
                    if (permanentlyDenied) {
                        context.startActivity(
                            Intent(
                                Settings.ACTION_APPLICATION_DETAILS_SETTINGS,
                                Uri.fromParts("package", context.packageName, null)
                            )
                        )
                    } else {
                        permissionLauncher.launch(Manifest.permission.CAMERA)
                    }
                }
            )

            cameraError -> CameraMessage(
                text = "Не удалось запустить камеру. Закройте другие приложения, которые её используют, и попробуйте снова"
            )
        }

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

                // Кнопка фонарика скрыта до реализации

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
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .navigationBarsPadding()
                // над кнопкой «Отменить сканирование»
                .padding(bottom = 72.dp)
        )
    }
}

@Composable
private fun BoxScope.CameraMessage(
    text: String,
    buttonText: String? = null,
    onButtonClick: () -> Unit = {}
) {

    Column(
        modifier = Modifier
            .align(Alignment.Center)
            .padding(horizontal = 32.dp),

        horizontalAlignment =
            Alignment.CenterHorizontally
    ) {

        Text(
            text = text,
            color = Color.White,
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center
        )

        if (buttonText != null) {

            Spacer(modifier = Modifier.height(16.dp))

            Button(
                onClick = onButtonClick,
                shape = RoundedCornerShape(18.dp),
                colors =
                    ButtonDefaults.buttonColors(
                        containerColor = Color.White,
                        contentColor = Color.Black
                    )
            ) {

                Text(text = buttonText)
            }
        }
    }
}

private fun Context.hasCameraPermission(): Boolean =
    ContextCompat.checkSelfPermission(
        this,
        Manifest.permission.CAMERA
    ) == PackageManager.PERMISSION_GRANTED

private tailrec fun Context.findActivity(): Activity? = when (this) {
    is Activity -> this
    is ContextWrapper -> baseContext.findActivity()
    else -> null
}