"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { ActiveSession, ArchivedSession } from "@/types/attendance";
import { updateCode, disableEventUpdating, deleteEvent } from "@/lib/api/attendance";
import { storage } from "@/lib/storage";

const ROTATION_INTERVAL_MS = 10_000; // 10 seconds
const RETRY_INTERVAL_MS = 3_000; // 3 seconds on error retry
const TICK_INTERVAL_MS = 100; // 100ms update for smooth UI

interface UseAttendanceSessionProps {
  eventId: number;
  token: string | null;
  teacherName?: string;
  teacherEmail?: string;
  onUnauthorized?: () => void;
}

export function useAttendanceSession({
  eventId,
  token,
  teacherName = "",
  teacherEmail = "",
  onUnauthorized,
}: UseAttendanceSessionProps) {
  // Load session directly from storage by eventId as initial state
  const [session, setSession] = useState<ActiveSession | null>(() => {
    if (typeof window === "undefined") return null;
    return storage.getActiveSession(eventId);
  });

  const [secondsRemaining, setSecondsRemaining] = useState<number>(10);
  const [progressPercent, setProgressPercent] = useState<number>(100);
  const [isRotating, setIsRotating] = useState<boolean>(false);
  const [rotationError, setRotationError] = useState<string | null>(null);
  const [isStopped, setIsStopped] = useState<boolean>(() => {
    if (typeof window === "undefined") return false;
    const stored = storage.getActiveSession(eventId);
    return stored?.isStopped || false;
  });
  const [isStopping, setIsStopping] = useState<boolean>(false);
  const [newlyAddedCount, setNewlyAddedCount] = useState<number>(0);
  const [isCodeExpired, setIsCodeExpired] = useState<boolean>(false);

  // Stable references for handlers to prevent infinite render loops
  const onUnauthorizedRef = useRef(onUnauthorized);
  useEffect(() => {
    onUnauthorizedRef.current = onUnauthorized;
  }, [onUnauthorized]);

  const sessionRef = useRef<ActiveSession | null>(session);
  sessionRef.current = session;

  const tokenRef = useRef<string | null>(token);
  tokenRef.current = token;

  const isRotatingRef = useRef<boolean>(false);
  isRotatingRef.current = isRotating;

  const nextRotationTimeRef = useRef<number>(Date.now() + ROTATION_INTERVAL_MS);
  const prevStudentsCountRef = useRef<number>(session?.notedStudents?.length || 0);

  // Hydration sync: ensure session is loaded on client mount
  useEffect(() => {
    const stored = storage.getActiveSession(eventId);
    if (stored) {
      setSession(stored);
      setIsStopped(stored.isStopped || false);
      sessionRef.current = stored;
      prevStudentsCountRef.current = stored.notedStudents?.length || 0;
    }
  }, [eventId]);

  // Perform a single rotation call
  const performRotation = useCallback(async () => {
    const currentSession = sessionRef.current;
    if (!currentSession || currentSession.isStopped) return;
    if (isRotatingRef.current) return; // Prevent concurrent calls

    const currentToken = tokenRef.current || storage.getToken();
    if (!currentToken) return;

    isRotatingRef.current = true;
    setIsRotating(true);
    setRotationError(null);

    try {
      const response = await updateCode(currentSession.currentCode, currentToken);
      const incomingStudents = response.noted_students || [];

      // Check if new students appeared
      const diff = incomingStudents.length - prevStudentsCountRef.current;
      if (diff > 0) {
        setNewlyAddedCount(diff);
      }
      prevStudentsCountRef.current = incomingStudents.length;

      const updatedSession: ActiveSession = {
        ...currentSession,
        currentCode: response.new_code,
        notedStudents: incomingStudents,
        lastUpdatedAt: new Date().toISOString(),
      };

      // Update state and persistence
      sessionRef.current = updatedSession;
      setSession(updatedSession);
      storage.saveActiveSession(updatedSession);

      // Reset timer to full 10 seconds from now
      nextRotationTimeRef.current = Date.now() + ROTATION_INTERVAL_MS;
      setSecondsRemaining(10);
      setProgressPercent(100);
      setIsCodeExpired(false);
    } catch (err: unknown) {
      console.warn("Rotation request error:", err);
      const errorObj = err as { code?: string; message?: string };
      const errorCode = errorObj?.code;
      const errorMsg = err instanceof Error ? err.message : "Ошибка ротации кода";

      if (errorCode === "INVALID_QR") {
        // Backend invalidated the code or event was stopped on server
        setIsCodeExpired(true);
        setIsStopped(true);
        setRotationError("Код устарел или приём отметок был завершён на сервере.");
        // Mark as stopped in storage
        if (sessionRef.current) {
          const stoppedSession = { ...sessionRef.current, isStopped: true };
          sessionRef.current = stoppedSession;
          setSession(stoppedSession);
          storage.saveActiveSession(stoppedSession);
        }
      } else if (errorCode === "UNAUTHORIZED") {
        setRotationError("Сессия истекла. Требуется повторный вход.");
        onUnauthorizedRef.current?.();
      } else {
        // Transient network error: schedule retry in 3 seconds to avoid request flood
        setRotationError(errorMsg);
        nextRotationTimeRef.current = Date.now() + RETRY_INTERVAL_MS;
      }
    } finally {
      isRotatingRef.current = false;
      setIsRotating(false);
    }
  }, []); // Stable callback: no dynamic dependencies

  // Main countdown timer loop based on absolute timestamps (drift-free)
  useEffect(() => {
    // Do not run timer if session is missing, stopped or expired
    if (!session || session.isStopped || isStopped || isCodeExpired) return;

    // Initialize timer window if needed
    if (nextRotationTimeRef.current <= Date.now()) {
      nextRotationTimeRef.current = Date.now() + ROTATION_INTERVAL_MS;
    }

    const timer = setInterval(() => {
      const currentSession = sessionRef.current;
      if (!currentSession || currentSession.isStopped) {
        clearInterval(timer);
        return;
      }

      const now = Date.now();
      const remainingMs = Math.max(0, nextRotationTimeRef.current - now);
      const remainingSec = Math.max(0, Math.ceil(remainingMs / 1000));
      const progress = Math.max(0, Math.min(100, (remainingMs / ROTATION_INTERVAL_MS) * 100));

      setSecondsRemaining(remainingSec);
      setProgressPercent(progress);

      if (remainingMs <= 0 && !isRotatingRef.current) {
        // Advance nextRotationTimeRef immediately to prevent repeated triggers on subsequent ticks
        nextRotationTimeRef.current = Date.now() + ROTATION_INTERVAL_MS;
        performRotation();
      }
    }, TICK_INTERVAL_MS);

    return () => clearInterval(timer);
  }, [session?.eventId, isStopped, isCodeExpired, performRotation]);

  // Finish / Stop event
  const stopEvent = useCallback(async () => {
    const currentSession = sessionRef.current;
    const currentToken = tokenRef.current || storage.getToken();
    if (!currentSession || !currentToken) return;

    setIsStopping(true);
    try {
      await disableEventUpdating(currentSession.eventId, currentToken);

      const stoppedSession: ActiveSession = {
        ...currentSession,
        isStopped: true,
        lastUpdatedAt: new Date().toISOString(),
      };

      sessionRef.current = stoppedSession;
      setSession(stoppedSession);
      setIsStopped(true);
      storage.saveActiveSession(stoppedSession);

      // Archive session permanently in local storage
      const archiveItem: ArchivedSession = {
        id: `session-${currentSession.eventId}-${Date.now()}`,
        eventId: currentSession.eventId,
        eventName: currentSession.eventName,
        eventTime: currentSession.eventTime,
        teacherName,
        teacherEmail,
        createdAt: currentSession.createdAt,
        finishedAt: new Date().toISOString(),
        students: currentSession.notedStudents,
        totalCount: currentSession.notedStudents.length,
      };
      storage.archiveSession(archiveItem);
    } finally {
      setIsStopping(false);
    }
  }, [teacherName, teacherEmail]);

  // Delete event permanently from server and local storage
  const removeEvent = useCallback(async () => {
    const currentSession = sessionRef.current;
    const currentToken = tokenRef.current || storage.getToken();
    if (!currentSession || !currentToken) return;

    try {
      await deleteEvent(currentSession.eventId, currentToken);
      storage.clearActiveSession(currentSession.eventId);
      setSession(null);
    } catch (err) {
      console.error("Failed to delete event:", err);
      throw err;
    }
  }, []);

  // Force rotation manually if teacher clicks Refresh
  const forceRotate = useCallback(() => {
    nextRotationTimeRef.current = Date.now();
    performRotation();
  }, [performRotation]);

  return {
    session,
    secondsRemaining,
    progressPercent,
    isRotating,
    rotationError,
    isStopped,
    isStopping,
    isCodeExpired,
    newlyAddedCount,
    stopEvent,
    removeEvent,
    forceRotate,
  };
}
