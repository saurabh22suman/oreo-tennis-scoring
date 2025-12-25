# 🎾 Oreo Tennis Scoring (OTS)
## Tennis Scoring Logic, Match Modes & Random Doubles Team Selection

---

## 1. Purpose

This document defines the **complete tennis scoring logic** for Oreo Tennis Scoring (OTS), including:
- Standard tennis scoring (points, games, sets, match)
- Short-format recreational match mode (3-game rotational)
- Random team selection for doubles matches

This document is the **single source of truth** for scoring logic and match rules.

---

## 2. Tennis Scoring Model Overview

Tennis scoring progresses through four layers:

POINT → GAME → SET → MATCH

OTS stores points as integers internally but always displays **tennis symbols** to users.

---

## 3. Point-Level Logic (Within a Game)

### 3.1 Internal Representation

Each side (player or team):

points: integer (0, 1, 2, 3, ...)

---

### 3.2 Display Mapping

0 → 0  
1 → 15  
2 → 30  
3 → 40  

Points never display beyond 40.

---

### 3.3 Game Win Conditions

A side wins a game if:
- Points ≥ 4
- Lead by ≥ 2 points

Examples:
- 4–0, 4–2, 5–3 → Game

---

### 3.4 Deuce & Advantage

When both sides reach 40 (3 points):
- State becomes **Deuce**

From Deuce:
- Win point → Advantage
- Lose next point → Deuce
- Win point while in Advantage → Game

Display:
- Deuce
- Ad A / Ad B

---

### 3.5 Game Reset

After game win:
- Reset points to 0
- Increment games won
- Update server (mode dependent)

---

## 4. Standard Tennis Mode

### 4.1 Game-Level (Within a Set)

Each side tracks:

games: integer

Set win:
- Games ≥ 6
- Lead by ≥ 2 games

Examples:
- 6–0, 6–4, 7–5

---

### 4.2 Tie-Break (Optional)

Trigger: 6–6

Rules:
- First to 7 points
- Must lead by 2

Winning tie-break:
- Set score becomes 7–6

---

### 4.3 Match-Level

Default:
- Best of 3 sets
- First to 2 sets wins

---

## 5. Short-Format Rotational Match Mode

### 5.1 Overview

A recreational format commonly used:

- Maximum 3 games
- Best of 3 games
- Fixed serving order per game
- No sets involved

Hierarchy:

POINT → GAME → MATCH

---

### 5.2 Match Rules

- Normal tennis rules inside each game
- Match ends when a side wins 2 games
- If a side wins Games 1 and 2, Game 3 is skipped

---

### 5.3 Serving Order (Critical)

Before match start:
- Exactly 3 servers are selected in order

Example:
servers = [Player1, Player2, Player3]

Serving:
- Game 1 → Player1
- Game 2 → Player2
- Game 3 → Player3

Server does NOT depend on previous game outcome.

---

### 5.4 UI Display (Short Mode)

Example:

Game 2 of 3  
40 : 30  
Games 1 : 0  
Server: Player 2

---

## 6. Random Doubles Team Selection

### 6.1 Purpose

Allows quick and fair team assignment for doubles matches.

---

### 6.2 Eligibility Rules

- Match type: Doubles
- Player count must be even
- Default supported: 4 players

---

### 6.3 Randomization Logic

Steps:
1. Shuffle selected players
2. Split evenly:
   - First half → Team A
   - Second half → Team B

Teams are locked once match starts.

---

### 6.4 UI Behavior

- Button: “Randomize Teams”
- Visible only when eligible
- Allow re-randomize before match start
- Show confirmation before proceeding

---For this random  doubles team selection

## 7. Score Display Rules

- NEVER show raw numeric points
- Always show tennis symbols
- Priority:
  1. Point score
  2. Game score
  3. Set score (if applicable)

---

## 8. AI-Agent Implementation Prompt

You are implementing scoring logic for Oreo Tennis Scoring (OTS).

You MUST support:
- Standard Tennis Mode
- Short Rotational Mode (3-game best of 3)

Rules:For this random  doubles team selection
- Correct tennis scoring (0,15,30,40,Deuce,Advantage)
- Deterministic logic
- No UI logic inside scoring engine
- Configurable match modes

Do NOT:
- Simplify rules
- Display raw point counts
- Mix match modes

Deliver:
- Scoring state machine
- Pure, testable functions
- Clear transitions between states

---

## 9. Final Notes

Tennis scoring is stateful and non-linear.
Correctness > simplicity.
This document is authoritative.
