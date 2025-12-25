# Oreo Tennis Scoring (OTS)
Product Requirements Document

## 1. Product Overview
Oreo Tennis Scoring (OTS) is a lightweight, mobile-first Progressive Web App (PWA) for tracking tennis matches and serve statistics for a closed group of players.

The app is designed to be installed via Chrome on mobile devices and optimized for real-time match usage with minimal interaction.

## 2. Goals
- Enable fast, distraction-free tennis match scoring
- Capture meaningful serve-level statistics
- Provide post-match analysis automatically
- Allow centralized control of venues and players via a single admin
- Work reliably without network connectivity during matches

## 3. Non-Goals
- Public user accounts
- Social features or rankings
- Professional umpire-level scoring rules
- Video or sensor-based analytics

## 4. User Roles

### Admin (Authenticated)
- Single admin user
- Credentials stored as hashed secrets in environment variables
- Manages players and venues
- Views and deletes matches

### Players / Users (Unauthenticated)
- No login
- Select venue and players
- Track match scores and serve stats

## 5. Core Features (MVP)

### Pre-Match Setup
- Select venue
- Select match type (singles or doubles)
- Select players from admin-managed list

### Live Match Tracking
- Track points by serve type:
  - First serve (won / lost)
  - Second serve (won / lost)
  - Double fault
- Automatic server toggle
- Simple score display
- One-screen interaction

### Post-Match Analysis
- Serve percentages per player
- Points won on first/second serve
- Double fault counts
- Simple visual summaries

## 6. Key Metrics
- First serve in %
- First serve points won %
- Second serve points won %
- Double faults
- Total points won

## 7. UX Principles
- One-handed mobile usage
- Maximum 3 taps per point
- No text input during live match
- Large buttons and clear contrast
- Dark mode by default

## 8. Technical Constraints
- Offline-first during matches
- Installable PWA
- JS bundle should remain lightweight (<200KB target)

## 9. Success Criteria
- Full match playable offline
- App loads in under 1 second on mobile
- Zero dependency on network during live scoring

## 10. Future Enhancements (Out of Scope)
- Player history dashboards
- Venue-based analytics
- Data export and backups
- Cross-device sync
