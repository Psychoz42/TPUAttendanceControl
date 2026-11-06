package routes

import (
	"backend/auth"
	"backend/handlers"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
)

func SetupAttendanceRoutes(attendanceGroup fiber.Router) {
	teacherOnly := auth.RequireRole(utils.RoleTeacher)

	attendanceGroup.Use(auth.Authenticate)

	attendanceGroup.Post("/scan", auth.RequireRole(utils.RoleStudent), newLimiter(scanLimit, userOrIPKey), handlers.Scan)
	attendanceGroup.Post("/createEvent", teacherOnly, handlers.CreateEvent)
	attendanceGroup.Post("/getOwnerEvents", teacherOnly, handlers.GetOwnersEvents)
	attendanceGroup.Post("/disableEventUpdating", teacherOnly, handlers.DisableEventUpdating)
	attendanceGroup.Post("/deleteEvent", teacherOnly, handlers.DeleteEvent)
	attendanceGroup.Post("/updateCode", teacherOnly, handlers.UpdateCode)
}
