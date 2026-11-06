package apisim

import (
	"backend/utils"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type User struct {
	Username string
	Password string
	Email    string
	FullName string
	Group    string
	Role     string
}

var Users []User

func LoadUsers(filename string) ([]User, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line) // разбивает по любому количеству пробелов
		if len(fields) < 5 {
			continue
		}
		// роль необязательна, в старом формате из 5 полей все считаются студентами
		role := utils.RoleStudent
		if len(fields) > 5 {
			role = fields[5]
		}
		if role != utils.RoleStudent && role != utils.RoleTeacher {
			return nil, fmt.Errorf("user %s: unknown role %q", fields[0], role)
		}
		users = append(users, User{
			Username: fields[0],
			Password: fields[1],
			Email:    fields[2],
			FullName: strings.ReplaceAll(fields[3], "-", " "),
			Group:    fields[4],
			Role:     role,
		})
	}
	return users, scanner.Err()
}

func LoginCheckSim(login *utils.LoginRequest) (*utils.LoginResponse, error) {
	resp := utils.LoginResponse{}
	userFound := false

	for _, u := range Users {
		if login.Email == u.Email && login.Password == u.Password {
			resp = utils.LoginResponse{Username: u.Username, FullName: u.FullName, Email: u.Email, Group: u.Group, Role: u.Role}
			userFound = true
			break
		}
	}

	if !userFound {
		return &resp, fmt.Errorf("invalid email or password")
	} else {
		return &resp, nil
	}
}

func GetFullNameByUsername(username string) (string, error) {
	for _, u := range Users {
		if u.Username == username {
			return u.FullName, nil
		}
	}
	return "", fmt.Errorf("No such student")
}
