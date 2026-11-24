"use client";

import React from "react";
import { Clock, RefreshCw } from "lucide-react";

interface QrTimerProps {
  secondsRemaining: number;
  progressPercent: number;
  isRotating: boolean;
  onForceRotate?: () => void;
  disabled?: boolean;
}

export function QrTimer({
  secondsRemaining,
  progressPercent,
  isRotating,
  onForceRotate,
  disabled = false,
}: QrTimerProps) {
  const isUrgent = secondsRemaining <= 2;

  return (
    <div className="w-full bg-slate-900/90 border border-slate-800 rounded-2xl p-4 shadow-lg backdrop-blur-sm">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2 text-slate-300">
          <Clock className={`w-4 h-4 ${isUrgent ? "text-amber-400" : "text-emerald-400"}`} />
          <span className="text-sm font-medium">Ротация QR-кода:</span>
          <span className={`font-mono font-bold text-base ${isUrgent ? "text-amber-400" : "text-white"}`}>
            {secondsRemaining} с
          </span>
        </div>

        {onForceRotate && (
          <button
            onClick={onForceRotate}
            disabled={disabled || isRotating}
            title="Обновить код немедленно"
            className="inline-flex items-center space-x-1.5 px-2.5 py-1 text-xs font-medium rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isRotating ? "animate-spin text-emerald-400" : ""}`} />
            <span>Обновить</span>
          </button>
        )}
      </div>

      {/* Progress Bar Container */}
      <div className="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden p-0.5 border border-slate-700/50">
        <div
          className={`h-full rounded-full transition-all duration-100 ease-linear ${
            isUrgent
              ? "bg-amber-500 shadow-sm shadow-amber-500/50"
              : "bg-emerald-500 shadow-sm shadow-emerald-500/50"
          }`}
          style={{ width: `${progressPercent}%` }}
        />
      </div>

      <div className="flex items-center justify-between text-[11px] text-slate-400 mt-2">
        <span>Каждые ~10 секунд сервер выдаёт новый код для защиты от пересылки</span>
        <span className="font-medium text-slate-300">TTL: 30 с</span>
      </div>
    </div>
  );
}
