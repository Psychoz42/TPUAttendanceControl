"use client";

import React from "react";
import { QRCodeSVG } from "qrcode.react";
import { ShieldCheck, RefreshCw } from "lucide-react";

interface QrDisplayProps {
  code: string;
  isRotating?: boolean;
  size?: number;
}

export function QrDisplay({ code, isRotating = false, size = 320 }: QrDisplayProps) {
  return (
    <div className="flex flex-col items-center justify-center p-6 bg-white rounded-3xl shadow-2xl border border-slate-200">
      {/* Container for QR with white padding */}
      <div className="relative p-4 bg-white rounded-2xl border-2 border-slate-100 flex items-center justify-center">
        {/* Subtle update pulse indicator */}
        {isRotating && (
          <div className="absolute inset-0 bg-white/70 backdrop-blur-[2px] rounded-2xl flex items-center justify-center z-10 transition-opacity">
            <div className="flex items-center space-x-2 px-4 py-2 bg-slate-900/90 text-white rounded-xl shadow-lg text-sm font-medium">
              <RefreshCw className="w-4 h-4 animate-spin text-emerald-400" />
              <span>Ротация кода...</span>
            </div>
          </div>
        )}

        {/* The QR Code - STRICTLY the 16 characters string required by backend and mobile app */}
        <QRCodeSVG
          value={code || "TPU-ATTENDANCE-N"}
          size={size}
          level="M"
          marginSize={2}
          bgColor="#FFFFFF"
          fgColor="#0F172A"
        />
      </div>

      {/* Code verification string & security badge */}
      <div className="mt-5 flex flex-col items-center text-center">
        <div className="flex items-center space-x-2 text-xs font-mono text-slate-500 bg-slate-100 px-3 py-1.5 rounded-lg border border-slate-200">
          <ShieldCheck className="w-4 h-4 text-emerald-600" />
          <span>Код:</span>
          <span className="font-bold text-slate-800 tracking-wider select-all">{code}</span>
        </div>
        <p className="text-xs text-slate-400 mt-2">
          Студенты сканируют данный QR-код через мобильное приложение ТПУ
        </p>
      </div>
    </div>
  );
}
