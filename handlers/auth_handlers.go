package handlers

import (
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	//"backend/database"
)


func Login(c *fiber.Ctx) error {
	loginData := new(utils.LoginRequest)
	err := c.BodyParser(loginData)
	if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid data"})
    }

	userData, err := utils.LoginCheck(loginData)
	if err != nil {
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
    }
	
	return c.JSON(userData)
}