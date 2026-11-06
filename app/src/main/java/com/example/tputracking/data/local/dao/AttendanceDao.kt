package com.example.tputracking.data.local.dao

import androidx.paging.PagingSource
import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.example.tputracking.data.local.entity.AttendanceEntity

@Dao
interface AttendanceDao {

    @Query("""
        SELECT * FROM attendance
        ORDER BY id DESC
    """)
    fun getAttendancePaging():
            PagingSource<Int, AttendanceEntity>

    @Insert(
        onConflict =
            OnConflictStrategy.REPLACE
    )
    suspend fun insertAll(
        items: List<AttendanceEntity>
    )

    @Query("SELECT EXISTS(SELECT 1 FROM attendance WHERE subject = :subject AND dateTime = :dateTime)")
    suspend fun exists(
        subject: String,
        dateTime: String
    ): Boolean

    @Query("DELETE FROM attendance")
    suspend fun clearAll()
}