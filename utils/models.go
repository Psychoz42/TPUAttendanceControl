package utils

import "time"

type QrScanResponse struct {
	Success bool   `json:"success"`
	Subject string `json:"subject"`
	Time    string `json:"time"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type LoginRequest struct {
	email    string `json:"email"`
	password string `json:"password"`
}

type LoginResponse struct {
}

type Student struct {
	Login string `json:"login"`
}

type EventData struct {
	EventName  string `json:"event_name"`
	EventOwner string `json:"event_owner"`
	EventInfo  string `json:"event_info"`
}

type Event struct {
	Id             int       `json:"event_id"`
	Name           string    `json:"event_name"`
	Owner          string    `json:"event_owner"`
	Info           string    `json:"event_info"`
	Code           string    `json:"event_code"`
	Time           time.Time `json:"event_time"`
	ExpirationTime time.Time `json:"expiration_time"`
	UpdateNeed     bool      `json:"update_need"`
	NotedStudents  []Student `json:"noted_students"`
}
