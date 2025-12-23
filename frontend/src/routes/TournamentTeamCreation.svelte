<!-- Tournament Team Creation - Per OTS_Tournament_Spec.md Section 3.3 -->
<script>
  import { navigate } from '../stores/app.js';
  import { onMount } from 'svelte';
  
  let tournamentData = null;
  let teams = [];
  let creationMode = 'random'; // 'random' or 'manual'
  let loading = true;
  
  onMount(() => {
    const stored = localStorage.getItem('tournamentSetup');
    if (!stored) {
      navigate('tournament-setup');
      return;
    }
    
    tournamentData = JSON.parse(stored);
    
    // Initialize empty teams based on player count
    const numTeams = Math.floor(tournamentData.players.length / 2);
    teams = Array.from({ length: numTeams }, (_, i) => ({
      id: i + 1,
      name: `Team ${i + 1}`,
      player1: null,
      player2: null
    }));
    
    loading = false;
  });
  
  function shuffleArray(array) {
    const shuffled = [...array];
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }
    return shuffled;
  }
  
  function generateRandomTeams() {
    if (!tournamentData) return;
    
    const shuffledPlayers = shuffleArray(tournamentData.players);
    teams = teams.map((team, index) => ({
      ...team,
      player1: shuffledPlayers[index * 2] || null,
      player2: shuffledPlayers[index * 2 + 1] || null
    }));
  }
  
  function clearTeams() {
    teams = teams.map(team => ({
      ...team,
      player1: null,
      player2: null
    }));
  }
  
  // Get unassigned players
  $: assignedPlayerIds = teams.flatMap(t => [t.player1?.id, t.player2?.id]).filter(Boolean);
  $: unassignedPlayers = tournamentData?.players.filter(p => !assignedPlayerIds.includes(p.id)) || [];
  
  // Check if all teams are complete
  $: allTeamsComplete = teams.every(t => t.player1 && t.player2);
  
  function assignPlayer(teamIndex, slot, player) {
    teams = teams.map((team, i) => {
      if (i === teamIndex) {
        return { ...team, [slot]: player };
      }
      return team;
    });
  }
  
  function removePlayer(teamIndex, slot) {
    teams = teams.map((team, i) => {
      if (i === teamIndex) {
        return { ...team, [slot]: null };
      }
      return team;
    });
  }
  
  function proceed() {
    if (!allTeamsComplete) return;
    
    // Save teams to tournament data
    const fullData = {
      ...tournamentData,
      teams: teams.map(t => ({
        id: t.id,
        name: t.name,
        players: [t.player1, t.player2]
      }))
    };
    
    localStorage.setItem('tournamentData', JSON.stringify(fullData));
    navigate('tournament-dashboard');
  }
  
  function goBack() {
    navigate('tournament-setup');
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Create Teams</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="content">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else}
      <!-- Mode Selection -->
      <div class="mode-section">
        <div class="toggle-group">
          <button 
            class="toggle-btn" 
            class:active={creationMode === 'random'}
            on:click={() => creationMode = 'random'}
          >
            Random
          </button>
          <button 
            class="toggle-btn" 
            class:active={creationMode === 'manual'}
            on:click={() => creationMode = 'manual'}
          >
            Manual
          </button>
        </div>
        
        {#if creationMode === 'random'}
          <div class="action-buttons">
            <button class="btn btn-secondary" on:click={generateRandomTeams}>
              🎲 Shuffle Teams
            </button>
          </div>
        {/if}
      </div>
      
      <!-- Teams Display -->
      <div class="teams-container">
        {#each teams as team, index}
          <div class="team-card">
            <div class="team-header">
              <span class="team-badge">{team.id}</span>
              <span class="team-name">{team.name}</span>
            </div>
            
            <div class="team-players">
              <!-- Player 1 Slot -->
              <div class="player-slot" class:filled={team.player1}>
                {#if team.player1}
                  <span class="player-name">{team.player1.name}</span>
                  {#if creationMode === 'manual'}
                    <button class="remove-btn" on:click={() => removePlayer(index, 'player1')}>×</button>
                  {/if}
                {:else}
                  <span class="empty-slot">Empty</span>
                {/if}
              </div>
              
              <div class="player-divider">+</div>
              
              <!-- Player 2 Slot -->
              <div class="player-slot" class:filled={team.player2}>
                {#if team.player2}
                  <span class="player-name">{team.player2.name}</span>
                  {#if creationMode === 'manual'}
                    <button class="remove-btn" on:click={() => removePlayer(index, 'player2')}>×</button>
                  {/if}
                {:else}
                  <span class="empty-slot">Empty</span>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
      
      <!-- Unassigned Players (Manual Mode) -->
      {#if creationMode === 'manual' && unassignedPlayers.length > 0}
        <div class="unassigned-section">
          <h3>Unassigned Players</h3>
          <div class="player-chips">
            {#each unassignedPlayers as player}
              <div class="player-chip-draggable">
                {player.name}
                <div class="assign-dropdown">
                  {#each teams as team, teamIndex}
                    {#if !team.player1 || !team.player2}
                      <button 
                        class="assign-option"
                        on:click={() => {
                          const slot = !team.player1 ? 'player1' : 'player2';
                          assignPlayer(teamIndex, slot, player);
                        }}
                      >
                        → {team.name}
                      </button>
                    {/if}
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
      
      <!-- Status Message -->
      <div class="status-message">
        {#if allTeamsComplete}
          <p class="success">✓ All teams are ready!</p>
        {:else}
          <p class="warning">
            {unassignedPlayers.length} player{unassignedPlayers.length !== 1 ? 's' : ''} still unassigned
          </p>
        {/if}
      </div>
    {/if}
  </div>
  
  <!-- Footer Action -->
  <div class="footer-action">
    <button 
      class="btn btn-primary"
      disabled={!allTeamsComplete}
      on:click={proceed}
    >
      Start Tournament
    </button>
  </div>
</div>

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
  
  .loading-state {
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
  
  /* Mode Selection */
  .mode-section {
    margin-bottom: var(--space-lg);
  }
  
  .toggle-group {
    display: flex;
    gap: var(--space-xs);
    background: var(--surface);
    padding: var(--space-xs);
    border-radius: var(--radius-btn);
    margin-bottom: var(--space-md);
  }
  
  .toggle-btn {
    flex: 1;
    padding: var(--space-sm) var(--space-md);
    background: transparent;
    border: none;
    border-radius: 8px;
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .toggle-btn.active {
    background: var(--accent);
    color: white;
  }
  
  .action-buttons {
    display: flex;
    gap: var(--space-sm);
  }
  
  /* Teams Container */
  .teams-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    margin-bottom: var(--space-lg);
  }
  
  .team-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-card);
    padding: var(--space-md);
  }
  
  .team-header {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    margin-bottom: var(--space-md);
  }
  
  .team-badge {
    width: 28px;
    height: 28px;
    background: var(--accent);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 14px;
  }
  
  .team-name {
    font-weight: 600;
    color: var(--text-primary);
  }
  
  .team-players {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }
  
  .player-slot {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-sm) var(--space-md);
    background: var(--surface);
    border-radius: var(--radius-btn);
    min-height: 44px;
  }
  
  .player-slot.filled {
    background: rgba(34, 197, 94, 0.15);
    border: 1px solid rgba(34, 197, 94, 0.3);
  }
  
  .player-name {
    color: var(--text-primary);
    font-weight: 500;
  }
  
  .empty-slot {
    color: var(--text-secondary);
    font-style: italic;
  }
  
  .remove-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 18px;
    cursor: pointer;
    padding: 0 var(--space-xs);
  }
  
  .remove-btn:hover {
    color: var(--danger);
  }
  
  .player-divider {
    color: var(--text-secondary);
    font-size: 18px;
    padding: 0 var(--space-xs);
  }
  
  /* Unassigned Players */
  .unassigned-section {
    background: var(--bg-secondary);
    border-radius: var(--radius-card);
    padding: var(--space-md);
    margin-bottom: var(--space-lg);
  }
  
  .unassigned-section h3 {
    font: var(--font-section);
    color: var(--text-primary);
    margin-bottom: var(--space-md);
  }
  
  .player-chips {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }
  
  .player-chip-draggable {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    padding: var(--space-sm) var(--space-md);
    background: var(--surface);
    border-radius: 20px;
    color: var(--text-primary);
    font-size: 14px;
    cursor: pointer;
  }
  
  .player-chip-draggable:hover .assign-dropdown {
    display: flex;
  }
  
  .assign-dropdown {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    background: var(--bg-secondary);
    border: 1px solid var(--surface);
    border-radius: var(--radius-btn);
    flex-direction: column;
    min-width: 120px;
    z-index: 10;
    margin-top: var(--space-xs);
    box-shadow: var(--shadow-md);
  }
  
  .assign-option {
    background: none;
    border: none;
    color: var(--text-primary);
    padding: var(--space-sm) var(--space-md);
    text-align: left;
    cursor: pointer;
    font-size: 13px;
  }
  
  .assign-option:hover {
    background: var(--surface);
  }
  
  /* Status Message */
  .status-message {
    text-align: center;
    padding: var(--space-md);
  }
  
  .status-message p {
    margin: 0;
    font-size: 14px;
  }
  
  .status-message .success {
    color: var(--accent);
  }
  
  .status-message .warning {
    color: #F59E0B;
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
  
  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.15);
  }
</style>
