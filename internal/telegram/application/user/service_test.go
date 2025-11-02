package userapp

import (
	"testing"

	teleuserstorage "flash2fy/internal/adapters/storage/telegram/user"
	userstorage "flash2fy/internal/adapters/storage/user"
	appuserapp "flash2fy/internal/app/application/user"
	"flash2fy/internal/app/domain/user"
	telegrmdomain "flash2fy/internal/telegram/domain"
)

func TestEnsureUserCreatesNew(t *testing.T) {
	appRepo := userstorage.NewMemoryRepository()
	appService := appuserapp.NewService(appRepo)
	ctxRepo := teleuserstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)

	coreUser, ctxUser, err := service.EnsureUser(12345, "John", "johnny")
	if err != nil {
		t.Fatalf("ensure user failed: %v", err)
	}
	if coreUser.ID == "" {
		t.Fatalf("expected core user id")
	}
	if ctxUser.CoreUserID != coreUser.ID {
		t.Fatalf("expected projection to link core id")
	}
	if ctxUser.TelegramID != 12345 {
		t.Fatalf("expected telegram id")
	}
}

func TestEnsureUserIdempotent(t *testing.T) {
	appRepo := userstorage.NewMemoryRepository()
	appService := appuserapp.NewService(appRepo)
	ctxRepo := teleuserstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)

	coreUser1, ctxUser1, err := service.EnsureUser(777, "Doe", "doe")
	if err != nil {
		t.Fatalf("initial ensure failed: %v", err)
	}

	coreUser2, ctxUser2, err := service.EnsureUser(777, "Jane", "newdoe")
	if err != nil {
		t.Fatalf("second ensure failed: %v", err)
	}

	if coreUser1.ID != coreUser2.ID {
		t.Fatalf("expected same core user")
	}
	if ctxUser1.ID != ctxUser2.ID {
		t.Fatalf("expected same projection user")
	}
	if ctxUser2.Name != "Jane" || ctxUser2.Username != "newdoe" {
		t.Fatalf("expected context info updated, got %+v", ctxUser2)
	}
}

func TestEnsureUserValidation(t *testing.T) {
	appRepo := userstorage.NewMemoryRepository()
	appService := appuserapp.NewService(appRepo)
	ctxRepo := teleuserstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)

	if _, _, err := service.EnsureUser(0, "", ""); err != telegrmdomain.ErrEmptyTelegramID {
		t.Fatalf("expected ErrEmptyTelegramID, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	appRepo := userstorage.NewMemoryRepository()
	appService := appuserapp.NewService(appRepo)
	ctxRepo := teleuserstorage.NewMemoryRepository()
	service := NewService(appService, ctxRepo)

	coreUser, _, err := service.EnsureUser(999, "Jane", "jane")
	if err != nil {
		t.Fatalf("ensure failed: %v", err)
	}
	if err := service.DeleteUser(coreUser.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := appRepo.FindByID(coreUser.ID); err != user.ErrNotFound {
		t.Fatalf("expected user removed from app repo, got %v", err)
	}
	if _, err := ctxRepo.FindByCoreID(coreUser.ID); err != telegrmdomain.ErrUserNotFound {
		t.Fatalf("expected context projection removed, got %v", err)
	}
}
