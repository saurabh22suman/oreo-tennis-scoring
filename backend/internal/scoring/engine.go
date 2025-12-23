package scoring

import (
	"errors"
	"fmt"
)

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - CORE STATE MACHINE
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tennis_Scoring_Spec.md
// This is the main scoring engine that orchestrates match progression.
// All functions are PURE - they return new state without mutation.
// ═══════════════════════════════════════════════════════════════════════════

// NewMatchState creates a new tennis match with the specified mode and players.
//
// Parameters:
//   - mode: ModeStandard or ModeShortFormat
//   - players: Team assignments for all players
//   - servers: For short-format only, exactly 3 server IDs in order.
//     For standard mode, pass nil.
//
// Validation:
//   - Short-format requires exactly 3 servers
//   - Standard mode must have nil servers
//   - Teams must have players assigned
//
// Returns a new MatchState initialized to the start of the match.
func NewMatchState(mode MatchMode, players TeamPlayers, servers []string) (*MatchState, error) {
	// Validate mode
	if mode != ModeStandard && mode != ModeShortFormat {
		return nil, fmt.Errorf("invalid match mode: %s", mode)
	}

	// Validate players
	if len(players.TeamA) == 0 || len(players.TeamB) == 0 {
		return nil, errors.New("both teams must have at least one player")
	}

	// Validate servers based on mode
	if mode == ModeShortFormat {
		if len(servers) != 3 {
			return nil, errors.New("short-format mode requires exactly 3 servers")
		}
	} else {
		if servers != nil {
			return nil, errors.New("standard mode must not specify servers array")
		}
	}

	// Initialize match state
	state := &MatchState{
		Mode:    mode,
		Players: players,
		Servers: servers,

		CurrentGame: CurrentGameState{
			PointsA:     0,
			PointsB:     0,
			GameNumber:  1,
			ServerIndex: 0,
		},

		GamesA: 0,
		GamesB: 0,

		SetsA:      0,
		SetsB:      0,
		CurrentSet: 1,

		Winner:    nil,
		Completed: false,
	}

	return state, nil
}

// ScorePoint awards a point to the specified team and updates match state.
//
// This is the MAIN scoring function. It:
//  1. Awards the point to the specified team
//  2. Checks if the game is won
//  3. If won, handles game completion (which may trigger set/match win)
//  4. Returns updated state
//
// The function is PURE - it returns a new MatchState without mutation.
//
// Parameters:
//   - state: Current match state
//   - team: Team that won the point (TeamA or TeamB)
//
// Returns:
//   - Updated match state
//   - Error if match is already completed or invalid team
func ScorePoint(state *MatchState, team Team) (*MatchState, error) {
	// Validate
	if state.Completed {
		return nil, errors.New("cannot score point: match is already completed")
	}

	if team != TeamA && team != TeamB {
		return nil, fmt.Errorf("invalid team: %s", team)
	}

	// Create new state (immutable update)
	newState := copyMatchState(state)

	// Award point
	if team == TeamA {
		newState.CurrentGame.PointsA++
	} else {
		newState.CurrentGame.PointsB++
	}

	// Check if game is won
	winner := IsGameWon(newState.CurrentGame.PointsA, newState.CurrentGame.PointsB)

	if winner != nil {
		// Game won - handle game completion
		return handleGameWon(newState, *winner)
	}

	// Game still in progress
	return newState, nil
}

// handleGameWon handles the completion of a game.
//
// Routes to the appropriate handler based on match mode:
//   - Standard: May trigger set win, which may trigger match win
//   - Short-Format: May trigger match win directly
func handleGameWon(state *MatchState, winner Team) (*MatchState, error) {
	if state.Mode == ModeShortFormat {
		return handleShortFormatGameWon(state, winner), nil
	}

	return handleStandardGameWon(state, winner), nil
}

// IsMatchComplete checks if the match is over.
func IsMatchComplete(state *MatchState) bool {
	return state.Completed
}

// GetWinner returns the winning team if match is complete, nil otherwise.
func GetWinner(state *MatchState) *Team {
	if state.Completed {
		return state.Winner
	}
	return nil
}

// copyMatchState creates a deep copy of MatchState for immutable updates.
func copyMatchState(state *MatchState) *MatchState {
	newState := *state

	// Copy slices to avoid shared references
	if state.Servers != nil {
		newState.Servers = make([]string, len(state.Servers))
		copy(newState.Servers, state.Servers)
	}

	newState.Players = TeamPlayers{
		TeamA: make([]string, len(state.Players.TeamA)),
		TeamB: make([]string, len(state.Players.TeamB)),
	}
	copy(newState.Players.TeamA, state.Players.TeamA)
	copy(newState.Players.TeamB, state.Players.TeamB)

	return &newState
}

// resetGameState resets the current game to 0-0.
func resetGameState(state *MatchState) {
	state.CurrentGame.PointsA = 0
	state.CurrentGame.PointsB = 0
}
