<!-- Admin Login - Screen 6 from ui_design_spec.md -->
<script>
  import { navigate, isAdmin } from '../stores/app.js';
  import { login } from '../services/api.js';
  
  let username = '';
  let password = '';
  let loading = false;
  let error = '';
  
  function goBack() {
    navigate('home');
  }
  
  async function handleLogin(e) {
    e.preventDefault();
    error = '';
    loading = true;
    
    try {
      await login(username, password);
      isAdmin.set(true);
      navigate('admin-dashboard');
    } catch (err) {
      error = 'Invalid credentials';
      password = '';
      loading = false;
    }
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Admin Login</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; display: flex; flex-direction: column; justify-content: center; max-width: 360px; margin: 0 auto;">
    <form on:submit={handleLogin}>
      {#if error}
        <div class="alert alert-error mb-md">
          {error}
        </div>
      {/if}
      
      <div class="form-group">
        <label class="form-label" for="username">Username</label>
        <input 
          type="text" 
          id="username" 
          class="form-input"
          bind:value={username}
          autocomplete="username"
          required
          disabled={loading}
        />
      </div>
      
      <div class="form-group">
        <label class="form-label" for="password">Password</label>
        <input 
          type="password" 
          id="password" 
          class="form-input"
          bind:value={password}
          autocomplete="current-password"
          required
          disabled={loading}
        />
      </div>
      
      <button type="submit" class="btn btn-primary" disabled={loading}>
        {#if loading}
          Logging in...
        {:else}
          Login
        {/if}
      </button>
    </form>
  </div>
</div>
