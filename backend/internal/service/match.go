package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
)

// MatchService handles match business logic.
type MatchService struct {
	matchRepo  *repository.MatchRepository
	playerRepo *repository.PlayerRepository
	venueRepo  *repository.VenueRepository
}

// NewMatchService creates a new match service.
func NewMatchService(
	matchRepo *repository.MatchRepository,
	playerRepo *repository.PlayerRepository,
	venueRepo *repository.VenueRepository,
) *MatchService {
	return &MatchService{
		matchRepo:  matchRepo,
		playerRepo: playerRepo,
		venueRepo:  venueRepo,
	}
}

// CreateMatchRequest represents a request to create a new match.
type CreateMatchRequest struct {
	VenueID   uuid.UUID       `json:"venue_id"`
	MatchType model.MatchType `json:"match_type"`
	TeamA     []uuid.UUID     `json:"team_a"`
	TeamB     []uuid.UUID     `json:"team_b"`
}

// CreateMatch creates a new match and returns its ID.
func (s *MatchService) CreateMatch(ctx context.Context, req CreateMatchRequest) (*model.Match, error) {
	// Validate venue exists
	_, err := s.venueRepo.GetByID(ctx, req.VenueID)
	if err != nil {
		return nil, fmt.Errorf("invalid venue: %w", err)
	}

	// Validate match type and player count
	if req.MatchType == model.MatchTypeSingles {
		if len(req.TeamA) != 1 || len(req.TeamB) != 1 {
			return nil, fmt.Errorf("singles match requires exactly 1 player per team")
		}
	} else if req.MatchType == model.MatchTypeDoubles {
		if len(req.TeamA) != 2 || len(req.TeamB) != 2 {
			return nil, fmt.Errorf("doubles match requires exactly 2 players per team")
		}
	} else if req.MatchType == model.MatchTypeAustralianDoubles {
		// 1v2 format: Team A has 1 player, Team B has 2 players
		if len(req.TeamA) != 1 || len(req.TeamB) != 2 {
			return nil, fmt.Errorf("1v2 match requires exactly 1 player on team A and 2 players on team B")
		}
	} else {
		return nil, fmt.Errorf("invalid match type")
	}

	// Validate all players exist
	allPlayers := append(req.TeamA, req.TeamB...)
	for _, playerID := range allPlayers {
		if _, err := s.playerRepo.GetByID(ctx, playerID); err != nil {
			return nil, fmt.Errorf("invalid player %s: %w", playerID, err)
		}
	}

	// Create match
	match := &model.Match{
		ID:        uuid.New(),
		VenueID:   req.VenueID,
		MatchType: req.MatchType,
		StartedAt: time.Now(),
	}

	// Prepare match players
	var matchPlayers []model.MatchPlayer
	for _, playerID := range req.TeamA {
		matchPlayers = append(matchPlayers, model.MatchPlayer{
			MatchID:  match.ID,
			PlayerID: playerID,
			Team:     model.TeamA,
		})
	}
	for _, playerID := range req.TeamB {
		matchPlayers = append(matchPlayers, model.MatchPlayer{
			MatchID:  match.ID,
			PlayerID: playerID,
			Team:     model.TeamB,
		})
	}

	if err := s.matchRepo.Create(ctx, match, matchPlayers); err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return match, nil
}

// AddEvents adds point events to a match (idempotent).
func (s *MatchService) AddEvents(ctx context.Context, matchID uuid.UUID, events []model.PointEvent) (int, error) {
	// Verify match exists and is not completed
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return 0, fmt.Errorf("match not found: %w", err)
	}
	if match.EndedAt != nil {
		return 0, fmt.Errorf("cannot add events to completed match")
	}

	// Set match ID for all events
	for i := range events {
		events[i].MatchID = matchID
	}

	return s.matchRepo.InsertEvents(ctx, events)
}

// CompleteMatch marks a match as completed.
func (s *MatchService) CompleteMatch(ctx context.Context, matchID uuid.UUID) error {
	return s.matchRepo.Complete(ctx, matchID, time.Now())
}

