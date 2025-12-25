package model

import "github.com/google/uuid"

// VenueTeamTendency contains aggregate team tendency metrics at a venue.
// Per OTS_Venue_Team_Player_Tendencies_Spec.md Section 4.
// Eligibility: Team must have played at least 3 matches at the venue.
// Applies to doubles matches only.
type VenueTeamTendency struct {
	// TeamID is a composite key of sorted player IDs for doubles teams
	TeamID string `json:"team_id"`

	// Player1ID is the first player in the team (alphabetically by ID)
	Player1ID uuid.UUID `json:"player1_id"`

	// Player2ID is the second player in the team
	Player2ID uuid.UUID `json:"player2_id"`

	// Player1Name is the display name of player 1
	Player1Name string `json:"player1_name"`

	// Player2Name is the display name of player 2
	Player2Name string `json:"player2_name"`

	// MatchesPlayed is the total number of matches played at this venue
	MatchesPlayed int `json:"matches_played"`

	// MatchesWon is the number of matches won at this venue
	MatchesWon int `json:"matches_won"`

	// WinPercentage is (matches_won / matches_played) * 100
	WinPercentage float64 `json:"win_percentage"`

	// AvgGamesPerMatch is the average total games per match
	AvgGamesPerMatch float64 `json:"avg_games_per_match"`

	// DeucePercentage is the percentage of games that went to deuce
	DeucePercentage float64 `json:"deuce_percentage"`

	// FirstServePointsWonPct is team aggregate first serve points won percentage
	FirstServePointsWonPct float64 `json:"first_serve_points_won_pct"`
}

// VenuePlayerTendency contains aggregate player tendency metrics at a venue.
// Per OTS_Venue_Team_Player_Tendencies_Spec.md Section 5.
// Eligibility: Player must have played at least 5 matches at the venue.
// Note: Win percentage is explicitly NOT included per spec restrictions.
type VenuePlayerTendency struct {
	// PlayerID is the unique identifier for the player
	PlayerID uuid.UUID `json:"player_id"`

	// PlayerName is the display name of the player
	PlayerName string `json:"player_name"`

	// MatchesPlayed is the total number of matches played at this venue
	MatchesPlayed int `json:"matches_played"`

	// FirstServeInPct is the percentage of first serves that were in
	FirstServeInPct float64 `json:"first_serve_in_pct"`

	// DoubleFaultsPerGame is the average double faults per game served
	DoubleFaultsPerGame float64 `json:"double_faults_per_game"`

	// AvgPointsPerGame is the average points won per game (optional metric)
	AvgPointsPerGame float64 `json:"avg_points_per_game"`
}

// VenueTendencies is the response structure for venue tendencies endpoint.
// Per OTS_Venue_Team_Player_Tendencies_Spec.md Section 6.
type VenueTendencies struct {
	// VenueID is the venue these tendencies are for
	VenueID uuid.UUID `json:"venue_id"`

	// VenueName is the display name of the venue
	VenueName string `json:"venue_name"`

	// TeamTendencies contains eligible team tendencies (doubles only, 3+ matches)
	// Ordered alphabetically by team display name (neutral ordering per spec)
	TeamTendencies []VenueTeamTendency `json:"team_tendencies"`

	// PlayerTendencies contains eligible player tendencies (5+ matches)
	// Ordered alphabetically by player name (neutral ordering per spec)
	PlayerTendencies []VenuePlayerTendency `json:"player_tendencies"`
}

// Eligibility thresholds per OTS_Venue_Team_Player_Tendencies_Spec.md Section 3
const (
	// MinTeamMatchesForTendency is the minimum matches required for team eligibility
	MinTeamMatchesForTendency = 3

	// MinPlayerMatchesForTendency is the minimum matches required for player eligibility
	MinPlayerMatchesForTendency = 5
)
