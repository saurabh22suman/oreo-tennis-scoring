package tournament

import (
	"time"

	"github.com/google/uuid"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOURNAMENT ENGINE - TYPE DEFINITIONS
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tournament_Spec.md
// This file defines all types used by the tournament engine.
// The tournament engine has NO scoring logic - it uses the scoring engine.
// ═══════════════════════════════════════════════════════════════════════════

// TournamentStage represents the current stage of the tournament.
type TournamentStage string

const (
	// StageSetup: Tournament created, teams being formed
	StageSetup TournamentStage = "setup"

	// StageRoundRobin: Round-robin matches in progress
	StageRoundRobin TournamentStage = "round_robin"

	// StageKnockout: Semifinals/Final in progress
	StageKnockout TournamentStage = "knockout"

	// StageCompleted: Tournament finished, winner declared
	StageCompleted TournamentStage = "completed"
)

// MatchStage represents the stage/type of a match within a tournament.
type MatchStage string

const (
	StageRR    MatchStage = "round_robin" // Round-robin match
	StageSemi  MatchStage = "semi"        // Semifinal
	StageFinal MatchStage = "final"       // Final
)

// TournamentState represents the complete state of a doubles tournament.
type TournamentState struct {
	// ID: Unique tournament identifier
	ID uuid.UUID

	// VenueID: Where the tournament is being played
	VenueID uuid.UUID

	// PlayerIDs: All players participating in the tournament
	PlayerIDs []uuid.UUID

	// Teams: Generated doubles teams (pairs of players)
	Teams []Team

	// ─────────────────────────────────────────────────────────────────────
	// TOURNAMENT STAGE
	// ─────────────────────────────────────────────────────────────────────

	// Stage: Current tournament stage
	Stage TournamentStage

	// ─────────────────────────────────────────────────────────────────────
	// ROUND ROBIN
	// ─────────────────────────────────────────────────────────────────────

	// RoundRobinMatches: All matches in round-robin stage
	// Generated using formula: T × (T - 1) / 2
	RoundRobinMatches []Match

	// ─────────────────────────────────────────────────────────────────────
	// STANDINGS
	// ─────────────────────────────────────────────────────────────────────

	// Standings: Current tournament standings
	// Updated after each match completion
	Standings []TeamStanding

	// ─────────────────────────────────────────────────────────────────────
	// KNOCKOUT STAGE
	// ─────────────────────────────────────────────────────────────────────

	// KnockoutMatches: Semifinals and Final
	KnockoutMatches []Match

	// ─────────────────────────────────────────────────────────────────────
	// RESULT
	// ─────────────────────────────────────────────────────────────────────

	// Winner: Winning team ID (set after Final)
	Winner *uuid.UUID

	// Completed: True when tournament is finished
	Completed bool

	// ─────────────────────────────────────────────────────────────────────
	// METADATA
	// ─────────────────────────────────────────────────────────────────────

	CreatedAt time.Time
}

// Team represents a doubles team in the tournament.
type Team struct {
	// ID: Unique team identifier
	ID uuid.UUID

	// Player1ID: First player in the pair
	Player1ID uuid.UUID

	// Player2ID: Second player in the pair
	Player2ID uuid.UUID

	// TeamNumber: Display number (1, 2, 3, ...)
	TeamNumber int
}

// Match represents a match in the tournament.
type Match struct {
	// ID: Unique match identifier
	ID uuid.UUID

	// TournamentID: Parent tournament
	TournamentID uuid.UUID

	// TeamAID: First team
	TeamAID uuid.UUID

	// TeamBID: Second team
	TeamBID uuid.UUID

	// Stage: round_robin, semi, or final
	Stage MatchStage

	// MatchOrder: Order in which match should be played
	// (Optional, for scheduling)
	MatchOrder int

	// ─────────────────────────────────────────────────────────────────────
	// RESULT (Set after match completion)
	// ─────────────────────────────────────────────────────────────────────

	// ScoringMatchID: Link to the actual tennis match (scoring engine)
	// This connects tournament logic to match scoring logic
	ScoringMatchID *uuid.UUID

	// WinnerTeamID: Which team won this match
	WinnerTeamID *uuid.UUID

	// Completed: True when match is finished
	Completed bool
}

// TeamStanding represents a team's standing in the round-robin stage.
type TeamStanding struct {
	// TeamID: Team identifier
	TeamID uuid.UUID

	// Rank: Current rank (1 = first place)
	Rank int

	// ─────────────────────────────────────────────────────────────────────
	// STATS (Per OTS_Tournament_Spec.md Section 5.1)
	// ─────────────────────────────────────────────────────────────────────

	// Played: Matches played
	Played int

	// Won: Matches won
	Won int

	// Lost: Matches lost
	Lost int

	// Points: Tournament points (1 per win, 0 per loss)
	Points int
}

// MatchResult represents the outcome of a completed match.
// This is what the tournament engine receives from the scoring engine.
type MatchResult struct {
	// MatchID: Tournament match identifier
	MatchID uuid.UUID

	// WinnerTeamID: Which team won
	WinnerTeamID uuid.UUID

	// LoserTeamID: Which team lost
	LoserTeamID uuid.UUID
}

// TeamCreationMode specifies how teams are generated.
type TeamCreationMode string

const (
	// ModeRandom: Shuffle and pair sequentially
	ModeRandom TeamCreationMode = "random"

	// ModeManual: User manually assigns players to teams
	ModeManual TeamCreationMode = "manual"
)