// GetMatchSummary computes statistics for a match.
func (s *MatchService) GetMatchSummary(ctx context.Context, matchID uuid.UUID) (*model.MatchSummary, error) {
	// Get match
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found: %w", err)
	}

	// Get venue
	venue, err := s.venueRepo.GetByID(ctx, match.VenueID)
	if err != nil {
		return nil, fmt.Errorf("venue not found: %w", err)
	}

	// Get players
	matchPlayers, err := s.matchRepo.GetMatchPlayers(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	// Get events
	events, err := s.matchRepo.GetEvents(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Build player map
	playerMap := make(map[uuid.UUID]*model.Player)
	playerTeamMap := make(map[uuid.UUID]model.Team)
	for _, mp := range matchPlayers {
		player, err := s.playerRepo.GetByID(ctx, mp.PlayerID)
		if err != nil {
			return nil, fmt.Errorf("player not found: %w", err)
		}
		playerMap[mp.PlayerID] = player
		playerTeamMap[mp.PlayerID] = mp.Team
	}

	// Compute stats
	stats := make(map[uuid.UUID]*model.PlayerMatchStats)
	for playerID, player := range playerMap {
		stats[playerID] = &model.PlayerMatchStats{
			PlayerID:   playerID,
			PlayerName: player.Name,
			Team:       playerTeamMap[playerID],
		}
	}

	teamAScore := 0
	teamBScore := 0

	for _, event := range events {
		serverStats := stats[event.ServerPlayerID]
		if serverStats == nil {
			continue
		}

		// Track serve stats
		switch event.ServeType {
		case model.ServeTypeFirst:
			serverStats.FirstServesTotal++
			serverStats.FirstServesIn++
			if playerTeamMap[event.ServerPlayerID] == event.PointWinnerTeam {
				serverStats.FirstServeWon++
			}
		case model.ServeTypeSecond:
			serverStats.FirstServesTotal++ // First serve was out
			serverStats.SecondServesTotal++
			serverStats.SecondServesIn++
			if playerTeamMap[event.ServerPlayerID] == event.PointWinnerTeam {
				serverStats.SecondServeWon++
			}
		case model.ServeTypeDoubleFault:
			serverStats.FirstServesTotal++
			serverStats.SecondServesTotal++
			serverStats.DoubleFaults++
		}

		// Track point winner
		if event.PointWinnerTeam == model.TeamA {
			teamAScore++
		} else {
			teamBScore++
		}

		// Track total points won per player (for the winning team)
		for playerID, team := range playerTeamMap {
			if team == event.PointWinnerTeam {
				stats[playerID].TotalPointsWon++
			}
		}
	}

	// Convert stats map to slice
	var playerStats []model.PlayerMatchStats
	for _, stat := range stats {
		playerStats = append(playerStats, *stat)
	}

	// Compute games and sets by replaying events through scoring logic
	gamesA, gamesB, setsA, setsB := computeGamesAndSets(events)

	return &model.MatchSummary{
		MatchID:     matchID,
		Venue:       *venue,
		MatchType:   match.MatchType,
		StartedAt:   match.StartedAt,
		EndedAt:     match.EndedAt,
		TeamAScore:  teamAScore,
		TeamBScore:  teamBScore,
		GamesA:      gamesA,
		GamesB:      gamesB,
		SetsA:       setsA,
		SetsB:       setsB,
		PlayerStats: playerStats,
	}, nil
}

// computeGamesAndSets replays point events to calculate games and sets won.
// Uses simplified tennis scoring logic to derive game/set counts from raw points.
func computeGamesAndSets(events []model.PointEvent) (gamesA, gamesB, setsA, setsB int) {
	// Track points in current game
	pointsA := 0
	pointsB := 0
	
	// Track games in current set
	gamesInSetA := 0
	gamesInSetB := 0
	
	for _, event := range events {
		// Count point
		if event.PointWinnerTeam == model.TeamA {
			pointsA++
		} else {
			pointsB++
		}
		
		// Check if game is won (simplified: 4 points with 2 point lead)
		gameWonByA := pointsA >= 4 && pointsA-pointsB >= 2
		gameWonByB := pointsB >= 4 && pointsB-pointsA >= 2
		
		if gameWonByA {
			gamesA++
			gamesInSetA++
			pointsA = 0
			pointsB = 0
			
			// Check if set is won (6 games with 2 game lead, or 7-5, or tiebreak 7-6)
			if (gamesInSetA >= 6 && gamesInSetA-gamesInSetB >= 2) || gamesInSetA == 7 {
				setsA++
				gamesInSetA = 0
				gamesInSetB = 0
			}
		} else if gameWonByB {
			gamesB++
			gamesInSetB++
			pointsA = 0
			pointsB = 0
			
			// Check if set is won
			if (gamesInSetB >= 6 && gamesInSetB-gamesInSetA >= 2) || gamesInSetB == 7 {
				setsB++
				gamesInSetA = 0
				gamesInSetB = 0
			}
		}
	}
	
	return gamesA, gamesB, setsA, setsB
}

// DeleteMatch removes a match.
func (s *MatchService) DeleteMatch(ctx context.Context, matchID uuid.UUID) error {
	return s.matchRepo.Delete(ctx, matchID)
}
