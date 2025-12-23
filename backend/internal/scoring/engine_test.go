package scoring

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - UNIT TESTS
// ═══════════════════════════════════════════════════════════════════════════
// Comprehensive tests for all scoring logic following OTS spec.
// ═══════════════════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────────────────────
// HELPER FUNCTIONS
// ─────────────────────────────────────────────────────────────────────────────

func createTestPlayers() TeamPlayers {
	return TeamPlayers{
		TeamA: []string{"player1", "player2"},
		TeamB: []string{"player3", "player4"},
	}
}

func teamPtr(t Team) *Team {
	return &t
}

func scorePoints(t *testing.T, state *MatchState, sequence string) *MatchState {
	t.Helper()
	for _, char := range sequence {
		var team Team
		switch char {
		case 'A':
			team = TeamA
		case 'B':
			team = TeamB
		default:
			continue
		}

		newState, err := ScorePoint(state, team)
		if err != nil {
			t.Fatalf("Error scoring point for team %s: %v", team, err)
		}
		state = newState
	}
	return state
}

// ─────────────────────────────────────────────────────────────────────────────
// DISPLAY TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPointDisplay(t *testing.T) {
	tests := []struct {
		points   int
		expected string
	}{
		{0, "0"},
		{1, "15"},
		{2, "30"},
		{3, "40"},
		{4, "40"}, // Never goes beyond 40
		{5, "40"},
	}

	for _, tt := range tests {
		result := GetPointDisplay(tt.points)
		if result != tt.expected {
			t.Errorf("GetPointDisplay(%d) = %s, expected %s", tt.points, result, tt.expected)
		}
	}
}

func TestGetGameDisplayText(t *testing.T) {
	tests := []struct {
		pointsA   int
		pointsB   int
		expectedA string
		expectedB string
	}{
		{0, 0, "0", "0"},
		{1, 0, "15", "0"},
		{2, 1, "30", "15"},
		{3, 2, "40", "30"},
		{3, 3, "Deuce", "Deuce"},
		{4, 3, "Ad", "40"},
		{3, 4, "40", "Ad"},
		{4, 4, "Deuce", "Deuce"},
		{5, 4, "Ad", "40"},
	}

	for _, tt := range tests {
		result := GetGameDisplayText(tt.pointsA, tt.pointsB)
		if result.A != tt.expectedA || result.B != tt.expectedB {
			t.Errorf("GetGameDisplayText(%d, %d) = (%s, %s), expected (%s, %s)",
				tt.pointsA, tt.pointsB, result.A, result.B, tt.expectedA, tt.expectedB)
		}
	}
}

