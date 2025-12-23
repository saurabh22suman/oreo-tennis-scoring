package tournament

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - MAIN ORCHESTRATOR
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md
// This file orchestrates the complete tournament lifecycle.
//
// Tournament Flow:
//   Players → Teams → Round Robin → Knockout → Winner
//
// Stages:
//   1. Setup: Create tournament, generate teams
//   2. Round Robin: All teams play all teams
//   3. Knockout: Top teams compete in semis/final
//   4. Completed: Winner declared
// ═══════════════════════════════════════════════════════════════════════════

// NewTournament creates a new tournament in setup stage.
//
// Parameters:
//   - venueID: Where tournament is played
//   - playerIDs: All participating players
//
// Returns:
//   - TournamentState initialized in Setup stage
//   - Error if validation fails
func NewTournament(venueID uuid.UUID, playerIDs []uuid.UUID) (*TournamentState, error) {
	// Validation
	if venueID == uuid.Nil {
		return nil, errors.New("venue ID is required")
	}

	if len(playerIDs) < 4 {
		return nil, errors.New("minimum 4 players required for tournament")
	}

	if len(playerIDs)%2 != 0 {
		return nil, errors.New("player count must be even for doubles")
	}

	// Create tournament
	tournament := &TournamentState{
		ID:                uuid.New(),
		VenueID:           venueID,
		PlayerIDs:         playerIDs,
		Teams:             nil, // Teams created separately
		Stage:             StageSetup,
		RoundRobinMatches: nil,
		Standings:         nil,
		KnockoutMatches:   nil,
		Winner:            nil,
		Completed:         false,
		CreatedAt:         time.Now(),
	}

	return tournament, nil
}

// SetTeams assigns teams to the tournament and advances to round-robin stage.
//
// This must be called after NewTournament and before starting matches.
//
// Parameters:
//   - state: Current tournament state
//   - teams: Generated teams (from random or manual creation)
//
// Returns:
//   - Updated tournament state with teams set and matches generated
//   - Error if validation fails
func SetTeams(state *TournamentState, teams []Team) (*TournamentState, error) {
	// Validate current stage
	if state.Stage != StageSetup {
		return nil, errors.New("teams can only be set during setup stage")
	}

	// Validate teams
	if err := ValidateTeams(teams); err != nil {
		return nil, err
	}

	// Create new state (immutable update)
	newState := copyTournamentState(state)

	// Set teams
	newState.Teams = teams

	// Initialize standings
	newState.Standings = InitializeStandings(teams)

	// Generate round-robin matches
	newState.RoundRobinMatches = GenerateRoundRobinMatches(newState.ID, teams)

	// Advance to round-robin stage
	newState.Stage = StageRoundRobin

	return newState, nil
}

// RecordMatchResult records the result of a match and updates standings.
//
// This is called after each match completion.
//
// Parameters:
//   - state: Current tournament state
//   - result: Match result from scoring engine
//
// Returns:
//   - Updated tournament state with standings updated
//   - Error if match not found or already completed
func RecordMatchResult(state *TournamentState, result MatchResult) (*TournamentState, error) {
	// Create new state
	newState := copyTournamentState(state)

	// Find and update match in appropriate stage
	var matchFound bool

	// Check round-robin matches
	if newState.Stage == StageRoundRobin {
		for i := range newState.RoundRobinMatches {
			if newState.RoundRobinMatches[i].ID == result.MatchID {
				if newState.RoundRobinMatches[i].Completed {
					return nil, errors.New("match already completed")
				}

				newState.RoundRobinMatches[i].WinnerTeamID = &result.WinnerTeamID
				newState.RoundRobinMatches[i].Completed = true
				matchFound = true
				break
			}
		}

		if matchFound {
			// Update standings
			newState.Standings = UpdateStandingsWithResult(newState.Standings, result)

			// Check if round-robin complete
			if IsRoundRobinComplete(newState.RoundRobinMatches) {
				// Ready to advance to knockout
				// (Actual advancement is a separate call)
			}
		}
	}

	// Check knockout matches
	if newState.Stage == StageKnockout {
		for i := range newState.KnockoutMatches {
			if newState.KnockoutMatches[i].ID == result.MatchID {
				if newState.KnockoutMatches[i].Completed {
					return nil, errors.New("match already completed")
				}

				newState.KnockoutMatches[i].WinnerTeamID = &result.WinnerTeamID
				newState.KnockoutMatches[i].Completed = true
				matchFound = true

				// Check if it's the final
				if newState.KnockoutMatches[i].Stage == StageFinal {
					// Tournament complete!
					newState.Winner = &result.WinnerTeamID
					newState.Completed = true
					newState.Stage = StageCompleted
				}

				break
			}
		}
	}

	if !matchFound {
		return nil, errors.New("match not found in tournament")
	}

	return newState, nil
}

