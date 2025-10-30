package cardapp

import (
	"time"

	"github.com/google/uuid"

	"flash2fy/internal/domain/card"
	"flash2fy/internal/ports"
)

// CardService orchestrates card use-cases.
type CardService struct {
	repo ports.CardRepository
}

func NewCardService(repo ports.CardRepository) *CardService {
	return &CardService{repo: repo}
}

func (s *CardService) CreateCard(front, back, ownerID string) (card.Card, error) {
	newCard := card.Card{
		ID:        uuid.NewString(),
		Front:     front,
		Back:      back,
		OwnerID:   ownerID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := newCard.Validate(); err != nil {
		return card.Card{}, err
	}

	return s.repo.Save(newCard)
}

func (s *CardService) GetCard(id string) (card.Card, error) {
	return s.repo.FindByID(id)
}

func (s *CardService) ListCards() ([]card.Card, error) {
	return s.repo.FindAll()
}

func (s *CardService) UpdateCard(id, front, back string) (card.Card, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return card.Card{}, err
	}

	existing.Front = front
	existing.Back = back
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return card.Card{}, err
	}

	return s.repo.Update(existing)
}

func (s *CardService) DeleteCard(id string) error {
	return s.repo.Delete(id)
}
