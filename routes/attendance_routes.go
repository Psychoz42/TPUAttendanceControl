package routes

import (
	"backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupAttendanceRoutes(attendanceGroup fiber.Router) {
	attendanceGroup.Post("/scan", handlers.Scan)
	attendanceGroup.Post("/createEvent", handlers.CreateEvent)
	attendanceGroup.Post("/getOwnerEvents", handlers.GetOwnersEvents)
	attendanceGroup.Post("/disableEventUpdating", handlers.DisableEventUpdating)
	attendanceGroup.Post("/deleteEvent", handlers.DeleteEvent)
}
