"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { ActiveSession } from "@/types/attendance";
import { EventDetails } from "@/types/api";
import { updateCode, disableEventUpdating, deleteEvent, getEvent } from "@/lib/api/attendance";
import { storage } from "@/lib/storage";

const ROTATION_INTERVAL_MS = 10_000; // 10 seconds
const RETRY_INTERVAL_MS = 3_000; // 3 seconds on error retry
const TICK_INTERVAL_MS = 100; // 100ms update for smooth UI

function sessionFromServer(details: EventDetails): ActiveSession {
  return {
    eventId: details.event_id,
    eventName: details.event_name,
    eventTime: details.event_time,
    currentCode: details.code ?? "",
    notedStudents: details.noted_students,
    isStopped: !details.active,
    createdAt: details.created_at,
    lastUpdatedAt: new Date().toISOString(),
  };
}

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
  const [isSyncing, setIsSyncing] = useState<boolean>(!Number.isNaN(eventId));

  const teacherRef = useRef({ teacherName, teacherEmail });
  teacherRef.current = { teacherName, teacherEmail };

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

  const archive = useCallback((finished: ActiveSession) => {
    storage.archiveSession({
      id: `session-${finished.eventId}`,
      eventId: finished.eventId,
      eventName: finished.eventName,
      eventTime: finished.eventTime,
      teacherName: teacherRef.current.teacherName,
      teacherEmail: teacherRef.current.teacherEmail,
      createdAt: finished.createdAt,
      finishedAt: new Date().toISOString(),
      students: finished.notedStudents,
      totalCount: finished.notedStudents.length,
    });
  }, []);

  // The server is the source of truth. Returns true if this call discovered the event was closed
  const applyServerState = useCallback((details: EventDetails): boolean => {
    const previous = sessionRef.current;
    const synced = sessionFromServer(details);
    const justClosed = !!previous && !previous.isStopped && synced.isStopped;

    const diff = synced.notedStudents.length - prevStudentsCountRef.current;
    if (previous && diff > 0) {
      setNewlyAddedCount(diff);
    }
    prevStudentsCountRef.current = synced.notedStudents.length;

    sessionRef.current = synced;
    setSession(synced);
    setIsStopped(synced.isStopped);
    if (previous || !synced.isStopped) {
      storage.saveActiveSession(synced);
    }
    if (justClosed) {
      archive(synced);
    }
    return justClosed;
  }, [archive]);

  // Keep what is known locally when the server no longer has the event
  const markStoppedLocally = useCallback(() => {
    if (!sessionRef.current) return;
    const stopped: ActiveSession = {
      ...sessionRef.current,
      isStopped: true,
      lastUpdatedAt: new Date().toISOString(),
    };
    if (!sessionRef.current.isStopped) {
      archive(stopped);
    }
    sessionRef.current = stopped;
    setSession(stopped);
    setIsStopped(true);
    storage.saveActiveSession(stopped);
  }, [archive]);

  // INVALID_QR does not mean the event is over: the response to the previous rotation
  // may have been lost, or another tab may have rotated the code. Ask the server
  const recoverFromInvalidCode = useCallback(async (id: number, currentToken: string) => {
    try {
      const details = await getEvent(id, currentToken);
      const justClosed = applyServerState(details);
      if (details.active) {
        // The received code may already be expired, replace it right away
        nextRotationTimeRef.current = Date.now();
      } else if (justClosed) {
        setIsCodeExpired(true);
        setRotationError("Приём отметок был завершён на сервере.");
      }
    } catch (err: unknown) {
      const errorObj = err as { code?: string };
      if (errorObj?.code === "EVENT_NOT_FOUND") {
        setIsCodeExpired(true);
        setRotationError("Занятие удалено с сервера.");
        markStoppedLocally();
      } else {
        setRotationError(err instanceof Error ? err.message : "Ошибка ротации кода");
        nextRotationTimeRef.current = Date.now() + RETRY_INTERVAL_MS;
      }
    }
  }, [applyServerState, markStoppedLocally]);

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
        await recoverFromInvalidCode(currentSession.eventId, currentToken);
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
  }, [recoverFromInvalidCode]); // Stable callback: its dependencies never change

  // On open, take the state from the server: the page may have been reloaded after a long pause,
  // opened in another browser, or the event may have been closed in the meantime
  useEffect(() => {
    if (!token || Number.isNaN(eventId)) return;

    let cancelled = false;

    (async () => {
      try {
        const details = await getEvent(eventId, token);
        if (cancelled) return;
        if (applyServerState(details)) {
          setIsCodeExpired(true);
        }
        if (details.active) {
          // The saved code may already be expired, so show a fresh one right away
          performRotation();
        }
      } catch (err: unknown) {
        if (cancelled) return;
        const errorObj = err as { code?: string };
        if (errorObj?.code === "EVENT_NOT_FOUND") {
          markStoppedLocally();
        } else {
          console.warn("Could not load event from server:", err);
        }
      } finally {
        if (!cancelled) setIsSyncing(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [eventId, token, applyServerState, markStoppedLocally, performRotation]);

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

      try {
        // The local list is up to 10 seconds old, students who scanned since then are only on the server
        applyServerState(await getEvent(currentSession.eventId, currentToken));
      } catch (err) {
        console.warn("Could not load the final list from server:", err);
        markStoppedLocally();
      }
    } finally {
      setIsStopping(false);
    }
  }, [applyServerState, markStoppedLocally]);

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
    isSyncing,
    newlyAddedCount,
    stopEvent,
    removeEvent,
    forceRotate,
  };
}
