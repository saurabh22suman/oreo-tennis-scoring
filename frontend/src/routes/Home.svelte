<!-- Home Screen - Screen 1 from ui_design_spec.md -->
<script>
  import { navigate } from '../stores/app.js';
  import { getCurrentMatch } from '../services/db.js';
  import { onMount } from 'svelte';
  
  let hasIncompleteMatch = false;
  
  onMount(async () => {
    const match = await getCurrentMatch();
    hasIncompleteMatch = !!match && !match.completed;
  });
  
  function startNewMatch() {
    navigate('match-setup');
  }
  
  function startNewTournament() {
    navigate('tournament-setup');
  }
  
  function resumeMatch() {
    navigate('live-match');
  }
  
  function goToAdmin() {
    navigate('admin-login');
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
      
      {#if hasIncompleteMatch}
        <button class="btn btn-secondary" on:click={resumeMatch}>
          Resume Match
        </button>
      {/if}
      
      <button class="btn btn-ghost" on:click={goToAdmin} style="margin-top: var(--space-lg);">
        Admin
      </button>
    </div>
  </div>
</div>
