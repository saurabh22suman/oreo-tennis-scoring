package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrDuplicateID = errors.New("duplicate id")
	ErrInvalidData = errors.New("invalid data")
)

// PlayerRepository handles player database operations.
type PlayerRepository struct {
	pool *pgxpool.Pool
}

// NewPlayerRepository creates a new player repository.
func NewPlayerRepository(pool *pgxpool.Pool) *PlayerRepository {
	return &PlayerRepository{pool: pool}
}

// Create inserts a new player.
func (r *PlayerRepository) Create(ctx context.Context, player *model.Player) error {
	query := `
		INSERT INTO players (id, name, active)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`
	if player.ID == uuid.Nil {
		player.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query, player.ID, player.Name, player.Active).Scan(&player.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}
	return nil
}

// GetByID retrieves a player by ID.
func (r *PlayerRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Player, error) {
	query := `SELECT id, name, active, created_at FROM players WHERE id = $1`

	player := &model.Player{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&player.ID, &player.Name, &player.Active, &player.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return player, nil
}

// List retrieves all players.
func (r *PlayerRepository) List(ctx context.Context, activeOnly bool) ([]model.Player, error) {
	query := `SELECT id, name, active, created_at FROM players`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	defer rows.Close()

	var players []model.Player
	for rows.Next() {
		var p model.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.Active, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, p)
	}

	if players == nil {
		players = []model.Player{}
	}
	return players, nil
}

// Update modifies an existing player.
func (r *PlayerRepository) Update(ctx context.Context, player *model.Player) error {
	query := `
		UPDATE players
		SET name = $2, active = $3
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, player.ID, player.Name, player.Active)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
