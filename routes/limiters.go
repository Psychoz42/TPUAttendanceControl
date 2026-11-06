package routes

import (
	"backend/auth"
	"backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// лимиты запросов в минуту. Вся группа может сидеть за одним NAT университета,
// поэтому жёсткие лимиты считаются на аккаунт, а на один IP лимит мягкий
const (
	loginIPLimit      = 100
	loginAccountLimit = 10
	scanLimit         = 30
)

func newLimiter(max int, key func(c *fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          max,
		Expiration:   time.Minute,
		KeyGenerator: key,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(utils.ErrorResponse{Success: false, Code: "TOO_MANY_REQUESTS", Message: "Too many requests, try again later"})
		},
	})
}

func ipKey(c *fiber.Ctx) string {
	return c.IP()
}

// подбор пароля к одному аккаунту упирается в этот лимит, а соседи по NAT его не делят.
// Если тело не разобралось, email пустой, и такие запросы считаются вместе
func loginAccountKey(c *fiber.Ctx) string {
	var login utils.LoginRequest
	_ = c.BodyParser(&login)
	return c.IP() + "|" + login.Email
}

// запрос без токена считается по IP, пока AUTH_REQUIRED выключен
func userOrIPKey(c *fiber.Ctx) string {
	if username := auth.CurrentUsername(c); username != "" {
		return "user|" + username
	}
	return c.IP()
}
