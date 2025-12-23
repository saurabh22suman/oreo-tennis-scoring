// IndexedDB service for offline storage
import { openDB } from 'idb';

const DB_NAME = 'ots-db';
const DB_VERSION = 1;

let dbPromise = null;

export function getDB() {
    if (!dbPromise) {
        dbPromise = openDB(DB_NAME, DB_VERSION, {
            upgrade(db) {
                // Store for current match state
                if (!db.objectStoreNames.contains('currentMatch')) {
                    db.createObjectStore('currentMatch', { keyPath: 'id' });
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
            },
        });
    }
    return dbPromise;
}

// ═══════════════════════════════════════════════════
// CURRENT MATCH
// ═══════════════════════════════════════════════════

export async function saveCurrentMatch(match) {
    const db = await getDB();
    await db.put('currentMatch', { ...match, id: 'current' });
}

export async function getCurrentMatch() {
    const db = await getDB();
    return await db.get('currentMatch', 'current');
}

export async function clearCurrentMatch() {
    const db = await getDB();
    await db.delete('currentMatch', 'current');
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
