package userapp

import (
	"fmt"

	"github.com/google/uuid"

	appuser "flash2fy/internal/app/domain/user"
	telegrmdomain "flash2fy/internal/telegram/domain"
	telegrmports "flash2fy/internal/telegram/ports"
)

// AppUserService captures the upstream application contract used by Telegram.
type AppUserService interface {
	CreateUser(nickname string) (appuser.User, error)
	GetUser(id string) (appuser.User, error)
	DeleteUser(id string) error
}

// Service coordinates Telegram-specific user behaviour with the core model.
type Service struct {
	appUsers AppUserService
	ctxRepo  telegrmports.UserRepository
}

func NewService(appUsers AppUserService, ctxRepo telegrmports.UserRepository) *Service {
	return &Service{appUsers: appUsers, ctxRepo: ctxRepo}
}

func (s *Service) EnsureUser(telegramID int64, name, username string) (appuser.User, telegrmdomain.User, error) {
	if telegramID == 0 {
		return appuser.User{}, telegrmdomain.User{}, telegrmdomain.ErrEmptyTelegramID
	}

	existingCtx, err := s.ctxRepo.FindByTelegramID(telegramID)
	if err == nil {
		coreUser, err := s.appUsers.GetUser(existingCtx.CoreUserID)
		if err != nil {
			return appuser.User{}, telegrmdomain.User{}, err
		}
		if existingCtx.Name != name || existingCtx.Username != username {
			existingCtx.Name = name
			existingCtx.Username = username
			if _, err := s.ctxRepo.Save(existingCtx); err != nil {
				return appuser.User{}, telegrmdomain.User{}, err
			}
		}
		return coreUser, existingCtx, nil
	}
	if err != nil && err != telegrmdomain.ErrUserNotFound {
		return appuser.User{}, telegrmdomain.User{}, err
	}

	nickname := username
	if nickname == "" {
		nickname = fmt.Sprintf("tg-%d", telegramID)
	}

	createdCore, err := s.appUsers.CreateUser(nickname)
	if err != nil {
		return appuser.User{}, telegrmdomain.User{}, err
	}

	ctxUser := telegrmdomain.User{
		ID:         uuid.NewString(),
		CoreUserID: createdCore.ID,
		TelegramID: telegramID,
		Name:       name,
		Username:   username,
	}
	if err := ctxUser.Validate(); err != nil {
		return appuser.User{}, telegrmdomain.User{}, err
	}

	savedCtx, err := s.ctxRepo.Save(ctxUser)
	if err != nil {
		return appuser.User{}, telegrmdomain.User{}, err
	}

	return createdCore, savedCtx, nil
}

func (s *Service) DeleteUser(coreID string) error {
	if err := s.appUsers.DeleteUser(coreID); err != nil {
		return err
	}
	return s.ctxRepo.DeleteByCoreID(coreID)
}
