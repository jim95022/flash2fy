package ports

import "flash2fy/internal/domain/card"

// CardRepository defines the persistence behavior for cards.
type CardRepository interface {
	Save(card.Card) (card.Card, error)
	FindByID(id string) (card.Card, error)
	FindAll() ([]card.Card, error)
	Update(card.Card) (card.Card, error)
	Delete(id string) error
}
