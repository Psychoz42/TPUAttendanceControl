"use client";

import React from "react";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { LogOut, User, GraduationCap } from "lucide-react";

export function Header() {
  const { user, isAuthenticated, logout } = useAuth();

  return (
    <header className="bg-slate-900 border-b border-slate-800 text-white sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand / Logo */}
        <Link
          href="/dashboard"
          className="flex items-center space-x-3 group focus:outline-none focus:ring-2 focus:ring-emerald-500 rounded-lg p-1"
        >
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-700 flex items-center justify-center shadow-lg shadow-emerald-500/20 group-hover:scale-105 transition-transform">
            <GraduationCap className="w-6 h-6 text-white" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-lg tracking-tight text-white">ТПУ</span>
              <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                Преподаватель
              </span>
            </div>
            <p className="text-xs text-slate-400">Учёт посещаемости</p>
          </div>
        </Link>

        {/* User Info & Actions */}
        {isAuthenticated && user && (
          <div className="flex items-center space-x-4">
            <div className="hidden sm:flex items-center space-x-3 text-right">
              <div>
                <div className="text-sm font-medium text-slate-200">{user.fullName}</div>
                <div className="text-xs text-slate-400">{user.email}</div>
              </div>
              <div className="w-9 h-9 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-300">
                <User className="w-4 h-4" />
              </div>
            </div>

            <button
              onClick={logout}
              title="Выйти из аккаунта"
              className="inline-flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-red-500/10 hover:text-red-400 border border-slate-700 hover:border-red-500/30 text-slate-300 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-red-500"
            >
              <LogOut className="w-4 h-4" />
              <span className="hidden md:inline">Выход</span>
            </button>
          </div>
        )}
      </div>
    </header>
  );
}
