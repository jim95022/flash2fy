package cardapp

import (
	"github.com/google/uuid"

	appcard "flash2fy/internal/app/domain/card"
	telegrmdomain "flash2fy/internal/telegram/domain"
	telegrmports "flash2fy/internal/telegram/ports"
)

// AppCardService captures the upstream application contract used by Telegram.
type AppCardService interface {
	CreateCard(front, back, ownerID string) (appcard.Card, error)
	GetCard(id string) (appcard.Card, error)
	DeleteCard(id string) error
}

// Service orchestrates Telegram card workflows.
type Service struct {
	appCards AppCardService
	ctxRepo  telegrmports.CardRepository
}

func NewService(appCards AppCardService, ctxRepo telegrmports.CardRepository) *Service {
	return &Service{appCards: appCards, ctxRepo: ctxRepo}
}

func (s *Service) CreateCard(front, back string, owner telegrmdomain.User, chatID int64) (appcard.Card, error) {
	created, err := s.appCards.CreateCard(front, back, owner.CoreUserID)
	if err != nil {
		return appcard.Card{}, err
	}

	ctxCard := telegrmdomain.Card{
		ID:              uuid.NewString(),
		CoreCardID:      created.ID,
		OwnerTelegramID: owner.TelegramID,
		ChatID:          chatID,
	}
	if err := ctxCard.Validate(); err != nil {
		return appcard.Card{}, err
	}

	if _, err := s.ctxRepo.Save(ctxCard); err != nil {
		return appcard.Card{}, err
	}

	return created, nil
}

func (s *Service) GetCard(id string) (appcard.Card, error) {
	return s.appCards.GetCard(id)
}

func (s *Service) DeleteCard(id string) error {
	if err := s.appCards.DeleteCard(id); err != nil {
		return err
	}
	return s.ctxRepo.DeleteByCoreID(id)
}
