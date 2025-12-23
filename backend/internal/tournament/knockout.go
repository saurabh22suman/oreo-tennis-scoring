package tournament

import (
	"errors"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - KNOCKOUT STAGE
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md Section 6
// This file implements knockout stage (semifinals + final) logic.
//
// Advancement Rules (Section 6.1):
//
// Case A: 3 Teams
//   - Top 2 teams advance
//   - Final only: Rank 1 vs Rank 2
//
// Case B: 4 Teams
//   - All 4 advance
//   - SF1: Rank 1 vs Rank 4
//   - SF2: Rank 2 vs Rank 3
//   - Winners → Final
//
// Case C: 5+ Teams
//   - Top 4 advance to semifinals
//   - Remaining teams eliminated
//   - Same bracket as Case B
// ═══════════════════════════════════════════════════════════════════════════

// GenerateKnockoutMatches creates semifinals and/or final based on team count.
//
// Logic:
//   - 3 teams → Final only
//   - 4+ teams → Semifinals + Final
//
// Returns:
//   - Slice of knockout matches
//   - Error if standings incomplete or invalid
func GenerateKnockoutMatches(tournamentID uuid.UUID, standings []TeamStanding) ([]Match, error) {
	// Validate standings are complete
	if !IsStandingsComplete(standings) {
		return nil, errors.New("cannot generate knockout: round-robin not complete")
	}

	// Get ranked teams
	ranked := CalculateRankings(standings)
	numTeams := len(ranked)

	if numTeams < 2 {
		return nil, errors.New("minimum 2 teams required for knockout")
	}

	matches := make([]Match, 0)

	// Case A: 3 Teams → Final only
	if numTeams == 3 {
		final := createFinalMatch(tournamentID, ranked[0].TeamID, ranked[1].TeamID)
		matches = append(matches, final)
		return matches, nil
	}

	// Case B/C: 4+ Teams → Semifinals + Final
	// Take top 4 teams
	top4 := GetTopTeams(ranked, 4)

	// Semifinal 1: Rank 1 vs Rank 4
	sf1 := createSemifinalMatch(tournamentID, top4[0].TeamID, top4[3].TeamID, 1)

	// Semifinal 2: Rank 2 vs Rank 3
	sf2 := createSemifinalMatch(tournamentID, top4[1].TeamID, top4[2].TeamID, 2)

	// Final: Winners of SF1 and SF2 (to be determined)
	final := createPendingFinalMatch(tournamentID)

	matches = append(matches, sf1, sf2, final)
	return matches, nil
}

// UpdateFinalMatchup sets the teams for the final after semifinals complete.
//
// Parameters:
//   - matches: All knockout matches
//   - sf1Winner: Winner of Semifinal 1
//   - sf2Winner: Winner of Semifinal 2
//
// Returns:
//   - Updated matches with final matchup set
func UpdateFinalMatchup(matches []Match, sf1Winner, sf2Winner uuid.UUID) ([]Match, error) {
	updated := make([]Match, len(matches))
	copy(updated, matches)

	// Find the final match
	for i := range updated {
		if updated[i].Stage == StageFinal {
			updated[i].TeamAID = sf1Winner
			updated[i].TeamBID = sf2Winner
			return updated, nil
		}
	}

	return nil, errors.New("final match not found")
}

// AreSemifinalsComplete checks if both semifinals are finished.
func AreSemifinalsComplete(matches []Match) bool {
	semis := GetMatchesByStage(matches, StageSemi)

	if len(semis) == 0 {
		// No semifinals (3-team tournament)
		return true
	}

	for _, match := range semis {
		if !match.Completed {
			return false
		}
	}

	return len(semis) == 2 // Must have exactly 2 semifinals
}

// GetSemifinalWinners retrieves the winners of both semifinals.
//
// Returns:
//   - Winner of SF1
//   - Winner of SF2
//   - Error if semifinals not complete
func GetSemifinalWinners(matches []Match) (uuid.UUID, uuid.UUID, error) {
	semis := GetMatchesByStage(matches, StageSemi)

	if len(semis) != 2 {
		return uuid.Nil, uuid.Nil, errors.New("expected exactly 2 semifinals")
	}

	var sf1Winner, sf2Winner uuid.UUID

	for _, match := range semis {
		if !match.Completed {
			return uuid.Nil, uuid.Nil, errors.New("semifinals not complete")
		}

		if match.WinnerTeamID == nil {
			return uuid.Nil, uuid.Nil, errors.New("semifinal missing winner")
		}

		// SF1 has MatchOrder = 1, SF2 has MatchOrder = 2
		if match.MatchOrder == 1 {
			sf1Winner = *match.WinnerTeamID
		} else {
			sf2Winner = *match.WinnerTeamID
		}
	}

	if sf1Winner == uuid.Nil || sf2Winner == uuid.Nil {
		return uuid.Nil, uuid.Nil, errors.New("could not determine semifinal winners")
	}

	return sf1Winner, sf2Winner, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPER FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────

func createFinalMatch(tournamentID, teamA, teamB uuid.UUID) Match {
	return Match{
		ID:             uuid.New(),
		TournamentID:   tournamentID,
		TeamAID:        teamA,
		TeamBID:        teamB,
		Stage:          StageFinal,
		MatchOrder:     1,
		ScoringMatchID: nil,
		WinnerTeamID:   nil,
		Completed:      false,
	}
}

func createSemifinalMatch(tournamentID, teamA, teamB uuid.UUID, order int) Match {
	return Match{
		ID:             uuid.New(),
		TournamentID:   tournamentID,
		TeamAID:        teamA,
		TeamBID:        teamB,
		Stage:          StageSemi,
		MatchOrder:     order,
		ScoringMatchID: nil,
		WinnerTeamID:   nil,
		Completed:      false,
	}
}

func createPendingFinalMatch(tournamentID uuid.UUID) Match {
	return Match{
		ID:             uuid.New(),
		TournamentID:   tournamentID,
		TeamAID:        uuid.Nil, // To be determined
		TeamBID:        uuid.Nil, // To be determined
		Stage:          StageFinal,
		MatchOrder:     3,
		ScoringMatchID: nil,
		WinnerTeamID:   nil,
		Completed:      false,
	}
}
