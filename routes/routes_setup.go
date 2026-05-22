package routes

import "github.com/gofiber/fiber/v2"

func SetupApiRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	SetupAttendanceRoutes(api.Group("/attendance"))
}
