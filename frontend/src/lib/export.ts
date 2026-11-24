import * as XLSX from "xlsx";

export interface ExportData {
  eventName: string;
  eventTime: string;
  date?: string;
  teacherName?: string;
  students: string[];
}

export function exportToExcel(data: ExportData): void {
  const dateStr = data.date || new Date().toLocaleDateString("ru-RU");

  // Prepare header rows
  const headerData = [
    ["НАЦИОНАЛЬНЫЙ ИССЛЕДОВАТЕЛЬСКИЙ ТОМСКИЙ ПОЛИТЕХНИЧЕСКИЙ УНИВЕРСИТЕТ"],
    ["ВЕДОМОСТЬ УЧЁТА ПОСЕЩАЕМОСТИ ЗАНЯТИЯ"],
    [],
    ["Дисциплина / Предмет:", data.eventName],
    ["Время пары:", data.eventTime],
    ["Дата проведения:", dateStr],
    ["Преподаватель:", data.teacherName || "Не указан"],
    ["Всего отметившихся:", data.students.length],
    [],
    ["№", "ФИО Студента", "Статус"],
  ];

  // Student rows
  const studentRows = data.students.map((student, index) => [
    index + 1,
    student,
    "Присутствовал",
  ]);

  const fullData = [...headerData, ...studentRows];

  const worksheet = XLSX.utils.aoa_to_sheet(fullData);

  // Set column widths
  worksheet["!cols"] = [
    { wch: 6 }, // №
    { wch: 45 }, // ФИО
    { wch: 18 }, // Статус
  ];

  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Посещаемость");

  // Sanitize filename
  const cleanName = data.eventName.replace(/[/\\?%*:|"<>]/g, "_").slice(0, 30);
  const filename = `Посещаемость_${cleanName}_${dateStr.replace(/\./g, "-")}.xlsx`;

  XLSX.writeFile(workbook, filename);
}

export function exportToCsv(data: ExportData): void {
  const dateStr = data.date || new Date().toLocaleDateString("ru-RU");

  const lines: string[] = [
    `"Ведомость учёта посещаемости ТПУ"`,
    `"Предмет: ${data.eventName.replace(/"/g, '""')}"`,
    `"Время: ${data.eventTime}"`,
    `"Дата: ${dateStr}"`,
    `"Преподаватель: ${(data.teacherName || "").replace(/"/g, '""')}"`,
    `"Всего студентов: ${data.students.length}"`,
    "",
    `"№";"ФИО Студента";"Статус"`,
  ];

  data.students.forEach((student, idx) => {
    lines.push(`${idx + 1};"${student.replace(/"/g, '""')}";"Присутствовал"`);
  });

  // \uFEFF for UTF-8 BOM so Excel opens Cyrillic properly
  const csvContent = "\uFEFF" + lines.join("\r\n");
  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);

  const cleanName = data.eventName.replace(/[/\\?%*:|"<>]/g, "_").slice(0, 30);
  const filename = `Посещаемость_${cleanName}_${dateStr.replace(/\./g, "-")}.csv`;

  const link = document.createElement("a");
  link.setAttribute("href", url);
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export async function copyStudentsToClipboard(students: string[]): Promise<boolean> {
  if (students.length === 0) return false;
  const text = students.map((s, idx) => `${idx + 1}. ${s}`).join("\n");
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
