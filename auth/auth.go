package auth

import (
	"backend/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// сколько живёт токен, обновления нет: после истечения нужно войти заново
const tokenTTL = 12 * time.Hour

const minSecretLength = 32

const userLocalsKey = "user"

type User struct {
	Username string
	Role     string
}

type tokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

var secret []byte

// при false запросы без токена проходят как раньше. Нужно на время перехода клиентов на токены
var required bool

func Init(jwtSecret string, authRequired bool) error {
	if len(jwtSecret) < minSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d bytes long", minSecretLength)
	}

	secret = []byte(jwtSecret)
	required = authRequired

	return nil
}

// IssueToken не зависит от способа входа: так же будет вызываться и после входа через ТПУ
func IssueToken(user User) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
}

func parseToken(tokenString string) (User, error) {
	var claims tokenClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(*jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return User{}, fmt.Errorf("parse token: %w", err)
	}

	// пустой логин значит «пользователя нет», в токене такого быть не должно
	if claims.Subject == "" {
		return User{}, errors.New("parse token: empty subject")
	}

	return User{Username: claims.Subject, Role: claims.Role}, nil
}

func sendUnauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{Success: false, Code: "UNAUTHORIZED", Message: "Invalid or missing token"})
}

// Authenticate кладёт пользователя из токена в c.Locals. Неверный токен отклоняется
// всегда, даже при выключенном AUTH_REQUIRED, чтобы не откатиться к данным из тела запроса
func Authenticate(c *fiber.Ctx) error {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		if required {
			return sendUnauthorized(c)
		}
		return c.Next()
	}

	scheme, tokenString, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return sendUnauthorized(c)
	}

	user, err := parseToken(tokenString)
	if err != nil {
		return sendUnauthorized(c)
	}

	c.Locals(userLocalsKey, user)

	return c.Next()
}

// RequireRole ставится после Authenticate. Запрос без токена сюда доходит,
// только пока AUTH_REQUIRED выключен, и роль у него не проверяется
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals(userLocalsKey).(User)
		if ok && user.Role != role {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{Success: false, Code: "FORBIDDEN", Message: "Not enough rights"})
		}
		return c.Next()
	}
}

// CurrentUsername возвращает логин из токена или "", если запрос пришёл без токена
func CurrentUsername(c *fiber.Ctx) string {
	user, _ := c.Locals(userLocalsKey).(User)
	return user.Username
}
