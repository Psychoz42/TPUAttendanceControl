import { ApiErrorResponse } from "@/types/api";
import { storage } from "@/lib/storage";

export class ApiError extends Error {
  code: string;
  status: number;
  originalMessage: string;

  constructor(code: string, status: number, userFriendlyMessage: string, originalMessage: string) {
    super(userFriendlyMessage);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.originalMessage = originalMessage;
  }
}

export const ERROR_MESSAGES: Record<string, string> = {
  INVALID_DATA_FORMAT: "Некорректный формат данных в запросе",
  LOGIN_ERROR: "Неверный email или пароль",
  UNKNOWN_USER: "Пользователь не найден в системе",
  INVALID_QR: "QR-код устарел, ротация остановлена или занятие закрыто",
  EVENT_NOT_FOUND: "Занятие не найдено или принадлежит другому преподавателю",
  DB_ERROR: "Внутренняя ошибка базы данных сервера",
  INTERNAL_ERROR: "Внутренняя ошибка сервера",
  UNAUTHORIZED: "Сессия истекла или токен недействителен. Пожалуйста, войдите снова",
  FORBIDDEN: "Доступ запрещен. Веб-интерфейс предназначен только для преподавателей",
  TOO_MANY_REQUESTS: "Слишком много запросов. Пожалуйста, подождите минуту",
  HTTP_ERROR: "Ошибка сетевого соединения с сервером",
};

export function getFriendlyErrorMessage(code: string, rawMessage?: string): string {
  if (code && ERROR_MESSAGES[code]) {
    return ERROR_MESSAGES[code];
  }
  return rawMessage || "Произошла непредвиденная ошибка";
}

const DEFAULT_API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:80/api/v1";

export function getApiBaseUrl(): string {
  // If set in environment, use it
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL.replace(/\/$/, "");
  }
  return DEFAULT_API_URL.replace(/\/$/, "");
}

export async function requestApi<T>(
  path: string,
  body: Record<string, unknown> = {},
  token?: string | null
): Promise<T> {
  const baseUrl = getApiBaseUrl();
  const url = `${baseUrl}${path.startsWith("/") ? path : `/${path}`}`;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res: Response;
  try {
    res = await fetch(url, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
    });
  } catch (err) {
    throw new ApiError(
      "NETWORK_ERROR",
      0,
      "Не удалось связаться с сервером. Проверьте подключение к сети.",
      err instanceof Error ? err.message : String(err)
    );
  }

  let data: unknown;
  try {
    data = await res.json();
  } catch {
    throw new ApiError(
      "HTTP_ERROR",
      res.status,
      `Сервер вернул некорректный ответ (HTTP ${res.status})`,
      `HTTP status ${res.status}`
    );
  }

  // The saved token is no longer valid (it lives 12 hours): the only way forward is to log in again
  if (res.status === 401 && token && typeof window !== "undefined") {
    storage.clearAuth();
    window.location.assign("/login");
  }

  // Handle explicit HTTP 4xx/5xx errors
  if (!res.ok) {
    const errorBody = data as Partial<ApiErrorResponse>;
    const code = errorBody?.code || (res.status === 401 ? "UNAUTHORIZED" : res.status === 403 ? "FORBIDDEN" : res.status === 429 ? "TOO_MANY_REQUESTS" : "HTTP_ERROR");
    const rawMsg = errorBody?.message || res.statusText;
    throw new ApiError(code, res.status, getFriendlyErrorMessage(code, rawMsg), rawMsg);
  }

  // Handle business errors returned with HTTP 200
  if (data && typeof data === "object" && !Array.isArray(data)) {
    const maybeError = data as Record<string, unknown>;
    if (maybeError.success === false && typeof maybeError.code === "string") {
      const code = maybeError.code;
      const rawMsg = typeof maybeError.message === "string" ? maybeError.message : "";
      throw new ApiError(code, 200, getFriendlyErrorMessage(code, rawMsg), rawMsg);
    }
  }

  return data as T;
}
