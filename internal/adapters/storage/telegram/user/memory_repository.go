package teleuserstorage

import (
	"sync"

	"flash2fy/internal/telegram/domain"
)

// MemoryRepository stores Telegram user projections in memory.
type MemoryRepository struct {
	mu           sync.RWMutex
	byID         map[string]domain.User
	coreToID     map[string]string
	telegramToID map[int64]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byID:         make(map[string]domain.User),
		coreToID:     make(map[string]string),
		telegramToID: make(map[int64]string),
	}
}

func (r *MemoryRepository) Save(user domain.User) (domain.User, error) {
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[user.ID] = user
	r.coreToID[user.CoreUserID] = user.ID
	r.telegramToID[user.TelegramID] = user.ID
	return user, nil
}

func (r *MemoryRepository) FindByTelegramID(id int64) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctxID, ok := r.telegramToID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	user, ok := r.byID[ctxID]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *MemoryRepository) FindByCoreID(coreID string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctxID, ok := r.coreToID[coreID]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	user, ok := r.byID[ctxID]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *MemoryRepository) DeleteByCoreID(coreID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctxID, ok := r.coreToID[coreID]
	if !ok {
		return domain.ErrUserNotFound
	}

	user, ok := r.byID[ctxID]
	if ok {
		delete(r.telegramToID, user.TelegramID)
	}

	delete(r.coreToID, coreID)
	delete(r.byID, ctxID)
	return nil
}
