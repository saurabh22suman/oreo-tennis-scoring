package scoring

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - SHORT-FORMAT MODE
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tennis_Scoring_Spec.md Section 5
// This file implements the recreational 3-game "best of 3" format.
//
// Hierarchy: POINT → GAME → MATCH (no sets)
//
// Rules:
//   - Maximum 3 games
//   - First to win 2 games wins the match
//   - If a side wins Games 1 and 2, Game 3 is skipped
//   - Fixed serving order: Game 1 = Server[0], Game 2 = Server[1], Game 3 = Server[2]
//   - Server does NOT depend on previous game outcome
// ═══════════════════════════════════════════════════════════════════════════

// handleShortFormatGameWon handles game completion in short-format mode.
//
// Flow:
//  1. Increment games won for the winning team
//  2. Check if match is won (first to 2 games)
//  3. If match won, set winner and mark completed
//  4. If not won, advance to next game
//
// Match Win Condition:
//   - First team to win 2 games wins the match
//
// Server Rotation:
//   - Game N is served by Servers[N-1]
//   - Independent of game outcomes
func handleShortFormatGameWon(state *MatchState, winner Team) *MatchState {
	// Increment games won
	if winner == TeamA {
		state.GamesA++
	} else {
		state.GamesB++
	}

	// Check match win condition: First to 2 games
	if state.GamesA == 2 {
		// Team A wins match
		a := TeamA
		state.Winner = &a
		state.Completed = true
		return state
	}

	if state.GamesB == 2 {
		// Team B wins match
		b := TeamB
		state.Winner = &b
		state.Completed = true
		return state
	}

	// Match not won yet - advance to next game
	advanceToNextGame(state)

	return state
}

// advanceToNextGame sets up the next game in short-format mode.
//
// Actions:
//   - Reset points to 0-0
//   - Increment game number
//   - Increment server index (to use next server in rotation)
func advanceToNextGame(state *MatchState) {
	state.CurrentGame.PointsA = 0
	state.CurrentGame.PointsB = 0
	state.CurrentGame.GameNumber++
	state.CurrentGame.ServerIndex++

	// Safety check: Ensure we don't exceed 3 games
	// (This should never happen if match win logic is correct)
	if state.CurrentGame.GameNumber > 3 {
		state.CurrentGame.GameNumber = 3
	}

	if state.CurrentGame.ServerIndex >= len(state.Servers) {
		// Clamp to last server if we somehow go beyond
		// (Should never happen in valid short-format)
		state.CurrentGame.ServerIndex = len(state.Servers) - 1
	}
}

// GetCurrentServer returns the current server's ID for short-format mode.
//
// Returns:
//   - Server ID if in short-format mode and server is defined
//   - Empty string otherwise
func GetCurrentServer(state *MatchState) string {
	if state.Mode != ModeShortFormat {
		return ""
	}

	if state.Servers == nil || state.CurrentGame.ServerIndex >= len(state.Servers) {
		return ""
	}

	return state.Servers[state.CurrentGame.ServerIndex]
}
