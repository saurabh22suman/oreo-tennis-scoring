<!-- Tournament Dashboard - Per OTS_Tournament_Spec.md -->
<script>
  import { navigate, matchState } from '../stores/app.js';
  import { saveCurrentMatch } from '../services/db.js';
  import { onMount } from 'svelte';
  
  let tournamentData = null;
  let loading = true;
  let currentStage = 'round_robin'; // 'round_robin', 'knockout', 'completed'
  let matches = [];
  let standings = [];
  
  onMount(() => {
    const stored = localStorage.getItem('tournamentData');
    if (!stored) {
      navigate('tournament-setup');
      return;
    }
    
    tournamentData = JSON.parse(stored);
    
    // Generate round-robin matches if not already generated
    if (!tournamentData.matches) {
      tournamentData.matches = generateRoundRobinMatches(tournamentData.teams);
      localStorage.setItem('tournamentData', JSON.stringify(tournamentData));
    }
    
    matches = tournamentData.matches;
    updateStandings();
    loading = false;
  });
  
  function generateRoundRobinMatches(teams) {
    const matches = [];
    let matchId = 1;
    
    for (let i = 0; i < teams.length; i++) {
      for (let j = i + 1; j < teams.length; j++) {
        matches.push({
          id: matchId++,
          teamA: teams[i],
          teamB: teams[j],
          stage: 'round_robin',
          status: 'pending', // 'pending', 'in_progress', 'completed'
          winner: null,
          score: null
        });
      }
    }
    
    return matches;
  }
  
  function updateStandings() {
    if (!tournamentData?.teams) return;
    
    standings = tournamentData.teams.map(team => {
      const teamMatches = matches.filter(m => 
        m.status === 'completed' && 
        (m.teamA.id === team.id || m.teamB.id === team.id)
      );
      
      const won = teamMatches.filter(m => m.winner === team.id).length;
      const lost = teamMatches.length - won;
      
      return {
        team,
        played: teamMatches.length,
        won,
        lost,
        points: won // 1 point per win
      };
    }).sort((a, b) => b.points - a.points || b.won - a.won);
  }
  
  function getTeamNames(team) {
    return team.players.map(p => p.name).join(' & ');
  }
  
  $: roundRobinMatches = matches.filter(m => m.stage === 'round_robin');
  $: semiMatches = matches.filter(m => m.stage === 'semi');
  $: finalMatch = matches.find(m => m.stage === 'final');
  $: knockoutMatches = matches.filter(m => m.stage === 'semi' || m.stage === 'final');
  $: completedRoundRobin = roundRobinMatches.every(m => m.status === 'completed');
  $: completedSemis = semiMatches.length > 0 && semiMatches.every(m => m.status === 'completed');
  $: tournamentComplete = finalMatch?.status === 'completed';
  $: pendingMatches = matches.filter(m => m.status === 'pending');
  $: currentMatch = matches.find(m => m.status === 'in_progress');
  
  // Get tournament winner
  $: tournamentWinner = tournamentComplete && finalMatch ? 
    (finalMatch.winner === finalMatch.teamA.id ? finalMatch.teamA : finalMatch.teamB) : null;
  
  async function startMatch(match) {
    // Set match to in progress
    match.status = 'in_progress';
    tournamentData.matches = matches;
    localStorage.setItem('tournamentData', JSON.stringify(tournamentData));
    localStorage.setItem('currentTournamentMatch', JSON.stringify(match));
    
    // Navigate to live match with tournament context
    const teamAPlayerIds = match.teamA.players.map(p => p.id);
    const teamBPlayerIds = match.teamB.players.map(p => p.id);
    
    // Generate a local ID for tournament matches (prefixed with 't-')
    const tournamentMatchId = `t-${Date.now()}-${match.id}`;
    
    const matchData = {
      id: tournamentMatchId,
      venueId: tournamentData.venueId,
      venueName: tournamentData.venueName,
      matchType: 'doubles',
      matchMode: 'short',
      bestOf: 3,
      teamA: teamAPlayerIds,
      teamB: teamBPlayerIds,
      startedAt: new Date().toISOString(),
      currentServer: teamAPlayerIds[0],
      serverTeam: 'A',
      isTournamentMatch: true,
      tournamentMatchId: match.id,
      completed: false,
      events: [],
    };
    
    // Save to IndexedDB
    await saveCurrentMatch(matchData);
    
    // Update match state
    matchState.set(matchData);
    
    navigate('live-match');
  }
  
  function generateKnockoutMatches() {
    if (!completedRoundRobin) return;
    
    const numTeams = standings.length;
    let knockoutMatches = [];
    let matchId = matches.length + 1;
    
    if (numTeams === 3) {
      // Final only: Rank 1 vs Rank 2
      knockoutMatches.push({
        id: matchId++,
        teamA: standings[0].team,
        teamB: standings[1].team,
        stage: 'final',
        status: 'pending',
        winner: null,
        score: null
      });
    } else if (numTeams >= 4) {
      // Semifinals
      knockoutMatches.push({
        id: matchId++,
        teamA: standings[0].team,
        teamB: standings[3].team,
        stage: 'semi',
        status: 'pending',
        winner: null,
        score: null,
        matchLabel: 'Semi-Final 1'
      });
      knockoutMatches.push({
        id: matchId++,
        teamA: standings[1].team,
        teamB: standings[2].team,
        stage: 'semi',
        status: 'pending',
        winner: null,
        score: null,
        matchLabel: 'Semi-Final 2'
      });
    }
    
    matches = [...matches, ...knockoutMatches];
    tournamentData.matches = matches;
    currentStage = 'knockout';
    localStorage.setItem('tournamentData', JSON.stringify(tournamentData));
  }
  
  function generateFinalMatch() {
    if (!completedSemis || finalMatch) return;
    
    // Get winners from semifinals
    const semi1 = semiMatches[0];
    const semi2 = semiMatches[1];
    
    const winner1 = semi1.winner === semi1.teamA.id ? semi1.teamA : semi1.teamB;
    const winner2 = semi2.winner === semi2.teamA.id ? semi2.teamA : semi2.teamB;
    
    const finalMatchObj = {
      id: matches.length + 1,
      teamA: winner1,
      teamB: winner2,
      stage: 'final',
      status: 'pending',
      winner: null,
      score: null,
      matchLabel: '🏆 Final'
    };
    
    matches = [...matches, finalMatchObj];
    tournamentData.matches = matches;
    localStorage.setItem('tournamentData', JSON.stringify(tournamentData));
  }
  
  function goHome() {
    navigate('home');
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goHome}>
      ← Exit
    </button>
    <h1 class="header-title">Tournament</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="content">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else}
      <!-- Venue Info -->
      <div class="venue-badge">
        📍 {tournamentData.venueName} ({tournamentData.venueSurface})
      </div>
      
      <!-- Stage Indicator -->
      <div class="stage-tabs">
        <button 
          class="stage-tab" 
          class:active={currentStage === 'round_robin'}
          on:click={() => currentStage = 'round_robin'}
        >
          Round Robin
        </button>
        <button 
          class="stage-tab" 
          class:active={currentStage === 'knockout'}
          on:click={() => currentStage = 'knockout'}
          disabled={!completedRoundRobin && knockoutMatches.length === 0}
        >
          Knockout
        </button>
      </div>
      
      {#if currentStage === 'round_robin'}
        <!-- Standings Table -->
        <div class="section-card">
          <h3>Standings</h3>
          <div class="standings-table">
            <div class="standings-header">
              <span class="col-rank">#</span>
              <span class="col-team">Team</span>
              <span class="col-stat">P</span>
              <span class="col-stat">W</span>
              <span class="col-stat">L</span>
              <span class="col-stat">Pts</span>
            </div>
            {#each standings as row, index}
              <div class="standings-row" class:qualified={index < (standings.length >= 4 ? 4 : 2)}>
                <span class="col-rank">{index + 1}</span>
                <span class="col-team">
                  <span class="team-name-main">{row.team.name}</span>
                  <span class="team-players-sub">{getTeamNames(row.team)}</span>
                </span>
                <span class="col-stat">{row.played}</span>
                <span class="col-stat">{row.won}</span>
                <span class="col-stat">{row.lost}</span>
                <span class="col-stat points">{row.points}</span>
              </div>
            {/each}
          </div>
        </div>
        
        <!-- Round Robin Matches -->
        <div class="section-card">
          <h3>Matches</h3>
          <div class="matches-list">
            {#each roundRobinMatches as match}
              <div class="match-card" class:completed={match.status === 'completed'}>
                <div class="match-teams">
                  <div class="team-block" class:winner={match.winner === match.teamA.id}>
                    <span class="team-name">{match.teamA.name}</span>
                    <span class="team-players">{getTeamNames(match.teamA)}</span>
                  </div>
                  <span class="vs">vs</span>
                  <div class="team-block" class:winner={match.winner === match.teamB.id}>
                    <span class="team-name">{match.teamB.name}</span>
                    <span class="team-players">{getTeamNames(match.teamB)}</span>
                  </div>
                </div>
                
                <div class="match-action">
                  {#if match.status === 'completed'}
                    <span class="completed-badge">✓ Done</span>
                  {:else if match.status === 'in_progress'}
                    <span class="in-progress-badge">In Progress</span>
                  {:else}
                    <button class="btn btn-small btn-primary" on:click={() => startMatch(match)}>
                      Play
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
        
        <!-- Proceed to Knockout -->
        {#if completedRoundRobin && knockoutMatches.length === 0}
          <button class="btn btn-primary btn-full" on:click={generateKnockoutMatches}>
            Proceed to Knockout Stage
          </button>
        {/if}
        
      {:else if currentStage === 'knockout'}
        <!-- Tournament Winner Banner -->
        {#if tournamentWinner}
          <div class="winner-banner">
            <div class="trophy">🏆</div>
            <h2>Tournament Champion!</h2>
            <div class="winner-team">{tournamentWinner.name}</div>
            <div class="winner-players">
              {tournamentWinner.players.map(p => p.name).join(' & ')}
            </div>
          </div>
        {/if}
        
        <!-- Knockout Matches -->
        <div class="section-card">
          <h3>Knockout Stage</h3>
          
          {#if knockoutMatches.length === 0}
            <p class="text-secondary">Complete all round-robin matches first.</p>
          {:else}
            <div class="matches-list">
              {#each knockoutMatches as match}
                <div class="match-card knockout" class:completed={match.status === 'completed'}>
                  <div class="match-label">{match.matchLabel || (match.stage === 'final' ? '🏆 Final' : 'Semi-Final')}</div>
                  <div class="match-teams">
                    <div class="team-block" class:winner={match.winner === match.teamA.id}>
                      <span class="team-name">{match.teamA.name}</span>
                      <span class="team-players">{getTeamNames(match.teamA)}</span>
                    </div>
                    <span class="vs">vs</span>
                    <div class="team-block" class:winner={match.winner === match.teamB.id}>
                      <span class="team-name">{match.teamB.name}</span>
                      <span class="team-players">{getTeamNames(match.teamB)}</span>
                    </div>
                  </div>
                  
                  <div class="match-action">
                    {#if match.status === 'completed'}
                      <span class="completed-badge">✓ Done</span>
                    {:else if match.status === 'in_progress'}
                      <span class="in-progress-badge">In Progress</span>
                    {:else}
                      <button class="btn btn-small btn-primary" on:click={() => startMatch(match)}>
                        Play
                      </button>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
            
            <!-- Generate Final after Semis -->
            {#if completedSemis && !finalMatch}
              <button class="btn btn-primary btn-full" style="margin-top: var(--space-md);" on:click={generateFinalMatch}>
                Generate Final Match
              </button>
            {/if}
          {/if}
        </div>
        
        <!-- Finish Tournament Button -->
        {#if tournamentComplete}
          <button class="btn btn-secondary btn-full" style="margin-top: var(--space-md);" on:click={goHome}>
            Finish Tournament
          </button>
        {/if}
      {/if}
    {/if}
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
  
  /* Venue Badge */
  .venue-badge {
    text-align: center;
    padding: var(--space-sm) var(--space-md);
    background: var(--surface);
    border-radius: var(--radius-btn);
    color: var(--text-secondary);
    font-size: 13px;
    margin-bottom: var(--space-md);
  }
  
  /* Stage Tabs */
  .stage-tabs {
    display: flex;
    gap: var(--space-xs);
    background: var(--surface);
    padding: var(--space-xs);
    border-radius: var(--radius-btn);
    margin-bottom: var(--space-lg);
  }
  
  .stage-tab {
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
  
  .stage-tab.active {
    background: var(--accent);
    color: white;
  }
  
  .stage-tab:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  /* Section Card */
  .section-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-card);
    padding: var(--space-md);
    margin-bottom: var(--space-md);
  }
  
  .section-card h3 {
    font: var(--font-section);
    color: var(--text-primary);
    margin-bottom: var(--space-md);
  }
  
  /* Standings Table */
  .standings-table {
    font-size: 14px;
  }
  
  .standings-header, .standings-row {
    display: flex;
    align-items: center;
    padding: var(--space-sm) 0;
  }
  
  .standings-header {
    color: var(--text-secondary);
    font-size: 12px;
    text-transform: uppercase;
    border-bottom: 1px solid var(--surface);
  }
  
  .standings-row {
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }
  
  .standings-row.qualified {
    background: rgba(34, 197, 94, 0.1);
  }
  
  .col-rank {
    width: 30px;
    font-weight: 600;
  }
  
  .col-team {
    flex: 1;
  }
  
  .col-stat {
    width: 35px;
    text-align: center;
  }
  
  .col-stat.points {
    color: var(--accent);
    font-weight: 600;
  }
  
  /* Matches List */
  .matches-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  
  .match-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-md);
    background: var(--surface);
    border-radius: var(--radius-btn);
  }
  
  .match-card.completed {
    opacity: 0.7;
  }
  
  .match-card.knockout {
    border-left: 3px solid var(--accent);
  }
  
  .match-label {
    font-size: 11px;
    color: var(--accent);
    text-transform: uppercase;
    margin-bottom: var(--space-xs);
  }
  
  .match-teams {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    flex: 1;
  }
  
  .team {
    font-size: 14px;
    color: var(--text-primary);
  }
  
  .team.winner {
    color: var(--accent);
    font-weight: 600;
  }
  
  /* Team block with player names */
  .team-block {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  
  .team-block .team-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }
  
  .team-block .team-players {
    font-size: 11px;
    color: var(--text-secondary);
  }
  
  .team-block.winner .team-name {
    color: var(--accent);
  }
  
  .team-block.winner .team-players {
    color: var(--accent);
    opacity: 0.8;
  }
  
  /* Standings player names */
  .col-team {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  
  .team-name-main {
    font-weight: 600;
  }
  
  .team-players-sub {
    font-size: 11px;
    color: var(--text-secondary);
  }
  
  .vs {
    color: var(--text-secondary);
    font-size: 12px;
  }
  
  .match-action {
    margin-left: var(--space-md);
  }
  
  .completed-badge {
    color: var(--accent);
    font-size: 12px;
  }
  
  .in-progress-badge {
    color: #F59E0B;
    font-size: 12px;
  }
  
  /* Buttons */
  .btn {
    padding: var(--space-md) var(--space-lg);
    border-radius: var(--radius-btn);
    font: var(--font-button);
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }
  
  .btn-small {
    padding: var(--space-sm) var(--space-md);
    font-size: 13px;
  }
  
  .btn-primary {
    background: var(--accent);
    color: white;
  }
  
  .btn-primary:hover:not(:disabled) {
    background: var(--accent-hover);
  }
  
  .btn-full {
    width: 100%;
  }
  
  .text-secondary {
    color: var(--text-secondary);
    font-size: 14px;
    text-align: center;
    padding: var(--space-md);
  }
  
  /* Winner Banner */
  .winner-banner {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.2) 0%, rgba(34, 197, 94, 0.05) 100%);
    border: 2px solid var(--accent);
    border-radius: var(--radius-card);
    padding: var(--space-xl);
    text-align: center;
    margin-bottom: var(--space-lg);
  }
  
  .winner-banner .trophy {
    font-size: 48px;
    margin-bottom: var(--space-sm);
  }
  
  .winner-banner h2 {
    font-size: 20px;
    color: var(--accent);
    margin-bottom: var(--space-md);
  }
  
  .winner-team {
    font-size: 24px;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: var(--space-xs);
  }
  
  .winner-players {
    font-size: 14px;
    color: var(--text-secondary);
  }
  
  .btn-secondary {
    background: var(--surface);
    color: var(--text-primary);
  }
  
  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.15);
  }
</style>
