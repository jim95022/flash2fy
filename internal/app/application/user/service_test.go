package userapp

import (
	"testing"

	userstorage "flash2fy/internal/adapters/storage/user"
	"flash2fy/internal/app/domain/user"
)

func TestCreateUser(t *testing.T) {
	repo := userstorage.NewMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateUser("nickname")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected generated ID")
	}
	if created.Nickname != "nickname" {
		t.Fatalf("unexpected nickname: %s", created.Nickname)
	}
}

func TestCreateUserValidation(t *testing.T) {
	repo := userstorage.NewMemoryRepository()
	service := NewService(repo)

	if _, err := service.CreateUser(""); err != user.ErrEmptyNickname {
		t.Fatalf("expected ErrEmptyNickname, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := userstorage.NewMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateUser("nickname")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	updated, err := service.UpdateUser(created.ID, "new-nick")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Nickname != "new-nick" {
		t.Fatalf("expected nickname to change, got %s", updated.Nickname)
	}
}

func TestDeleteUser(t *testing.T) {
	repo := userstorage.NewMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateUser("nickname")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := service.DeleteUser(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := service.GetUser(created.ID); err != user.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
