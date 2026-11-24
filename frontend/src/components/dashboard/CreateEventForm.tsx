"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { createEvent } from "@/lib/api/attendance";
import { storage } from "@/lib/storage";
import { ActiveSession } from "@/types/attendance";
import { PlusCircle, Clock, BookOpen, AlertCircle, Sparkles } from "lucide-react";

// Official TPU Pair Schedule
const TPU_PAIRS = [
  { label: "1 пара", time: "08:30" },
  { label: "2 пара", time: "10:25" },
  { label: "3 пара", time: "12:45" },
  { label: "4 пара", time: "14:40" },
  { label: "5 пара", time: "16:35" },
  { label: "6 пара", time: "18:30" },
  { label: "7 пара", time: "20:25" },
];

export function CreateEventForm() {
  const router = useRouter();
  const { token } = useAuth();

  const [eventName, setEventName] = useState("");
  const [eventTime, setEventTime] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleSetCurrentTime = () => {
    const now = new Date();
    const hours = String(now.getHours()).padStart(2, "0");
    const minutes = String(now.getMinutes()).padStart(2, "0");
    setEventTime(`${hours}:${minutes}`);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    const trimmedName = eventName.trim();
    const trimmedTime = eventTime.trim();

    if (!trimmedName) {
      setErrorMessage("Введите название дисциплины или занятия");
      return;
    }

    if (trimmedName.length > 200) {
      setErrorMessage("Название не может превышать 200 символов");
      return;
    }

    // Backend requires H:MM or HH:MM (0:00 to 23:59)
    const timeRegex = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/;
    if (!timeRegex.test(trimmedTime)) {
      setErrorMessage("Укажите корректное время в формате ЧЧ:ММ (например, 09:05 или 14:40)");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      const res = await createEvent(
        {
          event_name: trimmedName,
          event_time: trimmedTime,
        },
        token
      );

      // Create new session object
      const newSession: ActiveSession = {
        eventId: res.event_id,
        eventName: trimmedName,
        eventTime: trimmedTime,
        currentCode: res.code,
        notedStudents: [],
        isStopped: false,
        createdAt: new Date().toISOString(),
        lastUpdatedAt: new Date().toISOString(),
      };

      // Save to active session storage
      storage.saveActiveSession(newSession);

      // Navigate to presentation session page
      router.push(`/session/${res.event_id}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Не удалось создать занятие";
      setErrorMessage(msg);
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-xl">
      <div className="flex items-center space-x-3 mb-6">
        <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
          <PlusCircle className="w-6 h-6" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">Создать новое занятие</h2>
          <p className="text-xs text-slate-400">
            Сгенерирует динамический QR-код для проектора и запустит приём отметок
          </p>
        </div>
      </div>

      {errorMessage && (
        <div className="mb-6 p-4 rounded-2xl bg-red-500/10 border border-red-500/30 flex items-start space-x-3 text-red-300 text-sm">
          <AlertCircle className="w-5 h-5 flex-shrink-0 text-red-400 mt-0.5" />
          <span>{errorMessage}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Event Name */}
        <div>
          <label className="block text-sm font-semibold text-slate-200 mb-2">
            Название дисциплины / Занятия
          </label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
              <BookOpen className="w-5 h-5" />
            </div>
            <input
              type="text"
              required
              maxLength={200}
              value={eventName}
              onChange={(e) => setEventName(e.target.value)}
              placeholder="Например: Высшая математика (лекция)"
              className="w-full pl-11 pr-4 py-3 bg-slate-800 border border-slate-700 rounded-2xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all"
            />
          </div>
          <div className="flex justify-between text-[11px] text-slate-400 mt-1.5 px-1">
            <span>Будет отображаться у студентов в мобильном приложении</span>
            <span>{eventName.length}/200</span>
          </div>
        </div>

        {/* Event Time */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-semibold text-slate-200">
              Время пары (ЧЧ:ММ)
            </label>
            <button
              type="button"
              onClick={handleSetCurrentTime}
              className="text-xs text-emerald-400 hover:text-emerald-300 font-medium inline-flex items-center space-x-1"
            >
              <Sparkles className="w-3.5 h-3.5" />
              <span>Текущее время</span>
            </button>
          </div>

          <div className="relative mb-3">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
              <Clock className="w-5 h-5" />
            </div>
            <input
              type="text"
              required
              value={eventTime}
              onChange={(e) => setEventTime(e.target.value)}
              placeholder="08:30"
              className="w-full pl-11 pr-4 py-3 bg-slate-800 border border-slate-700 rounded-2xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all font-mono"
            />
          </div>

          {/* TPU Pair Presets */}
          <div>
            <span className="text-xs text-slate-400 block mb-2">Быстрый выбор звонков ТПУ:</span>
            <div className="grid grid-cols-4 sm:grid-cols-7 gap-2">
              {TPU_PAIRS.map((pair) => (
                <button
                  key={pair.time}
                  type="button"
                  onClick={() => setEventTime(pair.time)}
                  className={`px-2 py-1.5 rounded-xl border text-xs font-medium transition-all ${
                    eventTime === pair.time
                      ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/50"
                      : "bg-slate-800/80 hover:bg-slate-800 text-slate-300 border-slate-700"
                  }`}
                >
                  <div className="font-bold">{pair.time}</div>
                  <div className="text-[10px] text-slate-400">{pair.label}</div>
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full py-3.5 px-6 rounded-2xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-semibold text-sm shadow-lg shadow-emerald-600/30 transition-all flex items-center justify-center space-x-2 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 focus:ring-offset-slate-900 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? (
            <span>Генерация занятия и QR-кода...</span>
          ) : (
            <>
              <PlusCircle className="w-5 h-5" />
              <span>Создать занятие и открыть экран QR</span>
            </>
          )}
        </button>
      </form>
    </div>
  );
}
