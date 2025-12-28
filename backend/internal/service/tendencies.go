package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
)

// DateFilter represents a date range filter for tendencies.
type DateFilter struct {
	Enabled   bool
	StartDate time.Time
	EndDate   time.Time
}

// TendenciesService handles venue tendency business logic.
type TendenciesService struct {
	tendenciesRepo *repository.TendenciesRepository
	venueRepo      *repository.VenueRepository
}

// NewTendenciesService creates a new tendencies service.
func NewTendenciesService(
	tendenciesRepo *repository.TendenciesRepository,
	venueRepo *repository.VenueRepository,
) *TendenciesService {
	return &TendenciesService{
		tendenciesRepo: tendenciesRepo,
		venueRepo:      venueRepo,
	}
}

// GetVenueTendencies retrieves team and player tendencies for a venue.
// Per OTS_Venue_Team_Player_Tendencies_Spec.md:
// - Teams: Doubles only, minimum 3 matches at venue
// - Players: Minimum 5 matches at venue
// - All metrics are aggregated and deterministic
// - Results ordered alphabetically (neutral ordering, no rankings)
func (s *TendenciesService) GetVenueTendencies(ctx context.Context, venueID uuid.UUID, dateFilter DateFilter) (*model.VenueTendencies, error) {
	// Validate venue exists
	venue, err := s.venueRepo.GetByID(ctx, venueID)
	if err != nil {
		return nil, fmt.Errorf("venue not found: %w", err)
	}

	// Convert to repository date filter
	repoFilter := repository.DateFilter{
		Enabled:   dateFilter.Enabled,
		StartDate: dateFilter.StartDate,
		EndDate:   dateFilter.EndDate,
	}

	// Get team tendencies
	teamTendencies, err := s.getTeamTendencies(ctx, venueID, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get team tendencies: %w", err)
	}

	// Get player tendencies
	playerTendencies, err := s.getPlayerTendencies(ctx, venueID, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get player tendencies: %w", err)
	}

	return &model.VenueTendencies{
		VenueID:          venue.ID,
		VenueName:        venue.Name,
		TeamTendencies:   teamTendencies,
		PlayerTendencies: playerTendencies,
	}, nil
}

// getTeamTendencies retrieves and filters team tendencies.
// Per spec Section 3: Team eligibility requires at least 3 matches at venue.
// Per spec Section 4: Applies to doubles matches only.
func (s *TendenciesService) getTeamTendencies(ctx context.Context, venueID uuid.UUID, dateFilter repository.DateFilter) ([]model.VenueTeamTendency, error) {
	rawStats, err := s.tendenciesRepo.GetTeamStatsAtVenue(ctx, venueID, dateFilter)
	if err != nil {
		return nil, err
	}

	var tendencies []model.VenueTeamTendency

	for _, ts := range rawStats {
		// Apply eligibility threshold - per spec Section 3
		if ts.MatchesPlayed < model.MinTeamMatchesForTendency {
			continue
		}

		// Get serve stats for this team
		firstServesIn, _, firstServePointsWon, err := s.tendenciesRepo.GetTeamServeStatsAtVenue(
			ctx, venueID, ts.Player1ID, ts.Player2ID, dateFilter,
		)
		if err != nil {
			return nil, err
		}

		// Calculate derived metrics
		tendency := model.VenueTeamTendency{
			TeamID:      formatTeamID(ts.Player1ID, ts.Player2ID),
			Player1ID:   ts.Player1ID,
			Player2ID:   ts.Player2ID,
			Player1Name: ts.Player1Name,
			Player2Name: ts.Player2Name,
			MatchesPlayed: ts.MatchesPlayed,
			MatchesWon:    ts.MatchesWon,
		}

		// Win percentage: (matches_won / matches_played) * 100
		if ts.MatchesPlayed > 0 {
			tendency.WinPercentage = float64(ts.MatchesWon) / float64(ts.MatchesPlayed) * 100
		}

		// Average games per match
		if ts.MatchesPlayed > 0 {
			tendency.AvgGamesPerMatch = float64(ts.TotalGames) / float64(ts.MatchesPlayed)
		}

		// First serve points won percentage
		if firstServesIn > 0 {
			tendency.FirstServePointsWonPct = float64(firstServePointsWon) / float64(firstServesIn) * 100
		}

		// Note: DeucePercentage requires game-level tracking which isn't available
		// in current data model. Setting to 0 as games are tracked as points.
		tendency.DeucePercentage = 0

		tendencies = append(tendencies, tendency)
	}

	// Sort alphabetically by team name (neutral ordering per spec Section 4)
	sort.Slice(tendencies, func(i, j int) bool {
		nameI := tendencies[i].Player1Name + " / " + tendencies[i].Player2Name
		nameJ := tendencies[j].Player1Name + " / " + tendencies[j].Player2Name
		return nameI < nameJ
	})

	if tendencies == nil {
		tendencies = []model.VenueTeamTendency{}
	}
	return tendencies, nil
}

// getPlayerTendencies retrieves and filters player tendencies.
// Per spec Section 3: Player eligibility requires at least 5 matches at venue.
// Per spec Section 5: NO win percentage - explicitly forbidden.
func (s *TendenciesService) getPlayerTendencies(ctx context.Context, venueID uuid.UUID, dateFilter repository.DateFilter) ([]model.VenuePlayerTendency, error) {
	rawStats, err := s.tendenciesRepo.GetPlayerStatsAtVenue(ctx, venueID, dateFilter)
	if err != nil {
		return nil, err
	}

	var tendencies []model.VenuePlayerTendency

	for _, ps := range rawStats {
		// Apply eligibility threshold - per spec Section 3
		if ps.MatchesPlayed < model.MinPlayerMatchesForTendency {
			continue
		}

		tendency := model.VenuePlayerTendency{
			PlayerID:      ps.PlayerID,
			PlayerName:    ps.PlayerName,
			MatchesPlayed: ps.MatchesPlayed,
		}

		// First serve in percentage
		if ps.FirstServesTotal > 0 {
			tendency.FirstServeInPct = float64(ps.FirstServesIn) / float64(ps.FirstServesTotal) * 100
		}

		// Double faults per game served
		// Using total serve points as proxy for games served
		if ps.TotalGamesServed > 0 {
			// Approximate games from serve points (roughly 4-6 points per game)
			estimatedGames := float64(ps.TotalGamesServed) / 4.0
			if estimatedGames > 0 {
				tendency.DoubleFaultsPerGame = float64(ps.DoubleFaults) / estimatedGames
			}
		}

		// Average points per game
		if ps.TotalGames > 0 {
			// Approximate games from total points
			estimatedGames := float64(ps.TotalGames) / 4.0
			if estimatedGames > 0 {
				tendency.AvgPointsPerGame = float64(ps.TotalPointsWon) / estimatedGames
			}
		}

		tendencies = append(tendencies, tendency)
	}

	// Sort alphabetically by player name (neutral ordering per spec Section 5)
	sort.Slice(tendencies, func(i, j int) bool {
		return tendencies[i].PlayerName < tendencies[j].PlayerName
	})

	if tendencies == nil {
		tendencies = []model.VenuePlayerTendency{}
	}
	return tendencies, nil
}

// formatTeamID creates a consistent team identifier from two player IDs.
// IDs are sorted to ensure the same team always has the same ID regardless
// of which player is listed first.
func formatTeamID(player1ID, player2ID uuid.UUID) string {
	if player1ID.String() < player2ID.String() {
		return player1ID.String() + ":" + player2ID.String()
	}
	return player2ID.String() + ":" + player1ID.String()
}
