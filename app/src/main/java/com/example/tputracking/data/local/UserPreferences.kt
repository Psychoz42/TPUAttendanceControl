package com.example.tputracking.data.local

import android.content.Context
import androidx.datastore.core.DataMigration
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.example.tputracking.data.local.security.CryptoManager
import com.example.tputracking.domain.model.User
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import org.json.JSONObject

// Старые ключи, в которых профиль хранился открытым текстом (до шифрования)
private val LEGACY_KEY_USERNAME = stringPreferencesKey("username")
private val LEGACY_KEY_FULL_NAME = stringPreferencesKey("full_name")
private val LEGACY_KEY_EMAIL = stringPreferencesKey("email")
private val LEGACY_KEY_GROUP = stringPreferencesKey("group")

private val KEY_USER_ENCRYPTED = stringPreferencesKey("user_encrypted")
private val KEY_TOKEN_ENCRYPTED = stringPreferencesKey("token_encrypted")

private val cryptoManager by lazy { CryptoManager() }

private val Context.dataStore by preferencesDataStore(
    name = "user_prefs",
    produceMigrations = { listOf(EncryptLegacyUserMigration) }
)

/** Переносит профиль из открытых ключей в зашифрованный и удаляет открытые. */
private object EncryptLegacyUserMigration : DataMigration<Preferences> {

    override suspend fun shouldMigrate(currentData: Preferences): Boolean =
        currentData.contains(LEGACY_KEY_USERNAME)

    override suspend fun migrate(currentData: Preferences): Preferences {
        val prefs = currentData.toMutablePreferences()
        val legacyUser = User(
            username = prefs[LEGACY_KEY_USERNAME] ?: "",
            fullName = prefs[LEGACY_KEY_FULL_NAME] ?: "",
            email = prefs[LEGACY_KEY_EMAIL] ?: "",
            group = prefs[LEGACY_KEY_GROUP] ?: "",
            role = ""
        )
        prefs[KEY_USER_ENCRYPTED] = cryptoManager.encrypt(legacyUser.toJson())
        prefs.remove(LEGACY_KEY_USERNAME)
        prefs.remove(LEGACY_KEY_FULL_NAME)
        prefs.remove(LEGACY_KEY_EMAIL)
        prefs.remove(LEGACY_KEY_GROUP)
        return prefs.toPreferences()
    }

    override suspend fun cleanUp() = Unit
}

private fun User.toJson(): String =
    JSONObject()
        .put("username", username)
        .put("fullName", fullName)
        .put("email", email)
        .put("group", group)
        .put("role", role)
        .toString()

private fun userFromJson(json: String): User? =
    try {
        val obj = JSONObject(json)
        User(
            username = obj.getString("username"),
            fullName = obj.optString("fullName"),
            email = obj.optString("email"),
            group = obj.optString("group"),
            role = obj.optString("role")
        )
    } catch (e: Exception) {
        null
    }

/**
 * Профиль пользователя и токен хранятся в DataStore в зашифрованном виде
 * (AES-GCM, ключ в Android Keystore).
 *
 * Токен хранится отдельно от профиля: [User] попадает в состояние UI, а токену там не место.
 * Профиль без токена означает истёкшую сессию: история и email сохраняются до нового входа.
 */
class UserPreferences(private val context: Context) {

    val user: Flow<User?> = context.dataStore.data.map { prefs ->
        val encrypted = prefs[KEY_USER_ENCRYPTED] ?: return@map null
        cryptoManager.decrypt(encrypted)?.let(::userFromJson)
    }.flowOn(Dispatchers.IO)

    val token: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[KEY_TOKEN_ENCRYPTED]?.let(cryptoManager::decrypt)
    }.flowOn(Dispatchers.IO)

    /** Профиль и токен пишутся одним edit, чтобы не получить половину сессии. */
    suspend fun saveSession(user: User, token: String) {
        val (encryptedUser, encryptedToken) = withContext(Dispatchers.IO) {
            cryptoManager.encrypt(user.toJson()) to cryptoManager.encrypt(token)
        }
        context.dataStore.edit { prefs ->
            prefs[KEY_USER_ENCRYPTED] = encryptedUser
            prefs[KEY_TOKEN_ENCRYPTED] = encryptedToken
        }
    }

    /** Сбрасывает только токен: профиль нужен, чтобы отличить истёкшую сессию от выхода. */
    suspend fun clearToken() {
        context.dataStore.edit { prefs ->
            prefs.remove(KEY_TOKEN_ENCRYPTED)
        }
    }

    suspend fun clear() {
        context.dataStore.edit { prefs ->
            prefs.clear()
        }
    }
}
