// Svelte stores for application state
import { writable, derived } from 'svelte/store';

// ═══════════════════════════════════════════════════
// NAVIGATION
// ═══════════════════════════════════════════════════

export const currentScreen = writable('home');

export function navigate(screen) {
    currentScreen.set(screen);
}

// ═══════════════════════════════════════════════════
// AUTH STATE
// ═══════════════════════════════════════════════════

export const isAdmin = writable(false);

// ═══════════════════════════════════════════════════
// MATCH STATE
// ═══════════════════════════════════════════════════

const initialMatchState = {
    id: null,
    venueId: null,
    venueName: null,
    matchType: 'singles',
    matchMode: 'standard',
    bestOf: null, // 3 or 5 for short format
    teamA: [],
    teamB: [],
    events: [],
    currentServer: null,
    serverTeam: 'A',
    startedAt: null,
};

export const matchState = writable({ ...initialMatchState });

export function resetMatch() {
    matchState.set({ ...initialMatchState });
}

// Derived store for score calculation
export const score = derived(matchState, ($match) => {
    let teamA = 0;
    let teamB = 0;

    for (const event of $match.events) {
        if (event.pointWinnerTeam === 'A') {
            teamA++;
        } else {
            teamB++;
        }
    }

    return { teamA, teamB };
});

// ═══════════════════════════════════════════════════
// ADMIN VIEW MATCH
// ═══════════════════════════════════════════════════

// Used when admin wants to view a completed match summary
export const viewMatchId = writable(null);

// ═══════════════════════════════════════════════════
// PLAYERS & VENUES CACHE
// ═══════════════════════════════════════════════════

export const players = writable([]);
export const venues = writable([]);

// ═══════════════════════════════════════════════════
// LOADING & ERROR STATES
// ═══════════════════════════════════════════════════

export const isLoading = writable(false);
export const error = writable(null);

export function setError(message) {
    error.set(message);
    setTimeout(() => error.set(null), 5000);
}

export function clearError() {
    error.set(null);
}
