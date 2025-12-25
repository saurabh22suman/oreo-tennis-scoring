# 📊 Oreo Tennis Scoring (OTS)
## Venue-Level Team & Player Tendencies Specification

---

## 1. Purpose

This document defines the **Team & Player Tendencies at Venue** feature for Oreo Tennis Scoring (OTS).

The goal of this feature is to surface **patterns and tendencies**, not rankings or judgments.
It is designed to help players understand how matches behave at a specific venue.

This feature is:
- Venue-centric
- Aggregate-first
- Psychologically safe for a closed group

This document is authoritative.

---

## 2. Scope & Principles

### What This Feature IS
- Aggregate analysis per venue
- Insight into doubles team performance
- Self-reflective player metrics

### What This Feature IS NOT
- Player rankings
- Leaderboards
- Public performance comparison
- Shareable personal stats

---

## 3. Eligibility & Gating Rules (Critical)

### Team Tendencies
A team is eligible only if:
- It has played **at least 3 matches** at the venue

### Player Tendencies
A player is eligible only if:
- They have played **at least 5 matches** at the venue

If criteria are not met:
- Do NOT display the team/player
- Do NOT show partial or placeholder stats

---

## 4. Team Tendencies (Doubles Only)

### Metrics (Per Venue)

For each eligible team:
- Matches played
- Matches won
- Win percentage
- Average games per match
- Deuce percentage in matches involving the team
- First serve points won percentage (team aggregate)

### Display Rules
- Teams must NOT be ranked by default
- Display in neutral order (e.g., alphabetical or recent)
- Clearly label as tendencies, not performance rating

Example:
```
Team: Ajit / Saurabh
Matches: 5
Win %: 60%
Avg Games / Match: 14.2
Deuce %: 28%
```

---

## 5. Player Tendencies (Aggregate, Neutral)

### Metrics (Per Venue)

For each eligible player:
- Matches played
- First serve in percentage
- Double faults per game
- Average points per game (optional)

### Restrictions
- Do NOT show player win percentage
- Do NOT rank players
- Do NOT compare players side-by-side

Example:
```
Player: Saurabh
Matches: 9
1st Serve In: 63%
DF / Game: 0.38
```

---

## 6. UI & UX Rules

### Placement
- Section appears on Venue Dashboard
- Always **collapsed by default**
- User must explicitly expand it

### Labeling
Section title:
```
Team & Player Tendencies
(based on matches at this venue)
```

### Presentation
- Card-based layout
- No charts
- No tables
- No leaderboards
- Limit visible entries to 3–5

---

## 7. Sharing & Privacy Rules

### Share Image Generation
- Team & Player Tendencies MUST NOT be included
- Shareable venue summary remains aggregate-only

### Rationale
- Prevents accidental shaming
- Preserves group harmony
- Keeps venue summaries neutral

---

## 8. Data & Aggregation Rules

This feature MUST be implemented using **derived aggregations** only.

No new raw data tables are allowed.

Suggested derived views:
- venue_team_stats
- venue_player_stats

All calculations must be deterministic and reproducible.

---

## 9. Performance Constraints

- Aggregations must be precomputed or cached
- Dashboard load must not exceed baseline venue dashboard time
- Avoid per-request heavy joins

---

## 10. Final Rules

- Venue > Team > Player (priority)
- Insight > Comparison
- Patterns > Rankings
- Respect group dynamics

This document is the single source of truth for this feature.
