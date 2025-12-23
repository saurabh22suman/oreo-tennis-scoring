<!-- Tournament Setup Screen - Per OTS_Tournament_Spec.md -->
<script>
  import { navigate } from '../stores/app.js';
  import { onMount } from 'svelte';
  
  let venues = [];
  let selectedVenueId = '';
  let selectedPlayers = [];
  let availablePlayers = [];
  let loading = true;
  let error = '';
  
  onMount(async () => {
    try {
      // Fetch venues
      const venuesRes = await fetch('/api/venues');
      if (venuesRes.ok) {
        const venuesData = await venuesRes.json();
        venues = venuesData.data;
        if (venues.length > 0) {
          selectedVenueId = venues[0].id;
        }
      }
      
      // Fetch active players
      const playersRes = await fetch('/api/players');
      if (playersRes.ok) {
        const playersData = await playersRes.json();
        availablePlayers = playersData.data;
      }
      
      loading = false;
    } catch (err) {
      error = 'Failed to load data';
      loading = false;
    }
  });
  
  function togglePlayer(playerId) {
    if (selectedPlayers.includes(playerId)) {
      selectedPlayers = selectedPlayers.filter(id => id !== playerId);
    } else {
      selectedPlayers = [...selectedPlayers, playerId];
    }
  }
  
  function canProceed() {
    // Need at least 4 players (minimum for tournament)
    // Must be even number for doubles
    return selectedPlayers.length >= 4 && 
           selectedPlayers.length % 2 === 0 &&
           selectedVenueId;
  }
  
  function proceed() {
    if (!canProceed()) return;
    
    // Store tournament setup data
    const tournamentData = {
      venueId: selectedVenueId,
      playerIds: selectedPlayers
    };
    
    localStorage.setItem('tournamentSetup', JSON.stringify(tournamentData));
    navigate('tournament-team-creation');
  }
  
  function goBack() {
    navigate('home');
  }
</script>

<div class="screen">
  <div class="container">
    <div class="screen-header">
      <button class="btn-back" on:click={goBack}>←</button>
      <h2>Tournament Setup</h2>
      <div style="width: 40px;"></div>
    </div>
    
    {#if loading}
      <p class="text-center text-secondary">Loading...</p>
    {:else if error}
      <p class="text-center text-danger">{error}</p>
    {:else}
      <!-- Venue Selection -->
      <div class="form-section">
        <label class="form-label">Select Venue</label>
        <select class="form-select" bind:value={selectedVenueId}>
          {#each venues as venue}
            <option value={venue.id}>{venue.name} ({venue.surface})</option>
          {/each}
        </select>
      </div>
      
      <!-- Player Selection -->
      <div class="form-section">
        <label class="form-label">
          Select Players 
          <span class="text-secondary">
            ({selectedPlayers.length} selected - need 4+ even number)
          </span>
        </label>
        
        {#if availablePlayers.length === 0}
          <p class="text-secondary">No players available. Please add players in Admin panel.</p>
        {:else}
          <div class="player-grid">
            {#each availablePlayers as player}
              <button
                class="player-card"
                class:selected={selectedPlayers.includes(player.id)}
                on:click={() => togglePlayer(player.id)}
              >
                <div class="player-name">{player.name}</div>
                {#if selectedPlayers.includes(player.id)}
                  <div class="check-icon">✓</div>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>
      
      <!-- Info Box -->
      <div class="info-box">
        <p class="text-small">
          <strong>What's next?</strong><br/>
          After selecting players, you'll create doubles teams (randomly or manually),
          then proceed to round-robin matches followed by knockout stage.
        </p>
      </div>
      
      <!-- Action Buttons -->
      <div class="button-group">
        <button 
          class="btn btn-primary btn-full"
          disabled={!canProceed()}
          on:click={proceed}
        >
          Continue to Team Creation
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .form-section {
    margin-bottom: var(--space-xl);
  }
  
  .form-label {
    display: block;
    margin-bottom: var(--space-sm);
    font-weight: 500;
    color: var(--text-primary);
  }
  
  .form-select {
    width: 100%;
    padding: var(--space-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-size: 16px;
  }
  
  .player-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: var(--space-sm);
  }
  
  .player-card {
    position: relative;
    padding: var(--space-md);
    background: var(--bg-elevated);
    border: 2px solid var(--border-color);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s;
    text-align: center;
  }
  
  .player-card:hover {
    border-color: var(--primary);
    background: var(--bg-hover);
  }
  
  .player-card.selected {
    border-color: var(--primary);
    background: var(--primary);
    color: white;
  }
  
  .player-name {
    font-weight: 500;
  }
  
  .check-icon {
    position: absolute;
    top: 4px;
    right: 8px;
    font-size: 18px;
  }
  
  .info-box {
    padding: var(--space-md);
    background: var(--bg-elevated);
    border-left: 3px solid var(--primary);
    border-radius: var(--radius-sm);
    margin-bottom: var(--space-xl);
  }
  
  .text-small {
    font-size: 14px;
    line-height: 1.5;
    margin: 0;
    color: var(--text-secondary);
  }
  
  .button-group {
    display: flex;
    gap: var(--space-md);
    margin-top: var(--space-xl);
  }
  
  .btn-full {
    flex: 1;
  }
</style>