func TestIsGameWon(t *testing.T) {
	tests := []struct {
		pointsA  int
		pointsB  int
		expected *Team
	}{
		{0, 0, nil},
		{3, 0, nil},
		{4, 0, teamPtr(TeamA)},
		{4, 2, teamPtr(TeamA)},
		{3, 4, nil},            // Not enough lead
		{4, 3, nil},            // Not enough lead (deuce territory)
		{5, 3, teamPtr(TeamA)}, // 2-point lead
		{3, 5, teamPtr(TeamB)},
		{6, 4, teamPtr(TeamA)},
		{7, 8, nil},            // Only 1-point lead
		{8, 6, teamPtr(TeamA)}, // 2-point lead
	}

	for _, tt := range tests {
		result := IsGameWon(tt.pointsA, tt.pointsB)

		if tt.expected == nil {
			if result != nil {
				t.Errorf("IsGameWon(%d, %d) = %v, expected nil", tt.pointsA, tt.pointsB, *result)
			}
		} else {
			if result == nil {
				t.Errorf("IsGameWon(%d, %d) = nil, expected %v", tt.pointsA, tt.pointsB, *tt.expected)
			} else if *result != *tt.expected {
				t.Errorf("IsGameWon(%d, %d) = %v, expected %v", tt.pointsA, tt.pointsB, *result, *tt.expected)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SHORT-FORMAT MODE TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestShortFormatBasicGame(t *testing.T) {
	players := createTestPlayers()
	servers := []string{"player1", "player2", "player3"}

	state, err := NewMatchState(ModeShortFormat, players, servers)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Team A wins first game: 4 points in a row
	state = scorePoints(t, state, "AAAA")

	if state.GamesA != 1 {
		t.Errorf("Expected GamesA = 1, got %d", state.GamesA)
	}

	if state.CurrentGame.GameNumber != 2 {
		t.Errorf("Expected GameNumber = 2, got %d", state.CurrentGame.GameNumber)
	}

	if state.CurrentGame.ServerIndex != 1 {
		t.Errorf("Expected ServerIndex = 1, got %d", state.CurrentGame.ServerIndex)
	}

	if state.Completed {
		t.Error("Match should not be completed after 1 game")
	}
}

func TestShortFormatEarlyWin(t *testing.T) {
	players := createTestPlayers()
	servers := []string{"player1", "player2", "player3"}

	state, err := NewMatchState(ModeShortFormat, players, servers)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Team A wins games 1 and 2 (no game 3 needed)
	// Game 1: A wins 4-0
	state = scorePoints(t, state, "AAAA")

	// Game 2: A wins 4-0
	state = scorePoints(t, state, "AAAA")

	if !state.Completed {
		t.Error("Match should be completed after Team A wins 2 games")
	}

	if state.Winner == nil || *state.Winner != TeamA {
		t.Errorf("Expected winner = TeamA, got %v", state.Winner)
	}

	if state.GamesA != 2 {
		t.Errorf("Expected GamesA = 2, got %d", state.GamesA)
	}

	if state.GamesB != 0 {
		t.Errorf("Expected GamesB = 0, got %d", state.GamesB)
	}
}

func TestShortFormatFullThreeGames(t *testing.T) {
	players := createTestPlayers()
	servers := []string{"player1", "player2", "player3"}

	state, err := NewMatchState(ModeShortFormat, players, servers)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Game 1: A wins
	state = scorePoints(t, state, "AAAA")

	// Game 2: B wins
	state = scorePoints(t, state, "BBBB")

	// Games tied 1-1, match not complete
	if state.Completed {
		t.Error("Match should not be complete at 1-1")
	}

	// Game 3: B wins
	state = scorePoints(t, state, "BBBB")

	// B wins 2-1
	if !state.Completed {
		t.Error("Match should be completed")
	}

	if state.Winner == nil || *state.Winner != TeamB {
		t.Errorf("Expected winner = TeamB, got %v", state.Winner)
	}

	if state.GamesA != 1 || state.GamesB != 2 {
		t.Errorf("Expected games 1-2, got %d-%d", state.GamesA, state.GamesB)
	}
}

func TestShortFormatDeuceGame(t *testing.T) {
	players := createTestPlayers()
	servers := []string{"player1", "player2", "player3"}

	state, err := NewMatchState(ModeShortFormat, players, servers)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Play to deuce (3-3)
	state = scorePoints(t, state, "AAABBB")

	display := GetMatchDisplay(state)
	if display.Points.A != "Deuce" || display.Points.B != "Deuce" {
		t.Errorf("Expected Deuce-Deuce, got %s-%s", display.Points.A, display.Points.B)
	}

	// A gets advantage
	state = scorePoints(t, state, "A")
	display = GetMatchDisplay(state)
	if display.Points.A != "Ad" || display.Points.B != "40" {
		t.Errorf("Expected Ad-40, got %s-%s", display.Points.A, display.Points.B)
	}

	// Back to deuce
	state = scorePoints(t, state, "B")
	display = GetMatchDisplay(state)
	if display.Points.A != "Deuce" || display.Points.B != "Deuce" {
		t.Errorf("Expected Deuce-Deuce, got %s-%s", display.Points.A, display.Points.B)
	}

	// B gets advantage and wins
	state = scorePoints(t, state, "BB")

	if state.GamesB != 1 {
		t.Errorf("Expected GamesB = 1, got %d", state.GamesB)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STANDARD MODE TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestStandardModeBasicSet(t *testing.T) {
	players := createTestPlayers()

	state, err := NewMatchState(ModeStandard, players, nil)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Team A wins 6 games straight (6-0 set)
	for i := 0; i < 6; i++ {
		state = scorePoints(t, state, "AAAA") // Win each game 4-0
	}

	// Set should be won
	if state.SetsA != 1 {
		t.Errorf("Expected SetsA = 1, got %d", state.SetsA)
	}

	if state.GamesA != 0 || state.GamesB != 0 {
		t.Errorf("Expected games reset to 0-0, got %d-%d", state.GamesA, state.GamesB)
	}

	if state.CurrentSet != 2 {
		t.Errorf("Expected CurrentSet = 2, got %d", state.CurrentSet)
	}

	if state.Completed {
		t.Error("Match should not be complete after 1 set")
	}
}

func TestStandardModeMatchWin(t *testing.T) {
	players := createTestPlayers()

	state, err := NewMatchState(ModeStandard, players, nil)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Team A wins first set 6-0
	for i := 0; i < 6; i++ {
		state = scorePoints(t, state, "AAAA")
	}

	// Team A wins second set 6-0
	for i := 0; i < 6; i++ {
		state = scorePoints(t, state, "AAAA")
	}

	// Match should be complete (2-0 in sets)
	if !state.Completed {
		t.Error("Match should be completed")
	}

	if state.Winner == nil || *state.Winner != TeamA {
		t.Errorf("Expected winner = TeamA, got %v", state.Winner)
	}

	if state.SetsA != 2 {
		t.Errorf("Expected SetsA = 2, got %d", state.SetsA)
	}
}

func TestStandardModeCloseSet(t *testing.T) {
	players := createTestPlayers()

	state, err := NewMatchState(ModeStandard, players, nil)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Play to 5-5
	for i := 0; i < 5; i++ {
		state = scorePoints(t, state, "AAAA") // A wins game
		state = scorePoints(t, state, "BBBB") // B wins game
	}

	if state.GamesA != 5 || state.GamesB != 5 {
		t.Errorf("Expected games 5-5, got %d-%d", state.GamesA, state.GamesB)
	}

	// A wins 6th game → 6-5
	state = scorePoints(t, state, "AAAA")

	// Set not won yet (need 2-game lead)
	if state.SetsA != 0 {
		t.Error("Set should not be won at 6-5")
	}

	// A wins 7th game → 7-5 (set won)
	state = scorePoints(t, state, "AAAA")

	if state.SetsA != 1 {
		t.Errorf("Expected SetsA = 1, got %d", state.SetsA)
	}
}

func TestSetWinConditions(t *testing.T) {
	tests := []struct {
		gamesA   int
		gamesB   int
		expected *Team
	}{
		{6, 0, teamPtr(TeamA)},
		{6, 4, teamPtr(TeamA)},
		{7, 5, teamPtr(TeamA)},
		{6, 5, nil}, // Not enough lead
		{5, 6, nil}, // Not enough lead
		{4, 6, teamPtr(TeamB)},
		{7, 6, teamPtr(TeamA)}, // Tie-break win
		{6, 7, teamPtr(TeamB)}, // Tie-break win
		{6, 6, nil},            // Tie-break state, no winner yet
	}

	for _, tt := range tests {
		result := IsSetWon(tt.gamesA, tt.gamesB)

		if tt.expected == nil {
			if result != nil {
				t.Errorf("IsSetWon(%d, %d) = %v, expected nil", tt.gamesA, tt.gamesB, *result)
			}
		} else {
			if result == nil {
				t.Errorf("IsSetWon(%d, %d) = nil, expected %v", tt.gamesA, tt.gamesB, *tt.expected)
			} else if *result != *tt.expected {
				t.Errorf("IsSetWon(%d, %d) = %v, expected %v", tt.gamesA, tt.gamesB, *result, *tt.expected)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// VALIDATION TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestInvalidMatchCreation(t *testing.T) {
	players := createTestPlayers()

	// Short format without servers
	_, err := NewMatchState(ModeShortFormat, players, nil)
	if err == nil {
		t.Error("Expected error for short format without servers")
	}

	// Short format with wrong number of servers
	_, err = NewMatchState(ModeShortFormat, players, []string{"p1", "p2"})
	if err == nil {
		t.Error("Expected error for short format with 2 servers")
	}

	// Standard format with servers
	_, err = NewMatchState(ModeStandard, players, []string{"p1", "p2", "p3"})
	if err == nil {
		t.Error("Expected error for standard format with servers array")
	}

	// Empty teams
	emptyPlayers := TeamPlayers{TeamA: []string{}, TeamB: []string{"p1"}}
	_, err = NewMatchState(ModeStandard, emptyPlayers, nil)
	if err == nil {
		t.Error("Expected error for empty team")
	}
}

func TestScoringAfterMatchComplete(t *testing.T) {
	players := createTestPlayers()
	servers := []string{"player1", "player2", "player3"}

	state, _ := NewMatchState(ModeShortFormat, players, servers)

	// Complete match
	state = scorePoints(t, state, "AAAA") // Game 1
	state = scorePoints(t, state, "AAAA") // Game 2

	// Try to score after match complete
	_, err := ScorePoint(state, TeamA)
	if err == nil {
		t.Error("Expected error when scoring after match completion")
	}
}
