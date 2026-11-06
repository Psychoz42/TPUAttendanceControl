package com.example.tputracking.data.remote.api

import com.example.tputracking.data.remote.dto.ScanQrRequest
import com.example.tputracking.data.remote.dto.ScanQrResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface AttendanceApiService {

    @POST("/api/v1/attendance/scan")
    suspend fun scanQrCode(
        @Body request: ScanQrRequest
    ): Response<ScanQrResponse>
}