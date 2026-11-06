package com.example.tputracking

import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import com.example.tputracking.data.local.UserPreferences
import com.example.tputracking.data.local.database.AppDatabase
import com.example.tputracking.domain.model.AuthResult
import com.example.tputracking.domain.model.ScanErrorReason
import com.example.tputracking.domain.model.ScanException
import com.example.tputracking.domain.model.User
import com.example.tputracking.domain.repository.AttendanceRepository
import com.example.tputracking.domain.repository.AuthRepository
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assume.assumeTrue
import org.junit.Before
import org.junit.FixMethodOrder
import org.junit.Test
import org.junit.runner.RunWith
import org.junit.runners.MethodSorters

/**
 * Сквозная проверка приложения против настоящего бэкенда (ветка backend) с AUTH_REQUIRED=true.
 *
 * По умолчанию пропускается. Запуск: поднять бэкенд и тестовых пользователей по разделу 3.1
 * MOBILE_PLAN.md, затем
 *   ./gradlew connectedDebugAndroidTest -Pandroid.testInstrumentationRunnerArguments.backendE2E=true
 * Адрес сервера берётся из BuildConfig.BASE_URL (по умолчанию http://10.0.2.2:8080).
 */
@RunWith(AndroidJUnit4::class)
@FixMethodOrder(MethodSorters.NAME_ASCENDING)
class BackendE2ETest {

    private lateinit var auth: AuthRepository
    private lateinit var attendance: AttendanceRepository
    private lateinit var app: TPUTrackingApplication

    private val http = OkHttpClient()

    @Before
    fun setUp() {
        assumeTrue(
            "Задайте backendE2E=true, чтобы запустить проверку против бэкенда",
            InstrumentationRegistry.getArguments().getString("backendE2E") == "true"
        )
        app = ApplicationProvider.getApplicationContext()
        auth = app.container.authRepository
        attendance = app.container.attendanceRepository
        runBlocking { auth.logout() }
    }

    @Test
    fun a_loginErrorsAndRoles() = runBlocking {
        assertEquals(
            AuthResult.Error("Неправильная почта или пароль"),
            auth.login("student1@tpu.ru", "wrong")
        )

        assertEquals(
            AuthResult.Error("Приложение предназначено для студентов. Преподавателям нужно использовать сайт."),
            auth.login("teacher1@tpu.ru", "teach1")
        )
        assertFalse("сессия преподавателя не сохраняется", auth.isLoggedIn())

        assertEquals(AuthResult.Success, auth.login("student1@tpu.ru", "pass1"))
        assertTrue(auth.isLoggedIn())
        assertTrue(auth.sessionActive().first())
        val user = auth.getUser().first()!!
        assertEquals("Иванов Иван Иванович", user.fullName)
        assertEquals("8К52Т", user.group)
        assertEquals(User.ROLE_STUDENT, user.role)
    }

    @Test
    fun b_scanLifecycle() = runBlocking {
        assertEquals(AuthResult.Success, auth.login("student1@tpu.ru", "pass1"))
        val teacher = teacherToken()
        val event = post("/attendance/createEvent", """{"event_name":"E2E урок","event_time":"9:05"}""", teacher)
        val code = event.getString("code")

        val first = attendance.scanQrCode(code)
        assertTrue("скан актуального кода: $first", first.isSuccess)
        assertEquals("E2E урок", first.getOrThrow().subject)
        assertEquals("09:05", first.getOrThrow().time)
        assertEquals(1, historyCount())

        assertTrue("повторный скан", attendance.scanQrCode(code).isSuccess)
        assertEquals("повторный скан не дублирует историю", 1, historyCount())

        val rotated = post("/attendance/updateCode", """{"code":"$code"}""", teacher)
        val newCode = rotated.getString("new_code")
        assertTrue(rotated.getJSONArray("noted_students").toString().contains("Иванов Иван Иванович"))

        // Сервер ещё 3 секунды принимает предыдущий код, ждём дольше
        Thread.sleep(4_000)
        assertEquals(ScanErrorReason.INVALID_QR, reasonOf(attendance.scanQrCode(code)))
        assertEquals(1, historyCount())

        assertTrue("скан нового кода", attendance.scanQrCode(newCode).isSuccess)
        assertEquals(1, historyCount())

        post("/attendance/disableEventUpdating", """{"event_id":${event.getInt("event_id")}}""", teacher)
        assertEquals(ScanErrorReason.INVALID_QR, reasonOf(attendance.scanQrCode(newCode)))

        assertEquals(
            "неизвестный код из 16 символов",
            ScanErrorReason.INVALID_QR,
            reasonOf(attendance.scanQrCode("AAAAAAAAAAAAAAAA"))
        )
        assertEquals(
            ScanErrorReason.INVALID_FORMAT,
            reasonOf(attendance.scanQrCode("https://example.com"))
        )
    }

