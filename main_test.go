package main

import (
	"backend/apisim"
	"backend/auth"
	"backend/database"
	"backend/utils"
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

const (
	testOwner      = "integration-test-owner"
	testOtherOwner = "integration-test-other-owner"
	testJWTSecret  = "integration-test-secret-integration-test-secret"
)

func setupTestAppWithAuth(t *testing.T, authRequired bool) *fiber.App {
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

	if err := auth.Init(testJWTSecret, authRequired); err != nil {
		t.Fatal(err)
	}

	apisim.Users = nil
	for i := 1; i <= 50; i++ {
		apisim.Users = append(apisim.Users, apisim.User{
			Username: fmt.Sprintf("student%d", i),
			Password: "pass",
			Email:    fmt.Sprintf("student%d@tpu.ru", i),
			FullName: fmt.Sprintf("Студент Номер %d", i),
			Group:    "test",
			Role:     utils.RoleStudent,
		})
	}
	for _, username := range []string{testOwner, testOtherOwner} {
		apisim.Users = append(apisim.Users, apisim.User{
			Username: username,
			Password: "pass",
			Email:    username + "@tpu.ru",
			FullName: username,
			Group:    "-",
			Role:     utils.RoleTeacher,
		})
	}

	t.Cleanup(func() {
		database.GetPool().Exec(context.Background(), "DELETE FROM events WHERE event_owner = ANY($1)", []string{testOwner, testOtherOwner})
	})

	return newApp()
}

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()
	return setupTestAppWithAuth(t, false)
}

// получает токен через настоящий login, как это сделает клиент
func loginToken(t *testing.T, app *fiber.App, email string) string {
	t.Helper()

	_, result := post(t, app, "/api/v1/auth/login", map[string]string{"email": email, "password": "pass"})
	token, _ := result["token"].(string)
	if token == "" {
		t.Fatalf("login %s: no token in response %v", email, result)
	}
	return token
}

