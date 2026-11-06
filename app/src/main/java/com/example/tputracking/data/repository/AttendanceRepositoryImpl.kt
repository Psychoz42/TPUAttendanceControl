package com.example.tputracking.data.repository

import androidx.paging.Pager
import androidx.paging.PagingConfig
import androidx.paging.PagingData
import androidx.paging.map
import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.local.dao.AttendanceDao
import com.example.tputracking.data.local.entity.AttendanceEntity
import com.example.tputracking.data.mapper.toDomain
import com.example.tputracking.data.remote.api.AttendanceApiService
import com.example.tputracking.data.remote.dto.ScanQrRequest
import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.domain.model.ScanResult
import com.example.tputracking.domain.repository.AttendanceRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.map
import java.io.IOException

class AttendanceRepositoryImpl(
    private val api: AttendanceApiService,
    private val dao: AttendanceDao,
    private val userPreferences: UserPreferences
) : AttendanceRepository {

    override suspend fun scanQrCode(
        qrCode: String
    ): Result<ScanResult> {

        // Проверка длины QR-кода (ровно 16 символов)
        if (qrCode.length != 16) {
            return Result.failure(Exception("Неверный формат QR. Пожалуйста, попытайтесь снова"))
        }

        return try {
            // Получаем имя пользователя для запроса
            val user = userPreferences.user.firstOrNull()
            val username = user?.username ?: ""

            val response =
                api.scanQrCode(
                    ScanQrRequest(
                        qrCode = qrCode,
                        username = username
                    )
                )

            if (response.isSuccessful) {

                val body = response.body()

                if (body != null) {

                    // Сохраняем данные, полученные ОТ СЕРВЕРА (body.subject, body.time)
                    dao.insertAll(
                        listOf(
                            AttendanceEntity(
                                id = System.currentTimeMillis(),
                                subject = body.subject,
                                teacher = "",
                                dateTime = body.time,
                                visited = body.success
                            )
                        )
                    )

                    Result.success(
                        ScanResult(
                            success = body.success,
                            subject = body.subject,
                            time = body.time
                        )
                    )

                } else {

                    Result.failure(
                        Exception("Пустой ответ от сервера")
                    )
                }

            } else {

                Result.failure(
                    Exception("Ошибка сканирования: ${response.code()}")
                )
            }

        } catch (e: IOException) {

            // Не показываем пользователю технические детали исключения
            Result.failure(Exception("Нет соединения с сервером"))

        } catch (e: Exception) {

            Result.failure(Exception("Не удалось отметить посещение"))
        }
    }

    override fun getAttendancePaging():
            Flow<PagingData<Attendance>> {

        return Pager(

            config = PagingConfig(
                pageSize = 20,
                enablePlaceholders = false
            ),

            pagingSourceFactory = {
                dao.getAttendancePaging()
            }

        ).flow.map { pagingData ->

            pagingData.map { entity ->

                entity.toDomain()
            }
        }
    }
}
