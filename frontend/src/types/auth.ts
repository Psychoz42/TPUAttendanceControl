import { Role } from "./api";

export interface UserProfile {
  username: string;
  fullName: string;
  email: string;
  group: string;
  role: Role;
}

export interface AuthState {
  user: UserProfile | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}
