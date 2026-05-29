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
        return c.JSON(utils.ErrorResponse{Success: false, Code: "INVALID_DATA_FORMAT", Message: "Invalid login data format"})
    }

	userData, err := utils.LoginCheck(loginData)
	if err != nil {
        return c.JSON(utils.ErrorResponse{Success: false, Code: "LOGIN_ERROR", Message: err.Error()})
    }
	
	return c.JSON(userData)
}