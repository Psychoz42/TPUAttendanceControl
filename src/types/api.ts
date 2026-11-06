export type Role = "student" | "teacher";

export interface ApiErrorResponse {
  success: false;
  code:
    | "INVALID_DATA_FORMAT"
    | "LOGIN_ERROR"
    | "UNKNOWN_USER"
    | "INVALID_QR"
    | "EVENT_NOT_FOUND"
    | "DB_ERROR"
    | "INTERNAL_ERROR"
    | "UNAUTHORIZED"
    | "FORBIDDEN"
    | "TOO_MANY_REQUESTS"
    | "HTTP_ERROR"
    | string;
  message: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  username: string;
  fullName: string;
  email: string;
  group: string;
  role: Role;
  token: string;
}

export interface CreateEventRequest {
  event_name: string;
  event_time: string;
  event_owner?: string;
}

export interface CreateEventResponse {
  event_id: number;
  code: string;
}

export interface UpdateCodeRequest {
  code: string;
}

export interface UpdateCodeResponse {
  new_code: string;
  noted_students: string[] | null;
}

export interface OwnerEventItem {
  event_id: number;
  event_name: string;
  event_time: string;
  created_at: string;
  active: boolean;
  noted_count: number;
}

export type OwnerEventsResponse = OwnerEventItem[];

export interface EventDetails {
  event_id: number;
  event_name: string;
  event_time: string;
  created_at: string;
  active: boolean;
  code?: string; // Present only while the event is open
  noted_students: string[];
}

export interface EventIdRequest {
  event_id: number;
}

export interface SuccessOnlyResponse {
  success: true;
}
