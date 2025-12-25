<!-- Match Summary - Screen 5 from ui_design_spec.md -->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { navigate, matchState, viewMatchId, isAdmin } from '../stores/app.js';
  import { getMatchSummary } from '../services/api.js';
  import { clearCurrentMatch, clearMatchEvents } from '../services/db.js';
  import Modal from '../lib/Modal.svelte';
  
  let summary = null;
  let loading = true;
  let isAdminView = false;
  let matchId = null;
  
  // Alert modal state
  let showAlertModal = false;
  let alertMessage = '';
  let navigateAfterAlert = '';
  
  function handleAlertClose() {
    showAlertModal = false;
    if (navigateAfterAlert) {
      navigate(navigateAfterAlert);
    }
  }
  
  onMount(async () => {
    // Check if this is an admin viewing a completed match
    if ($viewMatchId) {
      isAdminView = true;
      matchId = $viewMatchId;
    } else if ($matchState.id) {
      matchId = $matchState.id;
    } else {
      navigate('home');
      return;
    }
    
    try {
      summary = await getMatchSummary(matchId);
      loading = false;
    } catch (err) {
      alertMessage = 'Failed to load summary: ' + err.message;
      navigateAfterAlert = isAdminView ? 'admin-matches' : 'home';
      showAlertModal = true;
    }
  });
  
  onDestroy(() => {
    // Clear viewMatchId when leaving
    viewMatchId.set(null);
  });
  
  async function finish() {
    if (isAdminView) {
      viewMatchId.set(null);
      navigate('admin-matches');
    } else {
      await clearCurrentMatch();
      await clearMatchEvents(matchId);
      navigate('home');
    }
  }
  
  function getPercentage(value, total) {
    if (total === 0) return 0;
    return Math.round((value / total) * 100);
  }
  
  // Group players by team
  $: teamAPlayers = summary?.player_stats?.filter(p => p.team === 'A') || [];
  $: teamBPlayers = summary?.player_stats?.filter(p => p.team === 'B') || [];
  
  // Check if sets are applicable (standard mode has sets > 0)
  $: hasSets = (summary?.sets_a || 0) > 0 || (summary?.sets_b || 0) > 0;
</script>