func postAs(t *testing.T, app *fiber.App, token string, path string, body any) (int, map[string]any) {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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

func post(t *testing.T, app *fiber.App, path string, body any) (int, map[string]any) {
	t.Helper()
	return postAs(t, app, "", path, body)
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

// льготный срок идёт секунды, поэтому в тестах его заканчиваем напрямую, а не ждём
func endPreviousCodeGrace(t *testing.T, eventId int) {
	t.Helper()

	_, err := database.GetPool().Exec(context.Background(),
		"UPDATE events SET previous_code_expiration_time = now() - interval '1 second' WHERE event_id = $1", eventId)
	if err != nil {
		t.Fatal(err)
	}
}

func rotateCode(t *testing.T, app *fiber.App, oldCode string) string {
	t.Helper()

	_, result := post(t, app, "/api/v1/attendance/updateCode", map[string]string{"code": oldCode})
	newCode, _ := result["new_code"].(string)
	if len(newCode) != 16 {
		t.Fatalf("unexpected updateCode response: %v", result)
	}
	return newCode
}

func expectScanOK(t *testing.T, app *fiber.App, qrCode string, username string) {
	t.Helper()

	_, result := post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": qrCode, "username": username})
	if result["success"] != true {
		t.Fatalf("scan by %s: unexpected response %v", username, result)
	}
}

func expectScanRejected(t *testing.T, app *fiber.App, qrCode string, username string) {
	t.Helper()

	_, result := post(t, app, "/api/v1/attendance/scan", map[string]string{"qrCode": qrCode, "username": username})
	expectError(t, result, "INVALID_QR")
}

func TestPreviousCodeGrace(t *testing.T) {
	app := setupTestApp(t)
	eventId, firstCode := createTestEvent(t, app)

	// сразу после смены отметиться можно и старым, и новым кодом
	secondCode := rotateCode(t, app, firstCode)
	expectScanOK(t, app, firstCode, "student1")
	expectScanOK(t, app, secondCode, "student2")
	if n := notedCount(t, eventId); n != 2 {
		t.Fatalf("noted %d students, want 2", n)
	}

	// льготный срок есть только у одного предыдущего кода
	thirdCode := rotateCode(t, app, secondCode)
	expectScanRejected(t, app, firstCode, "student3")
	expectScanOK(t, app, secondCode, "student3")

	// сменить код по предыдущему нельзя, льготный срок касается только scan
	_, result := post(t, app, "/api/v1/attendance/updateCode", map[string]string{"code": secondCode})
	expectError(t, result, "INVALID_QR")

	endPreviousCodeGrace(t, eventId)
	expectScanRejected(t, app, secondCode, "student4")
	expectScanOK(t, app, thirdCode, "student4")

	// закрытие занятия отключает и предыдущий код
	fourthCode := rotateCode(t, app, thirdCode)
	_, result = post(t, app, "/api/v1/attendance/disableEventUpdating", map[string]int{"event_id": eventId})
	if result["success"] != true {
		t.Fatalf("unexpected disableEventUpdating response: %v", result)
	}
	expectScanRejected(t, app, thirdCode, "student5")
	expectScanRejected(t, app, fourthCode, "student5")
}

func TestExpiredCodeDoesNotRevive(t *testing.T) {
	app := setupTestApp(t)
	eventId, oldCode := createTestEvent(t, app)

	_, err := database.GetPool().Exec(context.Background(),
		"UPDATE events SET expiration_time = now() - interval '1 second' WHERE event_id = $1", eventId)
	if err != nil {
		t.Fatal(err)
	}
	expectScanRejected(t, app, oldCode, "student1")

	// updateCode работает и с истёкшим кодом, но отметиться им после этого нельзя
	newCode := rotateCode(t, app, oldCode)
	expectScanRejected(t, app, oldCode, "student1")
	expectScanOK(t, app, newCode, "student1")
}

func TestCodeLifetime(t *testing.T) {
	app := setupTestApp(t)
	eventId, eventCode := createTestEvent(t, app)

	check := func(stage string) {
		t.Helper()

		var ok bool
		query := "SELECT expiration_time > now() AND expiration_time <= now() + interval '30 seconds' FROM events WHERE event_id = $1"
		if err := database.GetPool().QueryRow(context.Background(), query, eventId).Scan(&ok); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("%s: code must live no longer than 30 seconds", stage)
		}
	}

	check("createEvent")
	rotateCode(t, app, eventCode)
	check("updateCode")
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

	endPreviousCodeGrace(t, eventId)
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

func TestLoginLimiterPerAccount(t *testing.T) {
	app := setupTestApp(t)

	// соседи по NAT: у одного кончились попытки, другой всё ещё может войти
	body := map[string]string{"email": "student1@tpu.ru", "password": "wrong"}
	for i := 0; i < 11; i++ {
		post(t, app, "/api/v1/auth/login", body)
	}

	status, result := post(t, app, "/api/v1/auth/login", map[string]string{"email": "student2@tpu.ru", "password": "pass"})
	if status != 200 || result["username"] != "student2" {
		t.Fatalf("login of other account: status %d, response %v", status, result)
	}
}

func TestLoginReturnsTokenAndRole(t *testing.T) {
	app := setupTestApp(t)

	_, result := post(t, app, "/api/v1/auth/login", map[string]string{"email": testOwner + "@tpu.ru", "password": "pass"})
	if result["role"] != utils.RoleTeacher || result["token"] == "" || result["token"] == nil {
		t.Fatalf("unexpected login response: %v", result)
	}

	_, result = post(t, app, "/api/v1/auth/login", map[string]string{"email": "student1@tpu.ru", "password": "pass"})
	if result["role"] != utils.RoleStudent {
		t.Fatalf("unexpected login response: %v", result)
	}
}

func TestAuthRequired(t *testing.T) {
	app := setupTestAppWithAuth(t, true)

	paths := []string{"scan", "createEvent", "getOwnerEvents", "disableEventUpdating", "deleteEvent", "updateCode"}
	for _, path := range paths {
		status, result := post(t, app, "/api/v1/attendance/"+path, map[string]string{})
		if status != fiber.StatusUnauthorized {
			t.Fatalf("%s without token: status %d, want 401", path, status)
		}
		expectError(t, result, "UNAUTHORIZED")

		status, result = postAs(t, app, "not-a-token", "/api/v1/attendance/"+path, map[string]string{})
		if status != fiber.StatusUnauthorized {
			t.Fatalf("%s with bad token: status %d, want 401", path, status)
		}
		expectError(t, result, "UNAUTHORIZED")
	}
}

func TestRoles(t *testing.T) {
	app := setupTestAppWithAuth(t, true)
	studentToken := loginToken(t, app, "student1@tpu.ru")
	teacherToken := loginToken(t, app, testOwner+"@tpu.ru")

	for _, path := range []string{"createEvent", "getOwnerEvents", "disableEventUpdating", "deleteEvent", "updateCode"} {
		status, result := postAs(t, app, studentToken, "/api/v1/attendance/"+path, map[string]string{})
		if status != fiber.StatusForbidden {
			t.Fatalf("student on %s: status %d, want 403", path, status)
		}
		expectError(t, result, "FORBIDDEN")
	}

	status, result := postAs(t, app, teacherToken, "/api/v1/attendance/scan", map[string]string{"qrCode": "x"})
	if status != fiber.StatusForbidden {
		t.Fatalf("teacher on scan: status %d, want 403", status)
	}
	expectError(t, result, "FORBIDDEN")
}

func TestIdentityComesFromToken(t *testing.T) {
	app := setupTestAppWithAuth(t, true)
	studentToken := loginToken(t, app, "student1@tpu.ru")
	teacherToken := loginToken(t, app, testOwner+"@tpu.ru")

	// владелец из тела игнорируется
	_, result := postAs(t, app, teacherToken, "/api/v1/attendance/createEvent", map[string]string{
		"event_name": "Тестовая пара", "event_owner": testOtherOwner, "event_time": "10:30",
	})
	eventId, ok := result["event_id"].(float64)
	eventCode, _ := result["code"].(string)
	if !ok {
		t.Fatalf("unexpected createEvent response: %v", result)
	}

	event, err := database.GetEventById(context.Background(), int(eventId))
	if err != nil {
		t.Fatal(err)
	}
	if event.Owner != testOwner {
		t.Fatalf("event owner %q, want %q", event.Owner, testOwner)
	}

	// логин из тела игнорируется: отмечается владелец токена, поле username необязательно
	_, result = postAs(t, app, studentToken, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode, "username": "student2"})
	if result["success"] != true {
		t.Fatalf("unexpected scan response: %v", result)
	}
	_, result = postAs(t, app, studentToken, "/api/v1/attendance/scan", map[string]string{"qrCode": eventCode})
	if result["success"] != true {
		t.Fatalf("unexpected scan response: %v", result)
	}

	event, err = database.GetEventById(context.Background(), int(eventId))
	if err != nil {
		t.Fatal(err)
	}
	if len(event.NotedStudents) != 1 || event.NotedStudents[0].Login != "student1" {
		t.Fatalf("noted students %v, want only student1", event.NotedStudents)
	}
}

func TestOwnerEventsFromToken(t *testing.T) {
	app := setupTestAppWithAuth(t, true)
	ownerToken := loginToken(t, app, testOwner+"@tpu.ru")
	otherToken := loginToken(t, app, testOtherOwner+"@tpu.ru")

	_, result := postAs(t, app, ownerToken, "/api/v1/attendance/createEvent", map[string]string{"event_name": "x", "event_time": "10:30"})
	eventId, ok := result["event_id"].(float64)
	if !ok {
		t.Fatalf("unexpected createEvent response: %v", result)
	}

	// чужой токен с owner в теле не должен увидеть событие
	req := httptest.NewRequest("POST", "/api/v1/attendance/getOwnerEvents", bytes.NewReader([]byte(`{"owner":"`+testOwner+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+otherToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var events []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		if e["event_id"] == eventId {
			t.Fatalf("other teacher sees event %v", e)
		}
	}
}

func TestOtherTeacherCannotTouchEvent(t *testing.T) {
	app := setupTestAppWithAuth(t, true)
	ownerToken := loginToken(t, app, testOwner+"@tpu.ru")
	otherToken := loginToken(t, app, testOtherOwner+"@tpu.ru")

	_, result := postAs(t, app, ownerToken, "/api/v1/attendance/createEvent", map[string]string{"event_name": "x", "event_time": "10:30"})
	eventId, ok := result["event_id"].(float64)
	eventCode, _ := result["code"].(string)
	if !ok {
		t.Fatalf("unexpected createEvent response: %v", result)
	}
	idBody := map[string]int{"event_id": int(eventId)}

	_, result = postAs(t, app, otherToken, "/api/v1/attendance/deleteEvent", idBody)
	expectError(t, result, "EVENT_NOT_FOUND")

	_, result = postAs(t, app, otherToken, "/api/v1/attendance/disableEventUpdating", idBody)
	expectError(t, result, "EVENT_NOT_FOUND")

	// код знает любой студент, но без прав на событие новый код и список он не получит
	_, result = postAs(t, app, otherToken, "/api/v1/attendance/updateCode", map[string]string{"code": eventCode})
	expectError(t, result, "INVALID_QR")

	event, err := database.GetEventById(context.Background(), int(eventId))
	if err != nil {
		t.Fatalf("event must still exist: %v", err)
	}
	if !event.UpdateNeed || event.Code != eventCode {
		t.Fatalf("event was changed by other teacher: %+v", event)
	}

	_, result = postAs(t, app, ownerToken, "/api/v1/attendance/updateCode", map[string]string{"code": eventCode})
	if newCode, _ := result["new_code"].(string); len(newCode) != 16 {
		t.Fatalf("unexpected updateCode response: %v", result)
	}

	_, result = postAs(t, app, ownerToken, "/api/v1/attendance/disableEventUpdating", idBody)
	if result["success"] != true {
		t.Fatalf("unexpected disableEventUpdating response: %v", result)
	}

	_, result = postAs(t, app, ownerToken, "/api/v1/attendance/deleteEvent", idBody)
	if result["success"] != true {
		t.Fatalf("unexpected deleteEvent response: %v", result)
	}
}

func TestScanLimiterPerUser(t *testing.T) {
	app := setupTestAppWithAuth(t, true)
	firstToken := loginToken(t, app, "student1@tpu.ru")
	secondToken := loginToken(t, app, "student2@tpu.ru")

	body := map[string]string{"qrCode": "unknown"}
	for i := 0; i < 30; i++ {
		postAs(t, app, firstToken, "/api/v1/attendance/scan", body)
	}

	status, result := postAs(t, app, firstToken, "/api/v1/attendance/scan", body)
	if status != fiber.StatusTooManyRequests {
		t.Fatalf("31st scan: status %d, want 429", status)
	}
	expectError(t, result, "TOO_MANY_REQUESTS")

	// другой студент с того же IP лимит не делит
	status, result = postAs(t, app, secondToken, "/api/v1/attendance/scan", body)
	if status != 200 {
		t.Fatalf("scan of other student: status %d, want 200", status)
	}
	expectError(t, result, "INVALID_QR")
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
