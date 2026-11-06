package com.example.tputracking.presentation.scanner.components

import android.annotation.SuppressLint
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview as CameraPreview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.example.tputracking.presentation.scanner.analyzer.QrCodeAnalyzer
import com.google.mlkit.vision.barcode.BarcodeScanning

@SuppressLint("UnsafeOptInUsageError")
@Composable
fun CameraPreview(
    modifier: Modifier = Modifier,
    onQrDetected: (String) -> Unit
) {

    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current

    AndroidView(
        modifier = modifier,

        factory = {

            val previewView = PreviewView(context)

            val cameraProviderFuture =
                ProcessCameraProvider.getInstance(context)

            cameraProviderFuture.addListener({

                val cameraProvider =
                    cameraProviderFuture.get()

                val preview =
                    CameraPreview.Builder()
                        .build()
                        .also {
                            it.setSurfaceProvider(
                                previewView.surfaceProvider
                            )
                        }

                val barcodeScanner =
                    BarcodeScanning.getClient()

                val imageAnalysis =
                    ImageAnalysis.Builder()
                        .build()
                        .also {

                            it.setAnalyzer(
                                ContextCompat.getMainExecutor(
                                    context
                                ),

                                QrCodeAnalyzer(
                                    scanner = barcodeScanner,
                                    onQrDetected = onQrDetected
                                )
                            )
                        }

                val cameraSelector =
                    CameraSelector.DEFAULT_BACK_CAMERA

                try {

                    cameraProvider.unbindAll()

                    cameraProvider.bindToLifecycle(
                        lifecycleOwner,
                        cameraSelector,
                        preview,
                        imageAnalysis
                    )

                } catch (_: Exception) {
                }

            }, ContextCompat.getMainExecutor(context))

            previewView
        }
    )
}