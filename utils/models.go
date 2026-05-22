package utils

type QrScanResponse struct {
	Success    bool    `json:"success"`
	Subject    string  `json:"subject"`
	Time       string  `json:"time"`
}