"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { getOwnerEvents, deleteEvent } from "@/lib/api/attendance";
import { storage } from "@/lib/storage";
import { OwnerEventItem } from "@/types/api";
import { ActiveSession, ArchivedSession } from "@/types/attendance";
import { exportToExcel, exportToCsv } from "@/lib/export";
import {
  History,
  FileSpreadsheet,
  FileText,
  Trash2,
  Calendar,
  Clock,
  Users,
  ExternalLink,
  Radio,
  AlertTriangle,
  RefreshCw,
} from "lucide-react";

export function ArchiveList() {
  const { token, user } = useAuth();
  const [activeBackendEvents, setActiveBackendEvents] = useState<OwnerEventItem[]>([]);
  const [archivedSessions, setArchivedSessions] = useState<ArchivedSession[]>([]);
  const [activeSessions, setActiveSessions] = useState<ActiveSession[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const loadData = useCallback(async () => {
    // 1. Load from local storage
    const localActive = storage.getAllActiveSessions();
    const localArchived = storage.getArchivedSessions();
    setActiveSessions(localActive);
    setArchivedSessions(localArchived);

    // 2. Fetch server events if token is available
    if (token) {
      try {
        const events = await getOwnerEvents(token);
        setActiveBackendEvents(events || []);
      } catch (err) {
        console.warn("Could not fetch owner events from backend:", err);
      }
    }
    setIsLoading(false);
    setIsRefreshing(false);
  }, [token]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleRefresh = () => {
    setIsRefreshing(true);
    loadData();
  };

  const handleDeleteBackendEvent = async (eventId: number) => {
    if (!token) return;
    if (confirm(`Удалить занятие #${eventId} с сервера?`)) {
      try {
        await deleteEvent(eventId, token);
        storage.clearActiveSession(eventId);
        setActiveBackendEvents((prev) => prev.filter((e) => e.event_id !== eventId));
        setActiveSessions((prev) => prev.filter((s) => s.eventId !== eventId));
      } catch (err) {
        console.error("Failed to delete event:", err);
      }
    }
  };

  const handleDeleteArchive = (id: string) => {
    if (confirm("Удалить данную запись из архива браузера?")) {
      storage.deleteArchivedSession(id);
      setArchivedSessions(storage.getArchivedSessions());
    }
  };

  const handleExcelExport = (session: ArchivedSession) => {
    exportToExcel({
      eventName: session.eventName,
      eventTime: session.eventTime,
      teacherName: session.teacherName || user?.fullName,
      date: new Date(session.createdAt).toLocaleDateString("ru-RU"),
      students: session.students,
    });
  };

  const handleCsvExport = (session: ArchivedSession) => {
    exportToCsv({
      eventName: session.eventName,
      eventTime: session.eventTime,
      teacherName: session.teacherName || user?.fullName,
      date: new Date(session.createdAt).toLocaleDateString("ru-RU"),
      students: session.students,
    });
  };

  // Find server events that don't have matching local QR code sessions
  const localEventIds = new Set(activeSessions.map((s) => s.eventId));
  const orphanServerEvents = activeBackendEvents.filter((e) => !localEventIds.has(e.event_id));

  // Active ongoing sessions (unstopped)
  const ongoingSessions = activeSessions.filter((s) => !s.isStopped);

  return (
    <div className="space-y-6">
      {/* 1. ONGOING / ACTIVE SESSIONS */}
      <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 shadow-xl">
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-slate-800">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <Radio className="w-5 h-5 animate-pulse" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white tracking-tight">
                Текущие активные пары
              </h3>
              <p className="text-xs text-slate-400">
                Занятия, на которых сейчас открыт приём отметок
              </p>
            </div>
          </div>

          <button
            onClick={handleRefresh}
            title="Обновить список"
            className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>

        {ongoingSessions.length === 0 && orphanServerEvents.length === 0 ? (
          <div className="text-center py-6 px-4 border border-dashed border-slate-800 rounded-2xl">
            <p className="text-sm text-slate-400">Сейчас нет активных занятий</p>
            <p className="text-xs text-slate-500 mt-1">
              Создайте занятие в форме слева, чтобы запустить QR-код для студентов.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {/* Locally stored active sessions with valid QR codes */}
            {ongoingSessions.map((session) => (
              <div
                key={session.eventId}
                className="p-4 rounded-2xl bg-gradient-to-r from-emerald-950/60 to-slate-800/80 border border-emerald-500/40 shadow-lg flex flex-col sm:flex-row sm:items-center justify-between gap-4"
              >
                <div className="flex items-center space-x-3">
                  <span className="relative flex h-3 w-3 flex-shrink-0">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
                  </span>
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-[11px] font-bold text-emerald-400 uppercase tracking-wider">
                        Идёт сейчас
                      </span>
                      <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-slate-800 text-slate-300">
                        #{session.eventId}
                      </span>
                    </div>
                    <div className="text-base font-bold text-white mt-0.5">
                      {session.eventName} ({session.eventTime})
                    </div>
                    <div className="text-xs text-slate-300 mt-0.5">
                      Отмечено: <span className="font-semibold text-emerald-400">{session.notedStudents.length}</span> чел.
                    </div>
                  </div>
                </div>

                <Link
                  href={`/session/${session.eventId}`}
                  className="inline-flex items-center justify-center space-x-1.5 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold transition-all shadow-md shadow-emerald-600/30 flex-shrink-0"
                >
                  <span>К экрану QR</span>
                  <ExternalLink className="w-3.5 h-3.5" />
                </Link>
              </div>
            ))}

            {/* Orphan server events created outside this browser */}
            {orphanServerEvents.map((evt) => (
              <div
                key={evt.event_id}
                className="p-3.5 rounded-2xl bg-slate-800/50 border border-amber-500/30 flex items-center justify-between gap-3 text-xs"
              >
                <div className="flex items-center space-x-2.5">
                  <AlertTriangle className="w-4 h-4 text-amber-400 flex-shrink-0" />
                  <div>
                    <div className="text-white font-medium">
                      #{evt.event_id} — {evt.event_name}
                    </div>
                    <div className="text-slate-400 text-[11px]">
                      Код отсутствует в браузере (создано в другом месте)
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleDeleteBackendEvent(evt.event_id)}
                  title="Удалить неактивное занятие с сервера"
                  className="inline-flex items-center space-x-1 px-2.5 py-1.5 rounded-xl bg-slate-700 hover:bg-red-500/20 text-slate-300 hover:text-red-300 font-medium transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>Удалить</span>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 2. ARCHIVE OF FINISHED SESSIONS */}
      <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 shadow-xl">
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-slate-800">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-2xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400">
              <History className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white tracking-tight">
                Архив завершённых пар
              </h3>
              <p className="text-xs text-slate-400">
                Сохранённые ведомости с возможностью выгрузки
              </p>
            </div>
          </div>

          <div className="text-xs font-medium text-slate-400">
            Записей: <span className="text-white font-bold">{archivedSessions.length}</span>
          </div>
        </div>

        {isLoading ? (
          <div className="text-center py-8 text-slate-400 text-sm">Загрузка данных...</div>
        ) : archivedSessions.length === 0 ? (
          <div className="text-center py-8 px-4 border border-dashed border-slate-800 rounded-2xl">
            <p className="text-sm text-slate-300 font-medium">Архив пока пуст</p>
            <p className="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
              Когда вы нажмёте «Завершить приём» на экране занятия, итоговая ведомость со списком студентов появится здесь для выгрузки в Excel.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {archivedSessions.map((session) => {
              const formattedDate = new Date(session.createdAt).toLocaleDateString("ru-RU", {
                day: "2-digit",
                month: "short",
                year: "numeric",
              });

              return (
                <div
                  key={session.id}
                  className="p-4 rounded-2xl bg-slate-800/60 border border-slate-700/60 hover:bg-slate-800 transition-colors flex flex-col md:flex-row md:items-center justify-between gap-4"
                >
                  <div className="space-y-1">
                    <div className="flex items-center space-x-2">
                      <span className="text-base font-bold text-white">
                        {session.eventName}
                      </span>
                      <span className="px-2 py-0.5 rounded-md text-[11px] font-mono bg-slate-700 text-slate-300">
                        #{session.eventId}
                      </span>
                    </div>

                    <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-400">
                      <span className="inline-flex items-center space-x-1">
                        <Calendar className="w-3.5 h-3.5 text-slate-500" />
                        <span>{formattedDate}</span>
                      </span>
                      <span className="inline-flex items-center space-x-1">
                        <Clock className="w-3.5 h-3.5 text-slate-500" />
                        <span>{session.eventTime}</span>
                      </span>
                      <span className="inline-flex items-center space-x-1 text-emerald-400 font-medium">
                        <Users className="w-3.5 h-3.5" />
                        <span>Отмечено: {session.totalCount} чел.</span>
                      </span>
                    </div>
                  </div>

                  {/* Actions: Export Excel, CSV, Delete */}
                  <div className="flex items-center space-x-2 self-end md:self-auto">
                    <button
                      onClick={() => handleExcelExport(session)}
                      title="Выгрузить в Excel (.xlsx)"
                      className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-xl bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-xs font-semibold transition-colors"
                    >
                      <FileSpreadsheet className="w-4 h-4" />
                      <span>Excel</span>
                    </button>

                    <button
                      onClick={() => handleCsvExport(session)}
                      title="Выгрузить в CSV"
                      className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-xl bg-teal-500/10 hover:bg-teal-500/20 text-teal-400 border border-teal-500/30 text-xs font-semibold transition-colors"
                    >
                      <FileText className="w-4 h-4" />
                      <span>CSV</span>
                    </button>

                    <button
                      onClick={() => handleDeleteArchive(session.id)}
                      title="Удалить из архива"
                      className="p-1.5 rounded-xl text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
