package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
)

// MatchRepository handles match database operations.
type MatchRepository struct {
	pool *pgxpool.Pool
}

// NewMatchRepository creates a new match repository.
func NewMatchRepository(pool *pgxpool.Pool) *MatchRepository {
	return &MatchRepository{pool: pool}
}

// Create inserts a new match with its players.
func (r *MatchRepository) Create(ctx context.Context, match *model.Match, players []model.MatchPlayer) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert match
	if match.ID == uuid.Nil {
		match.ID = uuid.New()
	}

	matchQuery := `
		INSERT INTO matches (id, venue_id, match_type, started_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	err = tx.QueryRow(ctx, matchQuery, match.ID, match.VenueID, match.MatchType, match.StartedAt).Scan(&match.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}

	// Insert match players
	for _, mp := range players {
		playerQuery := `
			INSERT INTO match_players (match_id, player_id, team)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(ctx, playerQuery, match.ID, mp.PlayerID, mp.Team)
		if err != nil {
			return fmt.Errorf("failed to add player to match: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// GetByID retrieves a match by ID.
func (r *MatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Match, error) {
	query := `
		SELECT id, venue_id, match_type, started_at, ended_at, created_at
		FROM matches WHERE id = $1
	`
	match := &model.Match{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&match.ID, &match.VenueID, &match.MatchType,
		&match.StartedAt, &match.EndedAt, &match.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get match: %w", err)
	}
	return match, nil
}

// GetMatchPlayers retrieves all players for a match.
func (r *MatchRepository) GetMatchPlayers(ctx context.Context, matchID uuid.UUID) ([]model.MatchPlayer, error) {
	query := `
		SELECT match_id, player_id, team
		FROM match_players
		WHERE match_id = $1
	`
	rows, err := r.pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match players: %w", err)
	}
	defer rows.Close()

	var players []model.MatchPlayer
	for rows.Next() {
		var mp model.MatchPlayer
		if err := rows.Scan(&mp.MatchID, &mp.PlayerID, &mp.Team); err != nil {
			return nil, fmt.Errorf("failed to scan match player: %w", err)
		}
		players = append(players, mp)
	}

	if players == nil {
		players = []model.MatchPlayer{}
	}
	return players, nil
}

// Complete marks a match as completed.
func (r *MatchRepository) Complete(ctx context.Context, matchID uuid.UUID, endedAt time.Time) error {
	query := `UPDATE matches SET ended_at = $2 WHERE id = $1 AND ended_at IS NULL`
	result, err := r.pool.Exec(ctx, query, matchID, endedAt)
	if err != nil {
		return fmt.Errorf("failed to complete match: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a match and all related data.
func (r *MatchRepository) Delete(ctx context.Context, matchID uuid.UUID) error {
	query := `DELETE FROM matches WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete match: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves all matches with optional filtering.
func (r *MatchRepository) List(ctx context.Context, limit int) ([]model.Match, error) {
	query := `
		SELECT id, venue_id, match_type, started_at, ended_at, created_at
		FROM matches
		ORDER BY started_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list matches: %w", err)
	}
	defer rows.Close()

	var matches []model.Match
	for rows.Next() {
		var m model.Match
		if err := rows.Scan(&m.ID, &m.VenueID, &m.MatchType, &m.StartedAt, &m.EndedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, m)
	}

	if matches == nil {
		matches = []model.Match{}
	}
	return matches, nil
}

// InsertEvents adds point events to a match (idempotent - ignores duplicates).
func (r *MatchRepository) InsertEvents(ctx context.Context, events []model.PointEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	// Build bulk insert query with ON CONFLICT DO NOTHING for idempotency
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*6)

	for i, e := range events {
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d)",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6,
		))
		valueArgs = append(valueArgs, e.ID, e.MatchID, e.Timestamp, e.ServerPlayerID, e.ServeType, e.PointWinnerTeam)
	}

	query := fmt.Sprintf(`
		INSERT INTO point_events (id, match_id, timestamp, server_player_id, serve_type, point_winner_team)
		VALUES %s
		ON CONFLICT (id) DO NOTHING
	`, strings.Join(valueStrings, ","))

	result, err := r.pool.Exec(ctx, query, valueArgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert events: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// GetEvents retrieves all events for a match.
func (r *MatchRepository) GetEvents(ctx context.Context, matchID uuid.UUID) ([]model.PointEvent, error) {
	query := `
		SELECT id, match_id, timestamp, server_player_id, serve_type, point_winner_team
		FROM point_events
		WHERE match_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := r.pool.Query(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []model.PointEvent
	for rows.Next() {
		var e model.PointEvent
		if err := rows.Scan(&e.ID, &e.MatchID, &e.Timestamp, &e.ServerPlayerID, &e.ServeType, &e.PointWinnerTeam); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	if events == nil {
		events = []model.PointEvent{}
	}
	return events, nil
}
