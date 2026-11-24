package code

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 16

// Уникальность кода проверяет constraint events_code_unique,
// при совпадении database повторяет вставку с новым кодом
func GenerateNewCode() string {
	b := make([]byte, codeLength)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err) // crypto/rand не возвращает ошибок в Go 1.24+
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
