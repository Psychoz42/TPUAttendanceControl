package utils

type QrScanResponse struct {
	Success bool   `json:"success"`
	Subject string `json:"subject"`
	Time    string `json:"time"`
}

type ErrorResponse struct {
	Success    bool    `json:"success"`
	Code       string  `json:"code"`
	Message    string  `json:"message"`
}

type LoginRequest struct {
	email      string	`json:"email"`
	password   string   `json:"password"`
}

type LoginResponse struct {
	
}

type Student struct {
	Id       int    `json:"id"`
	FullName string `json:"full_name"`
}

type EventData struct {
	EventName  string `json:"event_name"`
	EventOwner string `json:"event_owner"`
	EventInfo  string `json:"event_info"`
}