package main

import (
	"backend/routes"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Start main.go")

	app := fiber.New()

	routes.SetupApiRoutes(app)

	app.Listen(":3000")

}
