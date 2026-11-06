package handlers

import (
	"backend/apisim"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	loginData := new(utils.LoginRequest)
	err := c.BodyParser(loginData)
	if err != nil {
		return sendError(c, "INVALID_DATA_FORMAT", "Invalid login data format")
	}

	userData, err := apisim.LoginCheckSim(loginData) //utils.LoginCheck(loginData)
	if err != nil {
		return sendError(c, "LOGIN_ERROR", err.Error())
	}

	return c.JSON(userData)
}
