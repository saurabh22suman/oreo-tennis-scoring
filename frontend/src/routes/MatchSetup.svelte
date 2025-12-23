<!-- Match Setup - Screens 2 from ui_design_spec.md -->
<script>
  import { navigate, matchState, venues } from '../stores/app.js';
  
  let selectedVenue = '';
  let selectedType = 'singles';
  let selectedMode = 'short'; // 'standard' or 'short'
  
  function goBack() {
    navigate('home');
  }
  
  function continueToPlayers() {
    if (!selectedVenue) {
      alert('Please select a venue');
      return;
    }
    
    matchState.update(m => ({
      ...m,
      venueId: selectedVenue,
      venueName: $venues.find(v => v.id === selectedVenue)?.name || '',
      matchType: selectedType,
      matchMode: selectedMode,
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
        </div>
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
            Best of 3 games (recreational)
          {:else}
            Full tennis scoring (sets, tie-breaks)
          {/if}
        </p>
      </div>
    </div>
    
    <button class="btn btn-primary" on:click={continueToPlayers}>
      Continue
    </button>
  </div>
</div>
