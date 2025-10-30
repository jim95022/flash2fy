package cardstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"flash2fy/internal/domain/card"
)

// PostgresRepository persists cards in a PostgreSQL database.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(c card.Card) (card.Card, error) {
	const query = `
		INSERT INTO cards (id, front, back, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := r.db.ExecContext(context.Background(), query, c.ID, c.Front, c.Back, c.OwnerID, c.CreatedAt, c.UpdatedAt); err != nil {
		return card.Card{}, fmt.Errorf("insert card: %w", err)
	}

	return c, nil
}

func (r *PostgresRepository) FindByID(id string) (card.Card, error) {
	const query = `
		SELECT id, front, back, owner_id, created_at, updated_at
		FROM cards
		WHERE id = $1`

	var c card.Card
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(&c.ID, &c.Front, &c.Back, &c.OwnerID, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return card.Card{}, card.ErrNotFound
	}
	if err != nil {
		return card.Card{}, fmt.Errorf("find card by id: %w", err)
	}

	return c, nil
}

func (r *PostgresRepository) FindAll() ([]card.Card, error) {
	const query = `
		SELECT id, front, back, owner_id, created_at, updated_at
		FROM cards
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	defer rows.Close()

	var cards []card.Card
	for rows.Next() {
		var c card.Card
		if err := rows.Scan(&c.ID, &c.Front, &c.Back, &c.OwnerID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cards: %w", err)
	}

	return cards, nil
}

func (r *PostgresRepository) Update(c card.Card) (card.Card, error) {
	const query = `
		UPDATE cards
		SET front = $1, back = $2, owner_id = $3, updated_at = $4
		WHERE id = $5`

	res, err := r.db.ExecContext(context.Background(), query, c.Front, c.Back, c.OwnerID, c.UpdatedAt, c.ID)
	if err != nil {
		return card.Card{}, fmt.Errorf("update card: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return card.Card{}, fmt.Errorf("update card rows affected: %w", err)
	}
	if affected == 0 {
		return card.Card{}, card.ErrNotFound
	}

	return c, nil
}

func (r *PostgresRepository) Delete(id string) error {
	const query = `
		DELETE FROM cards
		WHERE id = $1`

	res, err := r.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete card rows affected: %w", err)
	}
	if affected == 0 {
		return card.ErrNotFound
	}
	return nil
}
