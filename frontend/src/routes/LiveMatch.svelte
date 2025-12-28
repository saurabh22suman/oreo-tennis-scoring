<!-- Live Match Screen - Screen 4 from ui_design_spec.md (MOST IMPORTANT) -->
<script>
  import { onMount } from 'svelte';
  import { navigate, matchState, players } from '../stores/app.js';
  import { saveEvent, getCurrentMatch, saveCurrentMatch, clearCurrentMatch, deleteIncompleteMatch } from '../services/db.js';
  import { syncEvents, completeMatch as apiCompleteMatch } from '../services/api.js';
  import { v4 as uuidv4 } from 'uuid';
  import { createMatchState, scorePoint, getMatchDisplay, MatchMode, startDeuceTiebreaker } from '../services/scoring.js';
  import Modal from '../lib/Modal.svelte';
  
  let syncing = false;
  let autoSyncInterval;
  let scoringState = null;
  let showWinnerModal = false;
  
  // Modal states
  let showEndMatchModal = false;
  let showSyncFailedModal = false;
  let showSyncFailedAlertModal = false;
  
  $: serverPlayer = $players.find(p => p.id === $matchState.currentServer);
  $: serverTeam = $matchState.serverTeam;
  $: display = scoringState ? getMatchDisplay(scoringState) : null;
  $: canStartDeuceTiebreaker = display?.canStartDeuceTiebreaker || false;
  $: inDeuceTiebreaker = display?.inDeuceTiebreaker || false;
  
  onMount(async () => {
    // Load match from IndexedDB if needed
    if (!$matchState.id) {
      const saved = await getCurrentMatch();
      if (saved) {
        matchState.set(saved);
      } else {
        navigate('home');
        return;
      }
    }
    
    // Initialize scoring state
    initializeScoringState();
    
    // Attempt sync every 30 seconds when online
    autoSyncInterval = setInterval(() => {
      if (navigator.onLine && !syncing) {
        attemptSync();
      }
    }, 30000);
    
    return () => {
      clearInterval(autoSyncInterval);
    };
  });
  
  function initializeScoringState() {
    // Determine mode from match state (default to standard if not set)
    const mode = $matchState.matchMode === 'short' ? MatchMode.SHORT_FORMAT : MatchMode.STANDARD;
    
    // For short format, we need servers array - use team players as servers
    const servers = mode === MatchMode.SHORT_FORMAT 
      ? [...$matchState.teamA, ...$matchState.teamB].slice(0, 5)
      : null;
    
    // Get best of (default to 3 for short format)
    const bestOf = $matchState.bestOf || 3;
    
    // Create initial scoring state
    scoringState = createMatchState(
      mode,
      { teamA: $matchState.teamA, teamB: $matchState.teamB },
      servers,
      bestOf
    );
    
    // Replay existing events to rebuild state
    if ($matchState.events) {
      for (const event of $matchState.events) {
        scoringState = scorePoint(scoringState, event.pointWinnerTeam);
      }
    }
  }
  
  async function recordPoint(serveType, won) {
    // Don't allow scoring after match is complete
    if (scoringState?.completed) {
      return;
    }
    
    const pointWinnerTeam = won ? serverTeam : (serverTeam === 'A' ? 'B' : 'A');
    
    const event = {
      id: uuidv4(),
      matchId: $matchState.id,
      timestamp: new Date().toISOString(),
      serverPlayerId: $matchState.currentServer,
      serveType,
      pointWinnerTeam,
      synced: false,
    };
    
    // Save to IndexedDB immediately (offline-first)
    await saveEvent(event);
    
    // Update scoring state with tennis logic
    scoringState = scorePoint(scoringState, pointWinnerTeam);
    
    // Update local state
    matchState.update(m => ({
      ...m,
      events: [...m.events, event],
    }));
    
    // Save match state with current score for resume functionality
    saveCurrentMatch({
      ...$matchState,
      score: scoringState ? {
        pointsA: scoringState.score?.pointsA || 0,
        pointsB: scoringState.score?.pointsB || 0,
        gamesA: scoringState.score?.gamesA || 0,
        gamesB: scoringState.score?.gamesB || 0,
        setsA: scoringState.score?.setsA || 0,
        setsB: scoringState.score?.setsB || 0,
      } : null,
    });
    
    // Check if match is now complete
    if (scoringState?.completed) {
      // Show winner modal
      showWinnerModal = true;
    }
    
    // Attempt sync in background (non-blocking)
    attemptSync();
  }
  
  async function attemptSync() {
    // Skip sync for tournament matches (they use local storage only)
    if ($matchState.isTournamentMatch) return;
    if (syncing || !navigator.onLine) return;
    
    syncing = true;
    try {
      await syncEvents($matchState.id);
    } catch (err) {
      console.log('Sync failed, will retry:', err);
    } finally {
      syncing = false;
    }
  }
  
  function toggleServer() {
    const currentIdx = $matchState.teamA.concat($matchState.teamB).indexOf($matchState.currentServer);
    const allPlayers = $matchState.teamA.concat($matchState.teamB);
    const nextIdx = (currentIdx + 1) % allPlayers.length;
    const nextServer = allPlayers[nextIdx];
    const nextTeam = $matchState.teamA.includes(nextServer) ? 'A' : 'B';
    
    matchState.update(m => ({
      ...m,
      currentServer: nextServer,
      serverTeam: nextTeam,
    }));
    
    saveCurrentMatch({
      ...$matchState,
      score: scoringState ? {
        pointsA: scoringState.score?.pointsA || 0,
        pointsB: scoringState.score?.pointsB || 0,
        gamesA: scoringState.score?.gamesA || 0,
        gamesB: scoringState.score?.gamesB || 0,
        setsA: scoringState.score?.setsA || 0,
        setsB: scoringState.score?.setsB || 0,
      } : null,
    });
  }
  
  function activateDeuceTiebreaker() {
    // Switch to best-of-3 points tiebreaker
    scoringState = startDeuceTiebreaker(scoringState);
    
    // Save the updated state
    saveCurrentMatch({
      ...$matchState,
      deuceTiebreaker: scoringState.deuceTiebreaker,
    });
  }
  
  function requestEndMatch() {
    if (scoringState?.completed) {
      // Match is already complete, proceed directly
      doEndMatch();
    } else {
      // Show confirmation modal
      showEndMatchModal = true;
    }
  }
  
  async function doEndMatch() {
    showEndMatchModal = false;
    
    // Check if this is a tournament match
    const isTournamentMatch = $matchState.isTournamentMatch;
    
    // Track if sync was successful
    let syncSuccessful = false;
    
    // Only sync to backend for non-tournament matches
    if (!isTournamentMatch) {
      syncing = true;
      try {
        await syncEvents($matchState.id);
        await apiCompleteMatch($matchState.id);
        syncSuccessful = true;
      } catch (err) {
        syncing = false;
        showSyncFailedModal = true;
        return;
      }
      syncing = false;
    }
    
    showWinnerModal = false;
    
    if (isTournamentMatch) {
      // Update tournament data
      const tournamentDataStr = localStorage.getItem('tournamentData');
      if (tournamentDataStr) {
        const tournamentData = JSON.parse(tournamentDataStr);
        
        // Find and update the match
        const matchIndex = tournamentData.matches.findIndex(m => m.id === $matchState.tournamentMatchId);
        if (matchIndex !== -1) {
          // Determine winner based on scoring state
          const teamAWon = scoringState.winner === 'A';
          const winnerTeamId = teamAWon 
            ? tournamentData.matches[matchIndex].teamA.id 
            : tournamentData.matches[matchIndex].teamB.id;
          
          tournamentData.matches[matchIndex].status = 'completed';
          tournamentData.matches[matchIndex].winner = winnerTeamId;
          tournamentData.matches[matchIndex].score = display ? 
            `${display.games.a}-${display.games.b}` : null;
          
          localStorage.setItem('tournamentData', JSON.stringify(tournamentData));
        }
        
        // Clear the current tournament match
        localStorage.removeItem('currentTournamentMatch');
      }
      
      // Clear from IndexedDB so Resume Match doesn't show
      await clearCurrentMatch();
      if ($matchState.id) {
        await deleteIncompleteMatch($matchState.id);
      }
      
      navigate('tournament-dashboard');
    } else {
      // Mark match as completed in IndexedDB so Resume Match doesn't show
      matchState.update(m => ({ ...m, completed: true }));
      await saveCurrentMatch($matchState);
      
      // Only show summary if sync was successful, otherwise go home
      if (syncSuccessful) {
        // Remove from incomplete matches list
        if ($matchState.id) {
          await deleteIncompleteMatch($matchState.id);
        }
        navigate('match-summary');
      } else {
        // This shouldn't happen in normal flow since sync failure shows modal
        await clearCurrentMatch();
        if ($matchState.id) {
          await deleteIncompleteMatch($matchState.id);
        }
        navigate('home');
      }
    }
  }
  
  async function endMatchWithoutSync() {
    showSyncFailedModal = false;
    
    // Clear match data since we can't show a proper summary
    await clearCurrentMatch();
    if ($matchState.id) {
      await deleteIncompleteMatch($matchState.id);
    }
    
    showSyncFailedAlertModal = true;
  }
  
  function handleSyncFailedAlertClose() {
    showSyncFailedAlertModal = false;
    navigate('home');
  }
