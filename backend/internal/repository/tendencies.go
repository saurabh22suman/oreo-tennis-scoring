package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DateFilter represents a date range filter for tendencies queries.
type DateFilter struct {
	Enabled   bool
	StartDate time.Time
	EndDate   time.Time
}

// TendenciesRepository handles venue tendency database operations.
type TendenciesRepository struct {
	pool *pgxpool.Pool
}

// NewTendenciesRepository creates a new tendencies repository.
func NewTendenciesRepository(pool *pgxpool.Pool) *TendenciesRepository {
	return &TendenciesRepository{pool: pool}
}

// TeamMatchStats contains raw aggregated data for a team at a venue.
type TeamMatchStats struct {
	Player1ID           uuid.UUID
	Player2ID           uuid.UUID
	Player1Name         string
	Player2Name         string
	MatchesPlayed       int
	MatchesWon          int
	TotalGames          int
	TotalDeuces         int
	TotalGamesForDeuce  int
	FirstServesIn       int
	FirstServesTotal    int
	FirstServePointsWon int
}

// PlayerMatchStats contains raw aggregated data for a player at a venue.
type PlayerMatchStats struct {
	PlayerID         uuid.UUID
	PlayerName       string
	MatchesPlayed    int
	FirstServesIn    int
	FirstServesTotal int
	DoubleFaults     int
	TotalGamesServed int
	TotalPointsWon   int
	TotalGames       int
}

// GetTeamStatsAtVenue retrieves aggregated team statistics for a venue.
// Returns only doubles teams that have played at this venue.
func (r *TendenciesRepository) GetTeamStatsAtVenue(ctx context.Context, venueID uuid.UUID, dateFilter DateFilter) ([]TeamMatchStats, error) {
	// Build date filter condition
	dateCondition := ""
	args := []interface{}{venueID}
	if dateFilter.Enabled {
		dateCondition = "AND m.ended_at >= $2 AND m.ended_at < $3"
		args = append(args, dateFilter.StartDate, dateFilter.EndDate)
	}

	// Query to get all doubles teams and their match counts at the venue.
	// A team is identified by the pair of player IDs (sorted to ensure consistency).
	query := fmt.Sprintf(`
		WITH doubles_matches AS (
			-- Get all completed doubles matches at this venue
			SELECT m.id as match_id
			FROM matches m
			WHERE m.venue_id = $1
			  AND m.match_type = 'doubles'
			  AND m.ended_at IS NOT NULL
			  %s
		),
		team_compositions AS (
			-- Get team compositions for each match
			SELECT 
				dm.match_id,
				mp.team,
				ARRAY_AGG(mp.player_id ORDER BY mp.player_id) as player_ids,
				ARRAY_AGG(p.name ORDER BY mp.player_id) as player_names
			FROM doubles_matches dm
			JOIN match_players mp ON mp.match_id = dm.match_id
			JOIN players p ON p.id = mp.player_id
			GROUP BY dm.match_id, mp.team
		),
		match_outcomes AS (
			-- Compute winner for each match based on point events
			SELECT 
				pe.match_id,
				CASE 
					WHEN SUM(CASE WHEN pe.point_winner_team = 'A' THEN 1 ELSE 0 END) > 
					     SUM(CASE WHEN pe.point_winner_team = 'B' THEN 1 ELSE 0 END)
					THEN 'A'
					ELSE 'B'
				END as winning_team,
				COUNT(*) as total_points
			FROM point_events pe
			WHERE pe.match_id IN (SELECT match_id FROM doubles_matches)
			GROUP BY pe.match_id
		),
		team_matches AS (
			-- Join teams with their match outcomes
			SELECT 
				tc.player_ids[1] as player1_id,
				tc.player_ids[2] as player2_id,
				tc.player_names[1] as player1_name,
				tc.player_names[2] as player2_name,
				tc.match_id,
				tc.team,
				CASE WHEN mo.winning_team = tc.team THEN 1 ELSE 0 END as won,
				mo.total_points as total_games
			FROM team_compositions tc
			JOIN match_outcomes mo ON mo.match_id = tc.match_id
		)
		-- Aggregate by team
		SELECT 
			player1_id,
			player2_id,
			player1_name,
			player2_name,
			COUNT(*) as matches_played,
			SUM(won) as matches_won,
			SUM(total_games) as total_games
		FROM team_matches
		GROUP BY player1_id, player2_id, player1_name, player2_name
		ORDER BY player1_name, player2_name
	`, dateCondition)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats: %w", err)
	}
	defer rows.Close()

	var results []TeamMatchStats
	for rows.Next() {
		var ts TeamMatchStats
		if err := rows.Scan(
			&ts.Player1ID,
			&ts.Player2ID,
			&ts.Player1Name,
			&ts.Player2Name,
			&ts.MatchesPlayed,
			&ts.MatchesWon,
			&ts.TotalGames,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team stats: %w", err)
		}
		results = append(results, ts)
	}

	if results == nil {
		results = []TeamMatchStats{}
	}
	return results, nil
}

