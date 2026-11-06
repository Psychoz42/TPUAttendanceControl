package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var h = NewHTTPClient()

// адрес реального API ТПУ, пока вместо него используется apisim
var apiServerUrl = os.Getenv("TPU_API_URL")

func LoginCheck(login *LoginRequest) (*LoginResponse, error) {
	if apiServerUrl == "" {
		return &LoginResponse{}, fmt.Errorf("TPU_API_URL is not set")
	}

	url := fmt.Sprintf("%s/login", apiServerUrl)
	body, err := json.Marshal(login)
	if err != nil {
		return &LoginResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return &LoginResponse{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := h.Do(req)
	if err != nil {
		return &LoginResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &LoginResponse{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return &LoginResponse{}, fmt.Errorf("decode response: %w", err)
	}

	return &data, nil
}
