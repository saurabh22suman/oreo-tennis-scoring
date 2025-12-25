<!-- Home Screen - Screen 1 from ui_design_spec.md -->
<script>
  import { navigate, resetMatch, matchState } from '../stores/app.js';
  import { getAllIncompleteMatches, deleteIncompleteMatch, clearCurrentMatch } from '../services/db.js';
  import { onMount } from 'svelte';
  
  let incompleteMatches = [];
  let showMatchList = false;
  let matchToDelete = null;
  
  onMount(async () => {
    try {
      await loadIncompleteMatches();
    } catch (err) {
      console.error('Failed to load incomplete matches:', err);
      incompleteMatches = [];
    }
  });
  
  async function loadIncompleteMatches() {
    try {
      incompleteMatches = await getAllIncompleteMatches();
    } catch (err) {
      console.error('Error loading matches:', err);
      incompleteMatches = [];
    }
  }
  
  async function startNewMatch() {
    try {
      // Reset match state to clear any old data
      resetMatch();
      // Clear legacy current match
      await clearCurrentMatch();
      navigate('match-setup');
    } catch (err) {
      console.error('Error starting new match:', err);
      navigate('match-setup');
    }
  }
  
  function startNewTournament() {
    navigate('tournament-setup');
  }
  
  function toggleMatchList() {
    showMatchList = !showMatchList;
  }
  
  function resumeMatch(match) {
    // Load match state into store
    matchState.set({
      id: match.matchId,
      venue: match.venue,
      matchType: match.matchType,
      formatMode: match.formatMode,
      teamA: match.teamA,
      teamB: match.teamB,
      score: match.score,
      events: match.events || [],
      completed: false,
    });
    navigate('live-match');
  }
  
  function confirmDelete(match) {
    matchToDelete = match;
  }
  
  async function deleteMatch() {
    if (matchToDelete) {
      await deleteIncompleteMatch(matchToDelete.matchId);
      matchToDelete = null;
      await loadIncompleteMatches();
    }
  }
  
  function cancelDelete() {
    matchToDelete = null;
  }
  
  function goToAdmin() {
    navigate('admin-login');
  }
  
  function goToAnalytics() {
    navigate('analytics');
  }
  
  function formatMatchInfo(match) {
    const players = [];
    if (match.teamA) {
      players.push(match.teamA.map(p => p.name?.split(' ')[0] || 'Player').join(' & '));
    }
    if (match.teamB) {
      players.push(match.teamB.map(p => p.name?.split(' ')[0] || 'Player').join(' & '));
    }
    return players.join(' vs ');
  }
  
  function formatTimeAgo(timestamp) {
    if (!timestamp) return '';
    const mins = Math.floor((Date.now() - timestamp) / 60000);
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    return 'Expiring soon';
  }
  
  function getMatchScore(match) {
    if (!match.score) return '0-0';
    const s = match.score;
    if (s.setsA !== undefined) {
      return `${s.setsA}-${s.setsB}`;
    }
    return `${s.gamesA || 0}-${s.gamesB || 0}`;
  }
</script>

<div class="screen screen-center">
  <div class="container">
    <div style="text-align: center; margin-bottom: var(--space-xl);">
      <h1 style="font-size: 28px; margin-bottom: var(--space-sm);">OTS</h1>
      <p class="text-secondary">Oreo Tennis Scoring</p>
    </div>
    
    <div style="display: flex; flex-direction: column; gap: var(--space-md); max-width: 360px; margin: 0 auto;">
      <button class="btn btn-primary" on:click={startNewMatch}>
        Start New Match
      </button>
      
      <button class="btn btn-primary" on:click={startNewTournament}>
        Start New Tournament
      </button>
      
      {#if incompleteMatches.length > 0}
        <button class="btn btn-secondary" on:click={toggleMatchList}>
          Resume Match ({incompleteMatches.length})
          <span style="margin-left: auto;">{showMatchList ? '▲' : '▼'}</span>
        </button>
        
        {#if showMatchList}
          <div class="match-list">
            {#each incompleteMatches as match}
              <div class="match-item">
                <div class="match-info" on:click={() => resumeMatch(match)}>
                  <div class="match-players">{formatMatchInfo(match)}</div>
                  <div class="match-meta">
                    <span class="match-score">{getMatchScore(match)}</span>
                    <span class="match-venue">{match.venue?.name || 'Unknown venue'}</span>
                    <span class="match-time">{formatTimeAgo(match.updatedAt || match.createdAt)}</span>
                  </div>
                </div>
                <button class="btn-delete" on:click={() => confirmDelete(match)} title="Delete match">
                  🗑️
                </button>
              </div>
            {/each}
          </div>
        {/if}
      {/if}
      
      <button class="btn btn-secondary" on:click={goToAnalytics} style="margin-top: var(--space-md);">
        📊 Analytics
      </button>
      
      <button class="btn btn-ghost" on:click={goToAdmin} style="margin-top: var(--space-sm);">
        Admin
      </button>
    </div>
  </div>
</div>

<!-- Delete Confirmation Modal -->
{#if matchToDelete}
  <div class="modal-overlay" on:click={cancelDelete}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-icon">🗑️</div>
      <h2 class="modal-title">Delete Match?</h2>
      <p class="modal-text">
        {formatMatchInfo(matchToDelete)}<br/>
        <span class="text-secondary">Score: {getMatchScore(matchToDelete)}</span>
      </p>
      <p class="modal-warning">This cannot be undone.</p>
      <div class="modal-actions">
        <button class="btn btn-secondary" on:click={cancelDelete}>Cancel</button>
        <button class="btn btn-danger" on:click={deleteMatch}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .match-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    margin-top: calc(-1 * var(--space-sm));
  }
  
  .match-item {
    display: flex;
    align-items: center;
    background: var(--surface);
    border-radius: var(--radius-btn);
    overflow: hidden;
  }
  
  .match-info {
    flex: 1;
    padding: var(--space-md);
    cursor: pointer;
    transition: background 0.15s;
  }
  
  .match-info:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  
  .match-players {
    font-weight: 600;
    font-size: 14px;
    margin-bottom: var(--space-xs);
  }
  
  .match-meta {
    display: flex;
    gap: var(--space-sm);
    font-size: 12px;
    color: var(--text-secondary);
  }
  
  .match-score {
    color: var(--accent);
    font-weight: 600;
  }
  
  .match-venue {
    flex: 1;
  }
  
  .match-time {
    color: var(--text-secondary);
  }
  
  .btn-delete {
    background: none;
    border: none;
    padding: var(--space-md);
    cursor: pointer;
    opacity: 0.6;
    transition: opacity 0.15s;
    font-size: 16px;
  }
  
  .btn-delete:hover {
    opacity: 1;
  }
  
  /* Modal styles */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: var(--space-md);
  }
  
  .modal {
    background: var(--surface);
    border-radius: var(--radius-card);
    padding: var(--space-lg);
    max-width: 320px;
    width: 100%;
    text-align: center;
  }
  
  .modal-icon {
    font-size: 48px;
    margin-bottom: var(--space-md);
  }
  
  .modal-title {
    font-size: 18px;
    margin-bottom: var(--space-sm);
  }
  
  .modal-text {
    color: var(--text-primary);
    margin-bottom: var(--space-sm);
    font-size: 14px;
  }
  
  .modal-warning {
    color: var(--danger);
    font-size: 12px;
    margin-bottom: var(--space-lg);
  }
  
  .modal-actions {
    display: flex;
    gap: var(--space-sm);
  }
  
  .modal-actions .btn {
    flex: 1;
  }
</style>