<div class="screen">
  <div class="header">
    <div style="width: 60px;"></div>
    <h1 class="header-title">Match Summary</h1>
    <div style="width: 60px;"></div>
  </div>
  
  {#if loading}
    <div class="container" style="flex: 1; display: flex; align-items: center; justify-content: center;">
      <div class="loading-spinner"></div>
    </div>
  {:else if summary}
    <div class="container" style="flex: 1; overflow-y: auto;">
      <!-- Match Info -->
      <div class="card mb-md">
        <div class="text-center">
          <p class="text-secondary" style="font-size: 12px;">{summary.venue.name}</p>
          
          <!-- Score Table -->
          <div class="score-table-summary">
            <div class="score-table-row score-table-header">
              <div class="score-table-cell"></div>
              <div class="score-table-cell">Team A</div>
              <div class="score-table-cell">Team B</div>
            </div>
            {#if hasSets}
              <div class="score-table-row">
                <div class="score-table-cell score-label-cell">Sets</div>
                <div class="score-table-cell score-value-cell">{summary.sets_a}</div>
                <div class="score-table-cell score-value-cell">{summary.sets_b}</div>
              </div>
            {/if}
            <div class="score-table-row">
              <div class="score-table-cell score-label-cell">Games</div>
              <div class="score-table-cell score-value-cell">{summary.games_a}</div>
              <div class="score-table-cell score-value-cell">{summary.games_b}</div>
            </div>
            <div class="score-table-row">
              <div class="score-table-cell score-label-cell">Points</div>
              <div class="score-table-cell score-value-cell">{summary.team_a_score}</div>
              <div class="score-table-cell score-value-cell">{summary.team_b_score}</div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Team A Stats Card -->
      <div class="card mb-sm team-card">
        <div class="team-card-header">Team A</div>
        <div class="team-stats-row">
          {#each teamAPlayers as stat, i}
            {@const firstServeInPct = getPercentage(stat.first_serves_in, stat.first_serves_total)}
            {@const firstServeWonPct = getPercentage(stat.first_serve_won, stat.first_serves_in)}
            <div class="player-stats" class:player-stats-border={i === 0 && teamAPlayers.length > 1}>
              <div class="player-name">{stat.player_name.split(' ')[0]}</div>
              <div class="stat-row">
                <span class="stat-label">1st In</span>
                <span class="stat-value">{firstServeInPct}%</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">1st Won</span>
                <span class="stat-value">{firstServeWonPct}%</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">DF</span>
                <span class="stat-value stat-danger">{stat.double_faults}</span>
              </div>
            </div>
          {/each}
        </div>
      </div>
      
      <!-- Team B Stats Card -->
      <div class="card mb-sm team-card">
        <div class="team-card-header">Team B</div>
        <div class="team-stats-row">
          {#each teamBPlayers as stat, i}
            {@const firstServeInPct = getPercentage(stat.first_serves_in, stat.first_serves_total)}
            {@const firstServeWonPct = getPercentage(stat.first_serve_won, stat.first_serves_in)}
            <div class="player-stats" class:player-stats-border={i === 0 && teamBPlayers.length > 1}>
              <div class="player-name">{stat.player_name.split(' ')[0]}</div>
              <div class="stat-row">
                <span class="stat-label">1st In</span>
                <span class="stat-value">{firstServeInPct}%</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">1st Won</span>
                <span class="stat-value">{firstServeWonPct}%</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">DF</span>
                <span class="stat-value stat-danger">{stat.double_faults}</span>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
    
    <div class="container">
      <button class="btn btn-primary" on:click={finish}>
        {isAdminView ? 'Back to Matches' : 'Finish'}
      </button>
    </div>
  {/if}
</div>

<!-- Alert Modal -->
<Modal 
  bind:show={showAlertModal}
  title="Error"
  message={alertMessage}
  icon="❌"
  type="alert"
  confirmText="OK"
  on:confirm={handleAlertClose}
/>

<style>
  .team-card {
    padding: var(--space-sm);
  }
  
  .team-card-header {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: var(--space-xs);
    text-align: center;
  }
  
  .team-stats-row {
    display: flex;
    gap: var(--space-sm);
  }
  
  .player-stats {
    flex: 1;
    padding: var(--space-xs);
  }
  
  .player-stats-border {
    border-right: 1px solid var(--border);
    padding-right: var(--space-sm);
  }
  
  .player-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: var(--space-xs);
    text-align: center;
  }
  
  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 2px 0;
    font-size: 12px;
  }
  
  .stat-label {
    color: var(--text-secondary);
  }
  
  .stat-value {
    color: var(--text-primary);
    font-weight: 500;
  }
  
  .stat-danger {
    color: var(--danger);
  }
  
  .mb-sm {
    margin-bottom: var(--space-sm);
  }
  
  /* Score Table Styles */
  .score-table-summary {
    background: var(--bg-secondary);
    border-radius: var(--radius-md);
    padding: var(--space-md);
    margin-bottom: var(--space-md);
  }
  
  .score-table-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-xs) 0;
    border-bottom: 1px solid var(--bg-tertiary);
  }
  
  .score-table-row:last-child {
    border-bottom: none;
  }
  
  .score-table-header {
    font-weight: 600;
    color: var(--text-secondary);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  
  .score-table-cell {
    flex: 1;
    text-align: center;
  }
  
  .score-label-cell {
    flex: 1;
    text-align: left;
    font-weight: 500;
    color: var(--text-secondary);
  }
  
  .score-value-cell {
    flex: 1;
    text-align: center;
    font-weight: 600;
    font-size: 1.1rem;
    color: var(--text-primary);
  }
  
  .score-value-cell.winner {
    color: var(--success);
  }
</style>
