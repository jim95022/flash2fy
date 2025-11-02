package domain

import "errors"

var (
	ErrEmptyCardID   = errors.New("telegram card id must not be empty")
	ErrEmptyCoreCard = errors.New("telegram card core id must not be empty")
	ErrEmptyOwner    = errors.New("telegram card owner id must not be empty")
	ErrCardNotFound  = errors.New("telegram card not found")
)

// Card represents Telegram-specific projection of a flashcard.
type Card struct {
	ID              string
	CoreCardID      string
	OwnerTelegramID int64
	ChatID          int64
}

func (c *Card) Validate() error {
	if c.ID == "" {
		return ErrEmptyCardID
	}
	if c.CoreCardID == "" {
		return ErrEmptyCoreCard
	}
	if c.OwnerTelegramID == 0 {
		return ErrEmptyOwner
	}
	return nil
}
