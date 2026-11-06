export interface ActiveSession {
  eventId: number;
  eventName: string;
  eventTime: string;
  currentCode: string;
  notedStudents: string[];
  isStopped: boolean;
  createdAt: string; // ISO string
  lastUpdatedAt: string; // ISO string
}

export interface ArchivedSession {
  id: string; // Unique ID (e.g. session-eventId-timestamp)
  eventId: number;
  eventName: string;
  eventTime: string;
  teacherName: string;
  teacherEmail: string;
  createdAt: string;
  finishedAt: string;
  students: string[];
  totalCount: number;
}
