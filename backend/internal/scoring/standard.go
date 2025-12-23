package scoring

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - STANDARD MODE
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tennis_Scoring_Spec.md Section 4
// This file implements traditional tennis scoring.
//
// Hierarchy: POINT → GAME → SET → MATCH
//
// Rules:
//   - Set Win: Games ≥ 6 with lead ≥ 2
//   - Tie-Break: Triggered at 6-6 (winner gets 7-6)
//   - Match Win: Best of 3 sets (first to 2 sets)
// ═══════════════════════════════════════════════════════════════════════════

// handleStandardGameWon handles game completion in standard tennis mode.
//
// Flow:
//  1. Increment games won in current set
//  2. Check if set is won
//  3. If set won, increment sets and check match win
//  4. If set won, start new set OR complete match
//  5. If set not won, start next game in same set
//
// Set Win Condition:
//   - Games ≥ 6 AND lead by ≥ 2 games
//   - OR win tie-break at 6-6 (resulting in 7-6)
//
// Match Win Condition:
//   - Best of 3 sets: First to 2 sets wins
func handleStandardGameWon(state *MatchState, winner Team) *MatchState {
	// Increment games in current set
	if winner == TeamA {
		state.GamesA++
	} else {
		state.GamesB++
	}

	// Check if set is won
	setWinner := IsSetWon(state.GamesA, state.GamesB)

	if setWinner != nil {
		// Set is won - handle set completion
		handleSetWon(state, *setWinner)
	} else {
		// Set continues - start next game
		startNextGameInSet(state)
	}

	return state
}

// handleSetWon handles the completion of a set.
//
// Actions:
//  1. Increment sets won for the winning team
//  2. Check if match is won (first to 2 sets)
//  3. If match won, mark as completed
//  4. If not, start a new set
func handleSetWon(state *MatchState, winner Team) {
	// Increment sets won
	if winner == TeamA {
		state.SetsA++
	} else {
		state.SetsB++
	}

	// Check match win condition: Best of 3 sets
	if state.SetsA == 2 {
		// Team A wins match
		a := TeamA
		state.Winner = &a
		state.Completed = true
		return
	}

	if state.SetsB == 2 {
		// Team B wins match
		b := TeamB
		state.Winner = &b
		state.Completed = true
		return
	}

	// Match not won - start new set
	startNewSet(state)
}

// startNewSet initializes a new set.
//
// Actions:
//   - Reset games to 0-0
//   - Increment set number
//   - Reset game points to 0-0
//   - Reset game number
func startNewSet(state *MatchState) {
	state.CurrentSet++
	state.GamesA = 0
	state.GamesB = 0

	state.CurrentGame.PointsA = 0
	state.CurrentGame.PointsB = 0
	state.CurrentGame.GameNumber = 1
}

// startNextGameInSet starts the next game within the current set.
//
// Actions:
//   - Reset points to 0-0
//   - Increment game number
//
// Note: In standard mode, server rotation would typically alternate
// based on odd/even game numbers, but that's handled by the UI/event layer,
// not the scoring engine.
func startNextGameInSet(state *MatchState) {
	state.CurrentGame.PointsA = 0
	state.CurrentGame.PointsB = 0
	state.CurrentGame.GameNumber = state.GamesA + state.GamesB + 1
}

// GetSetScore returns the current set score for standard mode.
//
// Returns:
//   - Sets won by each team
//   - Games won in current set by each team
func GetSetScore(state *MatchState) (setsA, setsB, gamesA, gamesB int) {
	if state.Mode != ModeStandard {
		return 0, 0, 0, 0
	}

	return state.SetsA, state.SetsB, state.GamesA, state.GamesB
}
