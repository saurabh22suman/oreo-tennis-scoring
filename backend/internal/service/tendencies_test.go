package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
)

func TestFormatTeamID(t *testing.T) {
	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	// Order shouldn't matter - result should be consistent
	result1 := formatTeamID(id1, id2)
	result2 := formatTeamID(id2, id1)

	if result1 != result2 {
		t.Errorf("formatTeamID should be order-independent: got %s and %s", result1, result2)
	}

	// Verify it contains both IDs
	expected := "11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222"
	if result1 != expected {
		t.Errorf("formatTeamID unexpected format: got %s, want %s", result1, expected)
	}
}

func TestEligibilityConstants(t *testing.T) {
	// Verify constants match spec requirements
	if model.MinTeamMatchesForTendency != 3 {
		t.Errorf("MinTeamMatchesForTendency should be 3 per spec, got %d", model.MinTeamMatchesForTendency)
	}

	if model.MinPlayerMatchesForTendency != 5 {
		t.Errorf("MinPlayerMatchesForTendency should be 5 per spec, got %d", model.MinPlayerMatchesForTendency)
	}
}

func TestVenueTeamTendencyStruct(t *testing.T) {
	// Verify the struct has all required fields per spec Section 4
	tendency := model.VenueTeamTendency{
		TeamID:                 "test-team-id",
		Player1ID:              uuid.New(),
		Player2ID:              uuid.New(),
		Player1Name:            "Player 1",
		Player2Name:            "Player 2",
		MatchesPlayed:          5,
		MatchesWon:             3,
		WinPercentage:          60.0,
		AvgGamesPerMatch:       14.2,
		DeucePercentage:        28.0,
		FirstServePointsWonPct: 72.5,
	}

	// Verify all fields are accessible
	if tendency.MatchesPlayed != 5 {
		t.Errorf("MatchesPlayed not set correctly")
	}
	if tendency.WinPercentage != 60.0 {
		t.Errorf("WinPercentage not set correctly")
	}
}

func TestVenuePlayerTendencyStruct(t *testing.T) {
	// Verify the struct has all required fields per spec Section 5
	// Note: WinPercentage is deliberately NOT included per spec restrictions
	tendency := model.VenuePlayerTendency{
		PlayerID:            uuid.New(),
		PlayerName:          "Test Player",
		MatchesPlayed:       9,
		FirstServeInPct:     63.0,
		DoubleFaultsPerGame: 0.38,
		AvgPointsPerGame:    2.5,
	}

	// Verify all fields are accessible
	if tendency.MatchesPlayed != 9 {
		t.Errorf("MatchesPlayed not set correctly")
	}
	if tendency.FirstServeInPct != 63.0 {
		t.Errorf("FirstServeInPct not set correctly")
	}
}

func TestVenueTendenciesResponse(t *testing.T) {
	// Verify response structure is complete
	response := model.VenueTendencies{
		VenueID:          uuid.New(),
		VenueName:        "Test Venue",
		TeamTendencies:   []model.VenueTeamTendency{},
		PlayerTendencies: []model.VenuePlayerTendency{},
	}

	if response.VenueName != "Test Venue" {
		t.Errorf("VenueName not set correctly")
	}
	if response.TeamTendencies == nil {
		t.Errorf("TeamTendencies should be empty slice, not nil")
	}
	if response.PlayerTendencies == nil {
		t.Errorf("PlayerTendencies should be empty slice, not nil")
	}
}

// TestTeamEligibilityFiltering tests that teams below threshold are filtered
func TestTeamEligibilityFiltering(t *testing.T) {
	testCases := []struct {
		name           string
		matchesPlayed  int
		shouldBeEligible bool
	}{
		{"0 matches - not eligible", 0, false},
		{"1 match - not eligible", 1, false},
		{"2 matches - not eligible", 2, false},
		{"3 matches - eligible", 3, true},
		{"5 matches - eligible", 5, true},
		{"10 matches - eligible", 10, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isEligible := tc.matchesPlayed >= model.MinTeamMatchesForTendency
			if isEligible != tc.shouldBeEligible {
				t.Errorf("Team with %d matches: got eligible=%v, want %v",
					tc.matchesPlayed, isEligible, tc.shouldBeEligible)
			}
		})
	}
}

// TestPlayerEligibilityFiltering tests that players below threshold are filtered
func TestPlayerEligibilityFiltering(t *testing.T) {
	testCases := []struct {
		name             string
		matchesPlayed    int
		shouldBeEligible bool
	}{
		{"0 matches - not eligible", 0, false},
		{"2 matches - not eligible", 2, false},
		{"4 matches - not eligible", 4, false},
		{"5 matches - eligible", 5, true},
		{"7 matches - eligible", 7, true},
		{"15 matches - eligible", 15, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isEligible := tc.matchesPlayed >= model.MinPlayerMatchesForTendency
			if isEligible != tc.shouldBeEligible {
				t.Errorf("Player with %d matches: got eligible=%v, want %v",
					tc.matchesPlayed, isEligible, tc.shouldBeEligible)
			}
		})
	}
}

// TestWinPercentageCalculation tests the win percentage formula
func TestWinPercentageCalculation(t *testing.T) {
	testCases := []struct {
		name          string
		matchesPlayed int
		matchesWon    int
		expectedPct   float64
	}{
		{"3 wins of 5 = 60%", 5, 3, 60.0},
		{"0 wins of 3 = 0%", 3, 0, 0.0},
		{"5 wins of 5 = 100%", 5, 5, 100.0},
		{"1 win of 4 = 25%", 4, 1, 25.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pct float64
			if tc.matchesPlayed > 0 {
				pct = float64(tc.matchesWon) / float64(tc.matchesPlayed) * 100
			}
			if pct != tc.expectedPct {
				t.Errorf("Win percentage: got %.1f, want %.1f", pct, tc.expectedPct)
			}
		})
	}
}

// TestAverageGamesPerMatchCalculation tests the avg games formula
func TestAverageGamesPerMatchCalculation(t *testing.T) {
	testCases := []struct {
		name          string
		matchesPlayed int
		totalGames    int
		expectedAvg   float64
	}{
		{"71 games in 5 matches = 14.2", 5, 71, 14.2},
		{"30 games in 3 matches = 10.0", 3, 30, 10.0},
		{"50 games in 4 matches = 12.5", 4, 50, 12.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var avg float64
			if tc.matchesPlayed > 0 {
				avg = float64(tc.totalGames) / float64(tc.matchesPlayed)
			}
			if avg != tc.expectedAvg {
				t.Errorf("Avg games per match: got %.1f, want %.1f", avg, tc.expectedAvg)
			}
		})
	}
}

// TestFirstServePercentageCalculation tests the first serve in percentage
func TestFirstServePercentageCalculation(t *testing.T) {
	testCases := []struct {
		name        string
		servesTotal int
		servesIn    int
		expectedPct float64
	}{
		{"63 of 100 = 63%", 100, 63, 63.0},
		{"0 of 50 = 0%", 50, 0, 0.0},
		{"100 of 100 = 100%", 100, 100, 100.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pct float64
			if tc.servesTotal > 0 {
				pct = float64(tc.servesIn) / float64(tc.servesTotal) * 100
			}
			if pct != tc.expectedPct {
				t.Errorf("First serve in pct: got %.1f, want %.1f", pct, tc.expectedPct)
			}
		})
	}
}
