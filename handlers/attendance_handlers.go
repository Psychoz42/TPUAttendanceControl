package handlers

import (
	"backend/apisim"
	"backend/auth"
	"backend/database"
	"backend/utils"
	"context"
	"errors"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
)

// сколько живёт QR-код после создания события и после каждого updateCode. Сайт меняет код
// примерно каждые 10 секунд, чтобы его не успевали переслать. Этот срок ограничивает пересылку,
// если сайт перестал обновлять код: закрыли вкладку или пропала сеть
const codeTTL = 30 * time.Second

// сколько предыдущий код ещё принимается в scan: код мог смениться, пока запрос студента
// шёл к серверу, а экран сайта показывает старый код, пока ждёт ответа updateCode
const previousCodeGrace = 3 * time.Second

const dbTimeout = 5 * time.Second

const maxEventNameLength = 200

func dbContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), dbTimeout)
}

func sendError(c *fiber.Ctx, code string, message string) error {
	return c.JSON(utils.ErrorResponse{Success: false, Code: code, Message: message})
}

// час одной или двумя цифрами, минуты строго двумя: "9:05" подходит, "9:5" нет
func parseEventTime(s string) (time.Time, error) {
	return time.Parse("15:04", s)
}

// логин из токена важнее логина из тела запроса. Тело используется,
// только пока AUTH_REQUIRED выключен и клиент ещё не присылает токен
func usernameOrBody(c *fiber.Ctx, fromBody string) string {
	if username := auth.CurrentUsername(c); username != "" {
		return username
	}
	return fromBody
}

func Scan(c *fiber.Ctx) error {
	qr := new(struct {
		QrCode   string `json:"qrCode"`
		Username string `json:"username"`
	})

	if err := c.BodyParser(qr); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid scan data format")
	}

	username := usernameOrBody(c, qr.Username)

	if qr.QrCode == "" || username == "" {
		return sendError(c, "INVALID_DATA_FORMAT", "qrCode and username are required")
	}

	if _, err := apisim.GetFullNameByUsername(username); err != nil {
		return sendError(c, "UNKNOWN_USER", "No such student")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	event, err := database.GetActiveEventByCode(ctx, qr.QrCode)
	if errors.Is(err, database.ErrNotFound) {
		return sendError(c, "INVALID_QR", "QR code is invalid or expired")
	}
	if err != nil {
		log.Printf("scan: %v", err)
		return sendError(c, "DB_ERROR", "Event search error")
	}

	// повторный скан тем же студентом ничего не меняет и тоже считается успешным
	if _, err = database.AddNotedStudent(ctx, event.Id, username); err != nil {
		log.Printf("scan: %v", err)
		return sendError(c, "DB_ERROR", "Failed to note student")
	}

	resp := utils.QrScanResponse{Success: true, Subject: event.Name, Time: event.Time.Format("15:04")}

	return c.JSON(resp)
}

func CreateEvent(c *fiber.Ctx) error {
	eventData := new(struct {
		Name  string `json:"event_name"`
		Owner string `json:"event_owner"`
		Time  string `json:"event_time"`
	})

	if err := c.BodyParser(eventData); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid event data format")
	}

	name := strings.TrimSpace(eventData.Name)
	if name == "" || utf8.RuneCountInString(name) > maxEventNameLength {
		return sendError(c, "INVALID_DATA_FORMAT", "Event name must be from 1 to 200 characters")
	}

	owner := usernameOrBody(c, eventData.Owner)
	if strings.TrimSpace(owner) == "" {
		return sendError(c, "INVALID_DATA_FORMAT", "Event owner is required")
	}

	eventTime, err := parseEventTime(eventData.Time)
	if err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Event time must be in HH:MM format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	eventId, eventCode, err := database.CreateEvent(ctx, name, owner, "", eventTime, codeTTL)
	if err != nil {
		log.Printf("createEvent: %v", err)
		return sendError(c, "DB_ERROR", "Failed to create event")
	}

	eventResponse := new(struct {
		EventId int    `json:"event_id"`
		Code    string `json:"code"`
	})
	eventResponse.EventId = eventId
	eventResponse.Code = eventCode

	return c.JSON(eventResponse)
}

