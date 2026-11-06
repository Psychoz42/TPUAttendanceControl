"use client";

import React, { useState, useEffect, useCallback, use } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { Header } from "@/components/layout/Header";
import { QrDisplay } from "@/components/qr/QrDisplay";
import { QrTimer } from "@/components/qr/QrTimer";
import { AttendanceList } from "@/components/session/AttendanceList";
import { ExportModal } from "@/components/session/ExportModal";
import { FinishConfirmModal } from "@/components/session/FinishConfirmModal";
import { useAttendanceSession } from "@/hooks/useAttendanceSession";
import {
  Maximize,
  Minimize,
  FileSpreadsheet,
  StopCircle,
  ArrowLeft,
  AlertCircle,
  Calendar,
  Clock,
  Trash2,
  AlertTriangle,
} from "lucide-react";

interface SessionPageProps {
  params: Promise<{ id: string }>;
}

export default function SessionPage({ params }: SessionPageProps) {
  const resolvedParams = use(params);
  const router = useRouter();
  const { user, token, isAuthenticated, isLoading: authLoading } = useAuth();
  const eventIdNum = parseInt(resolvedParams.id, 10);

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isExportModalOpen, setIsExportModalOpen] = useState(false);
  const [isFinishModalOpen, setIsFinishModalOpen] = useState(false);

  const handleUnauthorized = useCallback(() => {
    router.replace("/login");
  }, [router]);

  // Hook for attendance loop, drift-free countdown and code rotation
  const {
    session,
    secondsRemaining,
    progressPercent,
    isRotating,
    rotationError,
    isStopped,
    isStopping,
    isCodeExpired,
    isSyncing,
    newlyAddedCount,
    stopEvent,
    removeEvent,
    forceRotate,
  } = useAttendanceSession({
    eventId: eventIdNum,
    token,
    teacherName: user?.fullName,
    teacherEmail: user?.email,
    onUnauthorized: handleUnauthorized,
  });

  // Warn user before leaving or closing tab while session is actively running
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (session && !isStopped && !isCodeExpired) {
        e.preventDefault();
        e.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [session, isStopped, isCodeExpired]);

  // Fullscreen state listener
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  const toggleFullscreen = async () => {
    try {
      if (!document.fullscreenElement) {
        await document.documentElement.requestFullscreen();
      } else {
        await document.exitFullscreen();
      }
    } catch (err) {
      console.warn("Fullscreen toggle failed:", err);
    }
  };

  const handleConfirmFinish = async () => {
    await stopEvent();
    setIsFinishModalOpen(false);
    setIsExportModalOpen(true);
  };

  const handleDelete = async () => {
    if (confirm("Вы действительно хотите удалить данное событие на сервере?")) {
      await removeEvent();
      router.push("/dashboard");
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="w-10 h-10 border-4 border-emerald-500/30 border-t-emerald-500 rounded-full animate-spin" />
      </div>
    );
  }

  if (!isAuthenticated) {
    router.replace("/login");
    return null;
  }

  if (!session && isSyncing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="w-10 h-10 border-4 border-emerald-500/30 border-t-emerald-500 rounded-full animate-spin" />
      </div>
    );
  }

  if (!session) {
    return (
      <div className="min-h-screen flex flex-col bg-slate-950 text-slate-100">
        <Header />
        <div className="flex-1 flex items-center justify-center p-6 text-center">
          <div className="max-w-md p-8 bg-slate-900 border border-slate-800 rounded-3xl shadow-xl">
            <AlertCircle className="w-12 h-12 text-amber-400 mx-auto mb-4" />
            <h2 className="text-xl font-bold text-white mb-2">Не удалось открыть занятие</h2>
            <p className="text-sm text-slate-400 mb-6">
              Занятие #{resolvedParams.id} не найдено, удалено или принадлежит другому преподавателю. Если пропала связь с сервером, обновите страницу.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <Link
                href="/dashboard"
                className="w-full sm:w-auto inline-flex items-center justify-center space-x-2 px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-sm transition-all shadow-lg shadow-emerald-600/30"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>К списку занятий</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen flex flex-col bg-slate-950 text-slate-100 ${isFullscreen ? "p-4 sm:p-6" : ""}`}>
      {/* Show header only when not in fullscreen mode */}
      {!isFullscreen && <Header />}

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* Top Control Bar */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-4 rounded-3xl bg-slate-900/90 border border-slate-800 shadow-xl backdrop-blur-sm">
          {/* Left: Event Details */}
          <div className="flex items-center space-x-3">
            {!isFullscreen && (
              <Link
                href="/dashboard"
                title="В дашборд"
                className="p-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors"
              >
                <ArrowLeft className="w-5 h-5" />
              </Link>
            )}
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="text-xl sm:text-2xl font-black text-white tracking-tight">
                  {session.eventName}
                </h1>
                <span className="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold bg-slate-800 text-emerald-400 border border-emerald-500/30">
                  ID: #{session.eventId}
                </span>
                {isStopped && (
                  <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-red-500/20 text-red-400 border border-red-500/30">
                    Завершено
                  </span>
                )}
                {isCodeExpired && (
                  <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-amber-500/20 text-amber-400 border border-amber-500/30">
                    Код истёк
                  </span>
                )}
              </div>
              <div className="flex items-center space-x-4 text-xs text-slate-400 mt-1">
                <span className="inline-flex items-center space-x-1">
                  <Clock className="w-3.5 h-3.5 text-slate-500" />
                  <span>Пара: {session.eventTime}</span>
                </span>
                <span className="inline-flex items-center space-x-1">
                  <Calendar className="w-3.5 h-3.5 text-slate-500" />
                  <span>{new Date(session.createdAt).toLocaleDateString("ru-RU")}</span>
                </span>
              </div>
            </div>
          </div>

          {/* Right: Actions */}
          <div className="flex flex-wrap items-center gap-2">
            {/* Fullscreen Button */}
            <button
              onClick={toggleFullscreen}
              className="inline-flex items-center space-x-1.5 px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition-colors"
            >
              {isFullscreen ? (
                <>
                  <Minimize className="w-4 h-4" />
                  <span>Обычный вид</span>
                </>
              ) : (
                <>
                  <Maximize className="w-4 h-4 text-emerald-400" />
                  <span>На проектор (F11)</span>
                </>
              )}
            </button>

            {/* Export Sheet Button */}
            <button
              onClick={() => setIsExportModalOpen(true)}
              className="inline-flex items-center space-x-1.5 px-3.5 py-2 rounded-xl bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/40 text-xs font-semibold transition-colors"
            >
              <FileSpreadsheet className="w-4 h-4 text-emerald-400" />
              <span>Ведомость ({session.notedStudents.length})</span>
            </button>

            {/* Finish / Stop Event Button */}
            {!isStopped && !isCodeExpired ? (
              <button
                onClick={() => setIsFinishModalOpen(true)}
                className="inline-flex items-center space-x-1.5 px-4 py-2 rounded-xl bg-red-600/20 hover:bg-red-600/30 text-red-300 border border-red-500/40 text-xs font-bold transition-colors"
              >
                <StopCircle className="w-4 h-4 text-red-400" />
                <span>Завершить приём</span>
              </button>
            ) : (
              <button
                onClick={handleDelete}
                className="inline-flex items-center space-x-1.5 px-3 py-2 rounded-xl bg-slate-800 hover:bg-red-500/10 hover:text-red-400 border border-slate-700 text-xs font-medium transition-colors text-slate-400"
              >
                <Trash2 className="w-4 h-4" />
                <span>Удалить с сервера</span>
              </button>
            )}
          </div>
        </div>

        {/* Warning if rotation error occurred (e.g. temporary network issue) */}
        {rotationError && !isStopped && !isCodeExpired && (
          <div className="p-4 rounded-2xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-between text-amber-300 text-xs">
            <div className="flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0 text-amber-400" />
              <span>
                Предупреждение: {rotationError}. Студенты всё ещё могут сканировать текущий код на экране.
              </span>
            </div>
            <button
              onClick={forceRotate}
              className="px-2.5 py-1 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 font-semibold text-xs transition-colors"
            >
              Повторить сейчас
            </button>
          </div>
        )}

        {/* Code Expired Banner */}
        {isCodeExpired && (
          <div className="p-4 rounded-2xl bg-amber-500/15 border border-amber-500/40 flex items-center justify-between text-amber-200 text-xs">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="w-4 h-4 flex-shrink-0 text-amber-400" />
              <span>
                Код занятия устарел или приём отметок был закрыт на сервере. Ротация остановлена. Список отметившихся студентов сохранён.
              </span>
            </div>
            <button
              onClick={() => setIsExportModalOpen(true)}
              className="px-3 py-1.5 rounded-xl bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold transition-colors"
            >
              Сохранить ведомость
            </button>
          </div>
        )}

        {/* Stopped Event Banner */}
        {isStopped && !isCodeExpired && (
          <div className="p-4 rounded-2xl bg-slate-900 border border-slate-800 flex items-center justify-between text-slate-300 text-xs">
            <div className="flex items-center space-x-2">
              <StopCircle className="w-4 h-4 text-red-400" />
              <span>
                Приём отметок завершён. QR-код более не принимается сервером. Не забудьте выгрузить ведомость посещаемости.
              </span>
            </div>
            <button
              onClick={() => setIsExportModalOpen(true)}
              className="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold transition-colors"
            >
              Выгрузить в Excel
            </button>
          </div>
        )}

        {/* Main Grid: QR & Timer (Left) + Student Attendance List (Right) */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Left Column: QR Presentation Box */}
          <div className="lg:col-span-5 flex flex-col items-center space-y-4">
            {/* A closed event has no code, there is nothing to scan */}
            {session.currentCode && (
              <QrDisplay
                code={session.currentCode}
                isRotating={isRotating}
                size={isFullscreen ? 380 : 320}
              />
            )}

            {/* Rotation Countdown Timer (active only when not stopped and not expired) */}
            {!isStopped && !isCodeExpired ? (
              <QrTimer
                secondsRemaining={secondsRemaining}
                progressPercent={progressPercent}
                isRotating={isRotating}
                onForceRotate={forceRotate}
                disabled={isStopped || isCodeExpired}
              />
            ) : (
              <div className="w-full text-center p-3 rounded-2xl bg-slate-900/60 border border-slate-800 text-xs text-slate-400 font-medium">
                Ротация QR-кода остановлена
              </div>
            )}
          </div>

          {/* Right Column: Attendance List */}
          <div className="lg:col-span-7 h-full">
            <AttendanceList
              students={session.notedStudents}
              newlyAddedCount={newlyAddedCount}
            />
          </div>
        </div>
      </main>

      {/* Modals */}
      <ExportModal
        isOpen={isExportModalOpen}
        onClose={() => setIsExportModalOpen(false)}
        eventName={session.eventName}
        eventTime={session.eventTime}
        teacherName={user?.fullName}
        students={session.notedStudents}
      />

      <FinishConfirmModal
        isOpen={isFinishModalOpen}
        onClose={() => setIsFinishModalOpen(false)}
        onConfirm={handleConfirmFinish}
        isStopping={isStopping}
        studentsCount={session.notedStudents.length}
      />
    </div>
  );
}
