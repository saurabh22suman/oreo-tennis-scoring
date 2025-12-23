package scoring

// ═══════════════════════════════════════════════════════════════════════════
// TENNIS SCORING ENGINE - TYPE DEFINITIONS
// ═══════════════════════════════════════════════════════════════════════════
// Source of Truth: OTS_Tennis_Scoring_Spec.md
// This file defines all types used by the tennis scoring state machine.
// The scoring engine is STATELESS and has NO knowledge of tournaments.
// ═══════════════════════════════════════════════════════════════════════════

// MatchMode defines the scoring format for a tennis match.
type MatchMode string

const (
	// ModeStandard represents traditional tennis scoring:
	// Points → Games → Sets → Match
	// Default: Best of 3 sets
	ModeStandard MatchMode = "standard"

	// ModeShortFormat represents recreational 3-game format:
	// Points → Games → Match
	// Best of 3 games, no sets
	// Fixed server rotation per game
	ModeShortFormat MatchMode = "short"
)

// Team represents one side in a tennis match.
type Team string

const (
	TeamA Team = "A"
	TeamB Team = "B"
)

// GameState represents the state of the current game being played.
type GameState int

const (
	// GameInProgress means the game is ongoing with no special conditions
	GameInProgress GameState = iota

	// GameDeuce means both sides are at 40-40 (3-3 points)
	GameDeuce

	// GameAdvantageA means Team A has advantage after deuce
	GameAdvantageA

	// GameAdvantageB means Team B has advantage after deuce
	GameAdvantageB
)

// MatchState represents the complete scoring state of a tennis match.
// This struct is IMMUTABLE - all scoring operations return a new instance.
type MatchState struct {
	// Mode determines the scoring rules (standard or short-format)
	Mode MatchMode

	// Players assigned to each team
	Players TeamPlayers

	// Servers (short-format only): Array of exactly 3 server IDs
	// servers[0] serves Game 1, servers[1] serves Game 2, servers[2] serves Game 3
	Servers []string

	// ─────────────────────────────────────────────────────────────────────
	// CURRENT GAME STATE
	// ─────────────────────────────────────────────────────────────────────

	// CurrentGame tracks the state within the current game
	CurrentGame CurrentGameState

	// ─────────────────────────────────────────────────────────────────────
	// GAMES TRACKING
	// ─────────────────────────────────────────────────────────────────────

	// GamesA: Games won by Team A
	// - Standard mode: Games in current set
	// - Short-format mode: Total games won in match
	GamesA int

	// GamesB: Games won by Team B
	// - Standard mode: Games in current set
	// - Short-format mode: Total games won in match
	GamesB int

	// ─────────────────────────────────────────────────────────────────────
	// STANDARD MODE ONLY (SETS)
	// ─────────────────────────────────────────────────────────────────────

	// SetsA: Sets won by Team A (standard mode only)
	SetsA int

	// SetsB: Sets won by Team B (standard mode only)
	SetsB int

	// CurrentSet: Current set number (1, 2, or 3) (standard mode only)
	CurrentSet int

	// ─────────────────────────────────────────────────────────────────────
	// MATCH RESULT
	// ─────────────────────────────────────────────────────────────────────

	// Winner: Pointer to winning team (nil if match ongoing)
	Winner *Team

	// Completed: True if match is over
	Completed bool
}

// CurrentGameState tracks scoring within the current game being played.
type CurrentGameState struct {
	// PointsA: Raw point count for Team A in current game (0, 1, 2, 3, ...)
	// NEVER displayed directly - always mapped to tennis notation (0, 15, 30, 40)
	PointsA int

	// PointsB: Raw point count for Team B in current game (0, 1, 2, 3, ...)
	// NEVER displayed directly - always mapped to tennis notation (0, 15, 30, 40)
	PointsB int

	// GameNumber: Current game number
	// - Standard mode: Total games in set (1st game = 1, 2nd game = 2, etc.)
	// - Short-format: 1, 2, or 3
	GameNumber int

	// ServerIndex: Index into the Servers array (short-format only)
	// Determines which player serves the current game
	ServerIndex int
}

// TeamPlayers represents the player assignments for both teams.
type TeamPlayers struct {
	TeamA []string // Player IDs for Team A
	TeamB []string // Player IDs for Team B
}

// MatchDisplay represents the user-facing display of the current match state.
// This is what gets shown in the UI - never raw point counts.
type MatchDisplay struct {
	// Points: Tennis notation for current game (e.g., "15", "30", "40", "Deuce", "Ad")
	Points PointDisplay

	// Games: Games won by each team
	Games ScoreCount

	// Sets: Sets won by each team (standard mode only, nil for short-format)
	Sets *ScoreCount

	// CurrentSet: Current set number (standard mode only, 0 for short-format)
	CurrentSet int

	// GameNumber: Current game number within the match
	GameNumber int

	// TotalGames: Total possible games (3 for short-format, variable for standard)
	TotalGames int

	// Server: ID of the current server (nil if not applicable)
	Server *string

	// IsTieBreak: True if currently in a tie-break (standard mode only)
	IsTieBreak bool
}

// PointDisplay represents the current point score in tennis notation.
type PointDisplay struct {
	A string // Team A's score ("0", "15", "30", "40", "Deuce", "Ad")
	B string // Team B's score ("0", "15", "30", "40", "Deuce", "Ad")
}

// ScoreCount represents a simple numeric score (games or sets).
type ScoreCount struct {
	A int // Team A's count
	B int // Team B's count
}
