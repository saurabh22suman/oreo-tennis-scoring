package tournament

import (
	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - ROUND ROBIN MATCH GENERATION
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md Section 4
// This file implements round-robin match generation.
//
// Round Robin Definition:
//   Every team plays every other team exactly once
//
// Formula (Section 4.2):
//   total_matches = T × (T − 1) / 2
//
// Examples:
//   3 teams → 3 matches
//   4 teams → 6 matches
//   5 teams → 10 matches
// ═══════════════════════════════════════════════════════════════════════════

// GenerateRoundRobinMatches creates all matches for round-robin stage.
//
// Algorithm:
//
//	For each team T1:
//	  For each team T2 where T2 comes after T1:
//	    Create match: T1 vs T2
//
// This ensures:
//   - Every team plays every other team exactly once
//   - No duplicate matches
//   - No team plays itself
//
// Parameters:
//   - tournamentID: Parent tournament identifier
//   - teams: All teams participating in round-robin
//
// Returns:
//   - Slice of Match structs (not yet played)
//   - Match order is sequential but can be rearranged
func GenerateRoundRobinMatches(tournamentID uuid.UUID, teams []Team) []Match {
	numTeams := len(teams)

	// Calculate expected number of matches
	// Formula: T × (T - 1) / 2
	expectedMatches := numTeams * (numTeams - 1) / 2

	matches := make([]Match, 0, expectedMatches)
	matchOrder := 1

	// Generate all pairwise combinations
	for i := 0; i < numTeams; i++ {
		for j := i + 1; j < numTeams; j++ {
			match := Match{
				ID:             uuid.New(),
				TournamentID:   tournamentID,
				TeamAID:        teams[i].ID,
				TeamBID:        teams[j].ID,
				Stage:          StageRR,
				MatchOrder:     matchOrder,
				ScoringMatchID: nil,
				WinnerTeamID:   nil,
				Completed:      false,
			}

			matches = append(matches, match)
			matchOrder++
		}
	}

	return matches
}

// GetMatchByID finds a match by its ID.
func GetMatchByID(matches []Match, matchID uuid.UUID) (*Match, int) {
	for i := range matches {
		if matches[i].ID == matchID {
			return &matches[i], i
		}
	}
	return nil, -1
}

// GetMatchesByStage filters matches by stage.
func GetMatchesByStage(matches []Match, stage MatchStage) []Match {
	filtered := make([]Match, 0)
	for _, match := range matches {
		if match.Stage == stage {
			filtered = append(filtered, match)
		}
	}
	return filtered
}

// GetCompletedMatches returns all completed matches.
func GetCompletedMatches(matches []Match) []Match {
	completed := make([]Match, 0)
	for _, match := range matches {
		if match.Completed {
			completed = append(completed, match)
		}
	}
	return completed
}

// GetPendingMatches returns all incomplete matches.
func GetPendingMatches(matches []Match) []Match {
	pending := make([]Match, 0)
	for _, match := range matches {
		if !match.Completed {
			pending = append(pending, match)
		}
	}
	return pending
}

// IsRoundRobinComplete checks if all round-robin matches are finished.
func IsRoundRobinComplete(matches []Match) bool {
	rrMatches := GetMatchesByStage(matches, StageRR)
	for _, match := range rrMatches {
		if !match.Completed {
			return false
		}
	}
	return len(rrMatches) > 0
}
