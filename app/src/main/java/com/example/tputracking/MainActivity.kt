package com.example.tputracking

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.compose.rememberNavController
import com.example.tputracking.navigation.AppNavGraph
import com.example.tputracking.presentation.AppViewModelProvider
import com.example.tputracking.presentation.splash.SplashViewModel
import com.example.tputracking.ui.theme.AttendanceTheme

class MainActivity : ComponentActivity() {

    override fun onCreate(
        savedInstanceState: Bundle?
    ) {
        installSplashScreen()
        super.onCreate(savedInstanceState)

        setContent {

            AttendanceTheme {

                val navController =
                    rememberNavController()

                AppNavGraph(
                    navController = navController,

                    splashViewModel = viewModel(factory = AppViewModelProvider.Factory)
                )
            }
        }
    }
}