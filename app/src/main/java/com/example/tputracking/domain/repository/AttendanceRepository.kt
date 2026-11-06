package com.example.tputracking.domain.repository

import androidx.paging.PagingData
import com.example.tputracking.domain.model.Attendance
import com.example.tputracking.domain.model.ScanResult
import kotlinx.coroutines.flow.Flow

interface AttendanceRepository {

    suspend fun scanQrCode(
        qrCode: String
    ): Result<ScanResult>

    fun getAttendancePaging():
            Flow<PagingData<Attendance>>
}