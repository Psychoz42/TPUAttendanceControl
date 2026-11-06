package com.example.tputracking.navigation.auth

import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavHostController
import androidx.navigation.compose.composable
import com.example.tputracking.navigation.Routes
import com.example.tputracking.presentation.AppViewModelProvider
import com.example.tputracking.presentation.login.LoginScreen
import com.example.tputracking.presentation.login.LoginViewModel

fun NavGraphBuilder.authNavGraph(
    navController: NavHostController
) {

    composable(
        route = Routes.Login.route
    ) {

        LoginScreen(

            viewModel = viewModel<LoginViewModel>(factory = AppViewModelProvider.Factory),

            onLoginSuccess = {

                navController.navigate(
                    Routes.Home.route
                ) {

                    popUpTo(
                        Routes.Login.route
                    ) {
                        inclusive = true
                    }
                }
            }
        )
    }
}