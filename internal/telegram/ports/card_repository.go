package ports

import telegrmdomain "flash2fy/internal/telegram/domain"

type CardRepository interface {
	Save(telegrmdomain.Card) (telegrmdomain.Card, error)
	FindByCoreID(coreID string) (telegrmdomain.Card, error)
	DeleteByCoreID(coreID string) error
}
