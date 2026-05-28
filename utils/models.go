package utils

type QrScanResponse struct {
	Success    bool    `json:"success"`
	Subject    string  `json:"subject"`
	Time       string  `json:"time"`
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