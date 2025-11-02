package cardapp

import (
	"testing"

	cardstorage "flash2fy/internal/adapters/storage/card"
	telecardstorage "flash2fy/internal/adapters/storage/telegram/card"
	appcardapp "flash2fy/internal/app/application/card"
	"flash2fy/internal/app/domain/card"
	telegrmdomain "flash2fy/internal/telegram/domain"
)

func TestCreateCard(t *testing.T) {
	appRepo := cardstorage.NewMemoryRepository()
	appService := appcardapp.NewService(appRepo)
	ctxRepo := telecardstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)

	owner := telegrmdomain.User{ID: "tg-user-1", CoreUserID: "core-user-1", TelegramID: 42}

	created, err := service.CreateCard("Front", "Back", owner, 1234)
	if err != nil {
		t.Fatalf("create card failed: %v", err)
	}
	if created.OwnerID != owner.CoreUserID {
		t.Fatalf("expected owner id to link core user")
	}

	if _, err := ctxRepo.FindByCoreID(created.ID); err != nil {
		t.Fatalf("expected context projection, got %v", err)
	}
}

func TestCreateCardValidation(t *testing.T) {
	appRepo := cardstorage.NewMemoryRepository()
	appService := appcardapp.NewService(appRepo)
	ctxRepo := telecardstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)
	owner := telegrmdomain.User{ID: "tg-user-1", CoreUserID: "core-user-1", TelegramID: 42}

	if _, err := service.CreateCard("", "", owner, 0); err != card.ErrEmptyFront {
		t.Fatalf("expected card validation to trigger, got %v", err)
	}
}

func TestDeleteCard(t *testing.T) {
	appRepo := cardstorage.NewMemoryRepository()
	appService := appcardapp.NewService(appRepo)
	ctxRepo := telecardstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)
	owner := telegrmdomain.User{ID: "tg-user-1", CoreUserID: "core-user-1", TelegramID: 42}

	created, err := service.CreateCard("Front", "Back", owner, 1234)
	if err != nil {
		t.Fatalf("create card failed: %v", err)
	}

	if err := service.DeleteCard(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := appRepo.FindByID(created.ID); err != card.ErrNotFound {
		t.Fatalf("expected core card removed, got %v", err)
	}
	if _, err := ctxRepo.FindByCoreID(created.ID); err == nil {
		t.Fatalf("expected context projection removed")
	}
}
