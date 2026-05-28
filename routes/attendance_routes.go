package routes

import (
	"github.com/gofiber/fiber/v2"
	"backend/handlers"
)

func SetupAttendanceRoutes(attendanceGroup fiber.Router) {
	attendanceGroup.Post("/scan", handlers.Scan)
}