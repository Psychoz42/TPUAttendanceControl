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

    val sessionActive by
    splashViewModel
        .sessionActive
        .collectAsState()

    LaunchedEffect(
        splashState.isLoading
    ) {

        // Эффект перезапускается при пересоздании Activity (поворот экрана),
        // поэтому начальную навигацию выполняем только с экрана запуска
        if (!splashState.isLoading &&
            navController.currentDestination?.route == Routes.Splash.route
        ) {

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

    // Токен сброшен (401, выход): открываем вход и очищаем стек
    LaunchedEffect(
        sessionActive
    ) {

        val route = navController.currentDestination?.route

        if (sessionActive == false &&
            route != null &&
            route != Routes.Splash.route &&
            route != Routes.Login.route
        ) {

            navController.navigate(
                Graph.AUTH
            ) {

                popUpTo(
                    navController.graph.id
                ) {
                    inclusive = true
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