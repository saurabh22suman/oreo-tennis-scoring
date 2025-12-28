<!-- Analytics - Venue Team & Player Tendencies -->
<script>
  import { navigate } from '../stores/app.js';
  import { getVenues, getVenueTendencies } from '../services/api.js';
  import { onMount } from 'svelte';
  
  let venues = [];
  let selectedVenueId = '';
  let tendencies = null;
  let loading = false;
  let error = null;
  
  // Filters
  let showTeams = true;
  let showPlayers = true;
  let sortBy = 'name'; // 'name', 'matches', 'winrate'
  
  // Date filter
  let timePeriod = 'all'; // 'all', 'day', 'week', 'month'
  let selectedMonth = new Date().getMonth() + 1; // 1-12
  let selectedYear = new Date().getFullYear();
  
  // Month names for display
  const monthNames = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' }
  ];
  
  // Generate years for dropdown (current year and 2 previous)
  $: yearOptions = [selectedYear, selectedYear - 1, selectedYear - 2].map(y => ({ value: y, label: y.toString() }));
  
  onMount(async () => {
    try {
      venues = await getVenues();
    } catch (err) {
      console.error('Failed to load venues:', err);
    }
  });
  
  // Build date filter object based on selections
  function buildDateFilter() {
    if (timePeriod === 'all') {
      return {};
    }
    
    if (timePeriod === 'month') {
      return {
        period: 'month',
        month: selectedMonth,
        year: selectedYear
      };
    }
    
    return { period: timePeriod };
  }
  
  async function loadTendencies() {
    if (!selectedVenueId) {
      tendencies = null;
      return;
    }
    
    loading = true;
    error = null;
    
    try {
      const dateFilter = buildDateFilter();
      tendencies = await getVenueTendencies(selectedVenueId, dateFilter);
    } catch (err) {
      error = err.message || 'Failed to load tendencies';
      tendencies = null;
    } finally {
      loading = false;
    }
  }
  
  // Reload when venue or time period changes
  $: if (selectedVenueId) {
    // Need to make this reactive to time period changes too
    timePeriod, selectedMonth, selectedYear;
    loadTendencies();
  }
  
  // Sorted team tendencies
  $: sortedTeams = tendencies?.team_tendencies?.slice().sort((a, b) => {
    if (sortBy === 'matches') return b.matches_played - a.matches_played;
    if (sortBy === 'winrate') return b.win_percentage - a.win_percentage;
    return `${a.player1_name} & ${a.player2_name}`.localeCompare(`${b.player1_name} & ${b.player2_name}`);
  }) || [];
  
  // Sorted player tendencies
  $: sortedPlayers = tendencies?.player_tendencies?.slice().sort((a, b) => {
    if (sortBy === 'matches') return b.matches_played - a.matches_played;
    return a.player_name.localeCompare(b.player_name);
  }) || [];
  
  function goBack() {
    navigate('home');
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Analytics</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; overflow-y: auto;">
    <!-- Venue Selector -->
    <div class="form-group">
      <label class="form-label">Select Venue</label>
      <select class="form-select" bind:value={selectedVenueId}>
        <option value="">Choose a venue...</option>
        {#each venues as venue}
          <option value={venue.id}>{venue.name}</option>
        {/each}
      </select>
    </div>
    
    {#if selectedVenueId}
      <!-- Time Period Filter -->
      <div class="card mb-md filter-card">
        <div class="time-filter-section">
          <span class="filter-label">Time Period:</span>
          <div class="time-filter-buttons">
            <button 
              class="time-btn" 
              class:active={timePeriod === 'all'} 
              on:click={() => timePeriod = 'all'}
            >
              All Time
            </button>
            <button 
              class="time-btn" 
              class:active={timePeriod === 'day'} 
              on:click={() => timePeriod = 'day'}
            >
              Today
            </button>
            <button 
              class="time-btn" 
              class:active={timePeriod === 'week'} 
              on:click={() => timePeriod = 'week'}
            >
              Last 7 Days
            </button>
            <button 
              class="time-btn" 
              class:active={timePeriod === 'month'} 
              on:click={() => timePeriod = 'month'}
            >
              Month
            </button>
          </div>
          
          {#if timePeriod === 'month'}
            <div class="month-selectors">
              <select class="form-select month-select" bind:value={selectedMonth}>
                {#each monthNames as m}
                  <option value={m.value}>{m.label}</option>
                {/each}
              </select>
              <select class="form-select year-select" bind:value={selectedYear}>
                {#each yearOptions as y}
                  <option value={y.value}>{y.label}</option>
                {/each}
              </select>
            </div>
          {/if}
        </div>
      </div>
      
      <!-- Display Controls -->
      <div class="card mb-md filter-card">
        <div class="filter-row">
          <div class="filter-toggles">
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={showTeams} />
              <span>Teams</span>
            </label>
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={showPlayers} />
              <span>Players</span>
            </label>
          </div>
          
          <div class="sort-control">
            <span class="text-secondary" style="font-size: 12px;">Sort:</span>
            <select class="form-select sort-select" bind:value={sortBy}>
              <option value="name">Name</option>
              <option value="matches">Matches</option>
              {#if showTeams}
                <option value="winrate">Win Rate</option>
              {/if}
            </select>
          </div>
        </div>
      </div>
    {/if}
    
    {#if loading}
      <div style="display: flex; flex-direction: column; align-items: center; padding: var(--space-xl);">
        <div class="loading-spinner"></div>
        <p class="text-secondary" style="margin-top: var(--space-md);">Loading tendencies...</p>
      </div>
    {:else if error}
      <div class="card text-center">
        <p class="text-danger">{error}</p>
      </div>
    {:else if tendencies}
      <!-- Team Tendencies -->
      {#if showTeams}
        <div class="mb-lg">
          <h2 style="margin-bottom: var(--space-xs);">Team Tendencies</h2>
          <p class="text-secondary" style="font-size: 12px; margin-bottom: var(--space-md);">Doubles teams with 3+ matches</p>
          
          {#if sortedTeams.length === 0}
            <div class="card text-center">
              <p style="margin-bottom: var(--space-xs);">No eligible teams yet</p>
              <p class="text-secondary" style="font-size: 12px;">Teams need 3+ doubles matches at this venue</p>
            </div>
          {:else}
            <div class="list">
              {#each sortedTeams as team}
                <div class="list-item tendency-item">
                  <div class="tendency-content">
                    <div class="tendency-header">
                      <span class="list-item-title">{team.player1_name} & {team.player2_name}</span>
                      <span class="badge">{team.matches_played} matches</span>
                    </div>
                    <div class="stats-grid stats-grid-4">
                      <div class="stat-item">
                        <span class="stat-value text-accent">{team.win_percentage.toFixed(0)}%</span>
                        <span class="stat-label">Win Rate</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{team.matches_won}/{team.matches_played}</span>
                        <span class="stat-label">W/L</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{team.avg_games_per_match.toFixed(1)}</span>
                        <span class="stat-label">Avg Games</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{team.deuce_percentage.toFixed(0)}%</span>
                        <span class="stat-label">Deuce %</span>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
      
      <!-- Player Tendencies -->
      {#if showPlayers}
        <div class="mb-lg">
          <h2 style="margin-bottom: var(--space-xs);">Player Tendencies</h2>
          <p class="text-secondary" style="font-size: 12px; margin-bottom: var(--space-md);">Players with 5+ matches</p>
          
          {#if sortedPlayers.length === 0}
            <div class="card text-center">
              <p style="margin-bottom: var(--space-xs);">No eligible players yet</p>
              <p class="text-secondary" style="font-size: 12px;">Players need 5+ matches at this venue</p>
            </div>
          {:else}
            <div class="list">
              {#each sortedPlayers as player}
                <div class="list-item tendency-item">
                  <div class="tendency-content">
                    <div class="tendency-header">
                      <span class="list-item-title">{player.player_name}</span>
                      <span class="badge">{player.matches_played} matches</span>
                    </div>
                    <div class="stats-grid stats-grid-3">
                      <div class="stat-item">
                        <span class="stat-value">{player.first_serve_in_pct.toFixed(0)}%</span>
                        <span class="stat-label">1st Serve In</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{player.double_faults_per_game.toFixed(2)}</span>
                        <span class="stat-label">DF/Game</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-value">{player.avg_points_per_game.toFixed(1)}</span>
                        <span class="stat-label">Pts/Game</span>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
      
      {#if !showTeams && !showPlayers}
        <div class="card text-center">
          <p>Select at least one category to view</p>
        </div>
      {/if}
    {:else if !selectedVenueId}
      <div class="card text-center" style="margin-top: var(--space-lg);">
        <p class="text-secondary">Select a venue to view tendencies</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-md);
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--surface);
  }
  
  .header-back {
    background: none;
    border: none;
    color: var(--text-secondary);
    font: var(--font-body);
    cursor: pointer;
    padding: var(--space-sm);
    min-width: 60px;
    text-align: left;
  }
  
  .header-back:hover {
    color: var(--text-primary);
  }
  
  .header-title {
    font: var(--font-section);
    font-size: 18px;
  }
  
  .filter-card {
    padding: var(--space-md);
  }
  
  .filter-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }
  
  .filter-toggles {
    display: flex;
    gap: var(--space-lg);
  }
  
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font: var(--font-body);
    color: var(--text-primary);
    cursor: pointer;
  }
  
  .checkbox-label input[type="checkbox"] {
    width: 18px;
    height: 18px;
    accent-color: var(--accent);
  }
  
  .sort-control {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  
  .sort-select {
    width: auto;
    padding: var(--space-xs) var(--space-md);
    padding-right: var(--space-xl);
    font-size: 14px;
  }
  
  .tendency-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .tendency-content {
    width: 100%;
  }
  
  .tendency-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-md);
  }
  
  .badge {
    background: var(--bg-secondary);
    padding: var(--space-xs) var(--space-sm);
    border-radius: var(--radius-btn);
    font-size: 11px;
    color: var(--text-secondary);
  }
  
  .stats-grid {
    display: grid;
    gap: var(--space-sm);
  }
  
  .stats-grid-4 {
    grid-template-columns: repeat(4, 1fr);
  }
  
  .stats-grid-3 {
    grid-template-columns: repeat(3, 1fr);
  }
  
  .stat-item {
    text-align: center;
  }
  
  .stat-value {
    display: block;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }
  
  .stat-label {
    display: block;
    font-size: 10px;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.03em;
    margin-top: 2px;
  }
  
  /* Responsive: 2 columns on very small screens for 4-col grid */
  @media (max-width: 360px) {
    .stats-grid-4 {
      grid-template-columns: repeat(2, 1fr);
    }
  }
  
  /* Time Period Filter Styles */
  .time-filter-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  
  .filter-label {
    font-size: 12px;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }
  
  .time-filter-buttons {
    display: flex;
    gap: var(--space-xs);
    flex-wrap: wrap;
  }
  
  .time-btn {
    background: var(--bg-secondary);
    border: 1px solid var(--surface);
    color: var(--text-secondary);
    padding: var(--space-xs) var(--space-sm);
    border-radius: var(--radius-btn);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  
  .time-btn:hover {
    background: var(--surface);
    color: var(--text-primary);
  }
  
  .time-btn.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--bg-primary);
    font-weight: 500;
  }
  
  .month-selectors {
    display: flex;
    gap: var(--space-sm);
    margin-top: var(--space-xs);
  }
  
  .month-select {
    flex: 2;
    padding: var(--space-xs) var(--space-sm);
    font-size: 14px;
  }
  
  .year-select {
    flex: 1;
    padding: var(--space-xs) var(--space-sm);
    font-size: 14px;
  }
  
  @media (max-width: 360px) {
    .time-filter-buttons {
      flex-direction: column;
    }
    
    .time-btn {
      text-align: center;
    }
    
    .month-selectors {
      flex-direction: column;
    }
  }
</style>
