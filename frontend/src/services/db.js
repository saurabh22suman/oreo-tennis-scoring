// IndexedDB service for offline storage
import { openDB } from 'idb';

const DB_NAME = 'ots-db';
const DB_VERSION = 3; // Bumped for temp players support

// Match expiry time: 1 day in milliseconds
const MATCH_EXPIRY_MS = 24 * 60 * 60 * 1000;
// Temp player expiry time: 1 day in milliseconds
const TEMP_PLAYER_EXPIRY_MS = 24 * 60 * 60 * 1000;

let dbPromise = null;

export function getDB() {
    if (!dbPromise) {
        dbPromise = openDB(DB_NAME, DB_VERSION, {
            upgrade(db, oldVersion) {
                // Store for current match state (legacy - keep for migration)
                if (!db.objectStoreNames.contains('currentMatch')) {
                    db.createObjectStore('currentMatch', { keyPath: 'id' });
                }

                // Store for multiple incomplete matches (new in v2)
                if (!db.objectStoreNames.contains('incompleteMatches')) {
                    const matchStore = db.createObjectStore('incompleteMatches', { keyPath: 'matchId' });
                    matchStore.createIndex('createdAt', 'createdAt');
                }

                // Store for match events (offline queue)
                if (!db.objectStoreNames.contains('events')) {
                    const eventStore = db.createObjectStore('events', { keyPath: 'id' });
                    eventStore.createIndex('matchId', 'matchId');
                    eventStore.createIndex('synced', 'synced');
                }

                // Cache for players
                if (!db.objectStoreNames.contains('players')) {
                    db.createObjectStore('players', { keyPath: 'id' });
                }

                // Cache for venues
                if (!db.objectStoreNames.contains('venues')) {
                    db.createObjectStore('venues', { keyPath: 'id' });
                }

                // Store for temporary players (venue-specific, 1-day expiry)
                if (!db.objectStoreNames.contains('tempPlayers')) {
                    const tempStore = db.createObjectStore('tempPlayers', { keyPath: 'id' });
                    tempStore.createIndex('venueId', 'venueId');
                    tempStore.createIndex('createdAt', 'createdAt');
                }
            },
        });
    }
    return dbPromise;
}

// ═══════════════════════════════════════════════════
// CURRENT MATCH (Legacy - for backward compatibility)
// ═══════════════════════════════════════════════════

export async function saveCurrentMatch(match) {
    const db = await getDB();
    // Save to legacy store
    await db.put('currentMatch', { ...match, id: 'current' });
    
    // Also save to new incompleteMatches store if it has a matchId
    if (match.matchId || match.id) {
        const matchId = match.matchId || match.id;
        await saveIncompleteMatch(matchId, match);
    }
}

export async function getCurrentMatch() {
    try {
        const db = await getDB();
        return await db.get('currentMatch', 'current');
    } catch (err) {
        console.error('Failed to get current match:', err);
        return null;
    }
}

export async function clearCurrentMatch() {
    try {
        const db = await getDB();
        await db.delete('currentMatch', 'current');
    } catch (err) {
        console.error('Failed to clear current match:', err);
    }
}

// ═══════════════════════════════════════════════════
// INCOMPLETE MATCHES (Multiple match support)
// ═══════════════════════════════════════════════════

export async function saveIncompleteMatch(matchId, matchData) {
    const db = await getDB();
    const existing = await db.get('incompleteMatches', matchId);
    
    // Normalize venue data - it might be an object or just venueId/venueName
    const venue = matchData.venue || {
        id: matchData.venueId,
        name: matchData.venueName || 'Unknown venue',
    };
    
    await db.put('incompleteMatches', {
        matchId,
        venue,
        matchType: matchData.matchType,
        formatMode: matchData.matchMode || matchData.formatMode,
        teamA: matchData.teamA,
        teamB: matchData.teamB,
        score: matchData.score,
        events: matchData.events,
        currentServer: matchData.currentServer,
        serverTeam: matchData.serverTeam,
        completed: matchData.completed || false,
        createdAt: existing?.createdAt || Date.now(),
        updatedAt: Date.now(),
    });
}

export async function getIncompleteMatch(matchId) {
    const db = await getDB();
    return await db.get('incompleteMatches', matchId);
}

export async function getAllIncompleteMatches() {
    try {
        const db = await getDB();
        // Clean up expired matches first
        await cleanupExpiredMatches();
        // Return remaining matches sorted by most recent
        const matches = await db.getAll('incompleteMatches');
        return matches
            .filter(m => !m.completed)
            .sort((a, b) => (b.updatedAt || b.createdAt) - (a.updatedAt || a.createdAt));
    } catch (err) {
        console.error('Failed to get incomplete matches:', err);
        return [];
    }
}

export async function deleteIncompleteMatch(matchId) {
    try {
        const db = await getDB();
        await db.delete('incompleteMatches', matchId);
        // Also clear events for this match
        await clearMatchEvents(matchId);
    } catch (err) {
        console.error('Failed to delete incomplete match:', err);
    }
}

