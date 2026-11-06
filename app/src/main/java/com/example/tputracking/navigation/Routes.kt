package com.example.tputracking.navigation

sealed class Routes(
    val route: String
) {

    data object Splash :
        Routes("splash")

    data object Login :
        Routes("login")

    data object Home :
        Routes("home")

    data object Scanner :
        Routes("scanner")

    data object History :
        Routes("history")
}