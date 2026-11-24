import { UserProfile } from "@/types/auth";
import { ActiveSession, ArchivedSession } from "@/types/attendance";

const TOKEN_KEY = "tpu_auth_token";
const USER_KEY = "tpu_auth_user";
const ACTIVE_SESSIONS_MAP_KEY = "tpu_active_sessions_map";
const LEGACY_ACTIVE_SESSION_KEY = "tpu_active_session";
const ARCHIVED_SESSIONS_KEY = "tpu_archived_sessions";

export const storage = {
  // Auth
  getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(TOKEN_KEY);
  },

  setToken(token: string): void {
    if (typeof window === "undefined") return;
    localStorage.setItem(TOKEN_KEY, token);
  },

  clearToken(): void {
    if (typeof window === "undefined") return;
    localStorage.removeItem(TOKEN_KEY);
  },

  getUser(): UserProfile | null {
    if (typeof window === "undefined") return null;
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as UserProfile;
    } catch {
      return null;
    }
  },

  setUser(user: UserProfile): void {
    if (typeof window === "undefined") return;
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },

  clearUser(): void {
    if (typeof window === "undefined") return;
    localStorage.removeItem(USER_KEY);
  },

  clearAuth(): void {
    this.clearToken();
    this.clearUser();
  },

  // Active Sessions Storage (Supports multiple simultaneous active sessions by eventId)
  getActiveSessionsMap(): Record<number, ActiveSession> {
    if (typeof window === "undefined") return {};
    const raw = localStorage.getItem(ACTIVE_SESSIONS_MAP_KEY);
    if (!raw) {
      // Check legacy single-session key as fallback
      const legacyRaw = localStorage.getItem(LEGACY_ACTIVE_SESSION_KEY);
      if (legacyRaw) {
        try {
          const legacy = JSON.parse(legacyRaw) as ActiveSession;
          if (legacy && legacy.eventId) {
            return { [legacy.eventId]: legacy };
          }
        } catch {
          // ignore
        }
      }
      return {};
    }
    try {
      return JSON.parse(raw) as Record<number, ActiveSession>;
    } catch {
      return {};
    }
  },

  getAllActiveSessions(): ActiveSession[] {
    const map = this.getActiveSessionsMap();
    return Object.values(map).sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    );
  },

  getActiveSession(eventId?: number): ActiveSession | null {
    if (typeof window === "undefined") return null;
    const map = this.getActiveSessionsMap();

    if (eventId !== undefined) {
      return map[eventId] || null;
    }

    // If eventId not specified, return the most recent active unstopped session
    const all = Object.values(map);
    const active = all.filter((s) => !s.isStopped);
    if (active.length > 0) {
      active.sort(
        (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      );
      return active[0];
    }

    return all[0] || null;
  },

  saveActiveSession(session: ActiveSession): void {
    if (typeof window === "undefined") return;
    const map = this.getActiveSessionsMap();
    map[session.eventId] = session;
    localStorage.setItem(ACTIVE_SESSIONS_MAP_KEY, JSON.stringify(map));
    // Keep legacy key updated with the most recent session
    localStorage.setItem(LEGACY_ACTIVE_SESSION_KEY, JSON.stringify(session));
  },

  removeActiveSession(eventId: number): void {
    if (typeof window === "undefined") return;
    const map = this.getActiveSessionsMap();
    delete map[eventId];
    localStorage.setItem(ACTIVE_SESSIONS_MAP_KEY, JSON.stringify(map));

    const legacy = this.getActiveSession();
    if (legacy && legacy.eventId === eventId) {
      localStorage.removeItem(LEGACY_ACTIVE_SESSION_KEY);
    }
  },

  clearActiveSession(eventId?: number): void {
    if (eventId !== undefined) {
      this.removeActiveSession(eventId);
    } else {
      if (typeof window === "undefined") return;
      localStorage.removeItem(ACTIVE_SESSIONS_MAP_KEY);
      localStorage.removeItem(LEGACY_ACTIVE_SESSION_KEY);
    }
  },

  // Archived Sessions
  getArchivedSessions(): ArchivedSession[] {
    if (typeof window === "undefined") return [];
    const raw = localStorage.getItem(ARCHIVED_SESSIONS_KEY);
    if (!raw) return [];
    try {
      return JSON.parse(raw) as ArchivedSession[];
    } catch {
      return [];
    }
  },

  archiveSession(session: ArchivedSession): void {
    if (typeof window === "undefined") return;
    const current = this.getArchivedSessions();
    const filtered = current.filter((s) => s.id !== session.id);
    localStorage.setItem(
      ARCHIVED_SESSIONS_KEY,
      JSON.stringify([session, ...filtered])
    );
  },

  deleteArchivedSession(id: string): void {
    if (typeof window === "undefined") return;
    const current = this.getArchivedSessions();
    localStorage.setItem(
      ARCHIVED_SESSIONS_KEY,
      JSON.stringify(current.filter((s) => s.id !== id))
    );
  },
};
