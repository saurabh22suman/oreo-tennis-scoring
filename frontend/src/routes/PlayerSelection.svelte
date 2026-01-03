<!-- Player Selection - Screen 3 from ui_design_spec.md -->
<script>
  import { navigate, matchState, players } from '../stores/app.js';
  import { createMatch } from '../services/api.js';
  import { saveCurrentMatch, createTempPlayer, getTempPlayersForVenue } from '../services/db.js';
  import Modal from '../lib/Modal.svelte';
  import { onMount } from 'svelte';
  
  let teamAPlayer1 = '';
  let teamAPlayer2 = '';
  let teamBPlayer1 = '';
  let teamBPlayer2 = '';
  let loading = false;
  
  // Temp player creation
  let tempPlayers = [];
  let showTempPlayerModal = false;
  let newTempPlayerName = '';
  let creatingTempPlayer = false;
  
  // Modal states
  let showAlertModal = false;
  let alertMessage = '';
  
  onMount(async () => {
    // Load existing temp players for this venue
    if ($matchState.venueId) {
      tempPlayers = await getTempPlayersForVenue($matchState.venueId);
    }
  });
  
  $: isSingles = $matchState.matchType === 'singles';
  $: isAustralianDoubles = $matchState.matchType === '1v2';
  $: isDoubles = $matchState.matchType === 'doubles';
  
  // Combine regular players with temp players
  $: allPlayers = [...$players.filter(p => p.active !== false), ...tempPlayers];
  
  // For 1v2: Team A has 1 player, Team B has 2 players
  // For singles: 1 player each
  // For doubles: 2 players each
  $: teamAPlayerCount = isSingles || isAustralianDoubles ? 1 : 2;
  $: teamBPlayerCount = isSingles ? 1 : 2;
  
  // Reactive available players for each dropdown
  // Always include the currently selected value to prevent selection loss during re-renders
  $: availableForA1 = allPlayers.filter(p => 
    p.id === teamAPlayer1 || // Always keep current selection
    (p.id !== teamAPlayer2 && p.id !== teamBPlayer1 && p.id !== teamBPlayer2)
  );
  $: availableForA2 = allPlayers.filter(p => 
    p.id === teamAPlayer2 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamBPlayer1 && p.id !== teamBPlayer2)
  );
  $: availableForB1 = allPlayers.filter(p => 
    p.id === teamBPlayer1 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamAPlayer2 && p.id !== teamBPlayer2)
  );
  $: availableForB2 = allPlayers.filter(p => 
    p.id === teamBPlayer2 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamAPlayer2 && p.id !== teamBPlayer1)
  );
  
  async function addTempPlayer() {
    if (!newTempPlayerName.trim()) {
      alertMessage = 'Please enter a player name';
      showAlertModal = true;
      return;
    }
    
    creatingTempPlayer = true;
    try {
      const tempPlayer = await createTempPlayer(newTempPlayerName.trim(), $matchState.venueId);
      tempPlayers = [...tempPlayers, tempPlayer];
      newTempPlayerName = '';
      showTempPlayerModal = false;
    } catch (err) {
      alertMessage = 'Failed to create temporary player';
      showAlertModal = true;
    }
    creatingTempPlayer = false;
  }
  
  function goBack() {
    navigate('match-setup');
  }
  
  async function startMatch() {
    // Validate selections based on match type
    const teamA = teamAPlayerCount === 1 ? [teamAPlayer1] : [teamAPlayer1, teamAPlayer2];
    const teamB = teamBPlayerCount === 1 ? [teamBPlayer1] : [teamBPlayer1, teamBPlayer2];
    
    const teamAFiltered = teamA.filter(p => p);
    const teamBFiltered = teamB.filter(p => p);
    
    if (teamAFiltered.length !== teamAPlayerCount || teamBFiltered.length !== teamBPlayerCount) {
      if (isAustralianDoubles) {
        alertMessage = 'Please select 1 player for Team A and 2 players for Team B';
      } else {
        alertMessage = `Please select ${teamAPlayerCount} player(s) for each team`;
      }
      showAlertModal = true;
      return;
    }
    
    // Check for duplicates
    const allSelected = [...teamAFiltered, ...teamBFiltered];
    if (new Set(allSelected).size !== allSelected.length) {
      alertMessage = 'Cannot select the same player multiple times';
      showAlertModal = true;
      return;
    }
    
    loading = true;
    
    // Check if any selected player is a temp player
    const hasTempPlayers = allSelected.some(id => id.startsWith('temp-'));
    
    // Build player info map for display (needed for temp players)
    const playerInfoMap = {};
    allPlayers.forEach(p => {
      playerInfoMap[p.id] = { id: p.id, name: p.name, isTemp: p.isTemp || false };
    });
    
    try {
      let matchId;
      
      if (hasTempPlayers) {
        // For matches with temp players, create an offline-only match
        matchId = `local-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      } else {
        // Create match on backend for regular players
        const match = await createMatch(
          $matchState.venueId,
          $matchState.matchType,
          teamAFiltered,
          teamBFiltered
        );
        matchId = match.id;
      }
      
      // Update match state with player info for display
      matchState.update(m => ({
        ...m,
        id: matchId,
        teamA: teamAFiltered,
        teamB: teamBFiltered,
        playerInfo: playerInfoMap, // Store player info for name lookups
        hasTempPlayers,
        startedAt: new Date().toISOString(),
        currentServer: teamAFiltered[0],
        serverTeam: 'A',
      }));
      
      // Save to IndexedDB
      await saveCurrentMatch({
        id: matchId,
        venueId: $matchState.venueId,
        venueName: $matchState.venueName,
        matchType: $matchState.matchType,
        matchMode: $matchState.matchMode,
        bestOf: $matchState.bestOf,
        teamA: teamAFiltered,
        teamB: teamBFiltered,
        playerInfo: playerInfoMap,
        hasTempPlayers,
        startedAt: new Date().toISOString(),
        currentServer: teamAFiltered[0],
        serverTeam: 'A',
        completed: false,
        events: [],
      });
      
      navigate('live-match');
    } catch (err) {
      alertMessage = 'Failed to create match: ' + err.message;
      showAlertModal = true;
      loading = false;
    }
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Select Players</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; display: flex; flex-direction: column;">
    <!-- Add Temp Player Button -->
    <div class="temp-player-section">
      <button class="btn-add-temp" on:click={() => showTempPlayerModal = true}>
        + Add Guest Player
      </button>
      {#if tempPlayers.length > 0}
        <span class="temp-count">{tempPlayers.length} guest player{tempPlayers.length > 1 ? 's' : ''} (24h valid)</span>
      {/if}
    </div>
    
    <div style="flex: 1; display: flex; flex-direction: column; gap: var(--space-lg);">
      <!-- Team A -->
      <div>
        <h2 class="mb-md">Team A {#if isAustralianDoubles}<span class="text-secondary" style="font-size: 14px;">(1 player)</span>{/if}</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-sm);">
          <select class="form-select" value={teamAPlayer1} on:change={(e) => teamAPlayer1 = e.target.value}>
            <option value="">Select player...</option>
            {#each availableForA1 as player (player.id)}
              <option value={player.id} selected={player.id === teamAPlayer1}>{player.name}{player.isTemp ? ' 👤' : ''}</option>
            {/each}
          </select>
          {#if isDoubles}
            <select class="form-select" value={teamAPlayer2} on:change={(e) => teamAPlayer2 = e.target.value}>
              <option value="">Select player...</option>
              {#each availableForA2 as player (player.id)}
                <option value={player.id} selected={player.id === teamAPlayer2}>{player.name}{player.isTemp ? ' 👤' : ''}</option>
              {/each}
            </select>
          {/if}
        </div>
      </div>
      
      <!-- Team B -->
      <div>
        <h2 class="mb-md">Team B {#if isAustralianDoubles}<span class="text-secondary" style="font-size: 14px;">(2 players)</span>{/if}</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-sm);">
          <select class="form-select" value={teamBPlayer1} on:change={(e) => teamBPlayer1 = e.target.value}>
            <option value="">Select player...</option>
            {#each availableForB1 as player (player.id)}
              <option value={player.id} selected={player.id === teamBPlayer1}>{player.name}{player.isTemp ? ' 👤' : ''}</option>
            {/each}
          </select>
          {#if !isSingles}
            <select class="form-select" value={teamBPlayer2} on:change={(e) => teamBPlayer2 = e.target.value}>
              <option value="">Select player...</option>
              {#each availableForB2 as player (player.id)}
                <option value={player.id} selected={player.id === teamBPlayer2}>{player.name}{player.isTemp ? ' 👤' : ''}</option>
              {/each}
            </select>
          {/if}
        </div>
      </div>
    </div>
    
    <div style="margin-top: var(--space-xl); padding-top: var(--space-lg);">
      <button class="btn btn-primary" on:click={startMatch} disabled={loading}>
        {#if loading}
          Starting Match...
        {:else}
          Start Match
        {/if}
      </button>
    </div>
  </div>
</div>

<!-- Alert Modal -->
<Modal 
  bind:show={showAlertModal}
  title="Oops!"
  message={alertMessage}
  icon="⚠️"
  type="alert"
  confirmText="OK"
/>

<!-- Add Temp Player Modal -->
{#if showTempPlayerModal}
  <div class="modal-overlay" on:click={() => showTempPlayerModal = false}>
    <div class="modal-content" on:click|stopPropagation>
      <h3>Add Guest Player</h3>
      <p class="modal-desc">Guest players are only valid for 24 hours at this venue.</p>
      <input 
        type="text" 
        class="form-input"
        placeholder="Enter player name..."
        bind:value={newTempPlayerName}
        on:keydown={(e) => e.key === 'Enter' && addTempPlayer()}
      />
      <div class="modal-actions">
        <button class="btn btn-secondary" on:click={() => showTempPlayerModal = false}>
          Cancel
        </button>
        <button class="btn btn-primary" on:click={addTempPlayer} disabled={creatingTempPlayer}>
          {creatingTempPlayer ? 'Adding...' : 'Add Player'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .temp-player-section {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    margin-bottom: var(--space-lg);
    padding: var(--space-sm) 0;
  }
  
  .btn-add-temp {
    background: var(--surface);
    border: 1px dashed var(--text-secondary);
    color: var(--text-secondary);
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-btn);
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-add-temp:hover {
    border-color: var(--accent);
    color: var(--accent);
  }
  
  .temp-count {
    font-size: 12px;
    color: var(--text-secondary);
  }
  
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: var(--space-md);
  }
  
  .modal-content {
    background: var(--bg-secondary);
    border-radius: var(--radius-card);
    padding: var(--space-lg);
    width: 100%;
    max-width: 350px;
  }
  
  .modal-content h3 {
    color: var(--text-primary);
    margin-bottom: var(--space-sm);
    font-size: 18px;
  }
  
  .modal-desc {
    color: var(--text-secondary);
    font-size: 13px;
    margin-bottom: var(--space-md);
  }
  
  .form-input {
    width: 100%;
    padding: var(--space-md);
    background: var(--surface);
    border: 1px solid var(--surface);
    border-radius: var(--radius-btn);
    color: var(--text-primary);
    font-size: 16px;
    margin-bottom: var(--space-md);
  }
  
  .form-input:focus {
    outline: none;
    border-color: var(--accent);
  }
  
  .modal-actions {
    display: flex;
    gap: var(--space-sm);
  }
  
  .modal-actions .btn {
    flex: 1;
    padding: var(--space-md);
  }
  
  .btn-secondary {
    background: var(--surface);
    color: var(--text-primary);
    border: none;
    border-radius: var(--radius-btn);
    cursor: pointer;
  }
</style>