// AdvanceToKnockout transitions from round-robin to knockout stage.
//
// Must be called after all round-robin matches are complete.
//
// Parameters:
//   - state: Current tournament state
//
// Returns:
//   - Updated tournament state in knockout stage
//   - Error if round-robin not complete
func AdvanceToKnockout(state *TournamentState) (*TournamentState, error) {
	// Validate current stage
	if state.Stage != StageRoundRobin {
		return nil, errors.New("can only advance to knockout from round-robin stage")
	}

	// Validate round-robin complete
	if !IsRoundRobinComplete(state.RoundRobinMatches) {
		return nil, errors.New("cannot advance: round-robin not complete")
	}

	// Create new state
	newState := copyTournamentState(state)

	// Calculate final rankings
	newState.Standings = CalculateRankings(newState.Standings)

	// Generate knockout matches
	knockoutMatches, err := GenerateKnockoutMatches(newState.ID, newState.Standings)
	if err != nil {
		return nil, err
	}

	newState.KnockoutMatches = knockoutMatches
	newState.Stage = StageKnockout

	return newState, nil
}

// PrepareFinal sets up the final match after semifinals complete.
//
// Only needed for 4+ team tournaments.
//
// Parameters:
//   - state: Current tournament state
//
// Returns:
//   - Updated state with final matchup set
//   - Error if semifinals not complete
func PrepareFinal(state *TournamentState) (*TournamentState, error) {
	if state.Stage != StageKnockout {
		return nil, errors.New("not in knockout stage")
	}

	if !AreSemifinalsComplete(state.KnockoutMatches) {
		return nil, errors.New("semifinals not complete")
	}

	// Get semifinal winners
	sf1Winner, sf2Winner, err := GetSemifinalWinners(state.KnockoutMatches)
	if err != nil {
		return nil, err
	}

	// Create new state
	newState := copyTournamentState(state)

	// Update final matchup
	newState.KnockoutMatches, err = UpdateFinalMatchup(newState.KnockoutMatches, sf1Winner, sf2Winner)
	if err != nil {
		return nil, err
	}

	return newState, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// QUERY FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────

// GetNextMatch returns the next match to be played.
func GetNextMatch(state *TournamentState) *Match {
	var matches []Match

	if state.Stage == StageRoundRobin {
		matches = state.RoundRobinMatches
	} else if state.Stage == StageKnockout {
		matches = state.KnockoutMatches
	} else {
		return nil
	}

	// Return first incomplete match
	for i := range matches {
		if !matches[i].Completed {
			return &matches[i]
		}
	}

	return nil
}

// GetAllMatches returns all matches in the tournament.
func GetAllMatches(state *TournamentState) []Match {
	all := make([]Match, 0)
	all = append(all, state.RoundRobinMatches...)
	all = append(all, state.KnockoutMatches...)
	return all
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPER FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────

func copyTournamentState(state *TournamentState) *TournamentState {
	newState := *state

	// Deep copy slices
	newState.PlayerIDs = make([]uuid.UUID, len(state.PlayerIDs))
	copy(newState.PlayerIDs, state.PlayerIDs)

	if state.Teams != nil {
		newState.Teams = make([]Team, len(state.Teams))
		copy(newState.Teams, state.Teams)
	}

	if state.RoundRobinMatches != nil {
		newState.RoundRobinMatches = make([]Match, len(state.RoundRobinMatches))
		copy(newState.RoundRobinMatches, state.RoundRobinMatches)
	}

	if state.Standings != nil {
		newState.Standings = make([]TeamStanding, len(state.Standings))
		copy(newState.Standings, state.Standings)
	}

	if state.KnockoutMatches != nil {
		newState.KnockoutMatches = make([]Match, len(state.KnockoutMatches))
		copy(newState.KnockoutMatches, state.KnockoutMatches)
	}

	return &newState
}
