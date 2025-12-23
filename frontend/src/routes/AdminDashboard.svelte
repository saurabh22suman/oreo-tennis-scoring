<!-- Admin Dashboard - Screen 7 from ui_design_spec.md -->
<script>
  import { navigate, isAdmin } from '../stores/app.js';
  import { logout } from '../services/api.js';
  
  async function handleLogout() {
    await logout();
    isAdmin.set(false);
    navigate('home');
  }
  
  function goTo(screen) {
    navigate(screen);
  }
</script>

<div class="screen">
  <div class="header">
    <div style="width: 60px;"></div>
    <h1 class="header-title">Admin</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; display: flex; flex-direction: column; gap: var(--space-md);">
    <button class="nav-card" on:click={() => goTo('admin-players')}>
      <div class="nav-card-icon">👥</div>
      <div class="nav-card-content">
        <div class="nav-card-title">Players</div>
        <div class="nav-card-subtitle">Manage player roster</div>
      </div>
    </button>
    
    <button class="nav-card" on:click={() => goTo('admin-venues')}>
      <div class="nav-card-icon">🎾</div>
      <div class="nav-card-content">
        <div class="nav-card-title">Venues</div>
        <div class="nav-card-subtitle">Manage courts</div>
      </div>
    </button>
    
    <button class="nav-card" on:click={() => goTo('admin-matches')}>
      <div class="nav-card-icon">📊</div>
      <div class="nav-card-content">
        <div class="nav-card-title">Matches</div>
        <div class="nav-card-subtitle">View and delete matches</div>
      </div>
    </button>
    
    <div class="mt-auto">
      <button class="btn btn-ghost" on:click={handleLogout}>
        Logout
      </button>
    </div>
  </div>
</div>
