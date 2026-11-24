import { LoginRequest, LoginResponse } from "@/types/api";
import { requestApi } from "./client";

export async function loginUser(credentials: LoginRequest): Promise<LoginResponse> {
  return requestApi<LoginResponse>("/auth/login", {
    email: credentials.email.trim(),
    password: credentials.password,
  });
}
