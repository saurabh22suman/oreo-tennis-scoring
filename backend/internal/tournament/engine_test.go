package tournament

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - UNIT TESTS
// ═══════════════════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────────────────────
// TEAM GENERATION TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateRandomTeamsBasic(t *testing.T) {
	players := []uuid.UUID{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	teams, err := GenerateRandomTeams(players, 12345)
	if err != nil {
		t.Fatalf("Failed to generate teams: %v", err)
	}

	if len(teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(teams))
	}

	// Check team numbers
	if teams[0].TeamNumber != 1 || teams[1].TeamNumber != 2 {
		t.Error("Team numbers incorrectly assigned")
	}
}

func TestGenerateRandomTeamsDeterministic(t *testing.T) {
	players := []uuid.UUID{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	// Same seed should produce same teams
	teams1, _ := GenerateRandomTeams(players, 42)
	teams2, _ := GenerateRandomTeams(players, 42)

	if teams1[0].Player1ID != teams2[0].Player1ID {
		t.Error("Same seed produced different shuffles")
	}
}

func TestGenerateRandomTeamsValidation(t *testing.T) {
	// Too few players
	_, err := GenerateRandomTeams([]uuid.UUID{uuid.New(), uuid.New()}, 123)
	if err == nil {
		t.Error("Expected error for too few players")
	}

	// Odd number of players
	_, err = GenerateRandomTeams([]uuid.UUID{uuid.New(), uuid.New(), uuid.New()}, 123)
	if err == nil {
		t.Error("Expected error for odd number of players")
	}
}

func TestGenerateManualTeams(t *testing.T) {
	p1, p2, p3, p4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	pairs := [][2]uuid.UUID{
		{p1, p2},
		{p3, p4},
	}

	teams, err := GenerateManualTeams(pairs)
	if err != nil {
		t.Fatalf("Failed to generate manual teams: %v", err)
	}

	if len(teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(teams))
	}

	// Verify pairs
	if teams[0].Player1ID != p1 || teams[0].Player2ID != p2 {
		t.Error("Team 1 has incorrect players")
	}
}

func TestGenerateManualTeamsDuplicatePlayer(t *testing.T) {
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()

	// p2 appears in both teams
	pairs := [][2]uuid.UUID{
		{p1, p2},
		{p2, p3},
	}

	_, err := GenerateManualTeams(pairs)
	if err == nil {
		t.Error("Expected error for duplicate player")
	}
}

