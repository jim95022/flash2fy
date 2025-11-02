package telecardstorage

import (
	"sync"

	"flash2fy/internal/telegram/domain"
)

// MemoryRepository stores Telegram card projections in memory.
type MemoryRepository struct {
	mu       sync.RWMutex
	byID     map[string]domain.Card
	coreToID map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byID:     make(map[string]domain.Card),
		coreToID: make(map[string]string),
	}
}

func (r *MemoryRepository) Save(card domain.Card) (domain.Card, error) {
	if err := card.Validate(); err != nil {
		return domain.Card{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[card.ID] = card
	r.coreToID[card.CoreCardID] = card.ID
	return card, nil
}

func (r *MemoryRepository) FindByCoreID(coreID string) (domain.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.coreToID[coreID]
	if !ok {
		return domain.Card{}, domain.ErrCardNotFound
	}
	card, ok := r.byID[id]
	if !ok {
		return domain.Card{}, domain.ErrCardNotFound
	}
	return card, nil
}

func (r *MemoryRepository) DeleteByCoreID(coreID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.coreToID[coreID]
	if !ok {
		return domain.ErrCardNotFound
	}

	delete(r.coreToID, coreID)
	delete(r.byID, id)

	return nil
}