</script>

<div class="screen" style="padding-top: var(--space-md);">
  <!-- Venue Header -->
  <div style="text-align: center; margin-bottom: var(--space-md);">
    <p class="text-secondary" style="font-size: 13px;">{$matchState.venueName}</p>
    {#if syncing}
      <p class="text-accent" style="font-size: 12px; margin-top: 4px;">⟳ Syncing...</p>
    {/if}
  </div>
  
  <!-- Score Display -->
  {#if display}
    <!-- Team Headers -->
    <div class="score-table">
      <div class="score-header">
        <div class="score-col"></div>
        <div class="score-col team-name">Team A</div>
        <div class="score-col team-name">Team B</div>
      </div>
      
      <!-- Sets Row (Standard Mode Only) -->
      {#if display.sets}
        <div class="score-row">
          <div class="score-col row-label">Sets</div>
          <div class="score-col">{display.sets.a}</div>
          <div class="score-col">{display.sets.b}</div>
        </div>
      {/if}
      
      <!-- Games Row -->
      <div class="score-row">
        <div class="score-col row-label">Games</div>
        <div class="score-col">{display.games.a}</div>
        <div class="score-col">{display.games.b}</div>
      </div>
      
      <!-- Points Row -->
      <div class="score-row points-row">
        <div class="score-col row-label">Points</div>
        <div class="score-col points-value">{display.points.a}</div>
        <div class="score-col points-value">{display.points.b}</div>
      </div>
    </div>
  {:else}
    <div class="score-table">
      <div class="score-header">
        <div class="score-col"></div>
        <div class="score-col team-name">Team A</div>
        <div class="score-col team-name">Team B</div>
      </div>
      <div class="score-row">
        <div class="score-col row-label">Games</div>
        <div class="score-col">0</div>
        <div class="score-col">0</div>
      </div>
      <div class="score-row points-row">
        <div class="score-col row-label">Points</div>
        <div class="score-col points-value">0</div>
        <div class="score-col points-value">0</div>
      </div>
    </div>
  {/if}
  
  <!-- Server Indicator -->
  <div class="server-indicator" on:click={toggleServer} style="cursor: pointer;">
    <div class="server-dot"></div>
    <span class="server-name">
      {serverPlayer?.name || 'Server'} (Team {serverTeam})
    </span>
    <span class="text-secondary" style="font-size: 12px; margin-left: auto;">tap to change</span>
  </div>
  
  <!-- Action Buttons -->
  <div style="flex: 1; display: flex; flex-direction: column; gap: var(--space-sm); margin-top: var(--space-lg);">
    <button 
      class="btn btn-action btn-primary" 
      on:click={() => recordPoint('first', true)}
    >
      ✓ First Serve Won
    </button>
    
    <button 
      class="btn btn-action btn-secondary" 
      on:click={() => recordPoint('first', false)}
    >
      ✗ First Serve Lost
    </button>
    
    <button 
      class="btn btn-action btn-primary" 
      on:click={() => recordPoint('second', true)}
    >
      ✓ Second Serve Won
    </button>
    
    <button 
      class="btn btn-action btn-secondary" 
      on:click={() => recordPoint('second', false)}
    >
      ✗ Second Serve Lost
    </button>
    
    <button 
      class="btn btn-action btn-danger" 
      on:click={() => recordPoint('double_fault', false)}
    >
      Double Fault
    </button>
    
    <!-- Deuce Tiebreaker Button - shown when in deuce/advantage -->
    {#if canStartDeuceTiebreaker}
      <button 
        class="btn btn-action btn-tiebreaker" 
        on:click={activateDeuceTiebreaker}
      >
        🎯 Switch to Best of 3 Tiebreaker
      </button>
    {/if}
    
    <!-- Tiebreaker mode indicator -->
    {#if inDeuceTiebreaker}
      <div class="tiebreaker-banner">
        ⚡ Best of 3 Tiebreaker - First to 2 points wins!
      </div>
    {/if}
  </div>
  
  <!-- End Match Button -->
  <button 
    class="btn btn-ghost" 
    on:click={requestEndMatch}
    style="margin-top: var(--space-md);"
    disabled={syncing}
  >
    End Match
  </button>
</div>

<!-- Winner Modal -->
{#if showWinnerModal}
  <div class="modal-overlay">
    <div class="modal">
      <div class="modal-icon">🏆</div>
      <h2 class="modal-title">{scoringState?.winner === 'A' ? 'Team A' : 'Team B'} Wins!</h2>
      <p class="modal-subtitle">Congratulations!</p>
      <button class="btn btn-primary modal-btn" on:click={requestEndMatch}>
        View Summary
      </button>
    </div>
  </div>
{/if}

<!-- End Match Confirmation Modal -->
<Modal 
  bind:show={showEndMatchModal}
  title="End Match?"
  message="The match is not complete yet. Are you sure you want to end it?"
  icon="⚠️"
  type="confirm"
  confirmText="End Match"
  cancelText="Continue Playing"
  on:confirm={doEndMatch}
/>

<!-- Sync Failed Modal -->
<Modal 
  bind:show={showSyncFailedModal}
  title="Sync Failed"
  message="Failed to sync match data to server. End match anyway?"
  icon="⚠️"
  type="danger"
  confirmText="End Anyway"
  cancelText="Keep Playing"
  on:confirm={endMatchWithoutSync}
/>

<!-- Sync Failed Alert Modal -->
<Modal 
  bind:show={showSyncFailedAlertModal}
  title="Match Ended"
  message="Summary not available because sync failed."
  icon="ℹ️"
  type="alert"
  confirmText="OK"
  on:confirm={handleSyncFailedAlertClose}
/>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  
  .modal {
    background: var(--surface);
    border-radius: 16px;
    padding: var(--space-xl);
    text-align: center;
    max-width: 320px;
    width: 90%;
    animation: modalIn 0.3s ease-out;
  }
  
  @keyframes modalIn {
    from {
      opacity: 0;
      transform: scale(0.9);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
  
  .modal-icon {
    font-size: 64px;
    margin-bottom: var(--space-md);
  }
  
  .modal-title {
    font-size: 24px;
    font-weight: 700;
    color: var(--primary);
    margin-bottom: var(--space-sm);
  }
  
  .modal-subtitle {
    color: var(--text-secondary);
    margin-bottom: var(--space-lg);
  }
  
  .modal-btn {
    width: 100%;
  }
  
  .score-table {
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border-radius: 12px;
    padding: var(--space-md);
    margin-bottom: var(--space-md);
  }
  
  .score-header, .score-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  
  .score-header {
    margin-bottom: var(--space-sm);
    padding-bottom: var(--space-sm);
    border-bottom: 1px solid var(--border);
  }
  
  .score-col {
    flex: 1;
    text-align: center;
    font-size: 20px;
    font-weight: 600;
  }
  
  .score-col.team-name {
    color: var(--text-primary);
    font-size: 16px;
  }
  
  .score-col.row-label {
    text-align: left;
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 400;
  }
  
  .score-row {
    padding: var(--space-xs) 0;
  }
  
  .points-row {
    margin-top: var(--space-sm);
    padding-top: var(--space-sm);
    border-top: 1px solid var(--border);
  }
  
  .points-value {
    font-size: 32px;
    color: var(--primary);
  }
  
  .match-complete-banner {
    background: var(--primary);
    color: var(--background);
    text-align: center;
    padding: var(--space-md);
    border-radius: 8px;
    font-size: 20px;
    font-weight: 600;
    margin-bottom: var(--space-md);
  }
  
  .btn-tiebreaker {
    background: linear-gradient(135deg, #f59e0b, #d97706);
    color: white;
    border: none;
    font-weight: 600;
  }
  
  .btn-tiebreaker:hover {
    background: linear-gradient(135deg, #d97706, #b45309);
  }
  
  .tiebreaker-banner {
    background: linear-gradient(135deg, #f59e0b, #d97706);
    color: white;
    text-align: center;
    padding: var(--space-sm) var(--space-md);
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    margin-top: var(--space-sm);
  }
</style>
