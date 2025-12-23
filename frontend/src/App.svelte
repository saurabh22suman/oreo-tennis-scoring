<script>
  import { onMount } from 'svelte';
  import { currentScreen, isAdmin } from './stores/app.js';
  import { getCachedPlayers, getCachedVenues } from './services/db.js';
  import { getPlayers, getVenues, checkAuth } from './services/api.js';
  import { players, venues } from './stores/app.js';
  
  // Screens
  import Home from './routes/Home.svelte';
  import MatchSetup from './routes/MatchSetup.svelte';
  import PlayerSelection from './routes/PlayerSelection.svelte';
  import LiveMatch from './routes/LiveMatch.svelte';
  import MatchSummary from './routes/MatchSummary.svelte';
  import TournamentSetup from './routes/TournamentSetup.svelte';
  import AdminLogin from './routes/AdminLogin.svelte';
  import AdminDashboard from './routes/AdminDashboard.svelte';
  import AdminPlayers from './routes/AdminPlayers.svelte';
  import AdminVenues from './routes/AdminVenues.svelte';
  import AdminMatches from './routes/AdminMatches.svelte';
  
  let screen = 'home';
  
  onMount(async () => {
    // Check admin auth status
    const authed = await checkAuth().catch(() => false);
    isAdmin.set(authed);
    
    // Load cached data first (instant load)
    const cachedPlayers = await getCachedPlayers();
    const cachedVenues = await getCachedVenues();
    
    if (cachedPlayers.length > 0) players.set(cachedPlayers);
    if (cachedVenues.length > 0) venues.set(cachedVenues);
    
    // Then fetch fresh data in background
    try {
      const freshPlayers = await getPlayers();
      const freshVenues = await getVenues();
      players.set(freshPlayers);
      venues.set(freshVenues);
    } catch (err) {
      console.log('Failed to fetch fresh data, using cache');
    }
  });
  
  // Subscribe to screen changes
  currentScreen.subscribe(value => {
    screen = value;
  });
</script>

<main>
  {#if screen === 'home'}
    <Home />
  {:else if screen === 'match-setup'}
    <MatchSetup />
  {:else if screen === 'player-selection'}
    <PlayerSelection />
  {:else if screen === 'live-match'}
    <LiveMatch />
  {:else if screen === 'match-summary'}
  {:else if screen === 'tournament-setup'}
    <TournamentSetup />
    <MatchSummary />
  {:else if screen === 'admin-login'}
    <AdminLogin />
  {:else if screen === 'admin-dashboard'}
    <AdminDashboard />
  {:else if screen === 'admin-players'}
    <AdminPlayers />
  {:else if screen === 'admin-venues'}
    <AdminVenues />
  {:else if screen === 'admin-matches'}
    <AdminMatches />
  {:else}
    <Home />
  {/if}
</main>
