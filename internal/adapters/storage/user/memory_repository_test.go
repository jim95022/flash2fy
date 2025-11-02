package userstorage

import (
	"testing"

	"flash2fy/internal/app/domain/user"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	repo := NewMemoryRepository()
	u := user.User{ID: "user-1", Nickname: "alice"}

	saved, err := repo.Save(u)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if saved.Nickname != u.Nickname {
		t.Fatalf("unexpected nickname: %s", saved.Nickname)
	}

	found, err := repo.FindByID(u.ID)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if found.ID != u.ID {
		t.Fatalf("unexpected id: %s", found.ID)
	}

	newNickname := "alice2"
	u.Nickname = newNickname
	updated, err := repo.Update(u)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Nickname != newNickname {
		t.Fatalf("expected nickname %s, got %s", newNickname, updated.Nickname)
	}

	if err := repo.Delete(u.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := repo.FindByID(u.ID); err != user.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
