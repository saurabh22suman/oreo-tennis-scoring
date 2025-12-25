# OTS Frontend Specification

## 1. Stack
- Framework: Svelte
- Build Tool: Vite
- Styling: CSS variables + plain CSS
- Charts: Chart.js (post-match only)

## 2. PWA Requirements
- Service Worker for offline support
- Web App Manifest
- Installable via Chrome on Android

## 3. Screens

### Home
- Start new match
- Resume last incomplete match

### Match Setup
- Venue selection
- Match type selection
- Player selection

### Live Match Screen
- Current score
- Server indicator
- Large action buttons:
  - First Serve Won
  - First Serve Lost
  - Second Serve Won
  - Second Serve Lost
  - Double Fault

### Match Summary
- Serve stats per player
- Simple bar/percentage charts
- Match highlights

## 4. Offline Strategy
- Store all match events in IndexedDB
- Never block UI on network
- Sync automatically when connection is restored

## 5. State Management
- Svelte stores
- Match state isolated per session

## 6. UX Constraints
- No scrolling during match
- No keyboard input during match
- Thumb-zone optimized layout

## 7. Performance Targets
- Initial load < 1s
- Minimal JS bundle
- Charts loaded only post-match
