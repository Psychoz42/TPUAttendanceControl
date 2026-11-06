package com.example.tputracking.di

import android.content.Context
import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.local.database.AppDatabase
import com.example.tputracking.data.repository.AttendanceRepositoryImpl
import com.example.tputracking.data.repository.AuthRepositoryImpl
import com.example.tputracking.domain.repository.AttendanceRepository
import com.example.tputracking.domain.repository.AuthRepository

interface AppContainer {
    val authRepository: AuthRepository
    val attendanceRepository: AttendanceRepository
}

class AppDataContainer(private val context: Context) : AppContainer {

    private val userPreferences: UserPreferences by lazy {
        NetworkModule.provideUserPreferences(context)
    }

    private val okHttpClient by lazy {
        NetworkModule.provideOkHttp(userPreferences)
    }

    private val retrofit by lazy {
        NetworkModule.provideRetrofit(okHttpClient)
    }

    private val authApiService by lazy {
        NetworkModule.provideAuthApi(retrofit)
    }

    private val attendanceApiService by lazy {
        NetworkModule.provideAttendanceApi(retrofit)
    }

    private val database by lazy {
        AppDatabase.getDatabase(context)
    }

    override val authRepository: AuthRepository by lazy {
        AuthRepositoryImpl(authApiService, userPreferences, database.attendanceDao())
    }

    override val attendanceRepository: AttendanceRepository by lazy {
        AttendanceRepositoryImpl(attendanceApiService, database.attendanceDao())
    }
}
