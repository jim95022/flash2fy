package domain

import (
	"errors"
)

var (
	ErrEmptyID         = errors.New("telegram user id must not be empty")
	ErrEmptyCoreID     = errors.New("telegram user core id must not be empty")
	ErrEmptyTelegramID = errors.New("telegram user telegram id must not be empty")
	ErrUserNotFound    = errors.New("telegram user not found")
)

// User represents Telegram-specific user data linked to the core model.
type User struct {
	ID         string
	CoreUserID string
	TelegramID int64
	Name       string
	Username   string
}

func (u *User) Validate() error {
	if u.ID == "" {
		return ErrEmptyID
	}
	if u.CoreUserID == "" {
		return ErrEmptyCoreID
	}
	if u.TelegramID == 0 {
		return ErrEmptyTelegramID
	}
	return nil
}
