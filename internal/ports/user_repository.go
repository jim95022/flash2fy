package ports

import "flash2fy/internal/domain/user"

// UserRepository defines persistence behavior for users.
type UserRepository interface {
	Save(user.User) (user.User, error)
	FindByID(id string) (user.User, error)
	FindAll() ([]user.User, error)
	Update(user.User) (user.User, error)
	Delete(id string) error
}
