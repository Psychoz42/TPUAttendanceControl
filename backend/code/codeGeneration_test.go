package code

import (
	"strings"
	"testing"
)

func TestGenerateNewCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		c := GenerateNewCode()
		if len(c) != codeLength {
			t.Fatalf("code %q has length %d, want %d", c, len(c), codeLength)
		}
		for _, r := range c {
			if !strings.ContainsRune(charset, r) {
				t.Fatalf("code %q contains %q not from charset", c, r)
			}
		}
		if seen[c] {
			t.Fatalf("code %q generated twice", c)
		}
		seen[c] = true
	}
}
