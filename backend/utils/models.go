package utils

import "time"

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
)

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
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Group    string `json:"group"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type Student struct {
	Login string `json:"login"`
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
	CreatedAt      time.Time `json:"created_at"`
}

type OwnersEventData struct {
	Id         int       `json:"event_id"`
	Name       string    `json:"event_name"`
	Time       string    `json:"event_time"`
	CreatedAt  time.Time `json:"created_at"`
	Active     bool      `json:"active"`
	NotedCount int       `json:"noted_count"`
}

// EventDetails отдаётся только владельцу события. Код есть только у открытого события
type EventDetails struct {
	Id            int       `json:"event_id"`
	Name          string    `json:"event_name"`
	Time          string    `json:"event_time"`
	CreatedAt     time.Time `json:"created_at"`
	Active        bool      `json:"active"`
	Code          string    `json:"code,omitempty"`
	NotedStudents []string  `json:"noted_students"`
}
