# OTS Core Logic Implementation Plan

## Overview
This document outlines the implementation of production-grade Tennis Scoring Engine and Tournament Engine for Oreo Tennis Scoring (OTS), strictly following the authoritative specifications.

---

## Architecture

### Separation of Concerns

```
┌─────────────────────────────┐
│   TOURNAMENT ENGINE         │
│  (Player → Team → Match)    │
│  - Team Generation          │
│  - Round Robin Matching     │
│  - Standings Calculation    │
│  - Knockout Progression     │
└──────────┬──────────────────┘
           │
           │ Uses as black box
           ▼
┌─────────────────────────────┐
│   SCORING ENGINE            │
│  (Point → Game → Set)       │
│  - Standard Tennis Mode     │
│  - Short Format Mode        │
│  - Deuce/Advantage Logic    │
│  - Display Mapping          │
└─────────────────────────────┘
```

**CRITICAL**: These two engines are COMPLETELY SEPARATE with NO shared mutable state.

---

## Module 1: Tennis Scoring Engine

### Location
`backend/internal/scoring/`

### Files to Create
```
scoring/
├── engine.go           # Core state machine
├── types.go            # Types and constants
├── display.go          # Display/presentation logic
├── standard.go         # Standard mode implementation
├── short_format.go     # Short-format mode implementation
└── engine_test.go      # Unit tests
```

### Public Interface

```go
package scoring

// MatchMode defines the scoring format
type MatchMode string

const (
    ModeStandard    MatchMode = "standard"
    ModeShortFormat MatchMode = "short"
)

// Team represents a side in the match
type Team string

const (
    TeamA Team = "A"
    TeamB Team = "B"
)

// MatchState represents the complete scoring state
type MatchState struct {
    Mode      MatchMode
    Players   TeamPlayers
    Servers   []string      // For short format only
    
    // Current game
    CurrentGame GameState
    
    // Games won (current set for standard, total for short)
    GamesA int
    GamesB int
    
    // Standard mode only
    SetsA      int
    SetsB      int
    CurrentSet int
    
    // Match result
    Winner    *Team
    Completed bool
}

// Core Functions (Pure, Stateless)
func NewMatchState(mode MatchMode, players TeamPlayers, servers []string) (*MatchState, error)
func ScorePoint(state *MatchState, team Team) (*MatchState, error)
func GetDisplay(state *MatchState) MatchDisplay
func IsMatchComplete(state *MatchState) bool
```

### Implementation Rules
1. **Pure Functions**: All scoring functions return new state, never mutate
2. **No Side Effects**: No I/O, no randomness, no time dependencies
3. **Deterministic**: Same input → Same output, always
4. **Self-Contained**: No knowledge of tournaments, databases, or UI
5. **Testable**: Every function has unit tests

---

## Module 2: Tournament Engine

### Location
`backend/internal/tournament/`

### Files to Create
```
tournament/
├── engine.go           # Main tournament orchestrator
├── types.go            # Types and constants
├── team_generator.go   # Random & manual team creation
├── round_robin.go      # Round-robin match generation
├── standings.go        # Standings calculation & ranking
├── knockout.go         # Knockout stage logic
└── engine_test.go      # Unit tests
```

### Public Interface

```go
package tournament

// TournamentState represents the full tournament state
type TournamentState struct {
    ID          string
    PlayerIDs   []string
    Teams       []Team
    
    // Stages
    Stage           TournamentStage
    RoundRobinMatch []Match
    StandingsTable  []TeamStanding
    KnockoutMatches []KnockoutMatch
    
    // Result
    Winner     *string  // Team ID
    Completed  bool
}

// Team represents a doubles team
type Team struct {
    ID       string
    Player1  string
    Player2  string
}

// Core Functions
func NewTournament(playerIDs []string) (*TournamentState, error)
func GenerateRandomTeams(playerIDs []string, seed int64) ([]Team, error)
func GenerateManualTeams(pairs [][2]string) ([]Team, error)
func GenerateRoundRobinMatches(teams []Team) []Match
func UpdateStandings(state *TournamentState, matchResult MatchResult) (*TournamentState, error)
func AdvanceToKnockout(state *TournamentState) (*TournamentState, error)
```

