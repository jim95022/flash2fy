package storage

import (
	"sync"

	"flash2fy/internal/domain/card"
)

// MemoryCardRepository persists cards in memory; useful for demos and tests.
type MemoryCardRepository struct {
	mu    sync.RWMutex
	store map[string]card.Card
}

func NewMemoryCardRepository() *MemoryCardRepository {
	return &MemoryCardRepository{
		store: make(map[string]card.Card),
	}
}

func (r *MemoryCardRepository) Save(c card.Card) (card.Card, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[c.ID] = c
	return c, nil
}

func (r *MemoryCardRepository) FindByID(id string) (card.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.store[id]
	if !ok {
		return card.Card{}, card.ErrNotFound
	}
	return c, nil
}

func (r *MemoryCardRepository) FindAll() ([]card.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cards := make([]card.Card, 0, len(r.store))
	for _, c := range r.store {
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *MemoryCardRepository) Update(c card.Card) (card.Card, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[c.ID]; !ok {
		return card.Card{}, card.ErrNotFound
	}
	r.store[c.ID] = c
	return c, nil
}

func (r *MemoryCardRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return card.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
