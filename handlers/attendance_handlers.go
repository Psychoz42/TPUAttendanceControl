package handlers

import (
	"backend/utils"
	"time"
	"backend/apisim"
	"backend/database"

	"backend/code"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Scan(c *fiber.Ctx) error {
	qr := new(struct {
		QrCode   string `json:"qrCode"`
		Username string `json:"username"`
	})

	if err := c.BodyParser(qr); err != nil {
		return err
	}

	eventId, err := database.GetEventIdByCode(qr.QrCode)
	if err != nil {
		c.JSON(utils.ErrorResponse{Success: false, Code: "INVALID_QR", Message: "No such code in database"})
	}

	event, err := database.GetEventById(eventId)
	if err != nil {
		c.JSON(utils.ErrorResponse{Success: false, Code: "DB_ERROR", Message: "Event search error"})
	}

	notedStudents, _ := database.GetNotedStudents(eventId)
	database.UpdateNotedStudents(eventId, append(notedStudents, utils.Student{Login: qr.Username}))

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
		return err
	}

	array := strings.Split(eventData.Time, ":")
	hours, _ := strconv.Atoi(array[0])
	minutes, _ := strconv.Atoi(array[1])

	eventTime := time.Time{}.Add(time.Hour * time.Duration(hours)).Add(time.Minute * time.Duration(minutes))
	code := code.GenerateNewCode()
	err := database.CreateEvent(eventData.Name, eventData.Owner, "", code, time.Now().Add(time.Minute*5), eventTime)
	if err != nil {
		return err
	}

	eventId, _ := database.GetEventIdByCode(code)

	eventResponse := new(struct {
		EventId int    `json:"event_id"`
		Code    string `json:"code"`
	})
	eventResponse.EventId = eventId
	eventResponse.Code = code

	return c.JSON(eventResponse)
}

func GetOwnersEvents(c *fiber.Ctx) error {
	owner := new(struct {
		Owner string `json:"owner"`
	})

	if err := c.BodyParser(owner); err != nil {
		return err
	}

	events, err := database.GetEventsByOwner(owner.Owner)
	if err != nil {
		return err
	}

	eventsResponse := []utils.OwnersEventData{}
	for _, e := range events {
		eventsResponse = append(eventsResponse, utils.OwnersEventData{Id: e.Id, Name: e.Name})
	}

	return c.JSON(eventsResponse)
}

func DisableEventUpdating(c *fiber.Ctx) error {
	eventId := new(struct {
		EventId int `json:"event_id"`
	})

	if err := c.BodyParser(eventId); err != nil {
		return err
	}

	err := database.UpdateUpdateNeedById(eventId.EventId, false)
	if err != nil {
		return err
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
		return err
	}

	_, err := database.DeleteEventById(eventId.EventId)
	if err != nil {
		return err
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
		return err
	}

	eventId, _ := database.GetEventIdByCode(eventCode.Code)
	newCode := code.GenerateNewCode()

	database.UpdateEventCode(eventId, newCode)

	notedStudents, _ := database.GetNotedStudents(eventId)
	var notedStudentsStrings []string

	for _, student := range notedStudents {
		fullName, _ := apisim.GetFullNameByUsername(student.Login)
		notedStudentsStrings = append(notedStudentsStrings, fullName)
	}

	resp := new(struct{
		NewCode        string   `json:"new_code"`
		NotedStudents  []string `json:"noted_students`
	})

	resp.NewCode = newCode
	resp.NotedStudents = notedStudentsStrings

	return c.JSON(resp)
}