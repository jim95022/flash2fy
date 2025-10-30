package userstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"flash2fy/internal/domain/user"
)

// PostgresRepository persists users in PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(u user.User) (user.User, error) {
	if err := u.Validate(); err != nil {
		return user.User{}, err
	}

	const query = `
		INSERT INTO users (id, nickname)
		VALUES ($1, $2)`

	if _, err := r.db.ExecContext(context.Background(), query, u.ID, u.Nickname); err != nil {
		return user.User{}, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

func (r *PostgresRepository) FindByID(id string) (user.User, error) {
	const query = `
		SELECT id, nickname
		FROM users
		WHERE id = $1`

	var u user.User
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(&u.ID, &u.Nickname)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return u, nil
}

func (r *PostgresRepository) FindAll() ([]user.User, error) {
	const query = `
		SELECT id, nickname
		FROM users
		ORDER BY nickname ASC`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.Nickname); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *PostgresRepository) Update(u user.User) (user.User, error) {
	if err := u.Validate(); err != nil {
		return user.User{}, err
	}

	const query = `
		UPDATE users
		SET nickname = $1
		WHERE id = $2`

	res, err := r.db.ExecContext(context.Background(), query, u.Nickname, u.ID)
	if err != nil {
		return user.User{}, fmt.Errorf("update user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return user.User{}, fmt.Errorf("update user rows affected: %w", err)
	}
	if affected == 0 {
		return user.User{}, user.ErrNotFound
	}

	return u, nil
}

func (r *PostgresRepository) Delete(id string) error {
	const query = `
		DELETE FROM users
		WHERE id = $1`

	res, err := r.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user rows affected: %w", err)
	}
	if affected == 0 {
		return user.ErrNotFound
	}

	return nil
}
