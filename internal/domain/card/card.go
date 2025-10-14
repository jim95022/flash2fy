package card

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyFront = errors.New("card front must not be empty")
	ErrEmptyBack  = errors.New("card back must not be empty")
	ErrNotFound   = errors.New("card not found")
)

// Card represents a flashcard with front and back content.
type Card struct {
	ID        string
	Front     string
	Back      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate ensures the card has the required fields.
func (c *Card) Validate() error {
	if strings.TrimSpace(c.Front) == "" {
		return ErrEmptyFront
	}
	if strings.TrimSpace(c.Back) == "" {
		return ErrEmptyBack
	}
	return nil
}
