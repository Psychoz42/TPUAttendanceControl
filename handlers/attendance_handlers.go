package handlers

import (
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	"backend/database"
)

func Scan(c *fiber.Ctx) error {
	qr := new(struct{
		QrCode	string	`json:"qrCode"`
	})
	
	if err := c.BodyParser(qr); err != nil { return err }

	eventId, err := database.GetEventIdByCode(qr.QrCode)
	if err != nil {
		c.JSON(utils.ErrorResponse{Success: false, Code: "INVALID_QR", Message: "No such code in database"})
	}

	event, err := database.GetEventById(eventId)
	if err != nil {
		c.JSON(utils.ErrorResponse{Success: false, Code: "DB_ERROR", Message: "Event search error"})
	}
	
	resp := utils.QrScanResponse{Success: true, Subject: event.Name, Time: event.Time.Format("15:04")}

	return c.JSON(resp)
}