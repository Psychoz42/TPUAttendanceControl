package com.example.tputracking.navigation

import androidx.compose.animation.AnimatedContentTransitionScope
import androidx.compose.animation.core.tween
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.navigation
import com.example.tputracking.navigation.auth.authNavGraph
import com.example.tputracking.navigation.main.mainNavGraph
import com.example.tputracking.presentation.splash.SplashScreen
import com.example.tputracking.presentation.splash.SplashViewModel

@Composable
fun AppNavGraph(
    navController: NavHostController,
    splashViewModel: SplashViewModel
) {

    val splashState by
    splashViewModel
        .uiState
        .collectAsState()

    LaunchedEffect(
        splashState.isLoading
    ) {

        if (!splashState.isLoading) {

            if (splashState.isAuthorized) {

                navController.navigate(
                    Graph.MAIN
                ) {

                    popUpTo(
                        Routes.Splash.route
                    ) {
                        inclusive = true
                    }
                }

            } else {

                navController.navigate(
                    Graph.AUTH
                ) {

                    popUpTo(
                        Routes.Splash.route
                    ) {
                        inclusive = true
                    }
                }
            }
        }
    }

    NavHost(
        navController = navController,

        route = Graph.ROOT,

        startDestination =
            Routes.Splash.route
    ) {

        composable(
            route = Routes.Splash.route,

            enterTransition = {

                slideIntoContainer(
                    AnimatedContentTransitionScope
                        .SlideDirection.Left,

                    animationSpec =
                        tween(300)
                )
            },

            exitTransition = {

                slideOutOfContainer(
                    AnimatedContentTransitionScope
                        .SlideDirection.Left,

                    animationSpec =
                        tween(300)
                )
            }
        ) {

            SplashScreen()
        }

        navigation(
            route = Graph.AUTH,
            startDestination =
                Routes.Login.route
        ) {

            authNavGraph(navController)
        }

        navigation(
            route = Graph.MAIN,
            startDestination =
                Routes.Home.route
        ) {

            mainNavGraph(navController)
        }
    }
}