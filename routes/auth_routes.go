package routes

import (
	"github.com/gofiber/fiber/v2"
	"backend/handlers"
)

func SetupAuthRoutes(authGroup fiber.Router) {
	authGroup.Post("/login", handlers.Login)
}