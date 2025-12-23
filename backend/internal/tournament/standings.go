package tournament

import (
	"sort"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - STANDINGS CALCULATION
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md Section 5
// This file implements tournament standings tracking and ranking.
//
// Stats Tracked (Section 5.1):
//   - played: Matches played
//   - won: Matches won
//   - lost: Matches lost
//   - points: Tournament points
//
// Points System (Section 5.2):
//   - Win → 1 point
//   - Loss → 0 points
//   - No draws
//
// Ranking Rules (Section 5.3):
//   1. Points (descending)
//   2. Head-to-head result
//   3. Games difference (optional)
// ═══════════════════════════════════════════════════════════════════════════

// InitializeStandings creates initial standings for all teams.
//
// All teams start with:
//   - Played: 0
//   - Won: 0
//   - Lost: 0
//   - Points: 0
//   - Rank: Unassigned (0)
func InitializeStandings(teams []Team) []TeamStanding {
	standings := make([]TeamStanding, len(teams))

	for i, team := range teams {
		standings[i] = TeamStanding{
			TeamID: team.ID,
			Rank:   0,
			Played: 0,
			Won:    0,
			Lost:   0,
			Points: 0,
		}
	}

	return standings
}

// UpdateStandingsWithResult updates standings after a match result.
//
// Updates:
//   - Winner: played +1, won +1, points +1
//   - Loser: played +1, lost +1, points +0
//
// Returns updated standings (immutable operation).
func UpdateStandingsWithResult(standings []TeamStanding, result MatchResult) []TeamStanding {
	// Copy standings (immutable update)
	updated := make([]TeamStanding, len(standings))
	copy(updated, standings)

	// Find winner and loser in standings
	for i := range updated {
		if updated[i].TeamID == result.WinnerTeamID {
			// Winner gets: +1 played, +1 won, +1 point
			updated[i].Played++
			updated[i].Won++
			updated[i].Points++
		} else if updated[i].TeamID == result.LoserTeamID {
			// Loser gets: +1 played, +1 lost, +0 points
			updated[i].Played++
			updated[i].Lost++
		}
	}

	return updated
}

// CalculateRankings sorts standings and assigns ranks.
//
// Ranking Rules (per spec Section 5.3):
//  1. Points (descending)
//  2. If tied, head-to-head result (requires match history)
//  3. If still tied, games difference (optional, not implemented)
//
// For simplicity in Phase 1: Rank by points only.
// Head-to-head can be added later by passing match results.
//
// Returns standings sorted by rank.
func CalculateRankings(standings []TeamStanding) []TeamStanding {
	// Copy standings
	ranked := make([]TeamStanding, len(standings))
	copy(ranked, standings)

	// Sort by points (descending)
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Points > ranked[j].Points
	})

	// Assign ranks
	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	return ranked
}

// GetStandingByTeamID retrieves a team's standing.
func GetStandingByTeamID(standings []TeamStanding, teamID uuid.UUID) *TeamStanding {
	for i := range standings {
		if standings[i].TeamID == teamID {
			return &standings[i]
		}
	}
	return nil
}

// GetTopTeams returns the top N teams by rank.
func GetTopTeams(standings []TeamStanding, n int) []TeamStanding {
	if n > len(standings) {
		n = len(standings)
	}

	// Ensure standings are ranked
	ranked := CalculateRankings(standings)

	return ranked[:n]
}

// IsStandingsComplete checks if all teams have played expected matches.
//
// In round-robin, each team plays (T-1) matches where T = total teams.
func IsStandingsComplete(standings []TeamStanding) bool {
	if len(standings) == 0 {
		return false
	}

	expectedMatches := len(standings) - 1

	for _, standing := range standings {
		if standing.Played != expectedMatches {
			return false
		}
	}

	return true
}
