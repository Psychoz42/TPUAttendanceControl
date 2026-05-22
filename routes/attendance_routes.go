package routes

import (
	"github.com/gofiber/fiber/v2"
	"backend/handlers"
)

func SetupAttendanceRoutes(attendance fiber.Router) {
	attendance.Post("/scan", handlers.Scan)
}