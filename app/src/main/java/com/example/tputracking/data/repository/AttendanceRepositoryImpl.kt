package com.example.tputracking.data.repository

import androidx.paging.Pager
import androidx.paging.PagingConfig
import androidx.paging.PagingData
import androidx.paging.map
import com.example.tputracking.data.local.dao.AttendanceDao
import com.example.tputracking.data.local.entity.AttendanceEntity
import com.example.tputracking.data.mapper.toDomain
import com.example.tputracking.data.remote.ApiErrorCode
import com.example.tputracking.data.remote.ApiErrorMessages
import com.example.tputracking.data.remote.api.AttendanceApiService
import com.example.tputracking.data.remote.dto.ScanQrRequest
import com.example.tputracking.data.remote.errorCode
import com.example.tputracking.data.remote.retryAfterSeconds
import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.domain.model.ScanErrorReason
import com.example.tputracking.domain.model.ScanException
import com.example.tputracking.domain.model.ScanResult
import com.example.tputracking.domain.repository.AttendanceRepository
import com.google.gson.JsonParseException
import com.google.gson.stream.MalformedJsonException
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import java.io.IOException
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class AttendanceRepositoryImpl(
    private val api: AttendanceApiService,
    private val dao: AttendanceDao,
    // Сервер присылает только время занятия (HH:MM), дату берём с устройства
    private val currentDate: () -> String = {
        SimpleDateFormat("dd.MM.yyyy", Locale("ru")).format(Date())
    }
) : AttendanceRepository {

    companion object {
        // Сервер выдаёт коды из 16 символов a-z, A-Z, 0-9; в QR только сам код
        private val QR_CODE_FORMAT = Regex("^[A-Za-z0-9]{16}$")
    }

    override suspend fun scanQrCode(
        qrCode: String
    ): Result<ScanResult> {

        if (!QR_CODE_FORMAT.matches(qrCode)) {
            return Result.failure(
                ScanException(
                    ScanErrorReason.INVALID_FORMAT,
                    "Неверный формат QR. Пожалуйста, попытайтесь снова"
                )
            )
        }

        return try {

            val response =
                api.scanQrCode(
                    ScanQrRequest(qrCode = qrCode)
                )

            val body = if (response.isSuccessful) response.body() else null
            val subject = body?.subject
            val time = body?.time

            // Сервер отдаёт большинство ошибок с HTTP 200, поэтому успех определяем по success
            if (body?.success == true && subject != null && time != null) {

                val dateTime = "${currentDate()}, $time"

                // Повторный скан того же занятия сервер принимает как успех; вторую запись не создаём
                if (!dao.exists(subject, dateTime)) {
                    dao.insertAll(
                        listOf(
                            AttendanceEntity(
                                id = System.currentTimeMillis(),
                                subject = subject,
                                teacher = "",
                                dateTime = dateTime,
                                visited = true
                            )
                        )
                    )
                }

                Result.success(
                    ScanResult(
                        success = true,
                        subject = subject,
                        time = time
                    )
                )

            } else {

                val code = if (response.isSuccessful) body?.code else response.errorCode()
                Result.failure(scanException(code, response.retryAfterSeconds()))
            }

        } catch (e: MalformedJsonException) {

            // Ответ не JSON (например, HTML от прокси)
            Result.failure(scanException(null, 0))

        } catch (e: JsonParseException) {

            Result.failure(scanException(null, 0))

        } catch (e: IOException) {

            // Не показываем пользователю технические детали исключения
            Result.failure(ScanException(ScanErrorReason.OTHER, ApiErrorMessages.NO_CONNECTION))

        } catch (e: Exception) {

            Result.failure(
                ScanException(ScanErrorReason.OTHER, "Не удалось отметить посещение")
            )
        }
    }

    private fun scanException(code: String?, retryAfterSeconds: Int): ScanException {
        val reason = when (code) {
            ApiErrorCode.INVALID_QR -> ScanErrorReason.INVALID_QR
            ApiErrorCode.UNAUTHORIZED -> ScanErrorReason.UNAUTHORIZED
            else -> ScanErrorReason.OTHER
        }
        return ScanException(reason, ApiErrorMessages.scan(code, retryAfterSeconds))
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
