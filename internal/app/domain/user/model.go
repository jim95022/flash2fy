package user

import (
	"errors"
	"strings"
)

var (
	ErrEmptyID       = errors.New("user id must not be empty")
	ErrEmptyNickname = errors.New("user nickname must not be empty")
	ErrNotFound      = errors.New("user not found")
)

// User represents an application user.
type User struct {
	ID       string
	Nickname string
}

// Validate ensures required fields are present.
func (u *User) Validate() error {
	if strings.TrimSpace(u.ID) == "" {
		return ErrEmptyID
	}
	if strings.TrimSpace(u.Nickname) == "" {
		return ErrEmptyNickname
	}
	return nil
}
