package main

import (
	"backend/apisim"
	"backend/auth"
	"backend/database"
	"backend/routes"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	cleanupInterval = 10 * time.Minute
	// запас на случай, если клиент преподавателя ненадолго пропал из сети:
	// пока событие не закрыто, updateCode может продолжить его
	closeGrace = 10 * time.Minute
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
	// задавать, только если сервер доступен исключительно через прокси, иначе клиент подделает свой IP
	proxyHeader := getenv("PROXY_HEADER", "")

	authRequired, err := strconv.ParseBool(getenv("AUTH_REQUIRED", "false"))
	if err != nil {
		log.Fatalf("неверное значение AUTH_REQUIRED: %v", err)
	}

	retentionDays, err := strconv.Atoi(getenv("EVENT_RETENTION_DAYS", "0"))
	if err != nil || retentionDays < 0 {
		log.Fatalf("неверное значение EVENT_RETENTION_DAYS: %q", os.Getenv("EVENT_RETENTION_DAYS"))
	}
	retention := time.Duration(retentionDays) * 24 * time.Hour

	err = auth.Init(os.Getenv("JWT_SECRET"), authRequired)
	if err != nil {
		log.Fatalf("не удалось настроить авторизацию: %v", err)
	}
	if !authRequired {
		log.Println("ВНИМАНИЕ: AUTH_REQUIRED выключен, запросы без токена принимаются без проверки прав")
	}

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

	go startCleanup(ctx, retention)

	app := newApp(proxyHeader)

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

func newApp(proxyHeader string) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:   64 * 1024,
		ProxyHeader: proxyHeader,
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

// закрывает брошенные события и удаляет устаревшие сразу при старте и дальше раз в cleanupInterval.
// retention 0 значит хранить события бессрочно
func startCleanup(ctx context.Context, retention time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		cleanupEvents(ctx, retention)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func cleanupEvents(ctx context.Context, retention time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	closed, err := database.CloseAbandonedEvents(ctx, closeGrace)
	if err != nil {
		log.Printf("cleanup: %v", err)
	}
	if len(closed) > 0 {
		fmt.Printf("Closed %d abandoned events: %v\n", len(closed), closed)
	}

	if retention == 0 {
		return
	}

	deleted, err := database.DeleteExpiredEvents(ctx, retention)
	if err != nil {
		log.Printf("cleanup: %v", err)
	}
	if len(deleted) > 0 {
		fmt.Printf("Deleted %d expired events: %v\n", len(deleted), deleted)
	}
}