func TestGenerateManualTeamsSamePlayerTwice(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()

	pairs := [][2]uuid.UUID{
		{p1, p1}, // Same player twice
		{p2, uuid.New()},
	}

	_, err := GenerateManualTeams(pairs)
	if err == nil {
		t.Error("Expected error for same player twice in team")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ROUND ROBIN TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateRoundRobinMatches3Teams(t *testing.T) {
	tournamentID := uuid.New()
	teams := []Team{
		{ID: uuid.New(), TeamNumber: 1},
		{ID: uuid.New(), TeamNumber: 2},
		{ID: uuid.New(), TeamNumber: 3},
	}

	matches := GenerateRoundRobinMatches(tournamentID, teams)

	// 3 teams → 3 matches
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// Verify all matches are round-robin
	for _, match := range matches {
		if match.Stage != StageRR {
			t.Error("Match should be round-robin stage")
		}
	}
}

func TestGenerateRoundRobinMatches4Teams(t *testing.T) {
	tournamentID := uuid.New()
	teams := make([]Team, 4)
	for i := range teams {
		teams[i] = Team{ID: uuid.New(), TeamNumber: i + 1}
	}

	matches := GenerateRoundRobinMatches(tournamentID, teams)

	// 4 teams → 6 matches
	if len(matches) != 6 {
		t.Errorf("Expected 6 matches, got %d", len(matches))
	}
}

func TestRoundRobinFormula(t *testing.T) {
	tests := []struct {
		numTeams        int
		expectedMatches int
	}{
		{3, 3},
		{4, 6},
		{5, 10},
		{6, 15},
	}

	for _, tt := range tests {
		teams := make([]Team, tt.numTeams)
		for i := range teams {
			teams[i] = Team{ID: uuid.New()}
		}

		matches := GenerateRoundRobinMatches(uuid.New(), teams)

		if len(matches) != tt.expectedMatches {
			t.Errorf("For %d teams: expected %d matches, got %d",
				tt.numTeams, tt.expectedMatches, len(matches))
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STANDINGS TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestInitializeStandings(t *testing.T) {
	teams := []Team{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	standings := InitializeStandings(teams)

	if len(standings) != 3 {
		t.Errorf("Expected 3 standings, got %d", len(standings))
	}

	for _, standing := range standings {
		if standing.Played != 0 || standing.Won != 0 || standing.Lost != 0 || standing.Points != 0 {
			t.Error("Initial standings should all be zero")
		}
	}
}

func TestUpdateStandingsWithResult(t *testing.T) {
	teamA, teamB := uuid.New(), uuid.New()

	standings := []TeamStanding{
		{TeamID: teamA, Played: 0, Won: 0, Lost: 0, Points: 0},
		{TeamID: teamB, Played: 0, Won: 0, Lost: 0, Points: 0},
	}

	result := MatchResult{
		MatchID:      uuid.New(),
		WinnerTeamID: teamA,
		LoserTeamID:  teamB,
	}

	updated := UpdateStandingsWithResult(standings, result)

	// Winner stats
	winnerStanding := GetStandingByTeamID(updated, teamA)
	if winnerStanding.Played != 1 || winnerStanding.Won != 1 || winnerStanding.Points != 1 {
		t.Errorf("Winner stats incorrect: played=%d, won=%d, points=%d",
			winnerStanding.Played, winnerStanding.Won, winnerStanding.Points)
	}

	// Loser stats
	loserStanding := GetStandingByTeamID(updated, teamB)
	if loserStanding.Played != 1 || loserStanding.Lost != 1 || loserStanding.Points != 0 {
		t.Errorf("Loser stats incorrect: played=%d, lost=%d, points=%d",
			loserStanding.Played, loserStanding.Lost, loserStanding.Points)
	}
}

func TestCalculateRankings(t *testing.T) {
	standings := []TeamStanding{
		{TeamID: uuid.New(), Points: 1},
		{TeamID: uuid.New(), Points: 3},
		{TeamID: uuid.New(), Points: 2},
	}

	ranked := CalculateRankings(standings)

	// Should be sorted by points descending
	if ranked[0].Points != 3 || ranked[1].Points != 2 || ranked[2].Points != 1 {
		t.Error("Rankings not sorted correctly by points")
	}

	// Check ranks assigned
	if ranked[0].Rank != 1 || ranked[1].Rank != 2 || ranked[2].Rank != 3 {
		t.Error("Ranks not assigned correctly")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// KNOCKOUT TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateKnockout3Teams(t *testing.T) {
	tournamentID := uuid.New()
	standings := []TeamStanding{
		{TeamID: uuid.New(), Rank: 1, Played: 2, Points: 2},
		{TeamID: uuid.New(), Rank: 2, Played: 2, Points: 1},
		{TeamID: uuid.New(), Rank: 3, Played: 2, Points: 0},
	}

	matches, err := GenerateKnockoutMatches(tournamentID, standings)
	if err != nil {
		t.Fatalf("Failed to generate knockout: %v", err)
	}

	// 3 teams → Final only
	if len(matches) != 1 {
		t.Errorf("Expected 1 match (final), got %d", len(matches))
	}

	if matches[0].Stage != StageFinal {
		t.Error("Match should be final stage")
	}

	// Final should be Rank 1 vs Rank 2
	if matches[0].TeamAID != standings[0].TeamID || matches[0].TeamBID != standings[1].TeamID {
		t.Error("Final matchup incorrect for 3 teams")
	}
}

func TestGenerateKnockout4Teams(t *testing.T) {
	tournamentID := uuid.New()
	standings := []TeamStanding{
		{TeamID: uuid.New(), Rank: 1, Played: 3, Points: 3},
		{TeamID: uuid.New(), Rank: 2, Played: 3, Points: 2},
		{TeamID: uuid.New(), Rank: 3, Played: 3, Points: 1},
		{TeamID: uuid.New(), Rank: 4, Played: 3, Points: 0},
	}

	matches, err := GenerateKnockoutMatches(tournamentID, standings)
	if err != nil {
		t.Fatalf("Failed to generate knockout: %v", err)
	}

	// 4 teams → 2 semis + 1 final = 3 matches
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// Check semifinals
	semis := GetMatchesByStage(matches, StageSemi)
	if len(semis) != 2 {
		t.Errorf("Expected 2 semifinals, got %d", len(semis))
	}

	// SF1: Rank 1 vs Rank 4
	// SF2: Rank 2 vs Rank 3
	sf1 := semis[0]
	if sf1.TeamAID != standings[0].TeamID || sf1.TeamBID != standings[3].TeamID {
		t.Error("SF1 matchup incorrect")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TOURNAMENT ENGINE TESTS
// ─────────────────────────────────────────────────────────────────────────────

func TestNewTournament(t *testing.T) {
	venueID := uuid.New()
	players := []uuid.UUID{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	tournament, err := NewTournament(venueID, players)
	if err != nil {
		t.Fatalf("Failed to create tournament: %v", err)
	}

	if tournament.Stage != StageSetup {
		t.Errorf("Expected Setup stage, got %s", tournament.Stage)
	}

	if len(tournament.PlayerIDs) != 4 {
		t.Errorf("Expected 4 players, got %d", len(tournament.PlayerIDs))
	}
}

func TestSetTeams(t *testing.T) {
	venueID := uuid.New()
	players := []uuid.UUID{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	tournament, _ := NewTournament(venueID, players)

	// Generate teams
	teams, _ := GenerateRandomTeams(players, time.Now().UnixNano())

	// Set teams
	updated, err := SetTeams(tournament, teams)
	if err != nil {
		t.Fatalf("Failed to set teams: %v", err)
	}

	if updated.Stage != StageRoundRobin {
		t.Errorf("Expected RoundRobin stage, got %s", updated.Stage)
	}

	if len(updated.RoundRobinMatches) != 1 {
		t.Errorf("Expected 1 round-robin match (2 teams), got %d", len(updated.RoundRobinMatches))
	}

	if len(updated.Standings) != 2 {
		t.Errorf("Expected 2 standings entries, got %d", len(updated.Standings))
	}
}

func TestFullTournamentFlow3Teams(t *testing.T) {
	// Setup
	venueID := uuid.New()
	players := []uuid.UUID{
		uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(),
	}

	tournament, _ := NewTournament(venueID, players)
	teams, _ := GenerateRandomTeams(players, 12345)
	tournament, _ = SetTeams(tournament, teams)

	// Should have 3 round-robin matches
	if len(tournament.RoundRobinMatches) != 3 {
		t.Fatalf("Expected 3 RR matches, got %d", len(tournament.RoundRobinMatches))
	}

	// Play all round-robin matches
	for i := range tournament.RoundRobinMatches {
		match := tournament.RoundRobinMatches[i]
		result := MatchResult{
			MatchID:      match.ID,
			WinnerTeamID: match.TeamAID,
			LoserTeamID:  match.TeamBID,
		}
		var err error
		tournament, err = RecordMatchResult(tournament, result)
		if err != nil {
			t.Fatalf("Failed to record match result: %v", err)
		}
	}

	// Advance to knockout
	tournament, err := AdvanceToKnockout(tournament)
	if err != nil {
		t.Fatalf("Failed to advance to knockout: %v", err)
	}

	if tournament.Stage != StageKnockout {
		t.Errorf("Expected Knockout stage, got %s", tournament.Stage)
	}

	// 3 teams → Final only
	if len(tournament.KnockoutMatches) != 1 {
		t.Errorf("Expected 1 knockout match, got %d", len(tournament.KnockoutMatches))
	}

	// Play final
	final := tournament.KnockoutMatches[0]
	finalResult := MatchResult{
		MatchID:      final.ID,
		WinnerTeamID: final.TeamAID,
		LoserTeamID:  final.TeamBID,
	}
	tournament, _ = RecordMatchResult(tournament, finalResult)

	// Tournament should be complete
	if !tournament.Completed {
		t.Error("Tournament should be completed")
	}

	if tournament.Stage != StageCompleted {
		t.Errorf("Expected Completed stage, got %s", tournament.Stage)
	}

	if tournament.Winner == nil {
		t.Error("Winner should be set")
	}
}
