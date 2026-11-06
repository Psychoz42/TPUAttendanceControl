package com.example.tputracking.presentation.history

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ArrowBack
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.paging.compose.LazyPagingItems
import androidx.paging.compose.collectAsLazyPagingItems
import com.example.tputracking.presentation.history.components.HistoryLessonCard
import com.example.tputracking.presentation.history.model.AttendanceHistoryUi
import com.example.tputracking.ui.theme.GreenPrimary

@Composable
fun HistoryScreen(
    viewModel: HistoryViewModel,
    onBackClick: () -> Unit = {},
) {
    val attendanceItems = viewModel.attendancePaging.collectAsLazyPagingItems()

    HistoryContent(
        attendanceItems = attendanceItems,
        onBackClick = onBackClick
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun HistoryContent(
    attendanceItems: LazyPagingItems<com.example.tputracking.domain.model.Attendance>,
    onBackClick: () -> Unit = {},
) {
    Scaffold(
        containerColor = MaterialTheme.colorScheme.background,

        topBar = {

            CenterAlignedTopAppBar(

                title = {

                    Text(
                        text = "Список посещенных занятий",

                        color = GreenPrimary,

                        style =
                            MaterialTheme.typography.titleMedium,

                        fontWeight = FontWeight.Medium
                    )
                },

                navigationIcon = {

                    IconButton(
                        onClick = onBackClick
                    ) {

                        Icon(
                            imageVector =
                                Icons.Outlined.ArrowBack,

                            contentDescription = null,

                            tint = GreenPrimary
                        )
                    }
                },

                colors =
                    TopAppBarDefaults.centerAlignedTopAppBarColors(
                        containerColor = MaterialTheme.colorScheme.background
                    )
            )
        }
    ) { padding ->

        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background),

            contentPadding = PaddingValues(
                start = 18.dp,
                end = 18.dp,
                top = padding.calculateTopPadding() + 14.dp,
                bottom = 24.dp
            ),

            verticalArrangement =
                Arrangement.spacedBy(16.dp)
        ) {

            items(count = attendanceItems.itemCount) { index ->
                val attendance = attendanceItems[index]
                attendance?.let {
                    HistoryLessonCard(
                        lesson = AttendanceHistoryUi(
                            id = it.id,
                            title = it.subject,
                            dateTime = it.dateTime,
                            visited = it.visited
                        )
                    )
                }
            }
        }
    }
}
