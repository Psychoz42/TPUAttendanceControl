package main

import (
	"backend/apisim"
	"backend/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Интеграционные тесты. Нужна БД с накатанными миграциями, например из docker compose:
// TEST_DATABASE_URL=postgresql://postgres:2367@localhost:5431/postgres go test ./...

const testOwner = "integration-test-owner"

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	if database.GetPool() == nil {
		if err := database.Connection(url); err != nil {
			t.Fatal(err)
		}
	}

	apisim.Users = nil
	for i := 1; i <= 50; i++ {
		apisim.Users = append(apisim.Users, apisim.User{
			Username: fmt.Sprintf("student%d", i),
			Password: "pass",
			Email:    fmt.Sprintf("student%d@tpu.ru", i),
			FullName: fmt.Sprintf("Студент Номер %d", i),
			Group:    "test",
		})
	}

	t.Cleanup(func() {
		database.GetPool().Exec(context.Background(), "DELETE FROM events WHERE event_owner = $1", testOwner)
	})

	return newApp()
}

func post(t *testing.T, app *fiber.App, path string, body any) (int, map[string]any) {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("%s: decode response: %v", path, err)
	}

	return resp.StatusCode, result
}

func expectError(t *testing.T, result map[string]any, code string) {
	t.Helper()
	if result["success"] != false || result["code"] != code {
		t.Fatalf("expected error %s, got %v", code, result)
	}
}

func createTestEvent(t *testing.T, app *fiber.App) (int, string) {
	t.Helper()

	_, result := post(t, app, "/api/v1/attendance/createEvent", map[string]string{
		"event_name": "Тестовая пара", "event_owner": testOwner, "event_time": "10:30",
	})

	id, ok := result["event_id"].(float64)
	eventCode, _ := result["code"].(string)
	if !ok || len(eventCode) != 16 {
		t.Fatalf("unexpected createEvent response: %v", result)
	}

	return int(id), eventCode
}

func notedCount(t *testing.T, eventId int) int {
	t.Helper()

	event, err := database.GetEventById(context.Background(), eventId)
	if err != nil {
		t.Fatal(err)
	}
	return len(event.NotedStudents)
}

func TestCreateEventValidation(t *testing.T) {
	app := setupTestApp(t)

	bodies := []map[string]string{
		{"event_name": "x", "event_owner": testOwner, "event_time": "10"},
		{"event_name": "x", "event_owner": testOwner, "event_time": ""},
		{"event_name": "  ", "event_owner": testOwner, "event_time": "10:30"},
		{"event_name": "x", "event_owner": "", "event_time": "10:30"},
	}
	for _, body := range bodies {
		_, result := post(t, app, "/api/v1/attendance/createEvent", body)
		expectError(t, result, "INVALID_DATA_FORMAT")
	}
}

func TestScanFlow(t *testing.T) {
	app := setupTestApp(t)
	eventId, eventCode := createTestEvent(t, app)

	_, result := post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": "unknown", "username": "student1"})
	expectError(t, result, "INVALID_QR")

	_, result = post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode, "username": "nobody"})
	expectError(t, result, "UNKNOWN_USER")

	_, result = post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode})
	expectError(t, result, "INVALID_DATA_FORMAT")

	for i := 0; i < 2; i++ {
		_, result = post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode, "username": "student1"})
		if result["success"] != true || result["subject"] != "Тестовая пара" || result["time"] != "10:30" {
			t.Fatalf("unexpected scan response: %v", result)
		}
	}
	if n := notedCount(t, eventId); n != 1 {
		t.Fatalf("after two scans of one student noted %d, want 1", n)
	}

	_, result = post(t, app, "/api/v1/attendance/updateCode", map[string]string{"code": eventCode})
	newCode, _ := result["new_code"].(string)
	students, _ := result["noted_students"].([]any)
	if len(newCode) != 16 || newCode == eventCode || len(students) != 1 || students[0] != "Студент Номер 1" {
		t.Fatalf("unexpected updateCode response: %v", result)
	}

	_, result = post(t, app, "/api/v1/attendance/updateCode", map[string]string{"code": eventCode})
	expectError(t, result, "INVALID_QR")

	_, result = post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode, "username": "student2"})
	expectError(t, result, "INVALID_QR")

	_, result = post(t, app, "/api/v1/attendance/disableEventUpdating", map[string]int{"event_id": eventId})
	if result["success"] != true {
		t.Fatalf("unexpected disableEventUpdating response: %v", result)
	}

	_, result = post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": newCode, "username": "student2"})
	expectError(t, result, "INVALID_QR")

	_, result = post(t, app, "/api/v1/attendance/updateCode", map[string]string{"code": newCode})
	expectError(t, result, "INVALID_QR")

	_, result = post(t, app, "/api/v1/attendance/deleteEvent", map[string]int{"event_id": eventId})
	if result["success"] != true {
		t.Fatalf("unexpected deleteEvent response: %v", result)
	}

	_, result = post(t, app, "/api/v1/attendance/deleteEvent", map[string]int{"event_id": eventId})
	expectError(t, result, "EVENT_NOT_FOUND")

	_, result = post(t, app, "/api/v1/attendance/disableEventUpdating", map[string]int{"event_id": eventId})
	expectError(t, result, "EVENT_NOT_FOUND")
}

