package routes

import (
	"backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// лимиты запросов в минуту на один IP. Вся группа может сидеть
// за одним NAT университета, поэтому scan ограничен мягко
const (
	loginLimit = 10
	scanLimit  = 30
)

func newLimiter(max int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(utils.ErrorResponse{Success: false, Code: "TOO_MANY_REQUESTS", Message: "Too many requests, try again later"})
		},
	})
}
