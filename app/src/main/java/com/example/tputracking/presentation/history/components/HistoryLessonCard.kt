package com.example.tputracking.presentation.history.components

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.example.tputracking.presentation.history.model.AttendanceHistoryUi
import com.example.tputracking.ui.theme.GreenPrimary

@Composable
fun HistoryLessonCard(
    lesson: AttendanceHistoryUi
) {

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .animateContentSize(),

        shape = RoundedCornerShape(18.dp),

        elevation = CardDefaults.cardElevation(
            defaultElevation = 3.dp
        )
    ) {

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(
                    horizontal = 18.dp,
                    vertical = 20.dp
                ),

            verticalAlignment = Alignment.CenterVertically,

            horizontalArrangement =
                Arrangement.SpaceBetween
        ) {

            Column(
                modifier = Modifier.weight(1f)
            ) {

                Text(
                    text = lesson.title,

                    style =
                        MaterialTheme.typography.titleMedium,

                    fontWeight = FontWeight.SemiBold,

                    color =
                        MaterialTheme.colorScheme.onSurface
                )

                Text(
                    text = lesson.dateTime,

                    style =
                        MaterialTheme.typography.bodySmall
                )
            }

            Spacer(
                modifier = Modifier.width(12.dp)
            )

            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(50))
                    .background(
                        GreenPrimary.copy(alpha = 0.15f)
                    )
                    .padding(
                        horizontal = 14.dp,
                        vertical = 7.dp
                    )
            ) {

                Text(
                    text = if (lesson.visited) {
                        "ПОСЕЩЕНО"
                    } else {
                        "ПРОПУЩЕНО"
                    },

                    color = GreenPrimary,

                    style =
                        MaterialTheme.typography.labelSmall,

                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}