export async function cleanupExpiredMatches() {
    try {
        const db = await getDB();
        const now = Date.now();
        const tx = db.transaction('incompleteMatches', 'readwrite');
        const matches = await tx.store.getAll();
        
        for (const match of matches) {
            const age = now - (match.createdAt || 0);
            if (age > MATCH_EXPIRY_MS) {
                await tx.store.delete(match.matchId);
                // Note: events will be cleaned up separately
            }
        }
        await tx.done;
    } catch (err) {
        console.error('Failed to cleanup expired matches:', err);
    }
}

// ═══════════════════════════════════════════════════
// EVENTS (OFFLINE QUEUE)
// ═══════════════════════════════════════════════════

export async function saveEvent(event) {
    const db = await getDB();
    await db.put('events', { ...event, synced: false });
}

export async function getUnsyncedEvents(matchId) {
    const db = await getDB();
    const tx = db.transaction('events', 'readonly');
    const index = tx.store.index('matchId');
    const events = await index.getAll(matchId);
    return events.filter(e => !e.synced);
}

export async function getAllMatchEvents(matchId) {
    const db = await getDB();
    const tx = db.transaction('events', 'readonly');
    const index = tx.store.index('matchId');
    return await index.getAll(matchId);
}

export async function deleteLastEvent(matchId) {
    const db = await getDB();
    const tx = db.transaction('events', 'readwrite');
    const index = tx.store.index('matchId');
    const events = await index.getAll(matchId);
    if (events.length > 0) {
        // Sort by timestamp and delete the last one
        events.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
        const lastEvent = events[events.length - 1];
        await tx.store.delete(lastEvent.id);
        await tx.done;
        return lastEvent;
    }
    return null;
}

export async function markEventsSynced(eventIds) {
    const db = await getDB();
    const tx = db.transaction('events', 'readwrite');
    for (const id of eventIds) {
        const event = await tx.store.get(id);
        if (event) {
            event.synced = true;
            await tx.store.put(event);
        }
    }
    await tx.done;
}

export async function clearMatchEvents(matchId) {
    const db = await getDB();
    const tx = db.transaction('events', 'readwrite');
    const index = tx.store.index('matchId');
    const events = await index.getAll(matchId);
    for (const event of events) {
        await tx.store.delete(event.id);
    }
    await tx.done;
}

// ═══════════════════════════════════════════════════
// PLAYERS & VENUES CACHE
// ═══════════════════════════════════════════════════

export async function cachePlayers(players) {
    const db = await getDB();
    const tx = db.transaction('players', 'readwrite');
    await tx.store.clear();
    for (const player of players) {
        await tx.store.put(player);
    }
    await tx.done;
}

export async function getCachedPlayers() {
    const db = await getDB();
    return await db.getAll('players');
}

export async function cacheVenues(venues) {
    const db = await getDB();
    const tx = db.transaction('venues', 'readwrite');
    await tx.store.clear();
    for (const venue of venues) {
        await tx.store.put(venue);
    }
    await tx.done;
}

export async function getCachedVenues() {
    const db = await getDB();
    return await db.getAll('venues');
}

// ═══════════════════════════════════════════════════
// TEMPORARY PLAYERS (Venue-specific, 1-day validity)
// ═══════════════════════════════════════════════════

export async function createTempPlayer(name, venueId) {
    const db = await getDB();
    const id = `temp-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const tempPlayer = {
        id,
        name,
        venueId,
        active: true,
        isTemp: true,
        createdAt: Date.now(),
        expiresAt: Date.now() + TEMP_PLAYER_EXPIRY_MS,
    };
    await db.put('tempPlayers', tempPlayer);
    return tempPlayer;
}

export async function getTempPlayersForVenue(venueId) {
    try {
        const db = await getDB();
        // First cleanup expired temp players
        await cleanupExpiredTempPlayers();
        
        const tx = db.transaction('tempPlayers', 'readonly');
        const index = tx.store.index('venueId');
        const players = await index.getAll(venueId);
        
        // Filter out expired players
        const now = Date.now();
        return players.filter(p => p.expiresAt > now);
    } catch (err) {
        console.error('Failed to get temp players:', err);
        return [];
    }
}

export async function getAllTempPlayers() {
    try {
        const db = await getDB();
        await cleanupExpiredTempPlayers();
        
        const players = await db.getAll('tempPlayers');
        const now = Date.now();
        return players.filter(p => p.expiresAt > now);
    } catch (err) {
        console.error('Failed to get all temp players:', err);
        return [];
    }
}

export async function deleteTempPlayer(id) {
    try {
        const db = await getDB();
        await db.delete('tempPlayers', id);
    } catch (err) {
        console.error('Failed to delete temp player:', err);
    }
}

export async function cleanupExpiredTempPlayers() {
    try {
        const db = await getDB();
        const now = Date.now();
        const tx = db.transaction('tempPlayers', 'readwrite');
        const players = await tx.store.getAll();
        
        for (const player of players) {
            if (player.expiresAt <= now) {
                await tx.store.delete(player.id);
            }
        }
        await tx.done;
    } catch (err) {
        console.error('Failed to cleanup expired temp players:', err);
    }
}