func GetOwnersEvents(c *fiber.Ctx) error {
	owner := new(struct {
		Owner string `json:"owner"`
	})

	if err := c.BodyParser(owner); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid owner data format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	events, err := database.GetEventsByOwner(ctx, usernameOrBody(c, owner.Owner))
	if err != nil {
		log.Printf("getOwnerEvents: %v", err)
		return sendError(c, "DB_ERROR", "Failed to get events")
	}

	eventsResponse := []utils.OwnersEventData{}
	for _, e := range events {
		eventsResponse = append(eventsResponse, utils.OwnersEventData{
			Id:         e.Id,
			Name:       e.Name,
			Time:       e.Time.Format("15:04"),
			CreatedAt:  e.CreatedAt,
			Active:     e.UpdateNeed,
			NotedCount: len(e.NotedStudents),
		})
	}

	return c.JSON(eventsResponse)
}

// GetEvent нужен, чтобы вернуться к занятию после обновления страницы или потери ответа
// updateCode и чтобы получить итоговый список после закрытия
func GetEvent(c *fiber.Ctx) error {
	eventId := new(struct {
		EventId int `json:"event_id"`
	})

	if err := c.BodyParser(eventId); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid event data format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	event, err := database.GetEventById(ctx, eventId.EventId, auth.CurrentUsername(c))
	if errors.Is(err, database.ErrNotFound) {
		return sendError(c, "EVENT_NOT_FOUND", "No such event")
	}
	if err != nil {
		log.Printf("getEvent: %v", err)
		return sendError(c, "DB_ERROR", "Failed to get event")
	}

	resp := utils.EventDetails{
		Id:            event.Id,
		Name:          event.Name,
		Time:          event.Time.Format("15:04"),
		CreatedAt:     event.CreatedAt,
		Active:        event.UpdateNeed,
		NotedStudents: []string{},
	}
	if event.UpdateNeed {
		resp.Code = event.Code
	}
	if names := studentFullNames(event.NotedStudents); names != nil {
		resp.NotedStudents = names
	}

	return c.JSON(resp)
}

// для неизвестного логина отдаёт сам логин. Для пустого списка возвращает nil
func studentFullNames(students []utils.Student) []string {
	var names []string

	for _, student := range students {
		fullName, err := apisim.GetFullNameByUsername(student.Login)
		if err != nil {
			fullName = student.Login
		}
		names = append(names, fullName)
	}

	return names
}

func DisableEventUpdating(c *fiber.Ctx) error {
	eventId := new(struct {
		EventId int `json:"event_id"`
	})

	if err := c.BodyParser(eventId); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid event data format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	err := database.UpdateUpdateNeedById(ctx, eventId.EventId, false, auth.CurrentUsername(c))
	if errors.Is(err, database.ErrNotFound) {
		return sendError(c, "EVENT_NOT_FOUND", "No such event")
	}
	if err != nil {
		log.Printf("disableEventUpdating: %v", err)
		return sendError(c, "DB_ERROR", "Failed to disable event updating")
	}

	resp := new(struct {
		Success bool `json:"success"`
	})
	resp.Success = true
	return c.JSON(resp)
}

func DeleteEvent(c *fiber.Ctx) error {
	eventId := new(struct {
		EventId int `json:"event_id"`
	})

	if err := c.BodyParser(eventId); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid event data format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	_, err := database.DeleteEventById(ctx, eventId.EventId, auth.CurrentUsername(c))
	if errors.Is(err, database.ErrNotFound) {
		return sendError(c, "EVENT_NOT_FOUND", "No such event")
	}
	if err != nil {
		log.Printf("deleteEvent: %v", err)
		return sendError(c, "DB_ERROR", "Failed to delete event")
	}

	resp := new(struct {
		Success bool `json:"success"`
	})
	resp.Success = true
	return c.JSON(resp)
}

func UpdateCode(c *fiber.Ctx) error {
	eventCode := new(struct {
		Code string `json:"code"`
	})

	if err := c.BodyParser(eventCode); err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid code data format")
	}

	ctx, cancel := dbContext(c)
	defer cancel()

	newCode, notedStudents, err := database.RotateEventCode(ctx, eventCode.Code, codeTTL, previousCodeGrace, auth.CurrentUsername(c))
	if errors.Is(err, database.ErrNotFound) {
		return sendError(c, "INVALID_QR", "Code is invalid or event updating is disabled")
	}
	if err != nil {
		log.Printf("updateCode: %v", err)
		return sendError(c, "DB_ERROR", "Failed to update code")
	}

	resp := new(struct {
		NewCode       string   `json:"new_code"`
		NotedStudents []string `json:"noted_students"`
	})

	resp.NewCode = newCode
	// пустой список уходит как null, фронт пока получает его в таком виде
	resp.NotedStudents = studentFullNames(notedStudents)

	return c.JSON(resp)
}
