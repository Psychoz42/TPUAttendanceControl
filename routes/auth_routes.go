package routes

import (
	"backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(authGroup fiber.Router) {
	authGroup.Post("/login", newLimiter(loginLimit), handlers.Login)
}
