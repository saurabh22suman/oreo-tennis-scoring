<!-- Match Setup - Screens 2 from ui_design_spec.md -->
<script>
  import { navigate, matchState, venues } from '../stores/app.js';
  
  let selectedVenue = '';
  let selectedType = 'singles';
  let selectedMode = 'short'; // 'standard' or 'short'
  let selectedBestOf = 3; // 3 or 5 games for short format
  
  function goBack() {
    navigate('home');
  }
  
  // Reactive validation
  $: canContinue = selectedVenue !== '';
  
  function continueToPlayers() {
    if (!selectedVenue) return;
    
    matchState.update(m => ({
      ...m,
      venueId: selectedVenue,
      venueName: $venues.find(v => v.id === selectedVenue)?.name || '',
      matchType: selectedType,
      matchMode: selectedMode,
      bestOf: selectedMode === 'short' ? selectedBestOf : null,
    }));
    
    navigate('player-selection');
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">New Match</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; display: flex; flex-direction: column;">
    <div style="flex: 1;">
      <div class="form-group">
        <label class="form-label" for="venue">Venue</label>
        <select class="form-select" id="venue" bind:value={selectedVenue}>
          <option value="">Select venue...</option>
          {#each $venues as venue}
            <option value={venue.id}>{venue.name} ({venue.surface})</option>
          {/each}
        </select>
      </div>
      
      <div class="form-group">
        <label class="form-label">Match Type</label>
        <div class="toggle-group">
          <button 
            class="toggle-btn" 
            class:active={selectedType === 'singles'}
            on:click={() => selectedType = 'singles'}
          >
            Singles
          </button>
          <button 
            class="toggle-btn" 
            class:active={selectedType === 'doubles'}
            on:click={() => selectedType = 'doubles'}
          >
            Doubles
          </button>
          <button 
            class="toggle-btn" 
            class:active={selectedType === '1v2'}
            on:click={() => selectedType = '1v2'}
          >
            1 vs 2
          </button>
        </div>
        {#if selectedType === '1v2'}
          <p class="text-secondary" style="font-size: 13px; margin-top: var(--space-sm);">
            Australian Doubles: 1 player vs 2 players
          </p>
        {/if}
      </div>
      
      <div class="form-group">
        <label class="form-label">Scoring Mode</label>
        <div class="toggle-group">
          <button 
            class="toggle-btn" 
            class:active={selectedMode === 'short'}
            on:click={() => selectedMode = 'short'}
          >
            Short Format
          </button>
          <button 
            class="toggle-btn" 
            class:active={selectedMode === 'standard'}
            on:click={() => selectedMode = 'standard'}
          >
            Standard
          </button>
        </div>
        <p class="text-secondary" style="font-size: 13px; margin-top: var(--space-sm);">
          {#if selectedMode === 'short'}
            Best of {selectedBestOf} games (recreational)
          {:else}
            Full tennis scoring (sets, tie-breaks)
          {/if}
        </p>
      </div>
      
      {#if selectedMode === 'short'}
        <div class="form-group">
          <label class="form-label">Games to Play</label>
          <div class="toggle-group">
            <button 
              class="toggle-btn" 
              class:active={selectedBestOf === 3}
              on:click={() => selectedBestOf = 3}
            >
              Best of 3
            </button>
            <button 
              class="toggle-btn" 
              class:active={selectedBestOf === 5}
              on:click={() => selectedBestOf = 5}
            >
              Best of 5
            </button>
          </div>
          <p class="text-secondary" style="font-size: 13px; margin-top: var(--space-sm);">
            First to {Math.ceil(selectedBestOf / 2)} games wins
          </p>
        </div>
      {/if}
    </div>
    
    <div class="footer-section">
      {#if !canContinue}
        <p class="hint-text">Please select a venue to continue</p>
      {/if}
      <button class="btn btn-primary" on:click={continueToPlayers} disabled={!canContinue}>
        Continue
      </button>
    </div>
  </div>
</div>

<style>
  .footer-section {
    margin-top: var(--space-lg);
  }
  
  .hint-text {
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
    margin-bottom: var(--space-sm);
  }
</style>
