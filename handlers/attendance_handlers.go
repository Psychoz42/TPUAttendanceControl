package handlers

import (
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	//"backend/database"
)

func Scan(c *fiber.Ctx) error {
	qr := new(struct{
		qrCode	string	`json:"qrCode"`
	})
	
	if err := c.BodyParser(qr); err != nil { return err }
	
	resp := new(utils.QrScanResponse)

	// Работа с бд, проверка кода

	return c.JSON(resp)
}