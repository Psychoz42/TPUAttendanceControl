package apisim

import (
	"backend/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsers(t *testing.T) {
	content := "user1 pass1 user1@tpu.ru Иванов-Иван-Иванович 8К52Т\n" +
		"\n" +
		"short line\n" +
		"  user2   pass2  user2@tpu.ru   Петров-Пётр   8К51  \n"

	filename := filepath.Join(t.TempDir(), "users.txt")
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	users, err := LoadUsers(filename)
	if err != nil {
		t.Fatal(err)
	}

	want := []User{
		{Username: "user1", Password: "pass1", Email: "user1@tpu.ru", FullName: "Иванов Иван Иванович", Group: "8К52Т"},
		{Username: "user2", Password: "pass2", Email: "user2@tpu.ru", FullName: "Петров Пётр", Group: "8К51"},
	}
	if len(users) != len(want) {
		t.Fatalf("got %d users, want %d: %+v", len(users), len(want), users)
	}
	for i := range want {
		if users[i] != want[i] {
			t.Errorf("user %d = %+v, want %+v", i, users[i], want[i])
		}
	}
}

func TestLoadUsersMissingFile(t *testing.T) {
	if _, err := LoadUsers(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoginCheckSim(t *testing.T) {
	Users = []User{{Username: "user1", Password: "pass1", Email: "user1@tpu.ru", FullName: "Иванов Иван", Group: "8К52Т"}}

	resp, err := LoginCheckSim(&utils.LoginRequest{Email: "user1@tpu.ru", Password: "pass1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "user1" || resp.FullName != "Иванов Иван" {
		t.Errorf("unexpected response %+v", resp)
	}

	if _, err := LoginCheckSim(&utils.LoginRequest{Email: "user1@tpu.ru", Password: "wrong"}); err == nil {
		t.Error("expected error for wrong password")
	}
}
