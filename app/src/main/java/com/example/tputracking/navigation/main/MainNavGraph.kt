package com.example.tputracking.navigation.main

import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavHostController
import androidx.navigation.compose.composable
import com.example.tputracking.navigation.Routes
import com.example.tputracking.presentation.AppViewModelProvider
import com.example.tputracking.presentation.history.HistoryScreen
import com.example.tputracking.presentation.history.HistoryViewModel
import com.example.tputracking.presentation.home.HomeScreen
import com.example.tputracking.presentation.home.HomeViewModel
import com.example.tputracking.presentation.scanner.QRScannerScreen
import com.example.tputracking.presentation.scanner.QRScannerViewModel

fun NavGraphBuilder.mainNavGraph(
    navController: NavHostController
) {

    composable(
        route = Routes.Home.route
    ) {

        HomeScreen(
            viewModel = viewModel<HomeViewModel>(factory = AppViewModelProvider.Factory),

            onScanClick = {

                navController.navigate(
                    Routes.Scanner.route
                )
            },

            onSeeAllClick = {

                navController.navigate(
                    Routes.History.route
                )
            }
        )
    }

    composable(
        route = Routes.Scanner.route
    ) {

        QRScannerScreen(

            viewModel = viewModel<QRScannerViewModel>(factory = AppViewModelProvider.Factory),

            onClose = {

                navController.popBackStack()
            }
        )
    }

    composable(
        route = Routes.History.route
    ) {

        HistoryScreen(
            viewModel = viewModel<HistoryViewModel>(factory = AppViewModelProvider.Factory),

            onBackClick = {

                navController.popBackStack()
            }
        )
    }
}