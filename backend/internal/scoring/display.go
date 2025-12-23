package scoring

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - DISPLAY LOGIC
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tennis_Scoring_Spec.md
// This file handles all display/presentation logic for tennis scoring.
// Converts internal state to user-facing tennis notation.
// ═══════════════════════════════════════════════════════════════════════════

// GetPointDisplay converts raw point count to tennis notation.
//
// Mapping (per spec):
//
//	0 → "0"
//	1 → "15"
//	2 → "30"
//	3+ → "40"
//
// IMPORTANT: Points never display beyond 40.
func GetPointDisplay(points int) string {
	switch points {
	case 0:
		return "0"
	case 1:
		return "15"
	case 2:
		return "30"
	default:
		// 3 or higher always shows as 40
		return "40"
	}
}

// GetGameState determines the current state of the game.
//
// Game Win Conditions (per spec):
//   - Points ≥ 4
//   - Lead by ≥ 2 points
//
// Deuce & Advantage Logic:
//   - Both at 40 (3+ points each) → Deuce or Advantage states
//   - From Deuce: Win point → Advantage
//   - From Advantage: Win point → Game, Lose point → Deuce
func GetGameState(pointsA, pointsB int) GameState {
	// Both sides at 40 or higher → Deuce/Advantage territory
	if pointsA >= 3 && pointsB >= 3 {
		diff := pointsA - pointsB

		switch diff {
		case 0:
			return GameDeuce
		case 1:
			return GameAdvantageA
		case -1:
			return GameAdvantageB
		default:
			// Difference of 2+ means game is won (handled by caller)
			return GameInProgress
		}
	}

	// Normal scoring (not in deuce territory)
	return GameInProgress
}

// IsGameWon checks if a game has been won by either team.
//
// Win Condition:
//   - Points ≥ 4
//   - Lead by ≥ 2 points
//
// Returns:
//   - nil: Game still in progress
//   - &TeamA: Team A won the game
//   - &TeamB: Team B won the game
func IsGameWon(pointsA, pointsB int) *Team {
	// Team A win condition
	if pointsA >= 4 && pointsA-pointsB >= 2 {
		a := TeamA
		return &a
	}

	// Team B win condition
	if pointsB >= 4 && pointsB-pointsA >= 2 {
		b := TeamB
		return &b
	}

	// Game still in progress
	return nil
}

// GetGameDisplayText returns the tennis notation for both teams' scores.
//
// Handles:
//   - Normal scoring: "0", "15", "30", "40"
//   - Deuce: Both show "Deuce"
//   - Advantage: Winner shows "Ad", loser shows "40"
func GetGameDisplayText(pointsA, pointsB int) PointDisplay {
	state := GetGameState(pointsA, pointsB)

	switch state {
	case GameDeuce:
		return PointDisplay{A: "Deuce", B: "Deuce"}

	case GameAdvantageA:
		return PointDisplay{A: "Ad", B: "40"}

	case GameAdvantageB:
		return PointDisplay{A: "40", B: "Ad"}

	default:
		// Normal scoring
		return PointDisplay{
			A: GetPointDisplay(pointsA),
			B: GetPointDisplay(pointsB),
		}
	}
}

// IsSetWon checks if a set has been won (standard mode only).
//
// Set Win Conditions (per spec):
//   - Games ≥ 6
//   - Lead by ≥ 2 games
//
// Tie-Break Handling:
//   - At 6-6, tie-break is played
//   - Winner of tie-break gets set at 7-6
//
// Returns:
//   - nil: Set still in progress
//   - &TeamA: Team A won the set
//   - &TeamB: Team B won the set
func IsSetWon(gamesA, gamesB int) *Team {
	// Normal set win: 6+ games with 2+ game lead
	if gamesA >= 6 && gamesA-gamesB >= 2 {
		a := TeamA
		return &a
	}

	if gamesB >= 6 && gamesB-gamesA >= 2 {
		b := TeamB
		return &b
	}

	// Tie-break win: 7-6
	if gamesA == 7 && gamesB == 6 {
		a := TeamA
		return &a
	}

	if gamesB == 7 && gamesA == 6 {
		b := TeamB
		return &b
	}

	// Set still in progress
	return nil
}

// IsTieBreak checks if the current game should be a tie-break.
//
// Tie-Break Trigger (per spec):
//   - Both teams at 6 games
//
// Note: Tie-break logic (first to 7 points, lead by 2) is handled
// separately as it's a different point-scoring system.
func IsTieBreak(gamesA, gamesB int) bool {
	return gamesA == 6 && gamesB == 6
}

// GetMatchDisplay returns the complete user-facing display of the match.
//
// This is the PRIMARY interface for UI rendering.
// It converts internal state to proper tennis notation.
func GetMatchDisplay(state *MatchState) MatchDisplay {
	// Get current point display (0, 15, 30, 40, Deuce, Ad)
	pointDisplay := GetGameDisplayText(
		state.CurrentGame.PointsA,
		state.CurrentGame.PointsB,
	)

	// Build base display
	display := MatchDisplay{
		Points:     pointDisplay,
		Games:      ScoreCount{A: state.GamesA, B: state.GamesB},
		CurrentSet: state.CurrentSet,
		GameNumber: state.CurrentGame.GameNumber,
	}

	// Mode-specific additions
	if state.Mode == ModeShortFormat {
		// Short-format mode
		display.TotalGames = 3
		display.Sets = nil

		// Server (if defined)
		if state.Servers != nil && state.CurrentGame.ServerIndex < len(state.Servers) {
			server := state.Servers[state.CurrentGame.ServerIndex]
			display.Server = &server
		}
	} else {
		// Standard mode
		sets := ScoreCount{A: state.SetsA, B: state.SetsB}
		display.Sets = &sets
		display.TotalGames = 0 // Variable in standard mode
		display.IsTieBreak = IsTieBreak(state.GamesA, state.GamesB)
	}

	return display
}
