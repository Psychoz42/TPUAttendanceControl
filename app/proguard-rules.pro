# Add project specific ProGuard rules here.
# You can control the set of applied configuration files using the
# proguardFiles setting in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# Gson заполняет DTO через рефлексию: имена полей должны совпадать с JSON-ключами,
# поэтому классы DTO не обфусцируем.
-keepattributes Signature, *Annotation*
-keep class com.example.tputracking.data.remote.dto.** { <fields>; <init>(...); }

# Скрываем имена исходных файлов в стектрейсах релизной сборки.
-renamesourcefileattribute SourceFile

# Вырезаем отладочные логи из релиза, чтобы в logcat не утекали данные.
-assumenosideeffects class android.util.Log {
    public static int v(...);
    public static int d(...);
    public static int i(...);
}