### Implementation Rules
1. **Uses Scoring Engine**: Never implements point/game/set logic itself
2. **Deterministic Randomization**: Accepts seed parameter for reproducibility
3. **Stateless Operations**: Functions operate on immutable tournament state
4. **Round-Robin Formula**: T × (T − 1) / 2 matches
5. **Clear Progression**: Setup → Round Robin → Knockout → Complete

---

## Integration Architecture

### Backend API Layer
`backend/internal/handler/tournament_handler.go`

```go
// Tournament lifecycle
POST   /api/tournaments                 # Create tournament
POST   /api/tournaments/:id/teams       # Generate teams (random/manual)
GET    /api/tournaments/:id/matches     # Get round-robin matches
POST   /api/tournaments/:id/matches/:matchId/result  # Submit match result
GET    /api/tournaments/:id/standings   # Get current standings
POST   /api/tournaments/:id/advance     # Advance to knockout stage
GET    /api/tournaments/:id             # Get tournament state
```

### Data Persistence
- Tournament state stored in PostgreSQL
- Match results link to scoring state snapshots
- Scoring engine remains stateless (doesn't know about DB)
- Tournament engine remains stateless (handlers manage persistence)

---

## Implementation Order

### Phase 1: Scoring Engine (Core) ✅
**Effort**: 3-4 hours

1. Create `scoring/types.go` - Define all types
2. Create `scoring/engine.go` - Core state machine
3. Create `scoring/display.go` - Display mapping
4. Create `scoring/standard.go` - Standard mode
5. Create `scoring/short_format.go` - Short format mode
6. Create `scoring/engine_test.go` - Comprehensive tests

**Success Criteria**:
- All unit tests pass
- Correctly handles Deuce/Advantage
- Supports both modes
- Display mappings are correct (0, 15, 30, 40, Deuce, Ad)
- No raw numeric points exposed

### Phase 2: Tournament Engine (Core) ✅
**Effort**: 3-4 hours

1. Create `tournament/types.go` - Define all types
2. Create `tournament/team_generator.go` - Team creation logic
3. Create `tournament/round_robin.go` - Match generation
4. Create `tournament/standings.go` - Standings calculator
5. Create `tournament/knockout.go` - Knockout logic
6. Create `tournament/engine.go` - Main orchestrator
7. Create `tournament/engine_test.go` - Comprehensive tests

**Success Criteria**:
- Random team generation works correctly
- Round-robin formula generates correct matches
- Standings calculation is accurate
- Knockout advancement follows rules
- All unit tests pass

### Phase 3: Database Schema ✅
**Effort**: 1 hour

Create migration: `backend/migrations/000X_tournament_schema.sql`

```sql
-- Tournament tables
CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id UUID REFERENCES venues(id),
    stage VARCHAR(20) NOT NULL,  -- setup, round_robin, knockout, completed
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE tournament_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    player1_id UUID NOT NULL REFERENCES players(id),
    player2_id UUID NOT NULL REFERENCES players(id),
    team_number INTEGER NOT NULL
);

CREATE TABLE tournament_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    team_a_id UUID NOT NULL REFERENCES tournament_teams(id),
    team_b_id UUID NOT NULL REFERENCES tournament_teams(id),
    stage VARCHAR(20) NOT NULL,  -- round_robin, semi, final
    match_id UUID REFERENCES matches(id),  -- Link to scoring system
    winner_team_id UUID REFERENCES tournament_teams(id),
    match_order INTEGER
);

CREATE TABLE tournament_standings (
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    played INTEGER NOT NULL DEFAULT 0,
    won INTEGER NOT NULL DEFAULT 0,
    lost INTEGER NOT NULL DEFAULT 0,
    points INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tournament_id, team_id)
);
```

### Phase 4: Backend Handlers ✅
**Effort**: 2-3 hours

1. Create `handler/tournament_handler.go`
2. Create `repository/tournament_repository.go`
3. Create `service/tournament_service.go`
4. Wire up routes in `cmd/api/main.go`

### Phase 5: Testing & Validation ✅
**Effort**: 2 hours

1. Unit tests for scoring engine
2. Unit tests for tournament engine
3. Integration tests for API endpoints
4. End-to-end tournament flow test

---

## Testing Strategy

### Scoring Engine Tests
```go
func TestStandardTennisScoring(t *testing.T)
func TestDeuceAndAdvantage(t *testing.T)
func TestShortFormatThreeGames(t *testing.T)
func TestShortFormatEarlyWin(t *testing.T)
func TestSetWinConditions(t *testing.T)
func TestMatchWinConditions(t *testing.T)
func TestServerRotation(t *testing.T)
func TestDisplayMapping(t *testing.T)
```

### Tournament Engine Tests
```go
func TestRandomTeamGeneration(t *testing.T)
func TestManualTeamCreation(t *testing.T)
func TestRoundRobinMatchGeneration(t *testing.T)
func TestStandingsCalculation(t *testing.T)
func TestRanking(t *testing.T)
func TestKnockoutAdvancement3Teams(t *testing.T)
func TestKnockoutAdvancement4Teams(t *testing.T)
func TestKnockoutAdvancement5PlusTeams(t *testing.T)
```

---

## Validation Checklist

### Scoring Engine ✅
- [ ] No tournament awareness
- [ ] Pure functions only
- [ ] Supports Standard mode
- [ ] Supports Short-Format mode
- [ ] Correct Deuce/Advantage logic
- [ ] Never shows raw point counts
- [ ] Server rotation works correctly
- [ ] All tests pass

### Tournament Engine ✅
- [ ] No scoring logic implementation
- [ ] Random team generation is deterministic with seed
- [ ] Round-robin uses T × (T − 1) / 2
- [ ] Standings track: played, won, lost, points
- [ ] Knockout rules for 3, 4, and 5+ teams
- [ ] All tests pass

### Integration ✅
- [ ] API endpoints work correctly
- [ ] Database schema is normalized
- [ ] Scoring state persisted correctly
- [ ] Tournament state persisted correctly
- [ ] End-to-end flow works

---

## Deliverables

1. ✅ `backend/internal/scoring/` - Complete scoring engine
2. ✅ `backend/internal/tournament/` - Complete tournament engine
3. ✅ Database migration for tournaments
4. ✅ API handlers for tournaments
5. ✅ Unit tests (>90% coverage)
6. ✅ Integration tests
7. ✅ Updated API documentation

---

## Timeline

- **Phase 1**: Scoring Engine - 4 hours
- **Phase 2**: Tournament Engine - 4 hours
- **Phase 3**: Database Schema - 1 hour
- **Phase 4**: Backend Handlers - 3 hours
- **Phase 5**: Testing - 2 hours

**Total Estimated Effort**: 14 hours

---

## Success Criteria

The implementation will be considered complete when:

1. **Scoring Engine**:
   - Passes all unit tests
   - Correctly implements both match modes
   - Has zero knowledge of tournaments
   - Display matches tennis notation exactly

2. **Tournament Engine**:
   - Passes all unit tests
   - Correctly generates teams and matches
   - Accurately calculates standings
   - Properly advances to knockout stage
   - Has zero scoring logic

3. **Integration**:
   - API endpoints work end-to-end
   - Database persists state correctly
   - Frontend can consume APIs
   - Full tournament can be completed

4. **Code Quality**:
   - All code is documented
   - All public functions have comments
   - Tests cover edge cases
   - No code duplication
   - Follows Go best practices

---

## Notes

- This implementation is **backend-first**
- Frontend integration is a separate phase
- Both engines are designed for **maximum testability**
- Scoring engine can be ported to frontend (JavaScript) later
- All logic is **deterministic** and **reproducible**

---

**STATUS**: Ready to implement Phase 1 (Scoring Engine)
