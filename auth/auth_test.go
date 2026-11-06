package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-test-secret-test-secret"

func initAuth(t *testing.T, authRequired bool) {
	t.Helper()
	if err := Init(testSecret, authRequired); err != nil {
		t.Fatal(err)
	}
}

func signClaims(t *testing.T, method jwt.SigningMethod, key any, claims tokenClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func validClaims(username string) tokenClaims {
	return tokenClaims{
		Role: "teacher",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
}

func TestInitRejectsShortSecret(t *testing.T) {
	if err := Init("short", false); err == nil {
		t.Fatal("expected error for short secret")
	}
}

func TestIssueAndParseToken(t *testing.T) {
	initAuth(t, false)

	token, err := IssueToken(User{Username: "user1", Role: "student"})
	if err != nil {
		t.Fatal(err)
	}

	user, err := parseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if user.Username != "user1" || user.Role != "student" {
		t.Errorf("unexpected user %+v", user)
	}
}

func TestParseTokenRejects(t *testing.T) {
	initAuth(t, false)

	expired := validClaims("user1")
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	noExpiration := validClaims("user1")
	noExpiration.ExpiresAt = nil

	valid := signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("user1"))

	tokens := map[string]string{
		"garbage":       "not-a-token",
		"tampered":      valid[:len(valid)-2] + "xx",
		"wrong secret":  signClaims(t, jwt.SigningMethodHS256, []byte("another-secret-another-secret-another"), validClaims("user1")),
		"alg none":      signClaims(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims("user1")),
		"other alg":     signClaims(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims("user1")),
		"expired":       signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), expired),
		"no expiration": signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), noExpiration),
		"empty subject": signClaims(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("")),
	}
	for name, token := range tokens {
		if _, err := parseToken(token); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestMiddleware(t *testing.T) {
	teacherToken := func() string {
		initAuth(t, false)
		token, err := IssueToken(User{Username: "teacher1", Role: "teacher"})
		if err != nil {
			t.Fatal(err)
		}
		return token
	}()
	studentToken, err := IssueToken(User{Username: "user1", Role: "student"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		required bool
		header   string
		status   int
		username string
	}{
		{name: "no token, optional", required: false, header: "", status: 200, username: ""},
		{name: "no token, required", required: true, header: "", status: 401},
		{name: "bad token, optional", required: false, header: "Bearer not-a-token", status: 401},
		{name: "wrong scheme", required: false, header: "Basic " + teacherToken, status: 401},
		{name: "no scheme", required: false, header: teacherToken, status: 401},
		{name: "right role", required: true, header: "Bearer " + teacherToken, status: 200, username: "teacher1"},
		{name: "scheme case", required: true, header: "bearer " + teacherToken, status: 200, username: "teacher1"},
		{name: "wrong role", required: true, header: "Bearer " + studentToken, status: 403},
		{name: "wrong role, optional", required: false, header: "Bearer " + studentToken, status: 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initAuth(t, tt.required)

			app := fiber.New()
			var username string
			app.Post("/", Authenticate, RequireRole("teacher"), func(c *fiber.Ctx) error {
				username = CurrentUsername(c)
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("POST", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Errorf("status %d, want %d", resp.StatusCode, tt.status)
			}
			if username != tt.username {
				t.Errorf("username %q, want %q", username, tt.username)
			}
		})
	}
}
