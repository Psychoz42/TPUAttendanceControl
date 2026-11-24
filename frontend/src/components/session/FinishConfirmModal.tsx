"use client";

import React from "react";
import { AlertTriangle, StopCircle, RefreshCw } from "lucide-react";

interface FinishConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isStopping: boolean;
  studentsCount: number;
}

export function FinishConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  isStopping,
  studentsCount,
}: FinishConfirmModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-sm animate-in fade-in duration-200">
      <div
        className="bg-slate-900 border border-slate-800 rounded-3xl max-w-md w-full p-6 shadow-2xl relative"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center space-x-3 mb-4">
          <div className="w-12 h-12 rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center flex-shrink-0">
            <AlertTriangle className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-white tracking-tight">
              Завершить приём отметок?
            </h3>
            <p className="text-xs text-slate-400">Это действие необратимо</p>
          </div>
        </div>

        <div className="text-sm text-slate-300 space-y-2 mb-6">
          <p>
            После завершения занятия ротация QR-кода будет остановлена, а студенты больше не смогут отметиться.
          </p>
          <div className="p-3 rounded-xl bg-slate-800/80 border border-slate-700/80 text-xs text-slate-300">
            Отмечено студентов: <span className="text-emerald-400 font-bold">{studentsCount}</span> чел.
            <br />
            Итоговый список будет сохранён в локальном архиве браузера, и вы сможете в любой момент выгрузить его в Excel или распечатать.
          </div>
        </div>

        <div className="flex items-center justify-end space-x-3">
          <button
            type="button"
            onClick={onClose}
            disabled={isStopping}
            className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-medium transition-colors disabled:opacity-50"
          >
            Отмена
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isStopping}
            className="inline-flex items-center space-x-2 px-5 py-2 rounded-xl bg-red-600 hover:bg-red-500 text-white text-sm font-semibold shadow-lg shadow-red-600/30 transition-all disabled:opacity-50"
          >
            {isStopping ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" />
                <span>Завершение...</span>
              </>
            ) : (
              <>
                <StopCircle className="w-4 h-4" />
                <span>Завершить занятие</span>
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
