package userapp

import (
	"strings"

	"github.com/google/uuid"

	"flash2fy/internal/app/domain/user"
	coreports "flash2fy/internal/app/ports"
)

// Service orchestrates user use-cases.
type Service struct {
	repo coreports.UserRepository
}

func NewService(repo coreports.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(nickname string) (user.User, error) {
	newUser := user.User{
		ID:       uuid.NewString(),
		Nickname: strings.TrimSpace(nickname),
	}
	if err := newUser.Validate(); err != nil {
		return user.User{}, err
	}

	return s.repo.Save(newUser)
}

func (s *Service) GetUser(id string) (user.User, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListUsers() ([]user.User, error) {
	return s.repo.FindAll()
}

func (s *Service) UpdateUser(id, nickname string) (user.User, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return user.User{}, err
	}

	existing.Nickname = strings.TrimSpace(nickname)
	if err := existing.Validate(); err != nil {
		return user.User{}, err
	}

	return s.repo.Update(existing)
}

func (s *Service) DeleteUser(id string) error {
	return s.repo.Delete(id)
}
