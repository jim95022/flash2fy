package userstorage

import (
	"sync"

	"flash2fy/internal/app/domain/user"
)

// MemoryRepository persists users in memory; suitable for tests and demos.
type MemoryRepository struct {
	mu    sync.RWMutex
	store map[string]user.User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		store: make(map[string]user.User),
	}
}

func (r *MemoryRepository) Save(u user.User) (user.User, error) {
	if err := u.Validate(); err != nil {
		return user.User{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[u.ID] = u
	return u, nil
}

func (r *MemoryRepository) FindByID(id string) (user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.store[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func (r *MemoryRepository) FindAll() ([]user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]user.User, 0, len(r.store))
	for _, u := range r.store {
		users = append(users, u)
	}
	return users, nil
}

func (r *MemoryRepository) Update(u user.User) (user.User, error) {
	if err := u.Validate(); err != nil {
		return user.User{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[u.ID]; !ok {
		return user.User{}, user.ErrNotFound
	}
	r.store[u.ID] = u
	return u, nil
}

func (r *MemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return user.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
