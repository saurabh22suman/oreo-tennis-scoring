<!-- Tournament Setup Screen - Per OTS_Tournament_Spec.md -->
<script>
  import { navigate, venues as venueStore, players as playerStore } from '../stores/app.js';
  import { createTempPlayer, getTempPlayersForVenue } from '../services/db.js';
  import { onMount } from 'svelte';
  
  let selectedVenueId = '';
  let selectedPlayers = [];
  let loading = true;
  let error = '';
  
  // Temp player support
  let tempPlayers = [];
  let showTempPlayerModal = false;
  let newTempPlayerName = '';
  let creatingTempPlayer = false;
  
  // Combine regular and temp players
  $: allPlayers = [...$playerStore, ...tempPlayers];
  
  // Computed values
  $: numberOfTeams = Math.floor(selectedPlayers.length / 2);
  $: totalMatches = numberOfTeams > 1 ? (numberOfTeams * (numberOfTeams - 1)) / 2 : 0;
  $: isValidSelection = selectedPlayers.length >= 6 && selectedPlayers.length % 2 === 0;
  $: validationMessage = getValidationMessage(selectedPlayers.length);
  
  function getValidationMessage(count) {
    if (count < 6) return `Select at least ${6 - count} more player${6 - count !== 1 ? 's' : ''} (min 6)`;
    if (count % 2 !== 0) return 'Select one more player (need even number)';
    return '';
  }
  
  async function loadTempPlayers() {
    if (selectedVenueId) {
      tempPlayers = await getTempPlayersForVenue(selectedVenueId);
      // Deselect any temp players that no longer exist
      selectedPlayers = selectedPlayers.filter(id => 
        allPlayers.some(p => p.id === id)
      );
    }
  }
  
  // Reload temp players when venue changes
  $: if (selectedVenueId) loadTempPlayers();
  
  onMount(async () => {
    try {
      // Use stores if already populated, otherwise fetch
      if ($venueStore.length === 0) {
        const venuesRes = await fetch('/api/venues');
        if (venuesRes.ok) {
          const venuesData = await venuesRes.json();
          venueStore.set(venuesData.data || []);
        }
      }
      
      if ($playerStore.length === 0) {
        const playersRes = await fetch('/api/players');
        if (playersRes.ok) {
          const playersData = await playersRes.json();
          playerStore.set(playersData.data || []);
        }
      }
      
      // Auto-select first venue if available
      if ($venueStore.length > 0 && !selectedVenueId) {
        selectedVenueId = $venueStore[0].id;
      }
      
      loading = false;
    } catch (err) {
      error = 'Failed to load data';
      loading = false;
    }
  });
  
  async function addTempPlayer() {
    if (!newTempPlayerName.trim()) return;
    
    creatingTempPlayer = true;
    try {
      const tempPlayer = await createTempPlayer(newTempPlayerName.trim(), selectedVenueId);
      tempPlayers = [...tempPlayers, tempPlayer];
      newTempPlayerName = '';
      showTempPlayerModal = false;
    } catch (err) {
      console.error('Failed to create temp player:', err);
    }
    creatingTempPlayer = false;
  }
  
  function togglePlayer(playerId) {
    if (selectedPlayers.includes(playerId)) {
      selectedPlayers = selectedPlayers.filter(id => id !== playerId);
    } else {
      selectedPlayers = [...selectedPlayers, playerId];
    }
  }
  
  function selectAll() {
    // Select all players (if odd, exclude last one)
    const allIds = allPlayers.map(p => p.id);
    selectedPlayers = allIds.length % 2 === 0 ? allIds : allIds.slice(0, -1);
  }
  
  function clearSelection() {
    selectedPlayers = [];
  }
  
  // Reactive validation for proceed button
  $: canProceed = isValidSelection && selectedVenueId !== '';
  
  function proceed() {
    if (!canProceed) return;
    
    const selectedVenue = $venueStore.find(v => v.id === selectedVenueId);
    
    // Store tournament setup data
    const tournamentData = {
      venueId: selectedVenueId,
      venueName: selectedVenue?.name || '',
      venueSurface: selectedVenue?.surface || '',
      playerIds: selectedPlayers,
      players: allPlayers.filter(p => selectedPlayers.includes(p.id))
    };
    
    localStorage.setItem('tournamentSetup', JSON.stringify(tournamentData));
    navigate('tournament-team-creation');
  }
  
  function goBack() {
    navigate('home');
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">New Tournament</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="content">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else if error}
      <div class="error-state">
        <p>{error}</p>
        <button class="btn btn-secondary" on:click={() => location.reload()}>
          Retry
        </button>
      </div>
    {:else}
      <!-- Venue Selection Card -->
      <div class="section-card">
        <div class="section-header">
          <span class="section-icon">📍</span>
          <h2>Select Venue</h2>
        </div>
        
        {#if $venueStore.length === 0}
          <p class="empty-message">No venues available. Add venues in Admin panel.</p>
        {:else}
          <select class="form-select" bind:value={selectedVenueId}>
            {#each $venueStore as venue}
              <option value={venue.id}>{venue.name} ({venue.surface})</option>
            {/each}
          </select>
        {/if}
      </div>
      
      <!-- Player Selection Card -->
      <div class="section-card">
        <div class="section-header">
          <span class="section-icon">👥</span>
          <h2>Select Players</h2>
        </div>
        
        <!-- Selection Status Bar -->
        <div class="status-bar">
          <div class="status-count">
            <span class="count-number">{selectedPlayers.length}</span>
            <span class="count-label">selected</span>
          </div>
          
          {#if isValidSelection}
            <div class="status-valid">
              <span class="valid-icon">✓</span>
              <span>{numberOfTeams} teams</span>
            </div>
          {:else}
            <div class="status-warning">
              {validationMessage}
            </div>
          {/if}
          
          <!-- Quick Actions -->
          <div class="quick-actions">
            {#if selectedPlayers.length > 0}
              <button class="action-btn" on:click={clearSelection}>Clear</button>
            {/if}
            {#if selectedPlayers.length < $playerStore.length}
              <button class="action-btn" on:click={selectAll}>All</button>
            {/if}
          </div>
        </div>
        
        {#if $playerStore.length === 0 && tempPlayers.length === 0}
          <p class="empty-message">No players available. Add players in Admin panel.</p>
        {:else}
          <!-- Add Guest Player Button -->
          <div class="add-guest-section">
            <button class="btn-add-guest" on:click={() => showTempPlayerModal = true}>
              + Add Guest Player
            </button>
            {#if tempPlayers.length > 0}
              <span class="guest-count">{tempPlayers.length} guest{tempPlayers.length > 1 ? 's' : ''} (24h valid)</span>
            {/if}
          </div>
          
          <div class="player-grid">
            {#each allPlayers as player}
              <button
                class="player-chip"
                class:selected={selectedPlayers.includes(player.id)}
                class:temp={player.isTemp}
                on:click={() => togglePlayer(player.id)}
              >
                {#if selectedPlayers.includes(player.id)}
                  <span class="chip-check">✓</span>
                {/if}
                <span class="chip-name">{player.name}</span>
                {#if player.isTemp}
                  <span class="chip-guest">👤</span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>
      
      <!-- Tournament Preview Card (shows when valid selection) -->
      {#if isValidSelection}
        <div class="preview-card">
          <div class="preview-header">
            <span class="preview-icon">🏆</span>
            <h3>Tournament Preview</h3>
          </div>
          
          <div class="preview-stats">
            <div class="stat-item">
              <span class="stat-value">{numberOfTeams}</span>
              <span class="stat-label">Teams</span>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-item">
              <span class="stat-value">{totalMatches}</span>
              <span class="stat-label">Round Robin Matches</span>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-item">
              <span class="stat-value">{numberOfTeams >= 4 ? '3' : '1'}</span>
              <span class="stat-label">{numberOfTeams >= 4 ? 'Semis + Final' : 'Final Only'}</span>
            </div>
          </div>
          
          <div class="preview-flow">
            <div class="flow-step">
              <div class="flow-dot active"></div>
              <span>Teams</span>
            </div>
            <div class="flow-line"></div>
            <div class="flow-step">
              <div class="flow-dot"></div>
              <span>Round Robin</span>
            </div>
            <div class="flow-line"></div>
            <div class="flow-step">
              <div class="flow-dot"></div>
              <span>{numberOfTeams >= 4 ? 'Semis' : 'Final'}</span>
            </div>
            {#if numberOfTeams >= 4}
              <div class="flow-line"></div>
              <div class="flow-step">
                <div class="flow-dot"></div>
                <span>Final</span>
              </div>
            {/if}
          </div>
        </div>
      {/if}
      
      <!-- Continue Button -->
      <div class="footer-action">
        <button 
          class="btn btn-primary"
          disabled={!canProceed}
          on:click={proceed}
        >
          Continue to Team Creation
        </button>
      </div>
    {/if}
  </div>
</div>

<!-- Add Guest Player Modal -->
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
  .screen {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
  }
  
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-md);
    background: var(--bg-secondary);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .header-back {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 14px;
    cursor: pointer;
    padding: var(--space-sm);
  }
  
  .header-title {
    font: var(--font-section);
    color: var(--text-primary);
  }
  
  .content {
    flex: 1;
    padding: var(--space-md);
    padding-bottom: 100px;
    overflow-y: auto;
  }
  
  /* Loading & Error States */
  .loading-state, .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-xl);
    gap: var(--space-md);
    color: var(--text-secondary);
  }
  
  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--surface);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  
  /* Section Cards */
  .section-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-card);
    padding: var(--space-lg);
    margin-bottom: var(--space-md);
  }
  
  .section-header {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    margin-bottom: var(--space-md);
  }
  
  .section-icon {
    font-size: 20px;
  }
  
  .section-header h2 {
    font: var(--font-section);
    color: var(--text-primary);
    margin: 0;
  }
  
  .empty-message {
    color: var(--text-secondary);
    font-size: 14px;
    text-align: center;
    padding: var(--space-md);
  }
  
  /* Form Select */
  .form-select {
    width: 100%;
    padding: var(--space-md);
    background: var(--surface);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: var(--radius-btn);
    color: var(--text-primary);
    font-size: 16px;
    cursor: pointer;
  }
  
  .form-select:focus {
    outline: none;
    border-color: var(--accent);
  }
  
  /* Status Bar */
  .status-bar {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    padding: var(--space-sm) var(--space-md);
    background: var(--surface);
    border-radius: var(--radius-btn);
    margin-bottom: var(--space-md);
    flex-wrap: wrap;
  }
  
  .status-count {
    display: flex;
    align-items: baseline;
    gap: var(--space-xs);
  }
  
  .count-number {
    font-size: 24px;
    font-weight: 600;
    color: var(--accent);
  }
  
  .count-label {
    font-size: 12px;
    color: var(--text-secondary);
  }
  
  .status-valid {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    color: var(--accent);
    font-size: 13px;
    font-weight: 500;
  }
  
  .valid-icon {
    width: 18px;
    height: 18px;
    background: var(--accent);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
  }
  
  .status-warning {
    color: #F59E0B;
    font-size: 13px;
  }
  
  .quick-actions {
    display: flex;
    gap: var(--space-xs);
    margin-left: auto;
  }
  
  .action-btn {
    background: rgba(255, 255, 255, 0.1);
    border: none;
    color: var(--text-secondary);
    padding: var(--space-xs) var(--space-sm);
    border-radius: 6px;
    font-size: 12px;
    cursor: pointer;
  }
  
  .action-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    color: var(--text-primary);
  }
  
  /* Player Grid */
  .player-grid {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }
  
  .player-chip {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    padding: var(--space-sm) var(--space-md);
    background: var(--surface);
    border: 2px solid transparent;
    border-radius: 20px;
    color: var(--text-primary);
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  
  .player-chip:hover {
    border-color: var(--accent);
    background: rgba(34, 197, 94, 0.1);
  }
  
  .player-chip.selected {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  
  .chip-check {
    font-size: 12px;
    font-weight: 600;
  }
  
  .chip-name {
    font-weight: 500;
  }
  
  /* Preview Card */
  .preview-card {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.15) 0%, rgba(34, 197, 94, 0.05) 100%);
    border: 1px solid rgba(34, 197, 94, 0.3);
    border-radius: var(--radius-card);
    padding: var(--space-lg);
    margin-bottom: var(--space-md);
  }
  
  .preview-header {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    margin-bottom: var(--space-lg);
  }
  
  .preview-icon {
    font-size: 24px;
  }
  
  .preview-header h3 {
    font: var(--font-section);
    color: var(--text-primary);
    margin: 0;
  }
  
  .preview-stats {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: var(--space-lg);
    margin-bottom: var(--space-lg);
    flex-wrap: wrap;
  }
  
  .stat-item {
    text-align: center;
  }
  
  .stat-value {
    display: block;
    font-size: 28px;
    font-weight: 600;
    color: var(--accent);
    line-height: 1;
  }
  
  .stat-label {
    display: block;
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: var(--space-xs);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  
  .stat-divider {
    width: 1px;
    height: 40px;
    background: rgba(255, 255, 255, 0.1);
  }
  
  /* Tournament Flow */
  .preview-flow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0;
    padding: var(--space-md) 0;
  }
  
  .flow-step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-xs);
  }
  
  .flow-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--surface);
    border: 2px solid var(--text-secondary);
  }
  
  .flow-dot.active {
    background: var(--accent);
    border-color: var(--accent);
  }
  
  .flow-step span {
    font-size: 11px;
    color: var(--text-secondary);
    white-space: nowrap;
  }
  
  .flow-line {
    width: 30px;
    height: 2px;
    background: var(--surface);
    margin-bottom: 20px;
  }
  
  /* Footer Action */
  .footer-action {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    padding: var(--space-md);
    background: linear-gradient(transparent, var(--bg-primary) 30%);
    padding-top: var(--space-xl);
  }
  
  .footer-action .btn {
    width: 100%;
  }
  
  /* Button Styles */
  .btn {
    padding: var(--space-md) var(--space-lg);
    border-radius: var(--radius-btn);
    font: var(--font-button);
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }
  
  .btn-primary {
    background: var(--accent);
    color: white;
  }
  
  .btn-primary:hover:not(:disabled) {
    background: var(--accent-hover);
  }
  
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .btn-secondary {
    background: var(--surface);
    color: var(--text-primary);
  }
  
  /* Guest Player Section */
  .add-guest-section {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    margin-bottom: var(--space-md);
  }
  
  .btn-add-guest {
    background: var(--surface);
    border: 1px dashed var(--text-secondary);
    color: var(--text-secondary);
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-btn);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-add-guest:hover {
    border-color: var(--accent);
    color: var(--accent);
  }
  
  .guest-count {
    font-size: 11px;
    color: var(--text-secondary);
  }
  
  .player-chip.temp {
    border: 1px dashed var(--text-secondary);
  }
  
  .player-chip.temp.selected {
    border-style: solid;
  }
  
  .chip-guest {
    font-size: 10px;
    margin-left: 2px;
  }
  
  /* Modal Styles */
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
</style>
