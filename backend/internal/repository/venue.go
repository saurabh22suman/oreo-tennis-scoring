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

// VenueRepository handles venue database operations.
type VenueRepository struct {
	pool *pgxpool.Pool
}

// NewVenueRepository creates a new venue repository.
func NewVenueRepository(pool *pgxpool.Pool) *VenueRepository {
	return &VenueRepository{pool: pool}
}

// Create inserts a new venue.
func (r *VenueRepository) Create(ctx context.Context, venue *model.Venue) error {
	query := `
		INSERT INTO venues (id, name, surface, active)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	if venue.ID == uuid.Nil {
		venue.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query, venue.ID, venue.Name, venue.Surface, venue.Active).Scan(&venue.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create venue: %w", err)
	}
	return nil
}

// GetByID retrieves a venue by ID.
func (r *VenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Venue, error) {
	query := `SELECT id, name, surface, active, created_at FROM venues WHERE id = $1`

	venue := &model.Venue{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&venue.ID, &venue.Name, &venue.Surface, &venue.Active, &venue.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}
	return venue, nil
}

// List retrieves all venues.
func (r *VenueRepository) List(ctx context.Context, activeOnly bool) ([]model.Venue, error) {
	query := `SELECT id, name, surface, active, created_at FROM venues`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}
	defer rows.Close()

	var venues []model.Venue
	for rows.Next() {
		var v model.Venue
		if err := rows.Scan(&v.ID, &v.Name, &v.Surface, &v.Active, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan venue: %w", err)
		}
		venues = append(venues, v)
	}

	if venues == nil {
		venues = []model.Venue{}
	}
	return venues, nil
}

// Update modifies an existing venue.
func (r *VenueRepository) Update(ctx context.Context, venue *model.Venue) error {
	query := `
		UPDATE venues
		SET name = $2, surface = $3, active = $4
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, venue.ID, venue.Name, venue.Surface, venue.Active)
	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
