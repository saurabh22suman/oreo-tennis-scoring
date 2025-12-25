// API service for backend communication
const API_BASE = import.meta.env.VITE_API_URL || '';

class ApiError extends Error {
    constructor(message, status) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
    }
}

async function request(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;

    const config = {
        headers: {
            'Content-Type': 'application/json',
        },
        credentials: 'include', // For cookies
        ...options,
    };

    if (config.body && typeof config.body === 'object') {
        config.body = JSON.stringify(config.body);
    }

    const response = await fetch(url, config);
    const data = await response.json();

    if (!response.ok) {
        throw new ApiError(data.error || 'Request failed', response.status);
    }

    return data.data;
}

// ═══════════════════════════════════════════════════
// AUTH
// ═══════════════════════════════════════════════════

export async function login(username, password) {
    return await request('/api/admin/login', {
        method: 'POST',
        body: { username, password },
    });
}

export async function logout() {
    return await request('/api/admin/logout', {
        method: 'POST',
    });
}

export async function checkAuth() {
    try {
        await request('/api/admin/check');
        return true;
    } catch {
        return false;
    }
}

// ═══════════════════════════════════════════════════
// PLAYERS
// ═══════════════════════════════════════════════════

export async function getPlayers() {
    return await request('/api/players');
}

export async function getAdminPlayers() {
    return await request('/api/admin/players');
}

export async function createPlayer(name) {
    return await request('/api/admin/players', {
        method: 'POST',
        body: { name },
    });
}

export async function updatePlayer(id, data) {
    return await request(`/api/admin/players/${id}`, {
        method: 'PATCH',
        body: data,
    });
}

// ═══════════════════════════════════════════════════
// VENUES
// ═══════════════════════════════════════════════════

export async function getVenues() {
    return await request('/api/venues');
}

export async function getAdminVenues() {
    return await request('/api/admin/venues');
}

export async function createVenue(name, surface) {
    return await request('/api/admin/venues', {
        method: 'POST',
        body: { name, surface },
    });
}

export async function updateVenue(id, data) {
    return await request(`/api/admin/venues/${id}`, {
        method: 'PATCH',
        body: data,
    });
}

export async function getVenueTendencies(venueId) {
    return await request(`/api/venues/${venueId}/tendencies`);
}

// ═══════════════════════════════════════════════════
// MATCHES
// ═══════════════════════════════════════════════════

export async function createMatch(venueId, matchType, teamA, teamB) {
    return await request('/api/matches', {
        method: 'POST',
        body: {
            venue_id: venueId,
            match_type: matchType,
            team_a: teamA,
            team_b: teamB,
        },
    });
}

export async function submitEvents(matchId, events) {
    const formattedEvents = events.map(e => ({
        id: e.id,
        timestamp: e.timestamp,
        server_player_id: e.serverPlayerId,
        serve_type: e.serveType,
        point_winner_team: e.pointWinnerTeam,
    }));

    return await request(`/api/matches/${matchId}/events`, {
        method: 'POST',
        body: { events: formattedEvents },
    });
}

export async function completeMatch(matchId) {
    return await request(`/api/matches/${matchId}/complete`, {
        method: 'POST',
    });
}

export async function getMatchSummary(matchId) {
    return await request(`/api/matches/${matchId}/summary`);
}

export async function getMatches() {
    return await request('/api/admin/matches');
}

export async function deleteMatch(matchId) {
    return await request(`/api/admin/matches/${matchId}`, {
        method: 'DELETE',
    });
}

// ═══════════════════════════════════════════════════
// SYNC SERVICE
// ═══════════════════════════════════════════════════

import { getUnsyncedEvents, markEventsSynced } from './db.js';

export async function syncEvents(matchId) {
    try {
        const unsyncedEvents = await getUnsyncedEvents(matchId);
        if (unsyncedEvents.length === 0) return { synced: 0 };

        const result = await submitEvents(matchId, unsyncedEvents);
        await markEventsSynced(unsyncedEvents.map(e => e.id));

        return { synced: result.inserted, total: unsyncedEvents.length };
    } catch (error) {
        console.error('Sync failed:', error);
        throw error;
    }
}
