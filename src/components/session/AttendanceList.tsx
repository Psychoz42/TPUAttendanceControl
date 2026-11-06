"use client";

import React, { useState, useMemo } from "react";
import { Users, Search, UserCheck, Sparkles } from "lucide-react";

interface AttendanceListProps {
  students: string[];
  newlyAddedCount?: number;
}

export function AttendanceList({ students, newlyAddedCount = 0 }: AttendanceListProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredStudents = useMemo(() => {
    if (!searchQuery.trim()) return students;
    const query = searchQuery.toLowerCase().trim();
    return students.filter((s) => s.toLowerCase().includes(query));
  }, [students, searchQuery]);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 shadow-xl flex flex-col h-full">
      {/* Header with Counter */}
      <div className="flex items-center justify-between pb-4 border-b border-slate-800">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white tracking-tight">Отметившиеся студенты</h2>
            <p className="text-xs text-slate-400">Обновляется автоматически</p>
          </div>
        </div>

        {/* Counter Badge */}
        <div className="flex items-center space-x-2">
          {newlyAddedCount > 0 && (
            <div className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 text-xs font-semibold animate-pulse">
              <Sparkles className="w-3.5 h-3.5 text-emerald-400" />
              <span>+{newlyAddedCount} нов.</span>
            </div>
          )}
          <div className="px-3.5 py-1.5 rounded-xl bg-slate-800 border border-slate-700 text-sm font-bold text-white shadow-inner">
            <span className="text-emerald-400 text-base">{students.length}</span> чел.
          </div>
        </div>
      </div>

      {/* Search Input */}
      {students.length > 0 && (
        <div className="mt-4 relative">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
            <Search className="w-4 h-4" />
          </div>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Поиск по фамилии или имени..."
            className="w-full pl-10 pr-4 py-2 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all"
          />
        </div>
      )}

      {/* Student List Content */}
      <div className="mt-4 flex-1 overflow-y-auto max-h-[500px] pr-1 space-y-2">
        {students.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
            <div className="w-14 h-14 rounded-2xl bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-500 mb-3 animate-pulse">
              <UserCheck className="w-7 h-7" />
            </div>
            <p className="text-sm font-medium text-slate-300">Список пока пуст</p>
            <p className="text-xs text-slate-500 mt-1 max-w-xs">
              Когда студенты отсканируют QR-код через мобильное приложение, их ФИО появятся здесь в реальном времени.
            </p>
          </div>
        ) : filteredStudents.length === 0 ? (
          <div className="py-12 text-center text-slate-400 text-sm">
            Студентов по запросу «{searchQuery}» не найдено
          </div>
        ) : (
          filteredStudents.map((student, idx) => (
            <div
              key={`${student}-${idx}`}
              className="flex items-center justify-between p-3 rounded-xl bg-slate-800/60 hover:bg-slate-800 border border-slate-700/60 transition-colors"
            >
              <div className="flex items-center space-x-3">
                <span className="w-6 text-xs font-mono text-slate-400 text-right">
                  {idx + 1}.
                </span>
                <span className="text-sm font-medium text-slate-200">
                  {student}
                </span>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Присутствует
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
