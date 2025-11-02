package cardstorage

import (
	"testing"
	"time"

	"flash2fy/internal/app/domain/card"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	repo := NewMemoryRepository()
	now := time.Now().UTC()
	c := card.Card{
		ID:        "card-1",
		Front:     "Question",
		Back:      "Answer",
		OwnerID:   "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	saved, err := repo.Save(c)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if saved.Front != c.Front {
		t.Fatalf("expected front %q, got %q", c.Front, saved.Front)
	}

	found, err := repo.FindByID(c.ID)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if found.Back != c.Back || found.OwnerID != c.OwnerID {
		t.Fatalf("unexpected card after find: %+v", found)
	}

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("findAll failed: %v", err)
	}
	if len(all) != 1 || all[0].ID != c.ID {
		t.Fatalf("unexpected findAll result: %+v", all)
	}

	updatedCard := found
	updatedCard.Front = "Updated question"
	updatedCard.Back = "Updated answer"
	updatedCard.UpdatedAt = now.Add(time.Minute)

	updated, err := repo.Update(updatedCard)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Front != updatedCard.Front || updated.OwnerID != updatedCard.OwnerID {
		t.Fatalf("unexpected card after update: %+v", updated)
	}

	if err := repo.Delete(updatedCard.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := repo.FindByID(updatedCard.ID); err != card.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryRepositoryNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	if _, err := repo.FindByID("missing"); err != card.ErrNotFound {
		t.Fatalf("expected ErrNotFound for missing card, got %v", err)
	}
	if err := repo.Delete("missing"); err != card.ErrNotFound {
		t.Fatalf("expected ErrNotFound when deleting missing card, got %v", err)
	}
}
