package domain

import (
	"errors"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("username already exists")
	ErrInvalidUsername = errors.New("invalid username")
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (u User) Validate() error {
	if len(u.Username) < 3 {
		return ErrInvalidUsername
	}
	return nil
}
