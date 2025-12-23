<!-- Admin Matches -->
<script>
  import { onMount } from 'svelte';
  import { navigate } from '../stores/app.js';
  import { getMatches, deleteMatch } from '../services/api.js';
  
  let matches = [];
  let loading = true;
  
  onMount(async () => {
    await loadMatches();
  });
  
  async function loadMatches() {
    loading = true;
    try {
      matches = await getMatches();
    } catch (err) {
      alert('Failed to load matches');
    }
    loading = false;
  }
  
  function goBack() {
    navigate('admin-dashboard');
  }
  
  async function handleDelete(matchId) {
    if (!confirm('Delete this match? This cannot be undone.')) return;
    
    try {
      await deleteMatch(matchId);
      await loadMatches();
    } catch (err) {
      alert('Failed to delete match');
    }
  }
  
  function formatDate(iso) {
    return new Date(iso).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Matches</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; overflow-y: auto;">
    {#if loading}
      <div style="display: flex; justify-content: center; padding: var(--space-xl);">
        <div class="loading-spinner"></div>
      </div>
    {:else if matches.length === 0}
      <div style="text-align: center; padding: var(--space-xl);">
        <p class="text-secondary">No matches yet</p>
      </div>
    {:else}
      <div class="list">
        {#each matches as match}
          <div class="list-item">
            <div class="list-item-content">
              <div class="list-item-title">
                {match.match_type === 'singles' ? 'Singles' : 'Doubles'}
                {#if match.ended_at}
                  ✓
                {/if}
              </div>
              <div class="list-item-subtitle">
                {formatDate(match.started_at)}
              </div>
            </div>
            <button 
              class="btn btn-danger" 
              on:click={() => handleDelete(match.id)}
              style="width: auto; padding: var(--space-sm) var(--space-md);"
            >
              Delete
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
