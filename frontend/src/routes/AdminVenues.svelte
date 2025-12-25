<!-- Admin Venues - Screen 9 from ui_design_spec.md -->
<script>
  import { onMount } from 'svelte';
  import { navigate } from '../stores/app.js';
  import { getAdminVenues, createVenue, updateVenue } from '../services/api.js';
  import Modal from '../lib/Modal.svelte';
  
  let venues = [];
  let loading = true;
  let showAdd = false;
  let newVenueName = '';
  let newVenueSurface = 'hard';
  
  // Modal state
  let showAlertModal = false;
  let alertMessage = '';
  
  onMount(async () => {
    await loadVenues();
  });
  
  async function loadVenues() {
    loading = true;
    try {
      venues = await getAdminVenues();
    } catch (err) {
      alertMessage = 'Failed to load venues';
      showAlertModal = true;
    }
    loading = false;
  }
  
  function goBack() {
    navigate('admin-dashboard');
  }
  
  async function handleAdd() {
    if (!newVenueName.trim()) return;
    
    try {
      await createVenue(newVenueName.trim(), newVenueSurface);
      newVenueName = '';
      newVenueSurface = 'hard';
      showAdd = false;
      await loadVenues();
    } catch (err) {
      alertMessage = 'Failed to create venue';
      showAlertModal = true;
    }
  }
  
  async function toggleActive(venue) {
    try {
      await updateVenue(venue.id, { active: !venue.active });
      await loadVenues();
    } catch (err) {
      alertMessage = 'Failed to update venue';
      showAlertModal = true;
    }
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Venues</h1>
    <button on:click={() => showAdd = !showAdd} class="btn btn-ghost" style="padding: var(--space-sm); width: auto; min-width: 60px;">
      {showAdd ? 'Cancel' : '+ Add'}
    </button>
  </div>
  
  <div class="container" style="flex: 1; overflow-y: auto;">
    {#if showAdd}
      <div class="card mb-md">
        <h3 class="mb-md">Add Venue</h3>
        <div class="form-group">
          <input 
            type="text" 
            class="form-input" 
            placeholder="Venue name"
            bind:value={newVenueName}
          />
        </div>
        <div class="form-group">
          <select class="form-select" bind:value={newVenueSurface}>
            <option value="hard">Hard Court</option>
            <option value="clay">Clay Court</option>
            <option value="grass">Grass Court</option>
          </select>
        </div>
        <button class="btn btn-primary" on:click={handleAdd}>
          Add Venue
        </button>
      </div>
    {/if}
    
    {#if loading}
      <div style="display: flex; justify-content: center; padding: var(--space-xl);">
        <div class="loading-spinner"></div>
      </div>
    {:else}
      <div class="list">
        {#each venues as venue}
          <div class="list-item">
            <div class="list-item-content">
              <div class="list-item-title">{venue.name}</div>
              <div class="list-item-subtitle">
                {venue.surface} • {venue.active ? 'Active' : 'Inactive'}
              </div>
            </div>
            <button 
              class="btn {venue.active ? 'btn-secondary' : 'btn-primary'}" 
              on:click={() => toggleActive(venue)}
              style="width: auto; padding: var(--space-sm) var(--space-md);"
            >
              {venue.active ? 'Disable' : 'Enable'}
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