// GetTeamServeStatsAtVenue retrieves first serve stats for teams at a venue.
func (r *TendenciesRepository) GetTeamServeStatsAtVenue(ctx context.Context, venueID uuid.UUID, player1ID, player2ID uuid.UUID, dateFilter DateFilter) (firstServesIn, firstServesTotal, firstServePointsWon int, err error) {
	// Build date filter condition
	dateCondition := ""
	args := []interface{}{venueID, player1ID, player2ID}
	if dateFilter.Enabled {
		dateCondition = "AND m.ended_at >= $4 AND m.ended_at < $5"
		args = append(args, dateFilter.StartDate, dateFilter.EndDate)
	}

	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN pe.serve_type = 'first' THEN 1 ELSE 0 END), 0) as first_serves_in,
			COALESCE(SUM(CASE WHEN pe.serve_type IN ('first', 'second', 'double_fault') THEN 1 ELSE 0 END), 0) as first_serves_total,
			COALESCE(SUM(CASE 
				WHEN pe.serve_type = 'first' AND mp.team = pe.point_winner_team THEN 1 
				ELSE 0 
			END), 0) as first_serve_points_won
		FROM point_events pe
		JOIN matches m ON m.id = pe.match_id
		JOIN match_players mp ON mp.match_id = pe.match_id AND mp.player_id = pe.server_player_id
		WHERE m.venue_id = $1
		  AND m.match_type = 'doubles'
		  AND m.ended_at IS NOT NULL
		  AND pe.server_player_id IN ($2, $3)
		  AND EXISTS (
			SELECT 1 FROM match_players mp1 
			JOIN match_players mp2 ON mp1.match_id = mp2.match_id AND mp1.team = mp2.team
			WHERE mp1.match_id = m.id 
			  AND mp1.player_id = $2 
			  AND mp2.player_id = $3
		  )
		  %s
	`, dateCondition)

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&firstServesIn,
		&firstServesTotal,
		&firstServePointsWon,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get team serve stats: %w", err)
	}
	return firstServesIn, firstServesTotal, firstServePointsWon, nil
}

// GetPlayerStatsAtVenue retrieves aggregated player statistics for a venue.
func (r *TendenciesRepository) GetPlayerStatsAtVenue(ctx context.Context, venueID uuid.UUID, dateFilter DateFilter) ([]PlayerMatchStats, error) {
	// Build date filter condition
	dateCondition := ""
	args := []interface{}{venueID}
	if dateFilter.Enabled {
		dateCondition = "AND m.ended_at >= $2 AND m.ended_at < $3"
		args = append(args, dateFilter.StartDate, dateFilter.EndDate)
	}

	query := fmt.Sprintf(`
		WITH venue_matches AS (
			-- Get all completed matches at this venue
			SELECT m.id as match_id
			FROM matches m
			WHERE m.venue_id = $1
			  AND m.ended_at IS NOT NULL
			  %s
		),
		player_matches AS (
			-- Get distinct matches per player at venue
			SELECT 
				mp.player_id,
				p.name as player_name,
				COUNT(DISTINCT mp.match_id) as matches_played
			FROM venue_matches vm
			JOIN match_players mp ON mp.match_id = vm.match_id
			JOIN players p ON p.id = mp.player_id
			GROUP BY mp.player_id, p.name
		),
		player_serve_stats AS (
			-- Get serve statistics when player was serving
			SELECT 
				pe.server_player_id as player_id,
				SUM(CASE WHEN pe.serve_type = 'first' THEN 1 ELSE 0 END) as first_serves_in,
				COUNT(*) as first_serves_total,
				SUM(CASE WHEN pe.serve_type = 'double_fault' THEN 1 ELSE 0 END) as double_faults
			FROM point_events pe
			JOIN venue_matches vm ON vm.match_id = pe.match_id
			GROUP BY pe.server_player_id
		),
		player_points AS (
			-- Get total points won by player
			SELECT 
				mp.player_id,
				SUM(CASE WHEN pe.point_winner_team = mp.team THEN 1 ELSE 0 END) as total_points_won,
				COUNT(*) as total_points_in_matches
			FROM point_events pe
			JOIN venue_matches vm ON vm.match_id = pe.match_id
			JOIN match_players mp ON mp.match_id = pe.match_id
			GROUP BY mp.player_id
		)
		SELECT 
			pm.player_id,
			pm.player_name,
			pm.matches_played,
			COALESCE(pss.first_serves_in, 0) as first_serves_in,
			COALESCE(pss.first_serves_total, 0) as first_serves_total,
			COALESCE(pss.double_faults, 0) as double_faults,
			COALESCE(pss.first_serves_total, 0) as total_games_served,
			COALESCE(pp.total_points_won, 0) as total_points_won,
			COALESCE(pp.total_points_in_matches, 0) as total_games
		FROM player_matches pm
		LEFT JOIN player_serve_stats pss ON pss.player_id = pm.player_id
		LEFT JOIN player_points pp ON pp.player_id = pm.player_id
		ORDER BY pm.player_name
	`, dateCondition)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player stats: %w", err)
	}
	defer rows.Close()

	var results []PlayerMatchStats
	for rows.Next() {
		var ps PlayerMatchStats
		if err := rows.Scan(
			&ps.PlayerID,
			&ps.PlayerName,
			&ps.MatchesPlayed,
			&ps.FirstServesIn,
			&ps.FirstServesTotal,
			&ps.DoubleFaults,
			&ps.TotalGamesServed,
			&ps.TotalPointsWon,
			&ps.TotalGames,
		); err != nil {
			return nil, fmt.Errorf("failed to scan player stats: %w", err)
		}
		results = append(results, ps)
	}

	if results == nil {
		results = []PlayerMatchStats{}
	}
	return results, nil
}
