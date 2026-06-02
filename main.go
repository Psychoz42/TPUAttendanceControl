package main

import (
	"backend/database"
	"backend/routes"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Start main.go")

	err := database.Connection()
	if err != nil {
		fmt.Println("failed to connect db")
	}
	defer database.GetPool().Close()

	//ids, _ := database.GetExpiredEventIds()
	//for _, id := range ids {
	//	database.DeleteEventById(id)
	//}
	//fmt.Println(ids)

	//e, _ := database.GetEventById(5)
	//fmt.Println(e)

	app := fiber.New()

	routes.SetupApiRoutes(app)

	app.Listen(":3000")

}
