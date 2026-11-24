package handlers

import (
	"backend/apisim"
	"backend/auth"
	"backend/utils"
	"log"

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

	userData.Token, err = auth.IssueToken(auth.User{Username: userData.Username, Role: userData.Role})
	if err != nil {
		log.Printf("login: %v", err)
		return sendError(c, "INTERNAL_ERROR", "Internal server error")
	}

	return c.JSON(userData)
}
