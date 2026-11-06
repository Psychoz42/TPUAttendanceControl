package main

import (
	"backend/apisim"
	"backend/database"
	"backend/routes"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	cleanupInterval = 10 * time.Minute
	// запас на случай, если клиент преподавателя ненадолго пропал из сети
	cleanupGrace = 10 * time.Minute
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	fmt.Println("Start main.go")

	databaseUrl := getenv("DATABASE_URL", "postgresql://postgres:2367@localhost:5431/postgres")
	listenAddr := getenv("LISTEN_ADDR", ":80")
	usersFile := getenv("USERS_FILE", "users.txt")

	var err error
	apisim.Users, err = apisim.LoadUsers(usersFile)
	if err != nil {
		log.Fatalf("не удалось загрузить пользователей: %v", err)
	}

	err = database.Connection(databaseUrl)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go startCleanup(ctx)

	app := newApp()

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down")
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	if err := app.Listen(listenAddr); err != nil {
		log.Fatal(err)
	}

	database.GetPool().Close()
}

func newApp() *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 64 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var fe *fiber.Error
			if errors.As(err, &fe) {
				// 404 неизвестного маршрута, 413 и т. п.: статус сохраняем
				return c.Status(fe.Code).JSON(utils.ErrorResponse{Success: false, Code: "HTTP_ERROR", Message: fe.Message})
			}
			// детали только в лог, клиенту не отдаём
			log.Printf("%s %s: %v", c.Method(), c.Path(), err)
			return c.JSON(utils.ErrorResponse{Success: false, Code: "INTERNAL_ERROR", Message: "Internal server error"})
		},
	})

	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // для теста разрешаем всем
	}))
	app.Static("/", "./dist")

	routes.SetupApiRoutes(app)

	return app
}

// удаляет просроченные события сразу при старте и дальше раз в cleanupInterval
func startCleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		deleteExpiredEvents(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func deleteExpiredEvents(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ids, err := database.DeleteExpiredEvents(ctx, cleanupGrace)
	if err != nil {
		log.Printf("cleanup: %v", err)
		return
	}
	if len(ids) > 0 {
		fmt.Printf("Deleted %d expired events: %v\n", len(ids), ids)
	}
}
