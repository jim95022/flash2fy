package ports

import telegrmdomain "flash2fy/internal/telegram/domain"

type UserRepository interface {
	Save(telegrmdomain.User) (telegrmdomain.User, error)
	FindByTelegramID(id int64) (telegrmdomain.User, error)
	FindByCoreID(coreID string) (telegrmdomain.User, error)
	DeleteByCoreID(coreID string) error
}
