// Tennis scoring engine - Pure state machine for tennis scoring logic
// Supports both Standard and Short-Format (Rotational) match modes

export const MatchMode = {
    STANDARD: 'standard',      // Full tennis: Points → Games → Sets → Match
    SHORT_FORMAT: 'short'      // Recreational: Points → Games (best of 3)
};

export const GameState = {
    IN_PROGRESS: 'in_progress',
    DEUCE: 'deuce',
    ADVANTAGE_A: 'advantage_a',
    ADVANTAGE_B: 'advantage_b',
    WON_A: 'won_a',
    WON_B: 'won_b'
};

// ═══════════════════════════════════════════════════
// POINT DISPLAY MAPPING
// ═══════════════════════════════════════════════════

export function getPointDisplay(points) {
    const mapping = { 0: '0', 1: '15', 2: '30', 3: '40' };
    return mapping[points] || '40';
}

// ═══════════════════════════════════════════════════
// GAME STATE MACHINE
// ═══════════════════════════════════════════════════

export function getGameState(pointsA, pointsB) {
    // Both at 40+ → Deuce or Advantage
    if (pointsA >= 3 && pointsB >= 3) {
        const diff = pointsA - pointsB;
        if (diff === 0) return GameState.DEUCE;
        if (diff === 1) return GameState.ADVANTAGE_A;
        if (diff === -1) return GameState.ADVANTAGE_B;
        if (diff >= 2) return GameState.WON_A;
        if (diff <= -2) return GameState.WON_B;
    }

    // Normal win condition: >= 4 points with 2+ lead
    if (pointsA >= 4 && pointsA - pointsB >= 2) return GameState.WON_A;
    if (pointsB >= 4 && pointsB - pointsA >= 2) return GameState.WON_B;

    return GameState.IN_PROGRESS;
}

export function getGameDisplayText(pointsA, pointsB) {
    const state = getGameState(pointsA, pointsB);

    switch (state) {
        case GameState.DEUCE:
            return { a: 'Deuce', b: 'Deuce' };
        case GameState.ADVANTAGE_A:
            return { a: 'Ad', b: '40' };
        case GameState.ADVANTAGE_B:
            return { a: '40', b: 'Ad' };
        default:
            return { a: getPointDisplay(pointsA), b: getPointDisplay(pointsB) };
    }
}

// ═══════════════════════════════════════════════════
// SET LOGIC (Standard Mode Only)
// ═══════════════════════════════════════════════════

export function isSetWon(gamesA, gamesB) {
    // Win condition: >= 6 games with 2+ lead
    if (gamesA >= 6 && gamesA - gamesB >= 2) return 'A';
    if (gamesB >= 6 && gamesB - gamesA >= 2) return 'B';

    // Tie-break at 6-6 results in 7-6 (handled separately)
    if (gamesA === 7 && gamesB === 6) return 'A';
    if (gamesB === 7 && gamesA === 6) return 'B';

    return null;
}

export function isTieBreak(gamesA, gamesB) {
    return gamesA === 6 && gamesB === 6;
}

// ═══════════════════════════════════════════════════
// MATCH STATE
// ═══════════════════════════════════════════════════

export function createMatchState(mode, players, servers = null) {
    const isShortFormat = mode === MatchMode.SHORT_FORMAT;

    return {
        mode,
        players,
        servers: isShortFormat ? servers : null, // [Player1, Player2, Player3] for short format

        // Current game state
        currentGame: {
            pointsA: 0,
            pointsB: 0,
            gameNumber: 1,           // 1, 2, or 3 for short format
            serverIndex: 0           // Index into servers array (short format)
        },

        // Games won (tracked differently per mode)
        gamesA: 0,
        gamesB: 0,

        // Standard mode only
        setsA: 0,
        setsB: 0,
        currentSet: 1,

        // Match result
        winner: null,
        completed: false
    };
}

// ═══════════════════════════════════════════════════
// POINT SCORING
// ═══════════════════════════════════════════════════

