package com.example.tputracking.presentation.scanner.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.unit.dp

@Composable
fun ScannerOverlay() {

    Canvas(
        modifier = Modifier.fillMaxSize()
    ) {

        drawRect(
            color = Color.Black.copy(alpha = 0.55f)
        )

        val frameWidth = 700f
        val frameHeight = 700f

        val left =
            (size.width - frameWidth) / 2

        val top =
            (size.height - frameHeight) / 2

        drawRoundRect(
            color = Color.Transparent,
            topLeft = Offset(left, top),
            size = Size(frameWidth, frameHeight),
            cornerRadius = CornerRadius(50f, 50f),
            blendMode = BlendMode.Clear
        )

        drawRoundRect(
            color = Color.White,
            topLeft = Offset(left, top),
            size = Size(frameWidth, frameHeight),
            cornerRadius = CornerRadius(50f, 50f),
            style = Stroke(
                width = 6f
            )
        )

        drawLine(
            color = Color.White.copy(alpha = 0.9f),
            start = Offset(
                left + 40f,
                top + frameHeight / 2
            ),
            end = Offset(
                left + frameWidth - 40f,
                top + frameHeight / 2
            ),
            strokeWidth = 3f
        )
    }
}