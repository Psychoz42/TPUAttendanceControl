package com.example.tputracking.presentation.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Logout
import androidx.compose.material.icons.outlined.School
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.paging.LoadState
import androidx.paging.compose.LazyPagingItems
import androidx.paging.compose.collectAsLazyPagingItems
import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.presentation.home.components.AttendanceCard
import com.example.tputracking.presentation.home.components.LessonCard
import com.example.tputracking.presentation.home.model.LessonUi
import com.example.tputracking.ui.theme.GreenPrimary

@Composable
fun HomeScreen(
    viewModel: HomeViewModel,
    onScanClick: () -> Unit = {},
    onSeeAllClick: () -> Unit = {},
) {
    val uiState by viewModel.uiState.collectAsState()
    val attendanceItems = viewModel.attendancePaging.collectAsLazyPagingItems()

    var showLogoutDialog by rememberSaveable { mutableStateOf(false) }

    if (showLogoutDialog) {
        AlertDialog(
            onDismissRequest = { showLogoutDialog = false },
            title = { Text("Выйти из аккаунта?") },
            text = { Text("История посещений на этом устройстве будет удалена.") },
            confirmButton = {
                TextButton(
                    onClick = {
                        showLogoutDialog = false
                        viewModel.logout()
                    }
                ) {
                    Text("Выйти", color = GreenPrimary)
                }
            },
            dismissButton = {
                TextButton(onClick = { showLogoutDialog = false }) {
                    Text("Отмена", color = Color.Gray)
                }
            }
        )
    }

    HomeContent(
        uiState = uiState,
        attendanceItems = attendanceItems,
        onScanClick = onScanClick,
        onSeeAllClick = onSeeAllClick,
        onLogoutClick = { showLogoutDialog = true },
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun HomeContent(
    uiState: HomeUiState,
    attendanceItems: LazyPagingItems<Attendance>,
    onScanClick: () -> Unit = {},
    onSeeAllClick: () -> Unit = {},
    onLogoutClick: () -> Unit = {},
) {
    Scaffold(
        containerColor = MaterialTheme.colorScheme.background,
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "Контроль посещаемости",
                            color = GreenPrimary,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold
                        )
                        uiState.user?.let {
                            Text(
                                text = "${it.fullName} • ${it.group}",
                                style = MaterialTheme.typography.labelSmall,
                                color = Color.Gray
                            )
                        }
                    }
                },
                navigationIcon = {
                    Icon(
                        imageVector = Icons.Outlined.School,
                        contentDescription = null,
                        tint = GreenPrimary,
                        modifier = Modifier.padding(start = 16.dp)
                    )
                },
                actions = {
                    IconButton(onClick = onLogoutClick) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Outlined.Logout,
                            contentDescription = "Выйти из аккаунта",
                            tint = GreenPrimary
                        )
                    }
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
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
                top = padding.calculateTopPadding() + 12.dp,
                bottom = 24.dp
            ),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {

            item {
                AttendanceCard(
                    onScanClick = onScanClick
                )
            }

            item {
                Spacer(modifier = Modifier.height(10.dp))
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "Последние занятия",
                        style = MaterialTheme.typography.titleLarge,
                        color = GreenPrimary,
                        fontWeight = FontWeight.Bold
                    )

                    Text(
                        text = "Архив",
                        color = Color.Gray,
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.clickable {
                            onSeeAllClick()
                        }
                    )
                }
            }

            // Show only first 3 items for Home screen
            val itemCount = minOf(attendanceItems.itemCount, 3)
            
            if (attendanceItems.loadState.refresh is LoadState.Loading) {
                item {
                    Box(modifier = Modifier.fillMaxWidth().padding(24.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = GreenPrimary)
                    }
                }
            } else if (itemCount == 0 && (attendanceItems.loadState.refresh is LoadState.NotLoading)) {
                item {
                    Text(
                        text = "История пуста",
                        modifier = Modifier.padding(vertical = 32.dp).fillMaxWidth(),
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color.Gray
                    )
                }
            } else {
                items(itemCount) { index ->
                    val attendance = attendanceItems[index]
                    attendance?.let {
                        LessonCard(
                            lesson = LessonUi(
                                id = it.id,
                                title = it.subject,
                                time = it.dateTime,
                                visited = it.visited
                            )
                        )
                    }
                }
            }
        }
    }
}
