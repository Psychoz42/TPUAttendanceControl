package com.example.tputracking.data.repository

import androidx.paging.PagingSource
import com.example.tputracking.data.local.dao.AttendanceDao
import com.example.tputracking.data.local.entity.AttendanceEntity
import com.example.tputracking.data.remote.api.AttendanceApiService
import com.example.tputracking.domain.model.ScanErrorReason
import com.example.tputracking.domain.model.ScanException
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

/** Разбор ответов /attendance/scan по контракту бэкенда (MOBILE_PLAN.md, раздел 2). */
class AttendanceRepositoryImplTest {

    private lateinit var server: MockWebServer
    private lateinit var dao: FakeAttendanceDao
    private lateinit var repository: AttendanceRepositoryImpl

    private val validCode = "cUItAEaYzhUjwCpm"

    @Before
    fun setUp() {
        server = MockWebServer()
        server.start()
        val api = Retrofit.Builder()
            .baseUrl(server.url("/"))
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(AttendanceApiService::class.java)
        dao = FakeAttendanceDao()
        repository = AttendanceRepositoryImpl(api, dao, currentDate = { "06.11.2026" })
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun enqueueJson(body: String, code: Int = 200) {
        server.enqueue(
            MockResponse()
                .setResponseCode(code)
                .setHeader("Content-Type", "application/json")
                .setBody(body)
        )
    }

    private fun failureOf(result: Result<*>): ScanException =
        result.exceptionOrNull() as ScanException

    @Test
    fun success_savesLessonWithDateAndSendsOnlyQrCode() = runTest {
        enqueueJson("""{"success": true, "subject": "Математика", "time": "09:05"}""")

        val result = repository.scanQrCode(validCode)

        assertTrue(result.isSuccess)
        assertEquals("Математика", result.getOrThrow().subject)
        assertEquals(1, dao.items.size)
        assertEquals("06.11.2026, 09:05", dao.items[0].dateTime)

        val request = server.takeRequest()
        assertEquals("/api/v1/attendance/scan", request.path)
        assertEquals("""{"qrCode":"$validCode"}""", request.body.readUtf8())
    }

    @Test
    fun repeatedScan_doesNotDuplicateHistory() = runTest {
        enqueueJson("""{"success": true, "subject": "Математика", "time": "09:05"}""")
        enqueueJson("""{"success": true, "subject": "Математика", "time": "09:05"}""")

        repository.scanQrCode(validCode)
        val second = repository.scanQrCode(validCode)

        assertTrue(second.isSuccess)
        assertEquals(1, dao.items.size)
    }

    @Test
    fun invalidQrWithHttp200_isTypedErrorAndNotSaved() = runTest {
        enqueueJson("""{"success": false, "code": "INVALID_QR", "message": "QR code is invalid or expired"}""")

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals(ScanErrorReason.INVALID_QR, error.reason)
        assertTrue(error.message!!.startsWith("QR-код устарел"))
        assertTrue(dao.items.isEmpty())
    }

    @Test
    fun errorCodesWithHttp200_mapToRussianMessages() = runTest {
        val expected = mapOf(
            "UNKNOWN_USER" to "Ваш аккаунт не найден в системе. Обратитесь к преподавателю",
            "DB_ERROR" to "Не удалось отметить посещение. Попробуйте ещё раз",
            "INTERNAL_ERROR" to "Не удалось отметить посещение. Попробуйте ещё раз",
            "INVALID_DATA_FORMAT" to "Не удалось отметить посещение. Попробуйте ещё раз",
            "SOMETHING_NEW" to "Ошибка сервера. Попробуйте позже"
        )
        for ((code, message) in expected) {
            enqueueJson("""{"success": false, "code": "$code", "message": "english text"}""")
            val error = failureOf(repository.scanQrCode(validCode))
            assertEquals(code, message, error.message)
            assertEquals(code, ScanErrorReason.OTHER, error.reason)
        }
        assertTrue(dao.items.isEmpty())
    }

    @Test
    fun unauthorized401_isTypedError() = runTest {
        enqueueJson("""{"success": false, "code": "UNAUTHORIZED", "message": "invalid token"}""", code = 401)

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals(ScanErrorReason.UNAUTHORIZED, error.reason)
        assertEquals("Сессия истекла. Войдите снова", error.message)
    }

    @Test
    fun forbidden403_showsStudentsOnly() = runTest {
        enqueueJson("""{"success": false, "code": "FORBIDDEN", "message": "forbidden"}""", code = 403)

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals("Отмечаться могут только студенты", error.message)
    }

    @Test
    fun tooManyRequests429_usesRetryAfter() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(429)
                .setHeader("Retry-After", "17")
                .setBody("""{"success": false, "code": "TOO_MANY_REQUESTS", "message": "too many"}""")
        )

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals("Слишком много попыток. Подождите 17 с", error.message)
    }

    @Test
    fun htmlFromProxy502_doesNotCrash() = runTest {
        server.enqueue(MockResponse().setResponseCode(502).setBody("<html>Bad Gateway</html>"))

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals("Ошибка сервера. Попробуйте позже", error.message)
    }

    @Test
    fun htmlWithHttp200_doesNotCrash() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("<html>parking</html>"))

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals("Ошибка сервера. Попробуйте позже", error.message)
        assertTrue(dao.items.isEmpty())
    }

    @Test
    fun foreignQr_isRejectedLocallyWithoutRequest() = runTest {
        for (content in listOf("https://example.com/abc", "short", "cUItAEaYzhUjwCp!", "")) {
            val error = failureOf(repository.scanQrCode(content))
            assertEquals(content, ScanErrorReason.INVALID_FORMAT, error.reason)
        }
        assertEquals(0, server.requestCount)
    }

    @Test
    fun serverDown_showsNoConnection() = runTest {
        server.shutdown()

        val error = failureOf(repository.scanQrCode(validCode))

        assertEquals("Нет соединения с сервером", error.message)
    }

    private class FakeAttendanceDao : AttendanceDao {

        val items = mutableListOf<AttendanceEntity>()

        override fun getAttendancePaging(): PagingSource<Int, AttendanceEntity> =
            throw UnsupportedOperationException()

        override suspend fun insertAll(items: List<AttendanceEntity>) {
            this.items += items
        }

        override suspend fun exists(subject: String, dateTime: String): Boolean =
            items.any { it.subject == subject && it.dateTime == dateTime }

        override suspend fun clearAll() {
            items.clear()
        }
    }
}
