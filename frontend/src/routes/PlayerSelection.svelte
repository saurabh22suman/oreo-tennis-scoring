<!-- Player Selection - Screen 3 from ui_design_spec.md -->
<script>
  import { navigate, matchState, players } from '../stores/app.js';
  import { createMatch } from '../services/api.js';
  import { saveCurrentMatch } from '../services/db.js';
  import Modal from '../lib/Modal.svelte';
  
  let teamAPlayer1 = '';
  let teamAPlayer2 = '';
  let teamBPlayer1 = '';
  let teamBPlayer2 = '';
  let loading = false;
  
  // Modal states
  let showAlertModal = false;
  let alertMessage = '';
  
  $: isSingles = $matchState.matchType === 'singles';
  $: isAustralianDoubles = $matchState.matchType === '1v2';
  $: isDoubles = $matchState.matchType === 'doubles';
  
  // For 1v2: Team A has 1 player, Team B has 2 players
  // For singles: 1 player each
  // For doubles: 2 players each
  $: teamAPlayerCount = isSingles || isAustralianDoubles ? 1 : 2;
  $: teamBPlayerCount = isSingles ? 1 : 2;
  
  // Filter active players
  $: activePlayers = $players.filter(p => p.active !== false);
  
  // Reactive available players for each dropdown
  // Always include the currently selected value to prevent selection loss during re-renders
  $: availableForA1 = activePlayers.filter(p => 
    p.id === teamAPlayer1 || // Always keep current selection
    (p.id !== teamAPlayer2 && p.id !== teamBPlayer1 && p.id !== teamBPlayer2)
  );
  $: availableForA2 = activePlayers.filter(p => 
    p.id === teamAPlayer2 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamBPlayer1 && p.id !== teamBPlayer2)
  );
  $: availableForB1 = activePlayers.filter(p => 
    p.id === teamBPlayer1 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamAPlayer2 && p.id !== teamBPlayer2)
  );
  $: availableForB2 = activePlayers.filter(p => 
    p.id === teamBPlayer2 || // Always keep current selection
    (p.id !== teamAPlayer1 && p.id !== teamAPlayer2 && p.id !== teamBPlayer1)
  );
  
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
    
    try {
      // Create match on backend
      const match = await createMatch(
        $matchState.venueId,
        $matchState.matchType,
        teamAFiltered,
        teamBFiltered
      );
      
      // Update match state
      matchState.update(m => ({
        ...m,
        id: match.id,
        teamA: teamAFiltered,
        teamB: teamBFiltered,
        startedAt: new Date().toISOString(),
        currentServer: teamAFiltered[0],
        serverTeam: 'A',
      }));
      
      // Save to IndexedDB
      await saveCurrentMatch({
        id: match.id,
        venueId: $matchState.venueId,
        venueName: $matchState.venueName,
        matchType: $matchState.matchType,
        matchMode: $matchState.matchMode,
        bestOf: $matchState.bestOf,
        teamA: teamAFiltered,
        teamB: teamBFiltered,
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
    <div style="flex: 1; display: flex; flex-direction: column; gap: var(--space-lg);">
      <!-- Team A -->
      <div>
        <h2 class="mb-md">Team A {#if isAustralianDoubles}<span class="text-secondary" style="font-size: 14px;">(1 player)</span>{/if}</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-sm);">
          <select class="form-select" value={teamAPlayer1} on:change={(e) => teamAPlayer1 = e.target.value}>
            <option value="">Select player...</option>
            {#each availableForA1 as player (player.id)}
              <option value={player.id} selected={player.id === teamAPlayer1}>{player.name}</option>
            {/each}
          </select>
          {#if isDoubles}
            <select class="form-select" value={teamAPlayer2} on:change={(e) => teamAPlayer2 = e.target.value}>
              <option value="">Select player...</option>
              {#each availableForA2 as player (player.id)}
                <option value={player.id} selected={player.id === teamAPlayer2}>{player.name}</option>
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
              <option value={player.id} selected={player.id === teamBPlayer1}>{player.name}</option>
            {/each}
          </select>
          {#if !isSingles}
            <select class="form-select" value={teamBPlayer2} on:change={(e) => teamBPlayer2 = e.target.value}>
              <option value="">Select player...</option>
              {#each availableForB2 as player (player.id)}
                <option value={player.id} selected={player.id === teamBPlayer2}>{player.name}</option>
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
