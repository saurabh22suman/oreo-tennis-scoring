package tournament

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - TEAM GENERATION
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md Section 3
// This file implements team creation for doubles tournaments.
//
// Two modes supported:
//   1. Random: Shuffle players and pair sequentially
//   2. Manual: User provides explicit pairs
// ═══════════════════════════════════════════════════════════════════════════

// GenerateRandomTeams creates doubles teams by shuffling and pairing players.
//
// Algorithm (per spec Section 3.3.A):
//  1. Shuffle player list randomly
//  2. Pair players sequentially
//
// Example:
//
//	Input:  [P1, P2, P3, P4, P5, P6]
//	Shuffle: [P4, P1, P6, P2, P5, P3]
//	Teams:
//	  T1 = P4 + P1
//	  T2 = P6 + P2
//	  T3 = P5 + P3
//
// Parameters:
//   - playerIDs: All players to be assigned to teams
//   - seed: Random seed for reproducibility (use time.Now().UnixNano() for random)
//
// Returns:
//   - Slice of Team structs
//   - Error if player count is odd or less than 4
func GenerateRandomTeams(playerIDs []uuid.UUID, seed int64) ([]Team, error) {
	// Validation
	if len(playerIDs) < 4 {
		return nil, errors.New("minimum 4 players required for tournament")
	}

	if len(playerIDs)%2 != 0 {
		return nil, errors.New("player count must be even for doubles")
	}

	// Create a random number generator with the given seed
	rng := rand.New(rand.NewSource(seed))

	// Shuffle player list (Fisher-Yates shuffle)
	shuffled := make([]uuid.UUID, len(playerIDs))
	copy(shuffled, playerIDs)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Pair players sequentially
	numTeams := len(shuffled) / 2
	teams := make([]Team, numTeams)

	for i := 0; i < numTeams; i++ {
		teams[i] = Team{
			ID:         uuid.New(),
			Player1ID:  shuffled[i*2],
			Player2ID:  shuffled[i*2+1],
			TeamNumber: i + 1,
		}
	}

	return teams, nil
}

// GenerateManualTeams creates doubles teams from explicit player pairs.
//
// Validation (per spec Section 3.3.B):
//   - Each team has exactly 2 players
//   - No player appears in more than one team
//
// Parameters:
//   - pairs: Slice of player ID pairs [[P1, P2], [P3, P4], ...]
//
// Returns:
//   - Slice of Team structs
//   - Error if validation fails
func GenerateManualTeams(pairs [][2]uuid.UUID) ([]Team, error) {
	if len(pairs) < 2 {
		return nil, errors.New("minimum 2 teams required for tournament")
	}

	// Track players to detect duplicates
	playerSet := make(map[uuid.UUID]bool)
	teams := make([]Team, len(pairs))

	for i, pair := range pairs {
		// Check for nil UUIDs
		if pair[0] == uuid.Nil || pair[1] == uuid.Nil {
			return nil, fmt.Errorf("team %d has invalid player ID", i+1)
		}

		// Check for same player twice in same team
		if pair[0] == pair[1] {
			return nil, fmt.Errorf("team %d has same player twice", i+1)
		}

		// Check if player already assigned
		if playerSet[pair[0]] {
			return nil, fmt.Errorf("player %s appears in multiple teams", pair[0])
		}
		if playerSet[pair[1]] {
			return nil, fmt.Errorf("player %s appears in multiple teams", pair[1])
		}

		// Mark players as assigned
		playerSet[pair[0]] = true
		playerSet[pair[1]] = true

		// Create team
		teams[i] = Team{
			ID:         uuid.New(),
			Player1ID:  pair[0],
			Player2ID:  pair[1],
			TeamNumber: i + 1,
		}
	}

	return teams, nil
}

// ValidateTeams ensures teams are valid for tournament play.
//
// Checks:
//   - Minimum 2 teams (for Final only)
//   - All teams have exactly 2 players
//   - No duplicate teams
func ValidateTeams(teams []Team) error {
	if len(teams) < 2 {
		return errors.New("minimum 2 teams required for tournament")
	}

	// Check for duplicate team IDs
	teamSet := make(map[uuid.UUID]bool)
	playerSet := make(map[uuid.UUID]bool)

	for _, team := range teams {
		// Check team ID
		if team.ID == uuid.Nil {
			return errors.New("team has invalid ID")
		}

		if teamSet[team.ID] {
			return fmt.Errorf("duplicate team ID: %s", team.ID)
		}
		teamSet[team.ID] = true

		// Check player IDs
		if team.Player1ID == uuid.Nil || team.Player2ID == uuid.Nil {
			return fmt.Errorf("team %d has invalid player ID", team.TeamNumber)
		}

		if team.Player1ID == team.Player2ID {
			return fmt.Errorf("team %d has same player twice", team.TeamNumber)
		}

		// Check for players in multiple teams
		if playerSet[team.Player1ID] {
			return fmt.Errorf("player %s appears in multiple teams", team.Player1ID)
		}
		if playerSet[team.Player2ID] {
			return fmt.Errorf("player %s appears in multiple teams", team.Player2ID)
		}

		playerSet[team.Player1ID] = true
		playerSet[team.Player2ID] = true
	}

	return nil
}

// GetTeamByID finds a team by its ID.
func GetTeamByID(teams []Team, teamID uuid.UUID) (*Team, error) {
	for i := range teams {
		if teams[i].ID == teamID {
			return &teams[i], nil
		}
	}
	return nil, fmt.Errorf("team not found: %s", teamID)
}
