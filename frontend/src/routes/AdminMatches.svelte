<!-- Admin Matches -->
<script>
  import { onMount } from 'svelte';
  import { navigate, viewMatchId } from '../stores/app.js';
  import { getMatches, deleteMatch } from '../services/api.js';
  import Modal from '../lib/Modal.svelte';
  
  let matches = [];
  let loading = true;
  let showDeleteModal = false;
  let deleteMatchId = null;
  let deleting = false;
  
  // Alert modal state
  let showAlertModal = false;
  let alertMessage = '';
  
  onMount(async () => {
    await loadMatches();
  });
  
  async function loadMatches() {
    loading = true;
    try {
      matches = await getMatches();
    } catch (err) {
      alertMessage = 'Failed to load matches';
      showAlertModal = true;
    }
    loading = false;
  }
  
  function goBack() {
    navigate('admin-dashboard');
  }
  
  function handleView(matchId) {
    viewMatchId.set(matchId);
    navigate('match-summary');
  }
  
  function confirmDelete(matchId) {
    deleteMatchId = matchId;
    showDeleteModal = true;
  }
  
  function cancelDelete() {
    showDeleteModal = false;
    deleteMatchId = null;
  }
  
  async function handleDelete() {
    if (!deleteMatchId) return;
    
    deleting = true;
    try {
      await deleteMatch(deleteMatchId);
      showDeleteModal = false;
      deleteMatchId = null;
      await loadMatches();
    } catch (err) {
      alertMessage = 'Failed to delete match';
      showAlertModal = true;
    }
    deleting = false;
  }
  
  function formatDate(iso) {
    return new Date(iso).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

<div class="screen">
  <div class="header">
    <button class="header-back" on:click={goBack}>
      ← Back
    </button>
    <h1 class="header-title">Matches</h1>
    <div style="width: 60px;"></div>
  </div>
  
  <div class="container" style="flex: 1; overflow-y: auto;">
    {#if loading}
      <div style="display: flex; justify-content: center; padding: var(--space-xl);">
        <div class="loading-spinner"></div>
      </div>
    {:else if matches.length === 0}
      <div style="text-align: center; padding: var(--space-xl);">
        <p class="text-secondary">No matches yet</p>
      </div>
    {:else}
      <div class="list">
        {#each matches as match}
          <div class="list-item">
            <div class="list-item-content">
              <div class="list-item-title">
                {match.match_type === 'singles' ? 'Singles' : 'Doubles'}
                {#if match.ended_at}
                  <span style="color: var(--color-success);">✓</span>
                {/if}
              </div>
              <div class="list-item-subtitle">
                {formatDate(match.started_at)}
              </div>
            </div>
            <div style="display: flex; gap: var(--space-sm);">
              {#if match.ended_at}
                <button 
                  class="btn btn-secondary" 
                  on:click={() => handleView(match.id)}
                  style="width: auto; padding: var(--space-sm) var(--space-md);"
                >
                  View
                </button>
              {/if}
              <button 
                class="btn btn-danger" 
                on:click={() => confirmDelete(match.id)}
                style="width: auto; padding: var(--space-sm) var(--space-md);"
              >
                Delete
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
  <div class="modal-overlay" on:click={cancelDelete}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-icon">🗑️</div>
      <h2 class="modal-title">Delete Match?</h2>
      <p class="modal-message">This action cannot be undone. All match data and statistics will be permanently removed.</p>
      <div class="modal-actions">
        <button class="btn btn-secondary" on:click={cancelDelete} disabled={deleting}>
          Cancel
        </button>
        <button class="btn btn-danger" on:click={handleDelete} disabled={deleting}>
          {deleting ? 'Deleting...' : 'Delete'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--space-md);
  }
  
  .modal {
    background: var(--color-surface);
    border-radius: 16px;
    padding: var(--space-xl);
    max-width: 400px;
    width: 100%;
    text-align: center;
    animation: modalSlideIn 0.2s ease-out;
  }
  
  @keyframes modalSlideIn {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(-10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }
  
  .modal-icon {
    font-size: 48px;
    margin-bottom: var(--space-md);
  }
  
  .modal-title {
    color: var(--color-text);
    margin-bottom: var(--space-sm);
    font-size: 24px;
  }
  
  .modal-message {
    color: var(--color-text-secondary);
    margin-bottom: var(--space-lg);
    line-height: 1.5;
  }
  
  .modal-actions {
    display: flex;
    gap: var(--space-sm);
    justify-content: center;
  }
  
  .modal-actions .btn {
    flex: 1;
    max-width: 150px;
  }
</style>

<!-- Alert Modal -->
<Modal 
  bind:show={showAlertModal}
  title="Error"
  message={alertMessage}
  icon="❌"
  type="alert"
  confirmText="OK"
/>