func TestConcurrentNotedStudents(t *testing.T) {
	app := setupTestApp(t)
	eventId, _ := createTestEvent(t, app)

	// напрямую через database, через HTTP упрёмся в limiter
	var wg sync.WaitGroup
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(login string) {
			defer wg.Done()
			// каждого студента отмечаем дважды, дублей быть не должно
			for j := 0; j < 2; j++ {
				if _, err := database.AddNotedStudent(context.Background(), eventId, login); err != nil {
					t.Error(err)
				}
			}
		}(fmt.Sprintf("student%d", i))
	}
	wg.Wait()

	if n := notedCount(t, eventId); n != 50 {
		t.Fatalf("noted %d students, want 50", n)
	}
}

func TestGetOwnerEvents(t *testing.T) {
	app := setupTestApp(t)
	eventId, _ := createTestEvent(t, app)

	req := httptest.NewRequest("POST", "/api/v1/attendance/getOwnerEvents", bytes.NewReader([]byte(`{"owner":"`+testOwner+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var events []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || int(events[0]["event_id"].(float64)) != eventId {
		t.Fatalf("unexpected getOwnerEvents response: %v", events)
	}
}

func TestDeleteExpiredEvents(t *testing.T) {
	setupTestApp(t)
	ctx := context.Background()

	var oldId, freshId int
	query := `
			INSERT INTO events (event_name, event_owner, event_code, expiration_time, event_time)
			VALUES ('x', $1, $2, now() - $3::interval, '10:00')
			RETURNING event_id
			`
	if err := database.GetPool().QueryRow(ctx, query, testOwner, "expired-test-1", time.Hour).Scan(&oldId); err != nil {
		t.Fatal(err)
	}
	if err := database.GetPool().QueryRow(ctx, query, testOwner, "expired-test-2", time.Minute).Scan(&freshId); err != nil {
		t.Fatal(err)
	}

	ids, err := database.DeleteExpiredEvents(ctx, cleanupGrace)
	if err != nil {
		t.Fatal(err)
	}

	deleted := map[int]bool{}
	for _, id := range ids {
		deleted[id] = true
	}
	if !deleted[oldId] || deleted[freshId] {
		t.Fatalf("deleted %v, want %d deleted and %d kept", ids, oldId, freshId)
	}
}

func TestLoginLimiter(t *testing.T) {
	app := setupTestApp(t)

	body := map[string]string{"email": "student1@tpu.ru", "password": "pass"}
	for i := 0; i < 10; i++ {
		status, result := post(t, app, "/api/v1/auth/login", body)
		if status != 200 || result["username"] != "student1" {
			t.Fatalf("login %d: status %d, response %v", i+1, status, result)
		}
	}

	status, result := post(t, app, "/api/v1/auth/login", body)
	if status != fiber.StatusTooManyRequests {
		t.Fatalf("11th login: status %d, want 429", status)
	}
	expectError(t, result, "TOO_MANY_REQUESTS")
}

func TestErrorHandler(t *testing.T) {
	app := setupTestApp(t)
	app.Post("/test/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})
	app.Post("/test/error", func(c *fiber.Ctx) error {
		return fmt.Errorf("secret db details")
	})

	for _, path := range []string{"/test/panic", "/test/error"} {
		status, result := post(t, app, path, nil)
		if status != 200 || result["message"] != "Internal server error" {
			t.Fatalf("%s: status %d, response %v", path, status, result)
		}
		expectError(t, result, "INTERNAL_ERROR")
	}

	status, result := post(t, app, "/api/v1/unknown", nil)
	if status != fiber.StatusNotFound {
		t.Fatalf("unknown route: status %d, want 404", status)
	}
	expectError(t, result, "HTTP_ERROR")
}
