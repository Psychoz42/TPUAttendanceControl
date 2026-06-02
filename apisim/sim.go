package apisim

import (
	"backend/utils"
	"fmt"
)

type User struct {
	Username   string
	Password   string
	Email      string
	FullName   string
	Group      string
}

var users = []User{
	User{Username: "vvo20", Password: "12431243", Email: "vvo20@tpu.ru", FullName: "Островский Владислав Валентинович", Group: "8К52Т"},
}

func LoginCheckSim(login *utils.LoginRequest) (*utils.LoginResponse, error) {
	resp := utils.LoginResponse{}
	userFound := false

	for _, u := range users {
		if login.Email == u.Email && login.Password == u.Password {
			resp = utils.LoginResponse{Username: u.Username, FullName: u.FullName, Email: u.Email, Group: u.Group}
			userFound = true
			break
		}
	}

	if userFound {
		return &resp, fmt.Errorf("user did not found")
	} else {
		return &resp, nil
	}
}

func GetFullNameByUsername(username string) (string, error) {
	for _, u := range users {
		if u.Username == username {
			return u.FullName, nil
		}
	}
	return "", fmt.Errorf("No such student")
}