    @Test
    fun c_expiredTokenKeepsHistory() = runBlocking {
        assertEquals(AuthResult.Success, auth.login("student1@tpu.ru", "pass1"))
        scanNewEvent("История до истечения")
        assertEquals(1, historyCount())

        // Подменяем токен на недействительный: так сервер отвечает после смены JWT_SECRET
        val prefs = UserPreferences(app)
        prefs.saveSession(prefs.user.first()!!, "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.invalid")

        val teacher = teacherToken()
        val code = post("/attendance/createEvent", """{"event_name":"После истечения","event_time":"10:00"}""", teacher)
            .getString("code")
        assertEquals(ScanErrorReason.UNAUTHORIZED, reasonOf(attendance.scanQrCode(code)))

        assertFalse("401 сбрасывает токен", auth.sessionActive().first())
        assertEquals("student1@tpu.ru", auth.getExpiredSessionEmail())
        assertEquals("история сохраняется при истечении токена", 1, historyCount())

        assertEquals(AuthResult.Success, auth.login("student1@tpu.ru", "pass1"))
        assertEquals("тот же пользователь видит свою историю", 1, historyCount())
        assertNull(auth.getExpiredSessionEmail())

        assertEquals(AuthResult.Success, auth.login("student2@tpu.ru", "pass2"))
        assertEquals("другой пользователь не видит чужую историю", 0, historyCount())
    }

    @Test
    fun d_logoutClearsEverything() = runBlocking {
        assertEquals(AuthResult.Success, auth.login("student1@tpu.ru", "pass1"))
        scanNewEvent("Перед выходом")

        auth.logout()

        assertFalse(auth.isLoggedIn())
        assertNull(auth.getUser().first())
        assertNull("после выхода email не подставляется", auth.getExpiredSessionEmail())
        assertEquals(0, historyCount())
    }

    // Последним: исчерпывает лимит student2 на минуту
    @Test
    fun z_rateLimit() = runBlocking {
        assertEquals(AuthResult.Success, auth.login("student2@tpu.ru", "pass2"))
        var last: Result<*>? = null
        repeat(31) { last = attendance.scanQrCode("BBBBBBBBBBBBBBBB") }
        val error = last!!.exceptionOrNull() as ScanException
        assertTrue(error.message!!, error.message!!.startsWith("Слишком много попыток. Подождите"))
    }

    private suspend fun scanNewEvent(name: String) {
        val code = post("/attendance/createEvent", """{"event_name":"$name","event_time":"11:30"}""", teacherToken())
            .getString("code")
        assertTrue(attendance.scanQrCode(code).isSuccess)
    }

    private fun reasonOf(result: Result<*>): ScanErrorReason? =
        (result.exceptionOrNull() as? ScanException)?.reason

    private fun historyCount(): Int =
        AppDatabase.getDatabase(app).openHelper.readableDatabase
            .query("SELECT COUNT(*) FROM attendance")
            .use { cursor ->
                cursor.moveToFirst()
                cursor.getInt(0)
            }

    private fun teacherToken(): String =
        post("/auth/login", """{"email":"teacher1@tpu.ru","password":"teach1"}""", null)
            .getString("token")

    private fun post(path: String, json: String, token: String?): JSONObject {
        val request = Request.Builder()
            .url(BuildConfig.BASE_URL + "/api/v1" + path)
            .post(json.toRequestBody("application/json".toMediaType()))
            .apply { if (token != null) header("Authorization", "Bearer $token") }
            .build()
        http.newCall(request).execute().use { response ->
            val body = JSONObject(response.body!!.string())
            assertFalse("$path: $body", body.optBoolean("success", true) == false)
            return body
        }
    }
}
