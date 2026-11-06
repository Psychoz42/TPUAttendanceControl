package com.example.tputracking.presentation.scanner.analyzer

import androidx.camera.core.ImageAnalysis
import androidx.camera.core.ImageProxy
import com.google.mlkit.vision.barcode.BarcodeScanner
import com.google.mlkit.vision.common.InputImage

class QrCodeAnalyzer(
    private val scanner: BarcodeScanner,
    private val onQrDetected: (String) -> Unit
) : ImageAnalysis.Analyzer {

    override fun analyze(
        imageProxy: ImageProxy
    ) {

        val mediaImage =
            imageProxy.image

        if (mediaImage != null) {

            val image =
                InputImage.fromMediaImage(
                    mediaImage,
                    imageProxy.imageInfo.rotationDegrees
                )

            scanner.process(image)
                .addOnSuccessListener { barcodes ->

                    barcodes.firstOrNull()
                        ?.rawValue
                        ?.let(onQrDetected)
                }

                .addOnCompleteListener {
                    imageProxy.close()
                }

        } else {

            imageProxy.close()
        }
    }
}