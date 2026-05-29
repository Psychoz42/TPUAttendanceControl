package utils

import (
	"fmt"
	"net/http"
	"encoding/json"
	"bytes"
)

var h = NewHTTPClient()
var apiServerUrl = "https://psychoz.site/TPUapi"

func LoginCheck(login *LoginRequest) (*LoginResponse, error) {
	url := fmt.Sprintf("%s/login", apiServerUrl)
	body, _ := json.Marshal(login)

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