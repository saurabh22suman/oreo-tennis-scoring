package model

import (
	"time"

	"github.com/google/uuid"
)

// Surface represents the court surface type.
type Surface string

const (
	SurfaceHard  Surface = "hard"
	SurfaceClay  Surface = "clay"
	SurfaceGrass Surface = "grass"
)

// MatchType represents singles or doubles.
type MatchType string

const (
	MatchTypeSingles MatchType = "singles"
	MatchTypeDoubles MatchType = "doubles"
)

// Team represents team A or B.
type Team string

const (
	TeamA Team = "A"
	TeamB Team = "B"
)

// ServeType represents the serve outcome.
type ServeType string

const (
	ServeTypeFirst       ServeType = "first"
	ServeTypeSecond      ServeType = "second"
	ServeTypeDoubleFault ServeType = "double_fault"
)

// Player represents a tennis player.
type Player struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// Venue represents a tennis venue.
type Venue struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Surface   Surface   `json:"surface"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// Match represents a tennis match.
type Match struct {
	ID        uuid.UUID  `json:"id"`
	VenueID   uuid.UUID  `json:"venue_id"`
	MatchType MatchType  `json:"match_type"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// MatchPlayer represents the association between a match and a player.
type MatchPlayer struct {
	MatchID  uuid.UUID `json:"match_id"`
	PlayerID uuid.UUID `json:"player_id"`
	Team     Team      `json:"team"`
}

// PointEvent represents a single point scored in a match.
type PointEvent struct {
	ID              uuid.UUID `json:"id"`
	MatchID         uuid.UUID `json:"match_id"`
	Timestamp       time.Time `json:"timestamp"`
	ServerPlayerID  uuid.UUID `json:"server_player_id"`
	ServeType       ServeType `json:"serve_type"`
	PointWinnerTeam Team      `json:"point_winner_team"`
}

// MatchWithDetails includes match info with related data.
type MatchWithDetails struct {
	Match   Match         `json:"match"`
	Venue   Venue         `json:"venue"`
	Players []MatchPlayer `json:"players"`
	Events  []PointEvent  `json:"events,omitempty"`
}

// MatchSummary contains computed statistics for a completed match.
type MatchSummary struct {
	MatchID     uuid.UUID          `json:"match_id"`
	Venue       Venue              `json:"venue"`
	MatchType   MatchType          `json:"match_type"`
	StartedAt   time.Time          `json:"started_at"`
	EndedAt     *time.Time         `json:"ended_at,omitempty"`
	TeamAScore  int                `json:"team_a_score"`  // Total points
	TeamBScore  int                `json:"team_b_score"`  // Total points
	GamesA      int                `json:"games_a"`       // Games won by Team A
	GamesB      int                `json:"games_b"`       // Games won by Team B
	SetsA       int                `json:"sets_a"`        // Sets won by Team A (standard mode only)
	SetsB       int                `json:"sets_b"`        // Sets won by Team B (standard mode only)
	PlayerStats []PlayerMatchStats `json:"player_stats"`
}

// PlayerMatchStats contains serve statistics for a player in a match.
type PlayerMatchStats struct {
	PlayerID          uuid.UUID `json:"player_id"`
	PlayerName        string    `json:"player_name"`
	Team              Team      `json:"team"`
	FirstServesIn     int       `json:"first_serves_in"`
	FirstServesTotal  int       `json:"first_serves_total"`
	FirstServeWon     int       `json:"first_serve_won"`
	SecondServesIn    int       `json:"second_serves_in"`
	SecondServesTotal int       `json:"second_serves_total"`
	SecondServeWon    int       `json:"second_serve_won"`
	DoubleFaults      int       `json:"double_faults"`
	TotalPointsWon    int       `json:"total_points_won"`
}
