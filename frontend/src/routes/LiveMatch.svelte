<!-- Live Match Screen - Screen 4 from ui_design_spec.md (MOST IMPORTANT) -->
<script>
  import { onMount } from 'svelte';
  import { navigate, matchState, score, players } from '../stores/app.js';
  import { saveEvent, getCurrentMatch, saveCurrentMatch } from '../services/db.js';
  import { syncEvents, completeMatch as apiCompleteMatch } from '../services/api.js';
  import { v4 as uuidv4 } from 'uuid';
  
  let syncing = false;
  let autoSyncInterval;
  
  $: serverPlayer = $players.find(p => p.id === $matchState.currentServer);
  $: serverTeam = $matchState.serverTeam;
  
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
  
  async function recordPoint(serveType, won) {
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
    
    // Update local state
    matchState.update(m => ({
      ...m,
      events: [...m.events, event],
    }));
    
    // Attempt sync in background (non-blocking)
    attemptSync();
  }
  
  async function attemptSync() {
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
    
    saveCurrentMatch($matchState);
  }
  
  async function endMatch() {
    if (!confirm('End match and view summary?')) return;
    
    // Force sync all events
    syncing = true;
    try {
      await syncEvents($matchState.id);
      await apiCompleteMatch($matchState.id);
    } catch (err) {
      if (!confirm('Failed to sync. End match anyway?')) {
        syncing = false;
        return;
      }
    }
    syncing = false;
    
    navigate('match-summary');
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
  <div class="score-display">
    <div class="score-team">
      <div class="score-label">Team A</div>
      <div class="score-value">{$score.teamA}</div>
    </div>
    <div class="score-divider">:</div>
    <div class="score-team">
      <div class="score-label">Team B</div>
      <div class="score-value">{$score.teamB}</div>
    </div>
  </div>
  
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
  </div>
  
  <!-- End Match Button -->
  <button 
    class="btn btn-ghost" 
    on:click={endMatch}
    style="margin-top: var(--space-md);"
    disabled={syncing}
  >
    End Match
  </button>
</div>
