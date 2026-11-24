"use client";

import React, { useState } from "react";
import { X, FileSpreadsheet, FileText, Copy, Check, Printer } from "lucide-react";
import { exportToExcel, exportToCsv, copyStudentsToClipboard } from "@/lib/export";

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  eventName: string;
  eventTime: string;
  teacherName?: string;
  students: string[];
}

export function ExportModal({
  isOpen,
  onClose,
  eventName,
  eventTime,
  teacherName,
  students,
}: ExportModalProps) {
  const [copied, setCopied] = useState(false);

  if (!isOpen) return null;

  const handleCopy = async () => {
    const success = await copyStudentsToClipboard(students);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleExcel = () => {
    exportToExcel({
      eventName,
      eventTime,
      teacherName,
      students,
    });
  };

  const handleCsv = () => {
    exportToCsv({
      eventName,
      eventTime,
      teacherName,
      students,
    });
  };

  const handlePrint = () => {
    const printWindow = window.open("", "_blank");
    if (!printWindow) {
      window.print();
      return;
    }
    const dateStr = new Date().toLocaleDateString("ru-RU");
    const rows =
      students.length > 0
        ? students
            .map(
              (s, i) =>
                `<tr><td style="border:1px solid #333;padding:6px;text-align:center;">${i + 1}</td><td style="border:1px solid #333;padding:6px;">${s}</td><td style="border:1px solid #333;padding:6px;text-align:center;">Присутствовал</td></tr>`
            )
            .join("")
        : `<tr><td colspan="3" style="border:1px solid #333;padding:12px;text-align:center;">Нет отметившихся студентов</td></tr>`;

    printWindow.document.write(`
      <!DOCTYPE html>
      <html>
        <head>
          <title>Ведомость посещаемости — ${eventName}</title>
          <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding: 30px; color: #111; max-width: 800px; margin: 0 auto; }
            h2, h3 { margin: 4px 0; text-align: center; }
            h2 { font-size: 16px; font-weight: bold; }
            h3 { font-size: 15px; margin-bottom: 20px; }
            .header-info { margin: 20px 0; font-size: 13px; line-height: 1.8; border-bottom: 1px solid #ccc; padding-bottom: 12px; }
            table { width: 100%; border-collapse: collapse; margin-top: 15px; font-size: 13px; }
            th { border: 1px solid #333; padding: 8px; background: #f2f2f2; text-align: left; }
            th.center { text-align: center; }
            .footer { margin-top: 50px; display: flex; justify-content: space-between; font-size: 13px; }
          </style>
        </head>
        <body>
          <h2>НАЦИОНАЛЬНЫЙ ИССЛЕДОВАТЕЛЬСКИЙ ТОМСКИЙ ПОЛИТЕХНИЧЕСКИЙ УНИВЕРСИТЕТ</h2>
          <h3>ВЕДОМОСТЬ УЧЁТА ПОСЕЩАЕМОСТИ ЗАНЯТИЯ</h3>
          <div class="header-info">
            <div><strong>Дисциплина:</strong> ${eventName}</div>
            <div><strong>Время пары:</strong> ${eventTime}</div>
            <div><strong>Дата:</strong> ${dateStr}</div>
            <div><strong>Преподаватель:</strong> ${teacherName || "Не указан"}</div>
            <div><strong>Всего присутствующих студентов:</strong> ${students.length} чел.</div>
          </div>
          <table>
            <thead>
              <tr>
                <th class="center" style="width: 50px;">№</th>
                <th>ФИО Студента</th>
                <th class="center" style="width: 140px;">Отметка</th>
              </tr>
            </thead>
            <tbody>
              ${rows}
            </tbody>
          </table>
          <div class="footer">
            <div>Подпись преподавателя: ____________________</div>
            <div>Дата: ${dateStr}</div>
          </div>
        </body>
      </html>
    `);
    printWindow.document.close();
    printWindow.focus();
    setTimeout(() => {
      printWindow.print();
      printWindow.close();
    }, 250);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div
        className="bg-slate-900 border border-slate-800 rounded-3xl max-w-lg w-full p-6 shadow-2xl relative"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-5 right-5 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Modal Title */}
        <div className="mb-6">
          <h3 className="text-xl font-bold text-white tracking-tight">
            Экспорт ведомости посещаемости
          </h3>
          <p className="text-xs text-slate-400 mt-1">
            Занятие: <span className="text-slate-200 font-medium">{eventName}</span> ({eventTime})
          </p>
          <p className="text-xs text-slate-400">
            Всего отмечено студентов: <span className="text-emerald-400 font-semibold">{students.length}</span>
          </p>
        </div>

        {/* Export Options Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-6">
          {/* Excel Export */}
          <button
            onClick={handleExcel}
            className="flex items-center space-x-3 p-4 rounded-2xl bg-slate-800/80 hover:bg-emerald-600/10 border border-slate-700 hover:border-emerald-500/40 text-left transition-all group"
          >
            <div className="w-10 h-10 rounded-xl bg-emerald-500/20 text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform">
              <FileSpreadsheet className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-semibold text-white group-hover:text-emerald-300">Excel (.xlsx)</div>
              <div className="text-[11px] text-slate-400">Готовая ведомость ТПУ</div>
            </div>
          </button>

          {/* CSV Export */}
          <button
            onClick={handleCsv}
            className="flex items-center space-x-3 p-4 rounded-2xl bg-slate-800/80 hover:bg-teal-600/10 border border-slate-700 hover:border-teal-500/40 text-left transition-all group"
          >
            <div className="w-10 h-10 rounded-xl bg-teal-500/20 text-teal-400 flex items-center justify-center group-hover:scale-105 transition-transform">
              <FileText className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-semibold text-white group-hover:text-teal-300">CSV таблица</div>
              <div className="text-[11px] text-slate-400">Для 1С и таблиц</div>
            </div>
          </button>

          {/* Copy to Clipboard */}
          <button
            onClick={handleCopy}
            className="flex items-center space-x-3 p-4 rounded-2xl bg-slate-800/80 hover:bg-blue-600/10 border border-slate-700 hover:border-blue-500/40 text-left transition-all group"
          >
            <div className="w-10 h-10 rounded-xl bg-blue-500/20 text-blue-400 flex items-center justify-center group-hover:scale-105 transition-transform">
              {copied ? <Check className="w-5 h-5 text-emerald-400" /> : <Copy className="w-5 h-5" />}
            </div>
            <div>
              <div className="text-sm font-semibold text-white group-hover:text-blue-300">
                {copied ? "Скопировано!" : "Скопировать список"}
              </div>
              <div className="text-[11px] text-slate-400">Нумерованный текст</div>
            </div>
          </button>

          {/* Print */}
          <button
            onClick={handlePrint}
            className="flex items-center space-x-3 p-4 rounded-2xl bg-slate-800/80 hover:bg-purple-600/10 border border-slate-700 hover:border-purple-500/40 text-left transition-all group"
          >
            <div className="w-10 h-10 rounded-xl bg-purple-500/20 text-purple-400 flex items-center justify-center group-hover:scale-105 transition-transform">
              <Printer className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-semibold text-white group-hover:text-purple-300">Распечатать</div>
              <div className="text-[11px] text-slate-400">Печать ведомости</div>
            </div>
          </button>
        </div>

        {/* Footer */}
        <div className="flex justify-end">
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-medium transition-colors"
          >
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
}
