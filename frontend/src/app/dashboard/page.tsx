"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { Header } from "@/components/layout/Header";
import { CreateEventForm } from "@/components/dashboard/CreateEventForm";
import { ArchiveList } from "@/components/dashboard/ArchiveList";

export default function DashboardPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading || !isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="w-10 h-10 border-4 border-emerald-500/30 border-t-emerald-500 rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-slate-950">
      <Header />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Welcome Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Панель управления преподавателя
            </h1>
            <p className="text-sm text-slate-400 mt-1">
              Здравствуйте, <span className="text-emerald-400 font-medium">{user?.fullName}</span>. Выберите пару или создайте новое занятие.
            </p>
          </div>
        </div>

        {/* Main Grid: Create Event Form (Left) & Active/Archived Events (Right) */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          <div className="lg:col-span-6">
            <CreateEventForm />
          </div>

          <div className="lg:col-span-6">
            <ArchiveList />
          </div>
        </div>
      </main>
    </div>
  );
}
