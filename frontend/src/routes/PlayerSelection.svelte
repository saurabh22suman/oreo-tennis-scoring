<!-- Player Selection - Screen 3 from ui_design_spec.md -->
<script>
  import { navigate, matchState, players } from '../stores/app.js';
  import { createMatch } from '../services/api.js';
  import { saveCurrentMatch } from '../services/db.js';
  
  let teamAPlayers = ['', ''];
  let teamBPlayers = ['', ''];
  let loading = false;
  
  $: isSingles = $matchState.matchType === 'singles';
  $: requiredPlayers = isSingles ? 1 : 2;
  
  // Reactive: Track all selected players
  $: allSelectedPlayers = [
    ...teamAPlayers.filter(p => p),
    ...teamBPlayers.filter(p => p)
  ];
  
  function goBack() {
    navigate('match-setup');
  }
  
  async function startMatch() {
    // Validate selections
    const teamA = teamAPlayers.slice(0, requiredPlayers).filter(p => p);
    const teamB = teamBPlayers.slice(0, requiredPlayers).filter(p => p);
    
    if (teamA.length !== requiredPlayers || teamB.length !== requiredPlayers) {
      alert(`Please select ${requiredPlayers} player(s) for each team`);
      return;
    }
    
    // Check for duplicates
    const allSelected = [...teamA, ...teamB];
    if (new Set(allSelected).size !== allSelected.length) {
      alert('Cannot select the same player multiple times');
      return;
    }
    
    loading = true;
    
    try {
      // Create match on backend
      const match = await createMatch(
        $matchState.venueId,
        $matchState.matchType,
        teamA,
        teamB
      );
      
      // Update match state
      matchState.update(m => ({
        ...m,
        id: match.id,
        teamA,
        teamB,
        startedAt: new Date().toISOString(),
        currentServer: teamA[0],
        serverTeam: 'A',
      }));
      
      // Save to IndexedDB
      await saveCurrentMatch({
        id: match.id,
        venueId: $matchState.venueId,
        venueName: $matchState.venueName,
        matchType: $matchState.matchType,
        teamA,
        teamB,
        startedAt: new Date().toISOString(),
        currentServer: teamA[0],
        serverTeam: 'A',
        completed: false,
      });
      
      navigate('live-match');
    } catch (err) {
      alert('Failed to create match: ' + err.message);
      loading = false;
    }
  }
  
  // Get available players for a specific dropdown
  // Excludes all selected players EXCEPT the current value (so you can see your own selection)
  function getAvailablePlayers(currentValue) {
    const otherSelected = allSelectedPlayers.filter(p => p !== currentValue);
    return $players.filter(p => !otherSelected.includes(p.id));
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
        <h2 class="mb-md">Team A</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-sm);">
          {#each Array(requiredPlayers) as _, i}
            {@const availablePlayers = getAvailablePlayers(teamAPlayers[i])}
            <select class="form-select" bind:value={teamAPlayers[i]}>
              <option value="">Player {i + 1}...</option>
              {#each availablePlayers as player}
                <option value={player.id}>{player.name}</option>
              {/each}
            </select>
          {/each}
        </div>
      </div>
      
      <!-- Team B -->
      <div>
        <h2 class="mb-md">Team B</h2>
        <div style="display: flex; flex-direction: column; gap: var(--space-sm);">
          {#each Array(requiredPlayers) as _, i}
            {@const availablePlayers = getAvailablePlayers(teamBPlayers[i])}
            <select class="form-select" bind:value={teamBPlayers[i]}>
              <option value="">Player {i + 1}...</option>
              {#each availablePlayers as player}
                <option value={player.id}>{player.name}</option>
              {/each}
            </select>
          {/each}
        </div>
      </div>
    </div>
    
    <button class="btn btn-primary" on:click={startMatch} disabled={loading}>
      {#if loading}
        Starting Match...
      {:else}
        Start Match
      {/if}
    </button>
  </div>
</div>
