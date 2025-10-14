package cardapp

import (
	"testing"
	"time"

	"flash2fy/internal/adapters/storage"
	"flash2fy/internal/domain/card"
)

func TestCreateCard(t *testing.T) {
	repo := storage.NewMemoryCardRepository()
	service := NewCardService(repo)

	created, err := service.CreateCard("Question", "Answer")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected generated ID, got empty string")
	}
	if created.Front != "Question" || created.Back != "Answer" {
		t.Fatalf("unexpected card payload: %+v", created)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestCreateCardValidation(t *testing.T) {
	repo := storage.NewMemoryCardRepository()
	service := NewCardService(repo)

	if _, err := service.CreateCard("", ""); err != card.ErrEmptyFront {
		t.Fatalf("expected ErrEmptyFront, got %v", err)
	}
	if _, err := service.CreateCard("front", ""); err != card.ErrEmptyBack {
		t.Fatalf("expected ErrEmptyBack, got %v", err)
	}
}

func TestUpdateCard(t *testing.T) {
	repo := storage.NewMemoryCardRepository()
	service := NewCardService(repo)

	created, err := service.CreateCard("Front", "Back")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	time.Sleep(time.Millisecond)

	updated, err := service.UpdateCard(created.ID, "New Front", "New Back")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Front != "New Front" || updated.Back != "New Back" {
		t.Fatalf("unexpected card payload after update: %+v", updated)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("expected UpdatedAt to change; created=%v updated=%v", updated.CreatedAt, updated.UpdatedAt)
	}
}

func TestUpdateCardNotFound(t *testing.T) {
	repo := storage.NewMemoryCardRepository()
	service := NewCardService(repo)

	if _, err := service.UpdateCard("missing", "front", "back"); err != card.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteCard(t *testing.T) {
	repo := storage.NewMemoryCardRepository()
	service := NewCardService(repo)

	created, err := service.CreateCard("Front", "Back")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := service.DeleteCard(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := service.GetCard(created.ID); err != card.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
