# 🏆 Oreo Tennis Scoring (OTS)
## Tournament Specification (Doubles – Round Robin + Knockouts)

---

## 1. Purpose

This document defines the **Tournament mode** for Oreo Tennis Scoring (OTS).

Tournament mode is designed for **club-style recreational play**, where:
- Players form doubles teams
- Teams play a round-robin stage
- Top teams advance to knockout matches (Final or Semifinals)

This spec is authoritative and must be followed exactly.

---

## 2. Tournament Overview

### Supported Tournament Type
- **Doubles only** (initial version)

### High-Level Flow
```
Players → Teams → Round Robin → Knockout → Winner
```

---

## 3. Player & Team Formation

### 3.1 Player Selection
- Users select all participating players
- Minimum players: 4
- Total players must be even

### 3.2 Number of Teams
```
number_of_teams = total_players / 2
```

Example:
- 6 players → 3 teams
- 8 players → 4 teams

---

### 3.3 Team Creation Modes

#### A. Random Team Generation
Steps:
1. Shuffle player list randomly
2. Pair players sequentially

Example:
```
Players: [P1, P2, P3, P4, P5, P6]
Shuffle → [P4, P1, P6, P2, P5, P3]

Teams:
T1 = P4 + P1
T2 = P6 + P2
T3 = P5 + P3
```

#### B. Manual Team Creation
- User manually assigns players to teams
- Validation:
  - Each team has exactly 2 players
  - No player appears in more than one team

### 3.4 Team Locking
- Teams are **locked** once the tournament starts
- No changes allowed mid-tournament

---

## 4. Round Robin Stage

### 4.1 Definition
In round robin:
> Every team plays every other team exactly once

---

### 4.2 Match Generation Formula
For `T` teams:
```
total_matches = T × (T − 1) / 2
```

Examples:
- 3 teams → 3 matches
- 4 teams → 6 matches
- 5 teams → 10 matches

---

### 4.3 Match List Generation (Example: 3 Teams)

Teams:
```
T1, T2, T3
```

Matches:
```
M1: T1 vs T2
M2: T1 vs T3
M3: T2 vs T3
```

Match order:
- Sequential OR
- User-selected

---

### 4.4 Match Rules
- Each match uses standard OTS match scoring
- Matches are independent
- Winner is reported back to tournament standings

---

## 5. Tournament Standings

### 5.1 Stats Tracked Per Team
```
played
won
lost
points
```

---

### 5.2 Points System
- Win → 1 point
- Loss → 0 points

(No draws supported)

---

### 5.3 Ranking Rules
Teams are ranked by:
1. Points (descending)
2. Head-to-head result
3. (Optional) Games difference

---

## 6. Knockout Stage

### 6.1 Advancement Rules

#### Case A: 3 Teams
- Top 2 teams advance
- **Final only**
```
Final: Rank 1 vs Rank 2
```

---

#### Case B: 4 Teams
- All 4 advance
- Semifinals:
  - SF1: Rank 1 vs Rank 4
  - SF2: Rank 2 vs Rank 3
- Winners → Final

---

#### Case C: 5+ Teams
- Top 4 advance to semifinals
- Remaining teams eliminated

---

## 7. Tournament Completion

- Tournament ends after Final
- Winner is declared
- No further matches allowed
- Match & tournament summary generated

---

## 8. UI Flow Summary

### Home Screen
```
[ Start New Match ]
[ Start New Tournament ]
[ Admin ]
```

### Tournament Screens
1. Tournament Setup (venue, players)
2. Team Creation (random / manual)
3. Tournament Dashboard
4. Live Match (reused from Match flow)
5. Final Summary

---

## 9. Data Model (Minimal)

### Tournament
```
tournament_id
venue_id
status = setup | round_robin | knockout | completed
created_at
```

### Team
```
team_id
tournament_id
players[]
```

### Tournament Match
```
match_id
tournament_id
team_a_id
team_b_id
stage = round_robin | semi | final
winner_team_id
```

---

## 10. Engineering Rules

- Tournament logic must be separate from match logic
- Match scoring engine must not be modified
- No UI assumptions in core tournament engine
- All logic must be deterministic and testable

---

## 11. AI-Agent Implementation Prompt

You are implementing Tournament mode for Oreo Tennis Scoring (OTS).

You MUST:
- Support doubles tournaments only
- Generate teams from selected players
- Generate round-robin matches correctly
- Maintain tournament standings
- Transition to knockout stage automatically

DO NOT:
- Invent new tournament formats
- Modify tennis scoring rules
- Add unnecessary complexity

Deliver:
- Team generator
- Round-robin match generator
- Standings calculator
- Knockout match generator

This document is the single source of truth.
