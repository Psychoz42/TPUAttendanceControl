package com.example.tputracking.di

import android.content.Context
import com.example.tputracking.BuildConfig
import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.remote.AuthInterceptor
import com.example.tputracking.data.remote.api.AuthApiService
import com.example.tputracking.data.remote.api.AttendanceApiService
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

object NetworkModule {

    private const val BASE_URL = BuildConfig.BASE_URL

    fun provideUserPreferences(
        context: Context
    ): UserPreferences {
        return UserPreferences(context)
    }

    fun provideOkHttp(
        userPreferences: UserPreferences
    ): OkHttpClient {

        val builder = OkHttpClient.Builder()
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(20, TimeUnit.SECONDS)
            .writeTimeout(20, TimeUnit.SECONDS)
            .addInterceptor(AuthInterceptor(userPreferences))

        // Логируем только в debug-сборке и без тел запросов:
        // уровень BODY писал бы в logcat пароль и токен из ответа входа.
        if (BuildConfig.DEBUG) {
            val logging = HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BASIC
                redactHeader("Authorization")
                redactHeader("Cookie")
                redactHeader("Set-Cookie")
            }
            builder.addInterceptor(logging)
        }

        return builder.build()
    }

    fun provideRetrofit(
        okHttpClient: OkHttpClient
    ): Retrofit {

        return Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(
                GsonConverterFactory.create()
            )
            .build()
    }

    fun provideAuthApi(
        retrofit: Retrofit
    ): AuthApiService {

        return retrofit.create(
            AuthApiService::class.java
        )
    }

    fun provideAttendanceApi(
        retrofit: Retrofit
    ): AttendanceApiService {

        return retrofit.create(
            AttendanceApiService::class.java
        )
    }
}
