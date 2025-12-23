<!-- Match Summary - Screen 5 from ui_design_spec.md -->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { navigate, matchState, viewMatchId, isAdmin } from '../stores/app.js';
  import { getMatchSummary } from '../services/api.js';
  import { clearCurrentMatch, clearMatchEvents } from '../services/db.js';
  
  let summary = null;
  let loading = true;
  let isAdminView = false;
  let matchId = null;
  
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
      alert('Failed to load summary: ' + err.message);
      navigate(isAdminView ? 'admin-matches' : 'home');
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
      <div class="card mb-lg">
        <div class="text-center">
          <p class="text-secondary">{summary.venue.name}</p>
          <div class="score-display" style="padding: var(--space-md) 0;">
            <div class="score-team">
              <div class="score-label">Team A</div>
              <div class="score-value" style="font-size: 36px;">{summary.team_a_score}</div>
            </div>
            <div class="score-divider" style="font-size: 24px;">:</div>
            <div class="score-team">
              <div class="score-label">Team B</div>
              <div class="score-value" style="font-size: 36px;">{summary.team_b_score}</div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Player Stats -->
      <h2 class="mb-md">Player Statistics</h2>
      
      {#each summary.player_stats as stat}
        {@const firstServeInPct = getPercentage(stat.first_serves_in, stat.first_serves_total)}
        {@const firstServeWonPct = getPercentage(stat.first_serve_won, stat.first_serves_in)}
        {@const secondServeWonPct = getPercentage(stat.second_serve_won, stat.second_serves_in)}
        
        <div class="card mb-md">
          <h3 style="margin-bottom: var(--space-md);">{stat.player_name} (Team {stat.team})</h3>
          
          <!-- First Serve In % -->
          <div class="stat-bar">
            <div class="stat-bar-label">
              <span class="text-secondary">First Serve In</span>
              <span class="text-primary">{firstServeInPct}%</span>
            </div>
            <div class="stat-bar-track">
              <div class="stat-bar-fill" style="width: {firstServeInPct}%;"></div>
            </div>
          </div>
          
          <!-- First Serve Points Won % -->
          <div class="stat-bar">
            <div class="stat-bar-label">
              <span class="text-secondary">First Serve Won</span>
              <span class="text-primary">{firstServeWonPct}%</span>
            </div>
            <div class="stat-bar-track">
              <div class="stat-bar-fill" style="width: {firstServeWonPct}%;"></div>
            </div>
          </div>
          
          <!-- Second Serve Points Won % -->
          <div class="stat-bar">
            <div class="stat-bar-label">
              <span class="text-secondary">Second Serve Won</span>
              <span class="text-primary">{secondServeWonPct}%</span>
            </div>
            <div class="stat-bar-track">
              <div class="stat-bar-fill" style="width: {secondServeWonPct}%;"></div>
            </div>
          </div>
          
          <!-- Double Faults -->
          <div class="stat-bar" style="margin-bottom: 0;">
            <div class="stat-bar-label">
              <span class="text-secondary">Double Faults</span>
              <span class="text-danger">{stat.double_faults}</span>
            </div>
          </div>
        </div>
      {/each}
    </div>
    
    <div class="container">
      <button class="btn btn-primary" on:click={finish}>
        {isAdminView ? 'Back to Matches' : 'Finish'}
      </button>
    </div>
  {/if}
</div>
