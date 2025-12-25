# OTS Core Logic - Quick Reference Guide

## 📋 Table of Contents
1. [Scoring Engine Usage](#scoring-engine-usage)
2. [Tournament Engine Usage](#tournament-engine-usage)
3. [Common Patterns](#common-patterns)
4. [Error Handling](#error-handling)

---

## Scoring Engine Usage

### Standard Tennis Match

```go
import "github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/scoring"

// Setup players
players := scoring.TeamPlayers{
    TeamA: []string{"player1", "player2"},
    TeamB: []string{"player3", "player4"},
}

// Create standard match
match, err := scoring.NewMatchState(scoring.ModeStandard, players, nil)
if err != nil {
    // Handle error
}

// Score points
match, _ = scoring.ScorePoint(match, scoring.TeamA)  // 15-0
match, _ = scoring.ScorePoint(match, scoring.TeamA)  // 30-0
match, _ = scoring.ScorePoint(match, scoring.TeamB)  // 30-15
match, _ = scoring.ScorePoint(match, scoring.TeamA)  // 40-15
match, _ = scoring.ScorePoint(match, scoring.TeamA)  // Game to A

// Get display
display := scoring.GetMatchDisplay(match)
fmt.Printf("Score: %s - %s\n", display.Points.A, display.Points.B)
fmt.Printf("Games: %d - %d\n", display.Games.A, display.Games.B)
fmt.Printf("Sets: %d - %d\n", display.Sets.A, display.Sets.B)

// Check if complete
if scoring.IsMatchComplete(match) {
    winner := scoring.GetWinner(match)
    fmt.Printf("Winner: Team %s\n", *winner)
}
```

### Short-Format Match (3-game best-of-3)

```go
// Define 3 servers in rotation
servers := []string{"player1", "player2", "player3"}

// Create short-format match
match, err := scoring.NewMatchState(
    scoring.ModeShortFormat,
    players,
    servers,
)

// Score points same as standard mode
match, _ = scoring.ScorePoint(match, scoring.TeamA)

// Display includes current server
display := scoring.GetMatchDisplay(match)
fmt.Printf("Game %d of %d\n", display.GameNumber, display.TotalGames)
fmt.Printf("Server: %s\n", *display.Server)
```

---

## Tournament Engine Usage

### Creating a Tournament

```go
import (
    "github.com/google/uuid"
    "github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/tournament"
)

// Setup
venueID := uuid.New()
playerIDs := []uuid.UUID{
    uuid.New(), // Player 1
    uuid.New(), // Player 2
    uuid.New(), // Player 3
    uuid.New(), // Player 4
    uuid.New(), // Player 5
    uuid.New(), // Player 6
}

// Create tournament
tourney, err := tournament.NewTournament(venueID, playerIDs)
if err != nil {
    // Handle error
}
// Tournament is now in "setup" stage
```

### Random Team Generation

```go
import "time"

// Generate random teams
seed := time.Now().UnixNano() // For true randomness
teams, err := tournament.GenerateRandomTeams(playerIDs, seed)
if err != nil {
    // Handle error
}

// For deterministic testing, use fixed seed
teams, _ = tournament.GenerateRandomTeams(playerIDs, 12345)
```

### Manual Team Creation

```go
// Manually create teams
p1, p2, p3, p4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

pairs := [][2]uuid.UUID{
    {p1, p2}, // Team 1
    {p3, p4}, // Team 2
}

teams, err := tournament.GenerateManualTeams(pairs)
if err != nil {
    // Handle validation errors (duplicate players, etc.)
}
```

### Starting the Tournament

```go
// Set teams (advances to round-robin stage)
tourney, err = tournament.SetTeams(tourney, teams)
if err != nil {
    // Handle error
}

// Get all round-robin matches
for _, match := range tourney.RoundRobinMatches {
    fmt.Printf("Match: Team %s vs Team %s\n", match.TeamAID, match.TeamBID)
}
```

### Recording Match Results

```go
// After a match is complete in the scoring engine
result := tournament.MatchResult{
    MatchID:      match.ID,
    WinnerTeamID: match.TeamAID,
    LoserTeamID:  match.TeamBID,
}

tourney, err = tournament.RecordMatchResult(tourney, result)
if err != nil {
    // Handle error
}

// Check standings
standings := tournament.CalculateRankings(tourney.Standings)
for _, standing := range standings {
    fmt.Printf("Rank %d: Team %s - %d points (%d-%d)\n",
        standing.Rank, standing.TeamID, standing.Points,
        standing.Won, standing.Lost)
}
```

### Advancing to Knockout Stage

```go
// After all round-robin matches complete
if tournament.IsRoundRobinComplete(tourney.RoundRobinMatches) {
    tourney, err = tournament.AdvanceToKnockout(tourney)
    if err != nil {
        // Handle error
    }
    
    // Get knockout matches
    for _, match := range tourney.KnockoutMatches {
        fmt.Printf("%s: Team %s vs Team %s\n",
            match.Stage, match.TeamAID, match.TeamBID)
    }
}
```

### Completing the Tournament

```go
// For 4+ team tournaments, after semifinals complete
if tournament.AreSemifinalsComplete(tourney.KnockoutMatches) {
    tourney, err = tournament.PrepareFinal(tourney)
    // Final matchup is now set
}

// Record final result
finalResult := tournament.MatchResult{
    MatchID:      finalMatch.ID,
    WinnerTeamID: winningTeamID,
    LoserTeamID:  losingTeamID,
}

tourney, _ = tournament.RecordMatchResult(tourney, finalResult)

// Tournament is now complete
if tourney.Completed {
    fmt.Printf("Tournament Winner: Team %s\n", *tourney.Winner)
}
```

---

## Common Patterns

### Pattern 1: Full Tournament Flow

```go
// 1. Create tournament
tourney, _ := tournament.NewTournament(venueID, playerIDs)

// 2. Generate teams
teams, _ := tournament.GenerateRandomTeams(playerIDs, seed)

// 3. Set teams
tourney, _ = tournament.SetTeams(tourney, teams)

// 4. Play all round-robin matches
for {
    nextMatch := tournament.GetNextMatch(tourney)
    if nextMatch == nil || tourney.Stage != tournament.StageRoundRobin {
        break
    }
    
    // Play match using scoring engine
    winner := playMatch(nextMatch) // Your implementation
    
    // Record result
    result := tournament.MatchResult{
        MatchID:      nextMatch.ID,
        WinnerTeamID: winner,
        LoserTeamID:  getLoser(nextMatch, winner),
    }
    tourney, _ = tournament.RecordMatchResult(tourney, result)
}

// 5. Advance to knockout
tourney, _ = tournament.AdvanceToKnockout(tourney)

// 6. Play semifinals (if any)
for {
    nextMatch := tournament.GetNextMatch(tourney)
    if nextMatch == nil || nextMatch.Stage != tournament.StageSemi {
        break
    }
    // Play and record result
}

// 7. Prepare final
if tournament.AreSemifinalsComplete(tourney.KnockoutMatches) {
    tourney, _ = tournament.PrepareFinal(tourney)
}

// 8. Play final
final := tournament.GetNextMatch(tourney)
// Play and record result

// 9. Winner declared
fmt.Printf("Winner: %s\n", *tourney.Winner)
```

### Pattern 2: Integrating Scoring with Tournament

```go
// Start a tournament match
tournamentMatch := tourney.RoundRobinMatches[0]

// Create scoring match
players := scoring.TeamPlayers{
    TeamA: getPlayersForTeam(tournamentMatch.TeamAID),
    TeamB: getPlayersForTeam(tournamentMatch.TeamBID),
}

scoringMatch, _ := scoring.NewMatchState(scoring.ModeStandard, players, nil)

// Play the match
for !scoring.IsMatchComplete(scoringMatch) {
    // User/system determines winner of point
    scoringMatch, _ = scoring.ScorePoint(scoringMatch, team)
}

// Match complete - record result in tournament
winner := scoring.GetWinner(scoringMatch)
var winnerTeamID uuid.UUID
if *winner == scoring.TeamA {
    winnerTeamID = tournamentMatch.TeamAID
} else {
    winnerTeamID = tournamentMatch.TeamBID
}

result := tournament.MatchResult{
    MatchID:      tournamentMatch.ID,
    WinnerTeamID: winnerTeamID,
    LoserTeamID:  getOtherTeam(tournamentMatch, winnerTeamID),
}

tourney, _ = tournament.RecordMatchResult(tourney, result)
```

---

## Error Handling

### Scoring Engine Errors

```go
// Invalid mode
_, err := scoring.NewMatchState("invalid_mode", players, nil)
// Error: "invalid match mode: invalid_mode"

// Short-format without servers
_, err = scoring.NewMatchState(scoring.ModeShortFormat, players, nil)
// Error: "short-format mode requires exactly 3 servers"

// Standard mode with servers
_, err = scoring.NewMatchState(scoring.ModeStandard, players, servers)
// Error: "standard mode must not specify servers array"

// Scoring after match complete
_, err = scoring.ScorePoint(completedMatch, scoring.TeamA)
// Error: "cannot score point: match is already completed"
```

### Tournament Engine Errors

```go
// Too few players
_, err := tournament.NewTournament(venueID, []uuid.UUID{uuid.New(), uuid.New()})
// Error: "minimum 4 players required for tournament"

// Odd number of players
_, err = tournament.GenerateRandomTeams(oddPlayers, seed)
// Error: "player count must be even for doubles"

// Duplicate player in manual teams
_, err = tournament.GenerateManualTeams(duplicatePairs)
// Error: "player xxx appears in multiple teams"

// Advancing before round-robin complete
_, err = tournament.AdvanceToKnockout(tourney)
// Error: "cannot advance: round-robin not complete"
```

---

## Best Practices

### 1. Always check errors
```go
match, err := scoring.ScorePoint(match, team)
if err != nil {
    log.Printf("Error scoring point: %v", err)
    return err
}
```

### 2. Use immutability correctly
```go
// CORRECT: Reassign returned value
match, _ = scoring.ScorePoint(match, team)

// INCORRECT: Ignoring returned value
scoring.ScorePoint(match, team) // State not updated!
```

### 3. Validate before operations
```go
if !scoring.IsMatchComplete(match) {
    match, _ = scoring.ScorePoint(match, team)
}

if tournament.IsRoundRobinComplete(tourney.RoundRobinMatches) {
    tourney, _ = tournament.AdvanceToKnockout(tourney)
}
```

### 4. Use deterministic seeds for testing
```go
// Production: random
teams, _ := tournament.GenerateRandomTeams(players, time.Now().UnixNano())

// Testing: fixed seed
teams, _ := tournament.GenerateRandomTeams(players, 12345)
```

---

## Testing

### Run all tests
```bash
cd backend
go test ./internal/scoring/... ./internal/tournament/... -v
```

### Test specific package
```bash
go test ./internal/scoring/... -v
go test ./internal/tournament/... -v
```

### Test coverage
```bash
go test ./internal/scoring/... -cover
go test ./internal/tournament/... -cover
```

---

## Summary

- **Scoring Engine**: `backend/internal/scoring`
  - Pure tennis logic
  - No tournament awareness
  - Immutable state updates

- **Tournament Engine**: `backend/internal/tournament`
  - Flow: Setup → Round Robin → Knockout → Complete
  - No scoring logic
  - Uses scoring engine as black box

Both engines are:
- ✅ Fully tested
- ✅ Production-ready
- ✅ 100% spec compliant
- ✅ Thread-safe (no shared mutable state)

---

For more details, see:
- `IMPLEMENTATION_COMPLETE.md` - Full implementation summary
- `OTS_Tennis_Scoring_Spec.md` - Tennis rules specification
- `OTS_Tournament_Spec.md` - Tournament rules specification
