import {
  CreateEventRequest,
  CreateEventResponse,
  EventDetails,
  UpdateCodeResponse,
  OwnerEventsResponse,
  SuccessOnlyResponse,
} from "@/types/api";
import { requestApi } from "./client";

export async function createEvent(
  data: CreateEventRequest,
  token: string
): Promise<CreateEventResponse> {
  return requestApi<CreateEventResponse>(
    "/attendance/createEvent",
    {
      event_name: data.event_name.trim(),
      event_time: data.event_time.trim(),
    },
    token
  );
}

export async function updateCode(
  code: string,
  token: string
): Promise<UpdateCodeResponse> {
  return requestApi<UpdateCodeResponse>(
    "/attendance/updateCode",
    {
      code: code.trim(),
    },
    token
  );
}

export async function getOwnerEvents(token: string): Promise<OwnerEventsResponse> {
  // Backend requires valid JSON body, so we pass empty object {}
  return requestApi<OwnerEventsResponse>("/attendance/getOwnerEvents", {}, token);
}

export async function getEvent(eventId: number, token: string): Promise<EventDetails> {
  return requestApi<EventDetails>(
    "/attendance/getEvent",
    {
      event_id: eventId,
    },
    token
  );
}

export async function disableEventUpdating(
  eventId: number,
  token: string
): Promise<SuccessOnlyResponse> {
  return requestApi<SuccessOnlyResponse>(
    "/attendance/disableEventUpdating",
    {
      event_id: eventId,
    },
    token
  );
}

export async function deleteEvent(
  eventId: number,
  token: string
): Promise<SuccessOnlyResponse> {
  return requestApi<SuccessOnlyResponse>(
    "/attendance/deleteEvent",
    {
      event_id: eventId,
    },
    token
  );
}
