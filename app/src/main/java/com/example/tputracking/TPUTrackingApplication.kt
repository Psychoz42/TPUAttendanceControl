package com.example.tputracking

import android.app.Application
import com.example.tputracking.di.AppContainer
import com.example.tputracking.di.AppDataContainer

class TPUTrackingApplication : Application() {

    /**
     * AppContainer instance used by the rest of classes to obtain dependencies
     */
    lateinit var container: AppContainer

    override fun onCreate() {
        super.onCreate()
        container = AppDataContainer(this)
    }
}
