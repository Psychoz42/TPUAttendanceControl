"use client";

import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { UserProfile } from "@/types/auth";
import { LoginRequest } from "@/types/api";
import { loginUser } from "@/lib/api/auth";
import { storage } from "@/lib/storage";

interface AuthContextType {
  user: UserProfile | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Initialize from storage
  useEffect(() => {
    const savedToken = storage.getToken();
    const savedUser = storage.getUser();

    if (savedToken && savedUser) {
      if (savedUser.role === "teacher") {
        setToken(savedToken);
        setUser(savedUser);
      } else {
        // If not teacher, purge
        storage.clearAuth();
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(async (credentials: LoginRequest) => {
    const res = await loginUser(credentials);

    if (res.role !== "teacher") {
      throw new Error(
        "Веб-интерфейс предназначен только для преподавателей. Студентам необходимо использовать мобильное приложение для сканирования QR-кодов."
      );
    }

    const profile: UserProfile = {
      username: res.username,
      fullName: res.fullName,
      email: res.email,
      group: res.group,
      role: res.role,
    };

    storage.setToken(res.token);
    storage.setUser(profile);

    setToken(res.token);
    setUser(profile);
  }, []);

  const logout = useCallback(() => {
    storage.clearAuth();
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        isAuthenticated: !!token && !!user,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