export function scorePoint(matchState, team) {
    const state = { ...matchState };
    const game = { ...state.currentGame };

    // Add point
    if (team === 'A') {
        game.pointsA++;
    } else {
        game.pointsB++;
    }

    // Check if game is won
    const gameState = getGameState(game.pointsA, game.pointsB);

    if (gameState === GameState.WON_A || gameState === GameState.WON_B) {
        const winner = gameState === GameState.WON_A ? 'A' : 'B';
        return handleGameWon(state, winner);
    }

    state.currentGame = game;
    return state;
}

// ═══════════════════════════════════════════════════
// GAME WIN HANDLING
// ═══════════════════════════════════════════════════

function handleGameWon(matchState, winner) {
    const state = { ...matchState };

    if (state.mode === MatchMode.SHORT_FORMAT) {
        return handleShortFormatGameWon(state, winner);
    } else {
        return handleStandardGameWon(state, winner);
    }
}

// Short format: Best of 3 games
function handleShortFormatGameWon(state, winner) {
    // Increment games
    if (winner === 'A') state.gamesA++;
    else state.gamesB++;

    // Check match win (first to 2 games)
    if (state.gamesA === 2) {
        state.winner = 'A';
        state.completed = true;
        return state;
    }
    if (state.gamesB === 2) {
        state.winner = 'B';
        state.completed = true;
        return state;
    }

    // Move to next game
    state.currentGame = {
        pointsA: 0,
        pointsB: 0,
        gameNumber: state.currentGame.gameNumber + 1,
        serverIndex: state.currentGame.serverIndex + 1
    };

    return state;
}

// Standard format: Games → Sets → Match
function handleStandardGameWon(state, winner) {
    // Increment games in current set
    if (winner === 'A') state.gamesA++;
    else state.gamesB++;

    // Check if set is won
    const setWinner = isSetWon(state.gamesA, state.gamesB);

    if (setWinner) {
        // Set won
        if (setWinner === 'A') state.setsA++;
        else state.setsB++;

        // Check match win (best of 3 sets = first to 2)
        if (state.setsA === 2) {
            state.winner = 'A';
            state.completed = true;
            return state;
        }
        if (state.setsB === 2) {
            state.winner = 'B';
            state.completed = true;
            return state;
        }

        // Start new set
        state.currentSet++;
        state.gamesA = 0;
        state.gamesB = 0;
    }

    // Reset game points
    state.currentGame = {
        pointsA: 0,
        pointsB: 0,
        gameNumber: state.gamesA + state.gamesB + 1,
        serverIndex: 0
    };

    return state;
}

// ═══════════════════════════════════════════════════
// DISPLAY HELPERS
// ═══════════════════════════════════════════════════

export function getMatchDisplay(matchState) {
    const { mode, currentGame, gamesA, gamesB, setsA, setsB } = matchState;
    const pointDisplay = getGameDisplayText(currentGame.pointsA, currentGame.pointsB);

    if (mode === MatchMode.SHORT_FORMAT) {
        return {
            points: pointDisplay,
            games: { a: gamesA, b: gamesB },
            gameNumber: currentGame.gameNumber,
            totalGames: 3,
            server: matchState.servers ? matchState.servers[currentGame.serverIndex] : null
        };
    } else {
        return {
            points: pointDisplay,
            games: { a: gamesA, b: gamesB },
            sets: { a: setsA, b: setsB },
            currentSet: matchState.currentSet,
            isTieBreak: isTieBreak(gamesA, gamesB)
        };
    }
}

// ═══════════════════════════════════════════════════
// RANDOM TEAM SELECTION (Doubles)
// ═══════════════════════════════════════════════════

export function randomizeTeams(playerIds) {
    if (playerIds.length % 2 !== 0) {
        throw new Error('Player count must be even for doubles');
    }

    // Shuffle using Fisher-Yates
    const shuffled = [...playerIds];
    for (let i = shuffled.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }

    // Split evenly
    const midpoint = Math.floor(shuffled.length / 2);
    return {
        teamA: shuffled.slice(0, midpoint),
        teamB: shuffled.slice(midpoint)
    };
}
