<!-- Admin Players - Screen 8 from ui_design_spec.md -->
<script>
  import { onMount } from 'svelte';
  import { navigate } from '../stores/app.js';
  import { getAdminPlayers, createPlayer, updatePlayer } from '../services/api.js';
  import Modal from '../lib/Modal.svelte';
  
  let players = [];
  let loading = true;
  let showAdd = false;
  let newPlayerName = '';
  
  // Modal state
  let showAlertModal = false;
  let alertMessage = '';
  
  onMount(async () => {
    await loadPlayers();
  });
  
  async function loadPlayers() {
    loading = true;
    try {
      players = await getAdminPlayers();
    } catch (err) {
      alertMessage = 'Failed to load players';
      showAlertModal = true;
    }
    loading = false;
  }
  
  function goBack() {
    navigate('admin-dashboard');
  }
  
  async function handleAdd() {
    if (!newPlayerName.trim()) return;
    
    try {
      await createPlayer(newPlayerName.trim());
      newPlayerName = '';
      showAdd = false;
      await loadPlayers();
    } catch (err) {
      alertMessage = 'Failed to create player';
      showAlertModal = true;
    }
  }
  
  async function toggleActive(player) {
    try {
      await updatePlayer(player.id, { active: !player.active });
      await loadPlayers();
    } catch (err) {
      alertMessage = 'Failed to update player';
      showAlertModal = true;
    }
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Players</h1>
    <button on:click={() => showAdd = !showAdd} class="btn btn-ghost" style="padding: var(--space-sm); width: auto; min-width: 60px;">
      {showAdd ? 'Cancel' : '+ Add'}
    </button>
  </div>
  
  <div class="container" style="flex: 1; overflow-y: auto;">
    {#if showAdd}
      <div class="card mb-md">
        <h3 class="mb-md">Add Player</h3>
        <div style="display: flex; gap: var(--space-sm);">
          <input 
            type="text" 
            class="form-input" 
            placeholder="Player name"
            bind:value={newPlayerName}
            on:keydown={(e) => e.key === 'Enter' && handleAdd()}
          />
          <button class="btn btn-primary" on:click={handleAdd} style="width: auto; min-width: 80px;">
            Add
          </button>
        </div>
      </div>
    {/if}
    
    {#if loading}
      <div style="display: flex; justify-content: center; padding: var(--space-xl);">
        <div class="loading-spinner"></div>
      </div>
    {:else}
      <div class="list">
        {#each players as player}
          <div class="list-item">
            <div class="list-item-content">
              <div class="list-item-title">{player.name}</div>
              <div class="list-item-subtitle">
                {player.active ? 'Active' : 'Inactive'}
              </div>
            </div>
            <button 
              class="btn {player.active ? 'btn-secondary' : 'btn-primary'}" 
              on:click={() => toggleActive(player)}
              style="width: auto; padding: var(--space-sm) var(--space-md);"
            >
              {player.active ? 'Disable' : 'Enable'}
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Alert Modal -->
<Modal 
  bind:show={showAlertModal}
  title="Error"
  message={alertMessage}
  icon="❌"
  type="alert"
  confirmText="OK"
/>
